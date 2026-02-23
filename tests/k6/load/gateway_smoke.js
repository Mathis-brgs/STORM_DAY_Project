import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.GATEWAY_URL || 'http://localhost:8080';

export const options = {
  vus: 5,
  duration: '15s',
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.01'],
  },
};

export default function () {
  // Test health endpoint
  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'health status 200': (r) => r.status === 200,
    'health body OK': (r) => r.body === 'OK',
  });

  // Test root endpoint
  const rootRes = http.get(`${BASE_URL}/`);
  check(rootRes, {
    'root status 200': (r) => r.status === 200,
  });

  sleep(0.5);
}
