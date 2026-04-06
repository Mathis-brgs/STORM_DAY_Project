import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';

const msgLatency = new Trend('msg_http_latency', true);
const msgErrors  = new Rate('msg_http_errors');
const msgSent    = new Counter('msg_http_sent');

const TARGET_VUS      = parseInt(__ENV.K6_VUS             || '500',  10);
const CONVERSATION_ID = parseInt(__ENV.K6_CONVERSATION_ID || '11',   10);
const TEST_DURATION   = __ENV.K6_DURATION                 || '3m';

export const options = {
  stages: [
    { duration: '30s', target: TARGET_VUS },
    { duration: TEST_DURATION, target: TARGET_VUS },
    { duration: '15s', target: 0 },
  ],
  thresholds: {
    'msg_http_latency': ['p(95)<500'],
    'msg_http_errors':  ['rate<0.05'],
  },
};

const BASE_URL = __ENV.BASE_URL      || 'http://localhost:8080';
const EMAIL    = __ENV.TEST_EMAIL    || 'k6-validate@storm.dev';
const PASSWORD = __ENV.TEST_PASSWORD || 'k6password123';

export function setup() {
  const res = http.post(
    `${BASE_URL}/auth/login`,
    JSON.stringify({ email: EMAIL, password: PASSWORD }),
    { headers: { 'Content-Type': 'application/json' } },
  );
  if (res.status !== 200 && res.status !== 201) {
    console.error(`[setup] Login failed: ${res.status}`);
    return {};
  }
  const body = JSON.parse(res.body);
  console.log(`[setup] Token OK, conversation_id=${CONVERSATION_ID}`);
  return { token: body.access_token };
}

export default function (data) {
  if (!data.token) { msgErrors.add(1); return; }

  const headers = {
    'Content-Type':  'application/json',
    'Authorization': `Bearer ${data.token}`,
  };

  const payload = JSON.stringify({
    conversation_id: CONVERSATION_ID,
    content: `k6 VU${__VU} iter${__ITER}`,
  });

  const start = Date.now();
  const res = http.post(`${BASE_URL}/api/messages`, payload, { headers });
  msgLatency.add(Date.now() - start);

  const ok = res.status === 200 || res.status === 201;
  msgErrors.add(!ok);
  if (ok) msgSent.add(1);

  check(res, { '201/200': (r) => ok });
}
