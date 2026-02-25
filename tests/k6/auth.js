/**
 * K6 Load Test — Auth Service (via Gateway)
 *
 * Teste le flow complet : register → login → refresh → logout
 * Cible : http://localhost:8080 (gateway)
 *
 * Usage :
 *   k6 run tests/k6/auth.js
 *   k6 run --env BASE_URL=http://localhost:8080 tests/k6/auth.js
 */

import http from 'k6/http'
import { check, sleep } from 'k6'
import { Counter } from 'k6/metrics'

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080'

const errors = new Counter('errors')

export const options = {
  stages: [
    { duration: '30s', target: 50 },
    { duration: '1m',  target: 50 },
    { duration: '30s', target: 100 },
    { duration: '1m',  target: 100 },
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed:   ['rate<0.01'],
    errors:            ['count<10'],
  },
}

export default function () {
  const suffix = `${__VU}_${__ITER}_${Date.now()}`

  // ── Register ────────────────────────────────────────────
  const regRes = http.post(
    `${BASE_URL}/auth/register`,
    JSON.stringify({
      username: `k6_user_${suffix}`,
      email:    `k6_${suffix}@storm.local`,
      password: 'Storm1234!',
    }),
    { headers: { 'Content-Type': 'application/json' } }
  )

  const regOk = check(regRes, {
    'register: status 201':        (r) => r.status === 201,
    'register: has access_token':  (r) => !!r.json('access_token'),
    'register: has refresh_token': (r) => !!r.json('refresh_token'),
  })
  if (!regOk) {
    errors.add(1)
    sleep(1)
    return
  }

  const accessToken  = regRes.json('access_token')
  const refreshToken = regRes.json('refresh_token')
  sleep(0.5)

  // ── Login ────────────────────────────────────────────────
  const loginRes = http.post(
    `${BASE_URL}/auth/login`,
    JSON.stringify({
      email:    `k6_${suffix}@storm.local`,
      password: 'Storm1234!',
    }),
    { headers: { 'Content-Type': 'application/json' } }
  )
  check(loginRes, {
    'login: status 201':       (r) => r.status === 201,
    'login: has access_token': (r) => !!r.json('access_token'),
  })
  sleep(0.5)

  // ── Refresh ──────────────────────────────────────────────
  const refreshRes = http.post(
    `${BASE_URL}/auth/refresh`,
    JSON.stringify({ refresh_token: refreshToken }),
    { headers: { 'Content-Type': 'application/json' } }
  )
  check(refreshRes, {
    'refresh: status 201':       (r) => r.status === 201,
    'refresh: has access_token': (r) => !!r.json('access_token'),
  })
  sleep(0.5)

  // ── Logout ───────────────────────────────────────────────
  const logoutRes = http.post(
    `${BASE_URL}/auth/logout`,
    null,
    { headers: { Authorization: `Bearer ${accessToken}` } }
  )
  check(logoutRes, {
    'logout: status 2xx': (r) => r.status >= 200 && r.status < 300,
  })

  sleep(1)
}