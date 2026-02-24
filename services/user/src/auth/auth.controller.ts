import { Controller } from '@nestjs/common';
import { MessagePattern } from '@nestjs/microservices';
import { AuthService } from './auth.service.js';
import { RegisterDto } from './dto/register.dto.js';
import { LoginDto } from './dto/login.dto.js';

@Controller('auth')
export class AuthController {
  constructor(private readonly authService: AuthService) {}

  // ── NATS Endpoints ────────────────────────────────────
  // (Remplacent les endpoints HTTP)

  @MessagePattern('auth.register')
  register(dto: RegisterDto) {
    return this.authService.register(dto);
  }

  @MessagePattern('auth.login')
  login(dto: LoginDto) {
    return this.authService.login(dto);
  }

  @MessagePattern('auth.refresh')
  refresh(data: { refresh_token: string }) {
    return this.authService.refresh(data.refresh_token);
  }

  @MessagePattern('auth.logout')
  logout(data: { userId: string }) {
    return this.authService.logout(data.userId);
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
