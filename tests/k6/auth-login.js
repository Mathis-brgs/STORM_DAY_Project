/**
 * Load Test : POST /auth/login
 * Objectif  : mesurer le throughput JWT en rafale
 *
 * Scénario :
 *   - 30s de montée en charge (0 → 50 VUs)
 *   - 1min à 50 VUs constants
 *   - 10s de descente
 *
 * Seuils :
 *   - 95% des requêtes < 500ms
 *   - Taux d'erreur HTTP < 1%
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate } from 'k6/metrics';

const loginDuration = new Trend('login_duration', true);
const loginErrors   = new Rate('login_errors');

export const options = {
  stages: [
    { duration: '30s', target: 50 },  // montée
    { duration: '1m',  target: 50 },  // plateau
    { duration: '10s', target: 0  },  // descente
  ],
  thresholds: {
    'login_duration': ['p(95)<500'],
    'login_errors':   ['rate<0.01'],
    'http_req_failed': ['rate<0.01'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Compte de test créé une fois avant la suite
// Assurez-vous qu'il existe en DB avant de lancer
const TEST_EMAIL    = __ENV.TEST_EMAIL    || 'k6-login@storm.dev';
const TEST_PASSWORD = __ENV.TEST_PASSWORD || 'k6password123';

export default function () {
  const payload = JSON.stringify({
    email:    TEST_EMAIL,
    password: TEST_PASSWORD,
  });

  const params = {
    headers: { 'Content-Type': 'application/json' },
    tags:    { name: 'auth_login' },
  };

  const res = http.post(`${BASE_URL}/auth/login`, payload, params);

  loginDuration.add(res.timings.duration);

  const ok = check(res, {
    'status 200 ou 201':     (r) => r.status === 200 || r.status === 201,
    'access_token présent':  (r) => {
      try { return !!JSON.parse(r.body).access_token; } catch { return false; }
    },
    'refresh_token présent': (r) => {
      try { return !!JSON.parse(r.body).refresh_token; } catch { return false; }
    },
  });

  loginErrors.add(!ok);

  sleep(1);
}
