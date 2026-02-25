import { Test, TestingModule } from '@nestjs/testing';
import { AuthController } from './auth.controller.js';
import { AuthService } from './auth.service.js';

describe('AuthController', () => {
  let controller: AuthController;

  const mockAuthService = {
    register: jest.fn(),
    login: jest.fn(),
    refresh: jest.fn(),
    logout: jest.fn(),
    validateToken: jest.fn(),
  };

  const mockTokens = {
    user: {
      id: 'user-uuid-1',
      username: 'testuser',
      email: 'test@example.com',
    },
    access_token: 'fake-access',
    refresh_token: 'fake-refresh',
  };

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      controllers: [AuthController],
      providers: [{ provide: AuthService, useValue: mockAuthService }],
    }).compile();

    controller = module.get<AuthController>(AuthController);
    jest.clearAllMocks();
  });

  // ── register ──────────────────────────────────────────────────────────────

  describe('register', () => {
    it('délègue à authService.register et retourne le résultat', async () => {
      const dto = {
        username: 'testuser',
        email: 'test@example.com',
        password: '123456',
      };
      mockAuthService.register.mockResolvedValue(mockTokens);

      const result = await controller.register(dto);

      expect(mockAuthService.register).toHaveBeenCalledWith(dto);
      expect(result).toBe(mockTokens);
    });
  });

  // ── login ─────────────────────────────────────────────────────────────────

  describe('login', () => {
    it('délègue à authService.login et retourne le résultat', async () => {
      const dto = { email: 'test@example.com', password: '123456' };
      mockAuthService.login.mockResolvedValue(mockTokens);

      const result = await controller.login(dto);

      expect(mockAuthService.login).toHaveBeenCalledWith(dto);
      expect(result).toBe(mockTokens);
    });
  });

  // ── refresh ───────────────────────────────────────────────────────────────

  describe('refresh', () => {
    it('délègue à authService.refresh avec le refresh_token', async () => {
      const refreshResult = {
        access_token: 'new-access',
        refresh_token: 'new-refresh',
      };
      mockAuthService.refresh.mockResolvedValue(refreshResult);

      const result = await controller.refresh({
        refresh_token: 'fake-refresh',
      });

      expect(mockAuthService.refresh).toHaveBeenCalledWith('fake-refresh');
      expect(result).toBe(refreshResult);
    });
  });

  // ── logout ────────────────────────────────────────────────────────────────

  describe('logout', () => {
    it('délègue à authService.logout avec le userId', async () => {
      const logoutResult = { message: 'Déconnexion réussie' };
      mockAuthService.logout.mockResolvedValue(logoutResult);

      const result = await controller.logout({ userId: 'user-uuid-1' });

      expect(mockAuthService.logout).toHaveBeenCalledWith('user-uuid-1');
      expect(result).toBe(logoutResult);
    });
  });

  // ── handleValidateToken ───────────────────────────────────────────────────

  describe('handleValidateToken', () => {
    it('retourne { valid: true, user } sur token valide', async () => {
      const validateResult = {
        valid: true,
        user: {
          id: 'user-uuid-1',
          username: 'testuser',
        },
      };
      mockAuthService.validateToken.mockResolvedValue(validateResult);

      const result = await controller.handleValidateToken({
        token: 'valid-token',
      });

      expect(mockAuthService.validateToken).toHaveBeenCalledWith('valid-token');
      expect(result).toBe(validateResult);
    });

    it('retourne { valid: false } sur token invalide', async () => {
      mockAuthService.validateToken.mockResolvedValue({ valid: false });

      const result = await controller.handleValidateToken({
        token: 'bad-token',
      });

      expect(result).toEqual({ valid: false });
    });
  });
});
