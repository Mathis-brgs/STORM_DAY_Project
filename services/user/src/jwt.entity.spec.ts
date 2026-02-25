import { Jwt } from './jwt.entity.js';

describe('Jwt entity', () => {
  it('generateId() assigne un UUIDv7 si id est absent', () => {
    const jwt = new Jwt();
    jwt.generateId();

    expect(jwt.id).toBeDefined();
    expect(jwt.id).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i,
    );
  });

  it('generateId() ne remplace pas un id déjà présent', () => {
    const jwt = new Jwt();
    const existingId = '019c8fa2-9239-7719-96ff-b2bca228206e';
    jwt.id = existingId;

    jwt.generateId();

    expect(jwt.id).toBe(existingId);
  });
});
