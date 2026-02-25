/**
 * K6 Load Test — Gateway WebSocket
 *
 * Teste les connexions WebSocket : connect → join room → send messages → close
 * Cible : ws://localhost:8080/ws (gateway)
 *
 * Usage :
 *   k6 run tests/k6/gateway-ws.js
 *   k6 run --env BASE_URL=http://localhost:8080 tests/k6/gateway-ws.js
 */

import ws    from 'k6/ws'
import { check, sleep } from 'k6'
import { Counter, Trend } from 'k6/metrics'

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080'
const WS_URL   = BASE_URL.replace(/^http/, 'ws') + '/ws'

const wsConnectTime      = new Trend('ws_connect_time_ms')
const wsMessagesReceived = new Counter('ws_messages_received')
const wsErrors           = new Counter('ws_errors')

export const options = {
  stages: [
    { duration: '20s', target: 50  },
    { duration: '1m',  target: 50  },
    { duration: '20s', target: 200 },
    { duration: '1m',  target: 200 },
    { duration: '10s', target: 0   },
  ],
  thresholds: {
    ws_connect_time_ms:  ['p(95)<1000'],
    ws_errors:           ['count<20'],
    ws_messages_received: ['count>0'],
  },
}

export default function () {
  // 10 rooms — les VUs se répartissent dessus
  const room     = `room_${(__VU % 10) + 1}`
  const username = `k6_user_${__VU}`
  const t0       = Date.now()

  const res = ws.connect(WS_URL, {}, function (socket) {
    wsConnectTime.add(Date.now() - t0)

    socket.on('open', function () {
      // Rejoindre la room
      socket.send(JSON.stringify({
        action:  'join',
        room:    room,
        user:    username,
        content: '',
      }))

      // Envoyer 5 messages à 1s d'intervalle puis se déconnecter
      let count = 0
      const interval = socket.setInterval(function () {
        socket.send(JSON.stringify({
          action:  'message',
          room:    room,
          user:    username,
          content: `msg ${count} de ${username}`,
        }))
        count++
        if (count >= 5) {
          socket.clearInterval(interval)
          socket.close()
        }
      }, 1000)
    })

    socket.on('message', function () {
      wsMessagesReceived.add(1)
    })

    socket.on('error', function (e) {
      wsErrors.add(1)
    })

    // Timeout de sécurité
    socket.setTimeout(function () {
      socket.close()
    }, 10000)
  })

  check(res, {
    'ws: upgrade 101': (r) => r && r.status === 101,
  })

  sleep(1)
}