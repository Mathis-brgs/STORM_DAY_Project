import { JwtStrategy } from './jwt.strategy.js';

describe('JwtStrategy', () => {
  let strategy: JwtStrategy;

  beforeEach(() => {
    strategy = new JwtStrategy();
  });

  it('retourne { id, username } à partir du payload JWT', () => {
    const payload = { sub: 'user-uuid-1', username: 'testuser' };

    const result = strategy.validate(payload);

    expect(result).toEqual({ id: 'user-uuid-1', username: 'testuser' });
  });

  it('mappe bien sub → id', () => {
    const payload = {
      sub: '019c8fa2-9239-7719-96ff-b2bca228206e',
      username: 'mathis',
    };

    const result = strategy.validate(payload);

    expect(result.id).toBe('019c8fa2-9239-7719-96ff-b2bca228206e');
    expect(result.username).toBe('mathis');
  });
});
