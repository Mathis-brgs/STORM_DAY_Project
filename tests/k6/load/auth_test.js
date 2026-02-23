import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.USER_URL || 'http://localhost:3000';

export const options = {
  stages: [
    { duration: '10s', target: 5 },
    { duration: '20s', target: 10 },
    { duration: '5s',  target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'],
    http_req_failed: ['rate<0.2'],
  },
};

export default function () {
  const uniqueId = `${__VU}_${__ITER}_${Date.now()}`;

  // Test register
  const registerRes = http.post(`${BASE_URL}/auth/register`, JSON.stringify({
    username: `user_${uniqueId}`,
    email: `user_${uniqueId}@test.com`,
    password: 'TestPassword123!',
  }), { headers: { 'Content-Type': 'application/json' } });

  check(registerRes, {
    'register status 2xx': (r) => r.status >= 200 && r.status < 300,
  });

  // Test login
  const loginRes = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
    email: `user_${uniqueId}@test.com`,
    password: 'TestPassword123!',
  }), { headers: { 'Content-Type': 'application/json' } });

  check(loginRes, {
    'login status 2xx': (r) => r.status >= 200 && r.status < 300,
    'login returns token': (r) => {
      try { return JSON.parse(r.body).access_token !== undefined; }
      catch(e) { return false; }
    },
  });

  sleep(0.5);
}
