import { Controller, Post, Body, UseGuards, Request } from '@nestjs/common';
import { AuthGuard } from '@nestjs/passport';
import { MessagePattern } from '@nestjs/microservices';
import { AuthService } from './auth.service.js';
import { RegisterDto } from './dto/register.dto.js';
import { LoginDto } from './dto/login.dto.js';

@Controller('auth')
export class AuthController {
  constructor(private readonly authService: AuthService) {}

  // ── HTTP Endpoints ────────────────────────────────────

  // POST /auth/register
  // Body : { username, email, password }
  // Retourne : { user, access_token, refresh_token }
  @Post('register')
  register(@Body() dto: RegisterDto) {
    return this.authService.register(dto);
  }

  // POST /auth/login
  // Body : { email, password }
  // Retourne : { user, access_token, refresh_token }
  @Post('login')
  login(@Body() dto: LoginDto) {
    return this.authService.login(dto);
  }

  // POST /auth/refresh
  // Body : { refresh_token }
  // Retourne : { access_token, refresh_token }
  @Post('refresh')
  refresh(@Body('refresh_token') refreshToken: string) {
    return this.authService.refresh(refreshToken);
  }

  // POST /auth/logout
  // Requiert un JWT valide
  // Révoque tous les refresh tokens de l'utilisateur
  @UseGuards(AuthGuard('jwt'))
  @Post('logout')
  logout(@Request() req: { user: { id: string } }) {
    return this.authService.logout(req.user.id);
  }

  // ── NATS Handlers ─────────────────────────────────────
  // Ces méthodes sont appelées par les autres services via NATS
  // (pas par HTTP)

  // Le Gateway envoie un JWT pour le valider
  // Retourne : { valid: true, user: {...} } ou { valid: false }
  @MessagePattern('auth.validate')
  handleValidateToken(data: { token: string }) {
    return this.authService.validateToken(data.token);
  }
}
