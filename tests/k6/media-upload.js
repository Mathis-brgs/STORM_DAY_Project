/**
 * K6 Load Test — Media Upload (via Gateway)
 *
 * Teste l'upload de fichiers : login → upload image → vérifier l'URL retournée
 * Cible : http://localhost:8080 (gateway → NATS → media service)
 *
 * PRÉREQUIS : le gateway doit exposer POST /media/upload qui forward vers NATS.
 *
 * Usage :
 *   k6 run tests/k6/media-upload.js
 *   k6 run --env BASE_URL=http://localhost:8080 tests/k6/media-upload.js
 */

import http from 'k6/http'
import { check, sleep } from 'k6'
import { Counter } from 'k6/metrics'

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080'

const uploadErrors = new Counter('upload_errors')

export const options = {
  stages: [
    { duration: '20s', target: 10 },
    { duration: '1m',  target: 10 },
    { duration: '20s', target: 30 },
    { duration: '30s', target: 30 },
    { duration: '10s', target: 0  },
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'],  // uploads peuvent être lents
    http_req_failed:   ['rate<0.05'],
    upload_errors:     ['count<5'],
  },
}

// Minimal 1×1 PNG blanc (67 bytes) encodé en base64
const TINY_PNG_B64 =
  'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=='

// setup() s'exécute une seule fois avant tous les VUs
export function setup() {
  // Créer le compte de test (idempotent, peut échouer si déjà existant)
  http.post(
    `${BASE_URL}/auth/register`,
    JSON.stringify({
      username: 'k6_media_tester',
      email:    'k6_media@storm.local',
      password: 'Storm1234!',
    }),
    { headers: { 'Content-Type': 'application/json' } }
  )

  const loginRes = http.post(
    `${BASE_URL}/auth/login`,
    JSON.stringify({ email: 'k6_media@storm.local', password: 'Storm1234!' }),
    { headers: { 'Content-Type': 'application/json' } }
  )

  const token = loginRes.status === 201 ? loginRes.json('access_token') : null
  return { token }
}

export default function (data) {
  const token = data?.token || null

  const headers = { 'Content-Type': 'application/json' }
  if (token) headers['Authorization'] = `Bearer ${token}`

  const payload = JSON.stringify({
    filename:    `k6_${__VU}_${__ITER}.png`,
    contentType: 'image/png',
    dataBase64:  TINY_PNG_B64,
  })

  const res = http.post(`${BASE_URL}/media/upload`, payload, { headers })

  const ok = check(res, {
    'upload: status 200 ou 201': (r) => r.status === 200 || r.status === 201,
    'upload: has url':           (r) => {
      try { return !!r.json('url') } catch { return false }
    },
    'upload: has mediaId':       (r) => {
      try { return !!r.json('mediaId') } catch { return false }
    },
  })
  if (!ok) uploadErrors.add(1)

  sleep(2)
}