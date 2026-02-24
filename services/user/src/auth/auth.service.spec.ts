import { Test, TestingModule } from '@nestjs/testing';
import { getRepositoryToken } from '@nestjs/typeorm';
import { JwtService } from '@nestjs/jwt';
import { ConflictException, UnauthorizedException } from '@nestjs/common';
import * as bcrypt from 'bcrypt';
import { AuthService } from './auth.service.js';
import { User } from '../user.entity.js';
import { Jwt } from '../jwt.entity.js';

jest.mock('bcrypt');

describe('AuthService', () => {
  let service: AuthService;

  const mockUserRepo = {
    findOne: jest.fn(),
    create: jest.fn(),
    save: jest.fn(),
    update: jest.fn(),
  };

  const mockJwtRepo = {
    findOne: jest.fn(),
    create: jest.fn(),
    save: jest.fn(),
    update: jest.fn(),
  };

  const mockJwtService = {
    sign: jest.fn(),
    verify: jest.fn(),
  };

  const mockUser: User = {
    id: 'user-uuid-1',
    username: 'testuser',
    email: 'test@example.com',
    password_hash: 'hashed_password',
    avatar_url: null as unknown as string,
    created_at: new Date(),
    generateId: jest.fn(),
  };

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        AuthService,
        { provide: getRepositoryToken(User), useValue: mockUserRepo },
        { provide: getRepositoryToken(Jwt), useValue: mockJwtRepo },
        { provide: JwtService, useValue: mockJwtService },
      ],
    }).compile();

    service = module.get<AuthService>(AuthService);
    jest.clearAllMocks();
  });

  // ── register ──────────────────────────────────────────────────────────────

  describe('register', () => {
    it('should throw ConflictException if email already exists', async () => {
      mockUserRepo.findOne.mockResolvedValue(mockUser);

      await expect(
        service.register({
          username: 'testuser',
          email: 'test@example.com',
          password: '123456',
        }),
      ).rejects.toThrow(ConflictException);
    });

    it('should create a user and return tokens', async () => {
      mockUserRepo.findOne.mockResolvedValue(null);
      mockUserRepo.create.mockReturnValue(mockUser);
      mockUserRepo.save.mockResolvedValue(mockUser);
      (bcrypt.hash as jest.Mock).mockResolvedValue('hashed_password');
      mockJwtService.sign.mockReturnValue('fake-token');
      mockJwtRepo.create.mockReturnValue({});
      mockJwtRepo.save.mockResolvedValue({});

      const result = await service.register({
        username: 'testuser',
        email: 'test@example.com',
        password: '123456',
      });

      expect(result.user.email).toBe('test@example.com');
      expect(result.access_token).toBe('fake-token');
    });
  });

  // ── login ─────────────────────────────────────────────────────────────────

  describe('login', () => {
    it('should throw UnauthorizedException if user not found', async () => {
      mockUserRepo.findOne.mockResolvedValue(null);

      await expect(
        service.login({ email: 'no@one.com', password: '123' }),
      ).rejects.toThrow(UnauthorizedException);
    });

    it('should throw UnauthorizedException if password is wrong', async () => {
      mockUserRepo.findOne.mockResolvedValue(mockUser);
      (bcrypt.compare as jest.Mock).mockResolvedValue(false);

      await expect(
        service.login({ email: 'test@example.com', password: 'wrong' }),
      ).rejects.toThrow(UnauthorizedException);
    });

    it('should return tokens on valid credentials', async () => {
      mockUserRepo.findOne.mockResolvedValue(mockUser);
      (bcrypt.compare as jest.Mock).mockResolvedValue(true);
      mockJwtService.sign.mockReturnValue('fake-token');
      mockJwtRepo.create.mockReturnValue({});
      mockJwtRepo.save.mockResolvedValue({});

      const result = await service.login({
        email: 'test@example.com',
        password: '123456',
      });

      expect(result.user.id).toBe('user-uuid-1');
      expect(result.access_token).toBe('fake-token');
    });
  });

  // ── validateToken ─────────────────────────────────────────────────────────

  describe('validateToken', () => {
    it('should return { valid: false } if token is invalid', async () => {
      mockJwtService.verify.mockImplementation(() => {
        throw new Error('invalid token');
      });

      const result = await service.validateToken('bad-token');

      expect(result.valid).toBe(false);
    });

    it('should return { valid: false } if user no longer exists', async () => {
      mockJwtService.verify.mockReturnValue({
        sub: 'user-uuid-1',
        username: 'testuser',
      });
      mockUserRepo.findOne.mockResolvedValue(null);

      const result = await service.validateToken('good-token');

      expect(result.valid).toBe(false);
    });

    it('should return { valid: true, user } on valid token', async () => {
      mockJwtService.verify.mockReturnValue({
        sub: 'user-uuid-1',
        username: 'testuser',
      });
      mockUserRepo.findOne.mockResolvedValue(mockUser);

      const result = await service.validateToken('good-token');

      expect(result.valid).toBe(true);
      expect(result.user?.id).toBe('user-uuid-1');
    });
  });

  // ── logout ────────────────────────────────────────────────────────────────

  describe('logout', () => {
    it('should revoke all tokens for the user', async () => {
      mockJwtRepo.update.mockResolvedValue({ affected: 2 });

      const result = await service.logout('user-uuid-1');

      expect(mockJwtRepo.update).toHaveBeenCalledWith(
        { user_id: 'user-uuid-1', revoked: false },
        { revoked: true },
      );
      expect(result.message).toBeDefined();
    });
  });

  // ── refresh ───────────────────────────────────────────────────────────────

  describe('refresh', () => {
    it('should throw UnauthorizedException if token not found', async () => {
      mockJwtRepo.findOne.mockResolvedValue(null);

      await expect(service.refresh('invalid-token')).rejects.toThrow(
        UnauthorizedException,
      );
    });

    it('should revoke expired token and throw UnauthorizedException', async () => {
      const expiredToken = {
        token: 'expired-token',
        revoked: false,
        expires_at: new Date(Date.now() - 1000),
        user_id: 'user-uuid-1',
      };
      mockJwtRepo.findOne.mockResolvedValue(expiredToken);
      mockJwtRepo.save.mockResolvedValue({});

      await expect(service.refresh('expired-token')).rejects.toThrow(
        UnauthorizedException,
      );

      expect(expiredToken.revoked).toBe(true);
      expect(mockJwtRepo.save).toHaveBeenCalledWith(expiredToken);
    });

    it('should return new tokens on valid refresh token', async () => {
      const storedToken = {
        token: 'valid-refresh',
        revoked: false,
        expires_at: new Date(Date.now() + 100_000),
        user_id: 'user-uuid-1',
      };
      mockJwtRepo.findOne.mockResolvedValue(storedToken);
      mockJwtRepo.save.mockResolvedValue({});
      mockUserRepo.findOne.mockResolvedValue(mockUser);
      mockJwtService.sign.mockReturnValue('new-token');
      mockJwtRepo.create.mockReturnValue({});

      const result = await service.refresh('valid-refresh');

      expect(storedToken.revoked).toBe(true);
      expect(mockJwtRepo.save).toHaveBeenCalledWith(
        expect.objectContaining({ token: 'valid-refresh', revoked: true }),
      );
      expect(result.access_token).toBe('new-token');
    });
  });
});
