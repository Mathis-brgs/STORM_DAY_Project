/**
 * k6 Load Test — Auth endpoints (login / register)
 * Target: Azure AKS gateway — http://51.138.200.66:8080
 *
 * Usage:
 *   k6 run auth-load-test.js
 *   k6 run --vus 50 --duration 60s auth-load-test.js   (surcharge)
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// ── Métriques custom ──────────────────────────────────────────────────────────
const loginErrorRate    = new Rate('login_errors');
const registerErrorRate = new Rate('register_errors');
const loginDuration     = new Trend('login_duration',    true);
const registerDuration  = new Trend('register_duration', true);

// ── Configuration des scénarios ───────────────────────────────────────────────
export const options = {
  scenarios: {
    // 1) Montée progressive (ramp-up) sur le login
    login_ramp: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 20 },  // montée à 20 VUs
        { duration: '60s', target: 20 },  // maintien 1 min
        { duration: '20s', target: 50 },  // pic à 50 VUs
        { duration: '30s', target: 50 },  // maintien pic
        { duration: '20s', target: 0  },  // descente
      ],
      exec: 'loginScenario',
    },

    // 2) Charge constante sur le register (10 VUs pendant 1 min)
    register_steady: {
      executor: 'constant-vus',
      vus: 10,
      duration: '90s',
      startTime: '30s',  // démarre après le début du login
      exec: 'registerScenario',
    },
  },

  thresholds: {
    // Login : p95 < 500ms, erreurs < 1%
    'login_duration':    ['p(95)<500'],
    'login_errors':      ['rate<0.01'],
    // Register : p95 < 800ms, erreurs < 5%
    'register_duration': ['p(95)<800'],
    'register_errors':   ['rate<0.05'],
    // Global HTTP
    'http_req_failed':   ['rate<0.05'],
    'http_req_duration': ['p(95)<1000'],
  },
};

const BASE_URL = 'http://51.138.200.66:8080';

const HEADERS = { 'Content-Type': 'application/json' };

// Comptes pré-existants en DB (créés via /auth/register)
const EXISTING_USERS = [
  { email: 'mathis@test.com',       password: 'Test1234!' },
  { email: 'pierre_final@test.com', password: 'Test1234!' },
  { email: 'test@storm.com',        password: 'Test1234!' },
  { email: 'test1@test.com',        password: 'Test1234!' },
  { email: 'dev@storm.io',          password: 'Storm2026!' },
];

// ── Scénario LOGIN ────────────────────────────────────────────────────────────
export function loginScenario() {
  // Choisit un utilisateur aléatoire parmi les comptes existants
  const user = EXISTING_USERS[Math.floor(Math.random() * EXISTING_USERS.length)];

  const res = http.post(
    `${BASE_URL}/auth/login`,
    JSON.stringify({ email: user.email, password: user.password }),
    { headers: HEADERS, tags: { name: 'auth_login' } },
  );

  loginDuration.add(res.timings.duration);

  const ok = check(res, {
    'login status 200':      (r) => r.status === 200,
    'login has token':       (r) => r.json('access_token') !== undefined,
    'login latency < 500ms': (r) => r.timings.duration < 500,
  });

  loginErrorRate.add(!ok);

  sleep(Math.random() * 2 + 0.5); // 0.5–2.5s entre requêtes
}

// ── Scénario REGISTER ─────────────────────────────────────────────────────────
let registerCounter = 0;

export function registerScenario() {
  // Génère un utilisateur unique par VU + itération
  const uid   = `${__VU}_${__ITER}_${Date.now()}`;
  const email = `loadtest_${uid}@storm.io`;

  const res = http.post(
    `${BASE_URL}/auth/register`,
    JSON.stringify({
      email:        email,
      password:     'LoadTest2026!',
      username:     `lt_${uid}`.slice(0, 30),
      display_name: `Load Tester ${__VU}`,
    }),
    { headers: HEADERS, tags: { name: 'auth_register' } },
  );

  registerDuration.add(res.timings.duration);

  const ok = check(res, {
    'register status 201':      (r) => r.status === 201,
    'register has token':       (r) => r.json('access_token') !== undefined,
    'register latency < 800ms': (r) => r.timings.duration < 800,
  });

  registerErrorRate.add(!ok);

  sleep(Math.random() * 3 + 1); // 1–4s entre requêtes (plus lent pour ne pas flood)
}
