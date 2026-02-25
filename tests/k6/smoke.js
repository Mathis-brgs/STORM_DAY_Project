/**
 * K6 Smoke Test — Vérification rapide que tous les services répondent
 *
 * Usage :
 *   k6 run tests/k6/smoke.js
 */

import http from 'k6/http'
import ws   from 'k6/ws'
import { check } from 'k6'

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080'
const WS_URL   = BASE_URL.replace(/^http/, 'ws') + '/ws'

export const options = {
  vus:        1,
  iterations: 1,
  thresholds: {
    http_req_failed: ['rate<0.01'],
  },
}

export default function () {
  // Health check gateway
  const health = http.get(`${BASE_URL}/`)
  check(health, { 'gateway: status 200': (r) => r.status === 200 })

  // Auth register (utilisateur unique)
  const ts = Date.now()
  const regRes = http.post(
    `${BASE_URL}/auth/register`,
    JSON.stringify({
      username:     `smoke_${ts}`,
      display_name: `Smoke ${ts}`,
      email:        `smoke_${ts}@storm.local`,
      password:     'Storm1234!',
    }),
    { headers: { 'Content-Type': 'application/json' } }
  )
  check(regRes, { 'auth register: status 201': (r) => r.status === 201 })

  // WebSocket connect
  const res = ws.connect(WS_URL, {}, function (socket) {
    socket.on('open', function () {
      socket.close()
    })
    socket.on('error', function () {
      socket.close()
    })
    socket.setTimeout(function () { socket.close() }, 3000)
  })
  check(res, { 'gateway ws: upgrade 101': (r) => r && r.status === 101 })
}