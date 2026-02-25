/**
 * Load Test : POST /auth/refresh
 * Objectif  : tester la rotation de tokens en parallèle
 *
 * Particularité :
 *   Chaque VU fait login lors de sa première itération pour obtenir
 *   son propre refresh_token. Ensuite il tourne refresh → refresh → ...
 *   Comme les refresh tokens sont à usage unique (rotation), chaque
 *   VU doit conserver son propre token courant — d'où le login individuel
 *   plutôt qu'un setup() partagé (qui donnerait le même token à tous les VUs).
 *
 * Stages :
 *   - 20s : 0 → 20 VUs   (rotation modérée)
 *   - 1m  : 20 VUs        (plateau)
 *   - 10s : retour à 0
 *
 * Seuils :
 *   - 95% des refresh < 400ms
 *   - Taux d'erreur < 1%
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate } from 'k6/metrics';

const refreshDuration = new Trend('refresh_duration', true);
const refreshErrors   = new Rate('refresh_errors');

export const options = {
  stages: [
    { duration: '20s', target: 20 },
    { duration: '1m',  target: 20 },
    { duration: '10s', target: 0  },
  ],
  thresholds: {
    'refresh_duration': ['p(95)<400'],
    'refresh_errors':   ['rate<0.01'],
    'http_req_failed':  ['rate<0.01'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

const TEST_EMAIL    = __ENV.TEST_EMAIL    || 'k6-refresh@storm.dev';
const TEST_PASSWORD = __ENV.TEST_PASSWORD || 'k6password123';

// Stockage du refresh_token courant — local à chaque VU (variable de module)
let currentRefreshToken = null;

export default function () {
  // Première itération : chaque VU se logue pour obtenir son propre token
  if (currentRefreshToken === null) {
    const loginRes = http.post(
      `${BASE_URL}/auth/login`,
      JSON.stringify({ email: TEST_EMAIL, password: TEST_PASSWORD }),
      { headers: { 'Content-Type': 'application/json' } },
    );
    if (loginRes.status === 200 || loginRes.status === 201) {
      currentRefreshToken = JSON.parse(loginRes.body).refresh_token;
    } else {
      console.error(`[VU ${__VU}] Login initial échoué: ${loginRes.status}`);
      return;
    }
  }

  if (!currentRefreshToken) {
    console.error(`[VU ${__VU}] Pas de refresh_token disponible, skip`);
    return;
  }

  const payload = JSON.stringify({ refresh_token: currentRefreshToken });

  const params = {
    headers: { 'Content-Type': 'application/json' },
    tags:    { name: 'auth_refresh' },
  };

  const res = http.post(`${BASE_URL}/auth/refresh`, payload, params);

  refreshDuration.add(res.timings.duration);

  let newToken = null;
  const ok = check(res, {
    'status 200 ou 201':    (r) => r.status === 200 || r.status === 201,
    'nouveau access_token': (r) => {
      try {
        const body = JSON.parse(r.body);
        newToken = body.refresh_token;
        return !!body.access_token;
      } catch { return false; }
    },
    'nouveau refresh_token': () => !!newToken,
  });

  if (ok && newToken) {
    // Rotation : le prochain appel utilisera le nouveau refresh_token
    currentRefreshToken = newToken;
  } else {
    // Token révoqué ou expiré → re-login pour repartir avec un token frais
    console.warn(`[VU ${__VU}] Refresh échoué (${res.status}), re-login...`);
    currentRefreshToken = null; // sera rechargé à la prochaine itération
  }

  refreshErrors.add(!ok);

  sleep(2);
}
