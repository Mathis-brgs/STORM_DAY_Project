/**
 * Load Test : validate token (via GET /users/:id protégé)
 * Objectif  : simuler un pic de validation JWT — cas réaliste Gateway
 *
 * Scénario :
 *   - setup() : login une fois → récupère l'access_token
 *   - Chaque VU envoie GET /users/:id avec le Bearer token
 *     → force le Gateway à appeler auth.validate via NATS à chaque requête
 *
 * Stages :
 *   - 20s : 0 → 100 VUs  (pic soudain)
 *   - 40s : 100 VUs       (plateau)
 *   - 10s : retour à 0
 *
 * Seuils :
 *   - 95% des validations < 300ms
 *   - Taux d'erreur < 1%
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate } from 'k6/metrics';

const validateDuration = new Trend('validate_duration', true);
const validateErrors   = new Rate('validate_errors');

export const options = {
  stages: [
    { duration: '20s', target: 100 },
    { duration: '40s', target: 100 },
    { duration: '10s', target: 0   },
  ],
  thresholds: {
    'validate_duration': ['p(95)<300'],
    'validate_errors':   ['rate<0.01'],
    'http_req_failed':   ['rate<0.01'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

const TEST_EMAIL    = __ENV.TEST_EMAIL    || 'k6-validate@storm.dev';
const TEST_PASSWORD = __ENV.TEST_PASSWORD || 'k6password123';

// setup() : exécuté une seule fois avant les VUs — récupère token + userId
export function setup() {
  const loginRes = http.post(
    `${BASE_URL}/auth/login`,
    JSON.stringify({ email: TEST_EMAIL, password: TEST_PASSWORD }),
    { headers: { 'Content-Type': 'application/json' } },
  );

  if (loginRes.status !== 200 && loginRes.status !== 201) {
    console.error(`[setup] Login échoué : ${loginRes.status} ${loginRes.body}`);
    return {};
  }

  const body = JSON.parse(loginRes.body);
  return {
    token:  body.access_token,
    userId: body.user?.id,
  };
}

export default function (data) {
  if (!data.token) {
    console.error('Pas de token — setup a échoué, skip VU');
    return;
  }

  const params = {
    headers: {
      'Content-Type':  'application/json',
      'Authorization': `Bearer ${data.token}`,
    },
    tags: { name: 'auth_validate' },
  };

  // GET /users/:id → le Gateway appelle auth.validate via NATS
  const res = http.get(`${BASE_URL}/users/${data.userId}`, params);

  validateDuration.add(res.timings.duration);

  const ok = check(res, {
    'status 200':       (r) => r.status === 200,
    'body non vide':    (r) => r.body && r.body.length > 0,
    'pas unauthorized': (r) => r.status !== 401,
  });

  validateErrors.add(!ok);

  sleep(0.5);
}
