import { User } from './user.entity.js';

describe('User entity', () => {
  it('generateId() assigne un UUIDv7 si id est absent', () => {
    const user = new User();
    user.generateId();

    expect(user.id).toBeDefined();
    expect(user.id).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i,
    );
  });

  it('generateId() ne remplace pas un id déjà présent', () => {
    const user = new User();
    const existingId = '019c8fa2-9239-7719-96ff-b2bca228206e';
    user.id = existingId;

    user.generateId();

    expect(user.id).toBe(existingId);
  });
});
