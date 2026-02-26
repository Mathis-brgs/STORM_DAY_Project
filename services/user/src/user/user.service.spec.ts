import { Test, TestingModule } from '@nestjs/testing';
import { getRepositoryToken } from '@nestjs/typeorm';
import { NotFoundException, ForbiddenException } from '@nestjs/common';
import { UserService } from './user.service.js';
import { User } from '../user.entity.js';
import { REDIS_CLIENT } from './user.module.js';

describe('UserService', () => {
  let service: UserService;

  const mockUserRepo = {
    findOne: jest.fn(),
    save: jest.fn(),
  };

  const mockRedis = {
    get: jest.fn(),
    set: jest.fn(),
    del: jest.fn(),
  };

  const mockUser: User = {
    id: 'user-uuid-1',
    username: 'testuser',
    display_name: 'Test User',
    email: 'test@example.com',
    password_hash: 'hashed_password',
    avatar_url: null as unknown as string,
    created_at: new Date(),
    generateId: jest.fn(),
  };

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        UserService,
        { provide: getRepositoryToken(User), useValue: mockUserRepo },
        { provide: REDIS_CLIENT, useValue: mockRedis },
      ],
    }).compile();

    service = module.get<UserService>(UserService);
    jest.clearAllMocks();
  });

  // ── findById ──────────────────────────────────────────────────────────────

  describe('findById', () => {
    it('should throw NotFoundException if user does not exist', async () => {
      mockUserRepo.findOne.mockResolvedValue(null);

      await expect(service.findById('unknown-id')).rejects.toThrow(
        NotFoundException,
      );
    });

    it('should return the user without password_hash', async () => {
      mockUserRepo.findOne.mockResolvedValue(mockUser);

      const result = await service.findById('user-uuid-1');

      expect(result.id).toBe('user-uuid-1');
      expect(result.email).toBe('test@example.com');
      expect(result).not.toHaveProperty('password_hash');
    });
  });

  // ── update ────────────────────────────────────────────────────────────────

  describe('update', () => {
    it('should throw ForbiddenException if requester is not the owner', async () => {
      await expect(
        service.update('user-uuid-1', 'other-user-id', { username: 'new' }),
      ).rejects.toThrow(ForbiddenException);
    });

    it('should throw NotFoundException if user does not exist', async () => {
      mockUserRepo.findOne.mockResolvedValue(null);

      await expect(
        service.update('user-uuid-1', 'user-uuid-1', { username: 'new' }),
      ).rejects.toThrow(NotFoundException);
    });

    it('should update username when provided', async () => {
      const userToUpdate = { ...mockUser };
      mockUserRepo.findOne.mockResolvedValue(userToUpdate);
      mockUserRepo.save.mockResolvedValue(userToUpdate);

      const result = await service.update('user-uuid-1', 'user-uuid-1', {
        username: 'newusername',
      });

      expect(userToUpdate.username).toBe('newusername');
      expect(result.username).toBe('newusername');
    });

    it('should update avatar_url when provided', async () => {
      const userToUpdate = { ...mockUser };
      mockUserRepo.findOne.mockResolvedValue(userToUpdate);
      mockUserRepo.save.mockResolvedValue(userToUpdate);

      const result = await service.update('user-uuid-1', 'user-uuid-1', {
        avatar_url: 'https://example.com/avatar.png',
      });

      expect(userToUpdate.avatar_url).toBe('https://example.com/avatar.png');
      expect(result.avatar_url).toBe('https://example.com/avatar.png');
    });

    it('should not modify fields that are not provided', async () => {
      const userToUpdate = { ...mockUser, username: 'original' };
      mockUserRepo.findOne.mockResolvedValue(userToUpdate);
      mockUserRepo.save.mockResolvedValue(userToUpdate);

      await service.update('user-uuid-1', 'user-uuid-1', {
        avatar_url: 'https://example.com/avatar.png',
      });

      expect(userToUpdate.username).toBe('original');
    });

    it('should return updated user without password_hash', async () => {
      const userToUpdate = { ...mockUser };
      mockUserRepo.findOne.mockResolvedValue(userToUpdate);
      mockUserRepo.save.mockResolvedValue(userToUpdate);

      const result = await service.update('user-uuid-1', 'user-uuid-1', {
        username: 'newusername',
      });

      expect(result).not.toHaveProperty('password_hash');
    });
  });

  // ── setStatus ─────────────────────────────────────────────────────────────

  describe('setStatus', () => {
    it('should call redis.set with EX when status is online', async () => {
      mockRedis.set.mockResolvedValue('OK');

      const result = await service.setStatus('user-uuid-1', 'online');

      expect(mockRedis.set).toHaveBeenCalledWith(
        'user:status:user-uuid-1',
        'online',
        'EX',
        300,
      );
      expect(result).toEqual({ ok: true });
    });

    it('should call redis.del when status is offline', async () => {
      mockRedis.del.mockResolvedValue(1);

      const result = await service.setStatus('user-uuid-1', 'offline');

      expect(mockRedis.del).toHaveBeenCalledWith('user:status:user-uuid-1');
      expect(mockRedis.set).not.toHaveBeenCalled();
      expect(result).toEqual({ ok: true });
    });
  });

  // ── getStatus ─────────────────────────────────────────────────────────────

  describe('getStatus', () => {
    it('should return online when Redis has the key', async () => {
      mockRedis.get.mockResolvedValue('online');

      const result = await service.getStatus('user-uuid-1');

      expect(mockRedis.get).toHaveBeenCalledWith('user:status:user-uuid-1');
      expect(result).toEqual({ status: 'online' });
    });

    it('should return offline when Redis key is missing', async () => {
      mockRedis.get.mockResolvedValue(null);

      const result = await service.getStatus('user-uuid-1');

      expect(result).toEqual({ status: 'offline' });
    });
  });
});
