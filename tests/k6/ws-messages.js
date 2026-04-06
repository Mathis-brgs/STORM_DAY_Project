import ws   from 'k6/ws';
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';

const msgSendLatency  = new Trend('msg_send_latency', true);
const msgRoundTrip    = new Trend('msg_roundtrip_ms', true);
const msgErrors       = new Rate('msg_errors');
const msgSent         = new Counter('msg_sent_total');
const msgReceived     = new Counter('msg_received_total');

const TARGET_VUS  = parseInt(__ENV.K6_VUS      || '100',  10);
const MSG_RATE    = parseInt(__ENV.K6_MSG_RATE  || '10',   10); // messages/s par VU
const TEST_DURATION = __ENV.K6_DURATION || '5m';

export const options = {
  stages: [
    { duration: '1m', target: TARGET_VUS },  // ramp up
    { duration: TEST_DURATION, target: TARGET_VUS }, // maintien
    { duration: '30s', target: 0 },           // descente
  ],
  thresholds: {
    'msg_send_latency':  ['p(95)<200'],
    'msg_errors':        ['rate<0.01'],
    'msg_roundtrip_ms':  ['p(95)<500'],
  },
};

const BASE_URL = __ENV.BASE_URL     || 'http://localhost:8080';
const WS_URL   = __ENV.WS_URL       || 'ws://localhost:8080';
const EMAIL    = __ENV.TEST_EMAIL   || 'k6-validate@storm.dev';
const PASSWORD = __ENV.TEST_PASSWORD || 'k6password123';

export function setup() {
  const res = http.post(
    `${BASE_URL}/auth/login`,
    JSON.stringify({ email: EMAIL, password: PASSWORD }),
    { headers: { 'Content-Type': 'application/json' } },
  );

  if (res.status !== 200 && res.status !== 201) {
    console.error(`[setup] Login échoué : ${res.status} — ${res.body}`);
    return {};
  }

  const body = JSON.parse(res.body);
  const token = body.access_token;

  // Créer un groupe de test partagé
  const groupRes = http.post(
    `${BASE_URL}/api/groups`,
    JSON.stringify({ name: 'k6-msg-load-test', is_private: false }),
    { headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` } },
  );

  let conversationId = parseInt(__ENV.K6_CONVERSATION_ID || '0', 10);
  if (groupRes.status === 200 || groupRes.status === 201) {
    const g = JSON.parse(groupRes.body);
    conversationId = g.data?.id || conversationId;
    console.log(`[setup] Groupe créé, conversation_id=${conversationId}`);
  } else {
    console.warn(`[setup] Groupe non créé (${groupRes.status}), utilise K6_CONVERSATION_ID env`);
  }

  return { token, conversationId };
}

export default function (data) {
  if (!data.token || !data.conversationId) {
    console.error(`[VU ${__VU}] Pas de token ou conversation_id`);
    msgErrors.add(1);
    return;
  }

  const { token, conversationId } = data;
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };

  // Map pour tracker les round-trips : msgId → timestamp envoi
  const pending = new Map();

  // Ouvrir WS pour recevoir les messages
  const wsRes = ws.connect(`${WS_URL}/ws?token=${token}`, {}, (socket) => {
    socket.on('open', () => {
      socket.send(JSON.stringify({ action: 'join', room: `conversation:${conversationId}` }));
    });

    socket.on('message', (raw) => {
      try {
        const msg = JSON.parse(raw);
        if (msg.type === 'new_message' && msg.data?.id) {
          const sentAt = pending.get(String(msg.data.id));
          if (sentAt) {
            msgRoundTrip.add(Date.now() - sentAt);
            msgReceived.add(1);
            pending.delete(String(msg.data.id));
          }
        }
      } catch (_) {}
    });

    socket.on('error', () => msgErrors.add(1));

    // Boucle d'envoi : MSG_RATE messages/s
    const intervalMs = 1000 / MSG_RATE;
    const totalDuration = 6 * 60 * 1000; 
    const start = Date.now();

    socket.setInterval(() => {
      if (Date.now() - start > totalDuration) {
        socket.close();
        return;
      }

      const sentAt = Date.now();
      const res = http.post(
        `${BASE_URL}/api/messages`,
        JSON.stringify({
          conversation_id: conversationId,
          content: `k6 msg from VU${__VU} t=${sentAt}`,
        }),
        { headers },
      );

      msgSendLatency.add(Date.now() - sentAt);

      const ok = res.status === 200 || res.status === 201;
      msgErrors.add(!ok);

      if (ok) {
        msgSent.add(1);
        try {
          const body = JSON.parse(res.body);
          const id = body.data?.id;
          if (id) pending.set(String(id), sentAt);
        } catch (_) {}
      }
    }, intervalMs);

    socket.setTimeout(() => socket.close(), totalDuration + 5000);
  });

  check(wsRes, { 'WS 101': (r) => r && r.status === 101 });
}
