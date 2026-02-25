import { Test, TestingModule } from '@nestjs/testing';
import {
  INestApplication,
  ValidationPipe,
  ConflictException,
  UnauthorizedException,
  NotFoundException,
  ForbiddenException,
} from '@nestjs/common';
import { DataSource } from 'typeorm';
import { AppModule } from '../src/app.module.js';
import { AuthService } from '../src/auth/auth.service.js';
import { UserService } from '../src/user/user.service.js';

// ─────────────────────────────────────────────────────────────────────────────
// Tests d'intégration : AuthService + UserService avec PostgreSQL réel
//
// Prérequis : PostgreSQL accessible via les env vars (ou valeurs par défaut) :
//   DB_HOST=localhost  DB_PORT=5432  DB_USER=storm
//   DB_PASSWORD=password  DB_NAME=storm_user_db
//   JWT_SECRET=storm-secret-key
//
// CI : GitHub Actions → service postgres + ces env vars dans le job
// ─────────────────────────────────────────────────────────────────────────────

const TEST_EMAIL = 'integ_test@storm.dev';
const TEST_PASSWORD = 'IntegPass123';
const TEST_USERNAME = 'integ_tester';

describe('Auth & User — Integration', () => {
  let app: INestApplication;
  let authService: AuthService;
  let userService: UserService;
  let dataSource: DataSource;

  // État partagé entre les tests (flow réaliste)
  let userId: string;
  let accessToken: string;
  let refreshToken: string;

  // ── Setup ─────────────────────────────────────────────────────────────────

  beforeAll(async () => {
    const module: TestingModule = await Test.createTestingModule({
      imports: [AppModule],
    }).compile();

    app = module.createNestApplication();
    app.useGlobalPipes(new ValidationPipe({ whitelist: true }));
    await app.init();

    authService = module.get<AuthService>(AuthService);
    userService = module.get<UserService>(UserService);
    dataSource = module.get<DataSource>(DataSource);

    // Nettoyer les données de test si un run précédent a échoué
    await dataSource.query(
      `DELETE FROM jwt WHERE user_id IN (SELECT id FROM users WHERE email = $1)`,
      [TEST_EMAIL],
    );
    await dataSource.query(`DELETE FROM users WHERE email = $1`, [TEST_EMAIL]);
  }, 30_000);

  afterAll(async () => {
    // Nettoyage systématique après les tests
    await dataSource.query(
      `DELETE FROM jwt WHERE user_id IN (SELECT id FROM users WHERE email = $1)`,
      [TEST_EMAIL],
    );
    await dataSource.query(`DELETE FROM users WHERE email = $1`, [TEST_EMAIL]);
    await app.close();
  });

  // ── 1. Register ───────────────────────────────────────────────────────────

  describe('register', () => {
    it('crée un utilisateur et retourne access_token + refresh_token', async () => {
      const result = await authService.register({
        username: TEST_USERNAME,
        email: TEST_EMAIL,
        password: TEST_PASSWORD,
      });

      expect(result.user.email).toBe(TEST_EMAIL);
      expect(result.user.username).toBe(TEST_USERNAME);
      expect(result.user).not.toHaveProperty('password_hash');
      expect(result.access_token).toBeDefined();
      expect(result.refresh_token).toBeDefined();

      userId = result.user.id;
      accessToken = result.access_token;
      refreshToken = result.refresh_token;
    });

    it('throw ConflictException sur email déjà utilisé', async () => {
      await expect(
        authService.register({
          username: 'autre',
          email: TEST_EMAIL,
          password: TEST_PASSWORD,
        }),
      ).rejects.toThrow(ConflictException);
    });
  });

  // ── 2. Login ──────────────────────────────────────────────────────────────

  describe('login', () => {
    it('retourne des tokens avec les bons identifiants', async () => {
      const result = await authService.login({
        email: TEST_EMAIL,
        password: TEST_PASSWORD,
      });

      expect(result.user.id).toBe(userId);
      expect(result.user).not.toHaveProperty('password_hash');
      expect(result.access_token).toBeDefined();
      expect(result.refresh_token).toBeDefined();

      refreshToken = result.refresh_token;
    });

    it('throw UnauthorizedException sur mauvais mot de passe', async () => {
      await expect(
        authService.login({ email: TEST_EMAIL, password: 'wrong' }),
      ).rejects.toThrow(UnauthorizedException);
    });

    it('throw UnauthorizedException sur email inconnu', async () => {
      await expect(
        authService.login({
          email: 'nobody@storm.dev',
          password: TEST_PASSWORD,
        }),
      ).rejects.toThrow(UnauthorizedException);
    });
  });

  // ── 3. ValidateToken ──────────────────────────────────────────────────────

  describe('validateToken', () => {
    it('retourne { valid: true, user } sur un access token valide', async () => {
      const result = await authService.validateToken(accessToken);

      expect(result.valid).toBe(true);
      expect(result.user?.id).toBe(userId);
      expect(result.user?.email).toBe(TEST_EMAIL);
      expect(result.user).not.toHaveProperty('password_hash');
    });

    it('retourne { valid: false } sur un token malformé', async () => {
      const result = await authService.validateToken('not-a-jwt');

      expect(result.valid).toBe(false);
      expect(result.user).toBeUndefined();
    });

    it('retourne { valid: false } sur un token signé avec une mauvaise clé', async () => {
      const result = await authService.validateToken(
        'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJmYWtlIiwiaWF0IjoxfQ.bad_signature',
      );

      expect(result.valid).toBe(false);
    });
  });

  // ── 4. Refresh token ──────────────────────────────────────────────────────

  describe('refresh', () => {
    it("retourne de nouveaux tokens et révoque l'ancien refresh token", async () => {
      const oldRefresh = refreshToken;
      const result = await authService.refresh(oldRefresh);

      expect(result.access_token).toBeDefined();
      expect(result.refresh_token).toBeDefined();
      expect(result.refresh_token).not.toBe(oldRefresh);

      accessToken = result.access_token;
      refreshToken = result.refresh_token;
    });

    it('throw UnauthorizedException si le refresh token est inconnu', async () => {
      await expect(authService.refresh('fake-refresh-token')).rejects.toThrow(
        UnauthorizedException,
      );
    });
  });

  // ── 5. UserService ────────────────────────────────────────────────────────

  describe('UserService.findById', () => {
    it('retourne le profil sans password_hash', async () => {
      const result = await userService.findById(userId);

      expect(result.id).toBe(userId);
      expect(result.email).toBe(TEST_EMAIL);
      expect(result.username).toBeDefined();
      expect(result).not.toHaveProperty('password_hash');
    });

    it('throw NotFoundException sur un id inconnu', async () => {
      await expect(
        userService.findById('00000000-0000-0000-0000-000000000000'),
      ).rejects.toThrow(NotFoundException);
    });
  });

  describe('UserService.update', () => {
    it('met à jour le username', async () => {
      const result = await userService.update(userId, userId, {
        username: 'updated_tester',
      });

      expect(result.username).toBe('updated_tester');
      expect(result).not.toHaveProperty('password_hash');
    });

    it("throw ForbiddenException si le requester n'est pas le propriétaire", async () => {
      await expect(
        userService.update(userId, 'other-user-id', { username: 'hacked' }),
      ).rejects.toThrow(ForbiddenException);
    });
  });

  // ── 6. Logout ─────────────────────────────────────────────────────────────

  describe('logout', () => {
    it('révoque tous les tokens et empêche le refresh suivant', async () => {
      const logoutResult = await authService.logout(userId);
      expect(logoutResult.message).toBeDefined();

      // Le refresh token doit être révoqué → throw
      await expect(authService.refresh(refreshToken)).rejects.toThrow(
        UnauthorizedException,
      );
    });

    it("l'access token reste valide pendant sa durée de vie (15min — stateless JWT)", async () => {
      // Design choice documenté : l'access token n'est pas révoqué au logout
      // Il expire naturellement après 15 min
      const result = await authService.validateToken(accessToken);
      expect(result.valid).toBe(true);
    });
  });
});
