/**
 * k6 Load Test — WebSocket simultaneous connections
 * Tests: login → get token → connect WS → stay connected
 *
 * Usage:
 *   k6 run ws-connections-test.js                        # default 100 VUs
 *   k6 run --vus 500 --duration 60s ws-connections-test.js
 */

import http from 'k6/http';
import ws from 'k6/ws';
import { check, sleep } from 'k6';
import { Counter, Gauge, Rate } from 'k6/metrics';

const wsConnected    = new Counter('ws_connected_total');
const wsErrors       = new Counter('ws_errors_total');
const loginErrors    = new Rate('login_error_rate');
const wsActiveConns  = new Gauge('ws_active_connections');

export const options = {
  scenarios: {
    ws_load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '20s', target: 50  },   // montée à 50 WS
        { duration: '30s', target: 100 },   // montée à 100 WS
        { duration: '60s', target: 100 },   // maintien 100 connexions simultanées
        { duration: '20s', target: 0   },   // descente
      ],
      exec: 'wsScenario',
    },
  },
  thresholds: {
    'login_error_rate':    ['rate<0.05'],
    'ws_errors_total':     ['count<20'],
    'http_req_duration':   ['p(95)<2000'],
  },
};

const BASE_URL = 'http://51.138.200.66:8080';
const WS_URL   = 'ws://51.138.200.66:8080/ws';

export function wsScenario() {
  // ── 1. Login pour obtenir un token ──────────────────────────────────────────
  const uid   = `${__VU}_${__ITER}`;
  const email = `loadtest_ws_${uid}@storm.io`;

  // Enregistrement d'un user unique par VU/iter
  const regRes = http.post(`${BASE_URL}/auth/register`,
    JSON.stringify({
      email,
      password:     'WsLoad2026!',
      username:     `ws_${uid}`.slice(0, 20),
      display_name: `WS User ${__VU}`,
    }),
    { headers: { 'Content-Type': 'application/json' }, timeout: '15s' },
  );

  const loginOk = check(regRes, {
    'register/login OK': (r) => r.status === 201 || r.status === 200,
  });

  if (!loginOk) {
    // Fallback: essayer de se connecter avec un compte existant
    const loginRes = http.post(`${BASE_URL}/auth/login`,
      JSON.stringify({ email: 'dev@storm.io', password: 'Storm2026!' }),
      { headers: { 'Content-Type': 'application/json' }, timeout: '15s' },
    );
    loginErrors.add(loginRes.status !== 200);
    if (loginRes.status !== 200) return;

    const token = loginRes.json('access_token');
    connectWS(token);
    return;
  }

  loginErrors.add(false);
  const token = regRes.json('access_token');
  connectWS(token);
}

function connectWS(token) {
  if (!token) { wsErrors.add(1); return; }

  const url = `${WS_URL}?token=${token}`;

  wsActiveConns.add(1);

  const res = ws.connect(url, { timeout: '10s' }, function(socket) {
    socket.on('open', () => {
      wsConnected.add(1);

      // Reste connecté 30s en envoyant un ping toutes les 10s
      socket.setInterval(() => {
        socket.ping();
      }, 10000);

      // Ferme proprement après 30s
      socket.setTimeout(() => {
        socket.close();
      }, 30000);
    });

    socket.on('message', (data) => {
      // Messages reçus du serveur (events NATS broadcastés)
      check(data, { 'message received': (d) => d !== undefined });
    });

    socket.on('error', (e) => {
      wsErrors.add(1);
    });

    socket.on('close', () => {
      wsActiveConns.add(-1);
    });
  });

  check(res, {
    'WS connected (101)': (r) => r && r.status === 101,
  });
}
