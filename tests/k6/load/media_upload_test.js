import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.MEDIA_URL || 'http://localhost:8080';

export const options = {
  stages: [
    { duration: '10s', target: 5 },   // ramp-up
    { duration: '20s', target: 10 },   // charge stable
    { duration: '5s',  target: 0 },    // ramp-down
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'],  // 95% des requetes < 2s
    http_req_failed: ['rate<0.1'],      // moins de 10% d'erreurs
  },
};

// Genere un petit fichier binaire simule (1KB JPEG header)
function generateFakeImage() {
  // JPEG magic bytes + random data
  const size = 1024;
  const data = new Uint8Array(size);
  data[0] = 0xFF; data[1] = 0xD8; data[2] = 0xFF; data[3] = 0xE0; // JPEG SOI + APP0
  for (let i = 4; i < size; i++) {
    data[i] = Math.floor(Math.random() * 256);
  }
  return data.buffer;
}

export default function () {
  const imageData = generateFakeImage();
  const filename = `test_${__VU}_${__ITER}.jpg`;

  const res = http.post(`${BASE_URL}/media/upload`, {
    file: http.file(imageData, filename, 'image/jpeg'),
  });

  check(res, {
    'upload status 200': (r) => r.status === 200,
    'response has mediaId': (r) => {
      try { return JSON.parse(r.body).mediaId !== undefined; }
      catch(e) { return false; }
    },
    'response has url': (r) => {
      try { return JSON.parse(r.body).url !== undefined; }
      catch(e) { return false; }
    },
  });

  sleep(0.5);
}
