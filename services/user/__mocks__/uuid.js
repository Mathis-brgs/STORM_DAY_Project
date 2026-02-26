'use strict';
const crypto = require('crypto');

// UUID v4 via crypto
function v4() {
  return crypto.randomUUID();
}

// UUID v7 : timestamp ms (48 bits) + version (4 bits) + random (74 bits)
function v7() {
  const ms = BigInt(Date.now());
  const rand = crypto.getRandomValues(new Uint8Array(10));

  const b = new Uint8Array(16);
  // Timestamp (48 bits)
  b[0] = Number((ms >> 40n) & 0xffn);
  b[1] = Number((ms >> 32n) & 0xffn);
  b[2] = Number((ms >> 24n) & 0xffn);
  b[3] = Number((ms >> 16n) & 0xffn);
  b[4] = Number((ms >> 8n) & 0xffn);
  b[5] = Number(ms & 0xffn);
  // Version 7
  b[6] = (rand[0] & 0x0f) | 0x70;
  b[7] = rand[1];
  // Variant 10xx
  b[8] = (rand[2] & 0x3f) | 0x80;
  b[9] = rand[3];
  // Random
  b[10] = rand[4]; b[11] = rand[5]; b[12] = rand[6];
  b[13] = rand[7]; b[14] = rand[8]; b[15] = rand[9];

  const h = Array.from(b).map((x) => x.toString(16).padStart(2, '0')).join('');
  return `${h.slice(0,8)}-${h.slice(8,12)}-${h.slice(12,16)}-${h.slice(16,20)}-${h.slice(20)}`;
}

module.exports = { v4, v7 };
