/**
 * Load Test : Connexions WebSocket simultanées
 * Objectif  : mesurer combien de connexions WS le Gateway Go peut maintenir
 *
 * Fonctionnement :
 *   setup()      → login une fois → récupère access_token + userId
 *   default()    → chaque VU ouvre une WS, reste connecté 30s (idle),
 *                  répond aux pings du Gateway, puis ferme
 *
 * 1 VU = 1 connexion WS maintenue ouverte pendant toute la durée du test.
 * Lancer avec : k6 run --vus N --duration 30s tests/k6/ws-connections.js
 *
 * Paliers recommandés :
 *   k6 run --vus 500   --duration 30s tests/k6/ws-connections.js
 *   k6 run --vus 1000  --duration 30s tests/k6/ws-connections.js
 *   k6 run --vus 5000  --duration 30s tests/k6/ws-connections.js
 *   k6 run --vus 10000 --duration 30s tests/k6/ws-connections.js
 *
 * Prérequis OS (macOS) — augmenter les file descriptors avant de lancer :
 *   ulimit -n 65536
 *
 * Seuils :
 *   - 95% des handshakes WS < 200ms
 *   - Taux d'erreur connexion < 1%
 */

import ws   from 'k6/ws';
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate, Gauge } from 'k6/metrics';

const wsConnectTime  = new Trend('ws_connect_time', true);
const wsConnErrors   = new Rate('ws_connection_errors');
const wsActiveSess   = new Gauge('ws_active_sessions');

const TARGET_VUS = parseInt(__ENV.K6_VUS || '30000', 10);

export const options = {
  stages: [
    { duration: '5m', target: TARGET_VUS },  // ramp up progressif
    { duration: '5m', target: TARGET_VUS },  // maintien
    { duration: '30s', target: 0 },          // descente
  ],
  thresholds: {
    'ws_connect_time':      ['p(95)<200'],
    'ws_connection_errors': ['rate<0.01'],
    'ws_connecting':        ['p(95)<200'],
  },
};

const BASE_URL = __ENV.BASE_URL    || 'http://localhost:8080';
const WS_URL   = __ENV.WS_URL     || 'ws://localhost:8080';
const EMAIL    = __ENV.TEST_EMAIL  || 'k6-validate@storm.dev';
const PASSWORD = __ENV.TEST_PASSWORD || 'k6password123';

// setup() : login une seule fois — token partagé entre tous les VUs
// (token valide 15min, suffit pour le test de connexion)
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
  console.log(`[setup] Token obtenu pour ${EMAIL}`);
  return {
    token:  body.access_token,
    userId: body.user?.id,
  };
}

const RAMP_SECONDS = parseInt(__ENV.K6_RAMP_SECONDS || '180', 10);

export default function (data) {
  if (!data.token) {
    console.error(`[VU ${__VU}] Pas de token — setup a échoué`);
    wsConnErrors.add(1);
    return;
  }

  // Étale les connexions sur RAMP_SECONDS pour un ramp progressif
  sleep(Math.random() * RAMP_SECONDS);

  const url = `${WS_URL}/ws?token=${data.token}`;

  const startTime = Date.now();

  const res = ws.connect(url, {}, (socket) => {
    socket.on('open', () => {
      const elapsed = Date.now() - startTime;
      wsConnectTime.add(elapsed);
      wsActiveSess.add(1);

      check(socket, { 'WS connecté': () => true });

      // Rejoindre sa room personnelle (le Gateway le fait auto côté serveur,
      // mais on peut aussi envoyer un join explicite pour une room partagée)
      socket.send(JSON.stringify({
        action: 'join',
        room:   'k6-load-test',
      }));

      // Rester connecté pendant toute la durée du test (géré par --duration)
      // On répond aux pings du Gateway (toutes les 30s)
    });

    socket.on('ping', () => {
      socket.ping(); // réponse pong automatique via k6
    });

    socket.on('error', (e) => {
      console.error(`[VU ${__VU}] WS error: ${e.error()}`);
      wsConnErrors.add(1);
      wsActiveSess.add(-1);
    });

    socket.on('close', () => {
      wsActiveSess.add(-1);
    });

    // Garder la connexion ouverte pendant toute la durée du test (11min)
    socket.setTimeout(() => {
      socket.close();
    }, 660000); // 11 min > durée totale (5m ramp + 5m hold + 30s descend)
  });

  check(res, {
    'handshake HTTP 101': (r) => r && r.status === 101,
  });

  const ok = res && res.status === 101;
  wsConnErrors.add(!ok);
}
