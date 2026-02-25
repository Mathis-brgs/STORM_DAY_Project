import { Test, TestingModule } from '@nestjs/testing';
import { UserController } from './user.controller.js';
import { UserService } from './user.service.js';

describe('UserController', () => {
  let controller: UserController;

  const mockUserService = {
    findById: jest.fn(),
    update: jest.fn(),
  };

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      controllers: [UserController],
      providers: [{ provide: UserService, useValue: mockUserService }],
    }).compile();

    controller = module.get<UserController>(UserController);
    jest.clearAllMocks();
  });

  // ── findById ──────────────────────────────────────────────────────────────

  describe('findById', () => {
    it('délègue à userService.findById et retourne le résultat', async () => {
      const user = { id: 'user-uuid-1', username: 'testuser' };
      mockUserService.findById.mockResolvedValue(user);

      const result = await controller.findById({ id: 'user-uuid-1' });

      expect(mockUserService.findById).toHaveBeenCalledWith('user-uuid-1');
      expect(result).toBe(user);
    });
  });

  // ── update ────────────────────────────────────────────────────────────────

  describe('update', () => {
    it('délègue à userService.update et retourne le résultat', async () => {
      const updated = { id: 'user-uuid-1', username: 'newname' };
      mockUserService.update.mockResolvedValue(updated);

      const result = await controller.update({
        id: 'user-uuid-1',
        userId: 'user-uuid-1',
        dto: { username: 'newname' },
      });

      expect(mockUserService.update).toHaveBeenCalledWith(
        'user-uuid-1',
        'user-uuid-1',
        { username: 'newname' },
      );
      expect(result).toBe(updated);
    });
  });
});
