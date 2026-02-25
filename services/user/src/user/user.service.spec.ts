import { Test, TestingModule } from '@nestjs/testing';
import { getRepositoryToken } from '@nestjs/typeorm';
import { NotFoundException, ForbiddenException } from '@nestjs/common';
import { UserService } from './user.service.js';
import { User } from '../user.entity.js';

describe('UserService', () => {
  let service: UserService;

  const mockUserRepo = {
    findOne: jest.fn(),
    save: jest.fn(),
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
});
