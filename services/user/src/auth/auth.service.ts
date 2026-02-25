import {
  Injectable,
  ConflictException,
  UnauthorizedException,
} from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { JwtService } from '@nestjs/jwt';
import * as bcrypt from 'bcrypt';
import { v7 as uuidv7 } from 'uuid';
import { User } from '../user.entity.js';
import { Jwt } from '../jwt.entity.js';
import { RegisterDto } from './dto/register.dto.js';
import { LoginDto } from './dto/login.dto.js';

interface JwtPayload {
  sub: string;
  username: string;
}

@Injectable()
export class AuthService {
  constructor(
    @InjectRepository(User)
    private readonly userRepo: Repository<User>,
    @InjectRepository(Jwt)
    private readonly jwtRepo: Repository<Jwt>,
    private readonly jwtService: JwtService,
  ) {}

  // ── Register ──
  async register(dto: RegisterDto) {
    const exists = await this.userRepo.findOne({
      where: { email: dto.email },
    });
    if (exists) {
      throw new ConflictException('Email déjà utilisé');
    }

    const user = this.userRepo.create({
      username: dto.username,
      display_name: dto.display_name,
      email: dto.email,
      password_hash: await bcrypt.hash(dto.password, 10),
    });
    await this.userRepo.save(user);

    const tokens = await this.generateTokens(user);
    return {
      user: {
        id: user.id,
        username: user.username,
        display_name: user.display_name,
        email: user.email,
      },
      ...tokens,
    };
  }

  // ── Login ──
  async login(dto: LoginDto) {
    const user = await this.userRepo.findOne({
      where: { email: dto.email },
    });
    if (!user) {
      throw new UnauthorizedException('Email ou mot de passe incorrect');
    }

    const valid = await bcrypt.compare(dto.password, user.password_hash);
    if (!valid) {
      throw new UnauthorizedException('Email ou mot de passe incorrect');
    }

    const tokens = await this.generateTokens(user);
    return {
      user: {
        id: user.id,
        username: user.username,
        display_name: user.display_name,
        email: user.email,
      },
      ...tokens,
    };
  }

  async validateToken(token: string) {
    try {
      const payload = this.jwtService.verify<JwtPayload>(token);
      const user = await this.userRepo.findOne({
        where: { id: payload.sub },
      });
      if (!user) {
        return { valid: false };
      }
      return {
        valid: true,
        user: {
          id: user.id,
          username: user.username,
          display_name: user.display_name,
          email: user.email,
        },
      };
    } catch {
      return { valid: false };
    }
  }

  async logout(userId: string) {
    await this.jwtRepo.update(
      { user_id: userId, revoked: false },
      { revoked: true },
    );
    return { message: 'Déconnexion réussie' };
  }

  async refresh(refreshToken: string) {
    const stored = await this.jwtRepo.findOne({
      where: { token: refreshToken, revoked: false },
    });
    if (!stored) {
      throw new UnauthorizedException('Refresh token invalide ou expiré');
    }
    if (stored.expires_at < new Date()) {
      stored.revoked = true;
      await this.jwtRepo.save(stored);
      throw new UnauthorizedException('Refresh token invalide ou expiré');
    }

    // Révoquer l'ancien refresh token
    stored.revoked = true;
    await this.jwtRepo.save(stored);

    const user = await this.userRepo.findOne({
      where: { id: stored.user_id },
    });
    if (!user) {
      throw new UnauthorizedException('Utilisateur introuvable');
    }

    return this.generateTokens(user);
  }

  private async generateTokens(user: User) {
    const payload = { sub: user.id, username: user.username, jti: uuidv7() };

    const accessToken = this.jwtService.sign(payload, { expiresIn: '15m' });
    const refreshToken = this.jwtService.sign(payload, { expiresIn: '7d' });

    // Stocker le refresh token en DB
    const jwt = this.jwtRepo.create({
      user_id: user.id,
      token: refreshToken,
      created_at: new Date(),
      expires_at: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000), // 7 jours
    });
    await this.jwtRepo.save(jwt);

    return { access_token: accessToken, refresh_token: refreshToken };
  }
}
