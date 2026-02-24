import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { JwtModule } from '@nestjs/jwt';
import { PassportModule } from '@nestjs/passport';
import { AuthService } from './auth.service.js';
import { AuthController } from './auth.controller.js';
import { JwtStrategy } from './jwt.strategy.js';
import { User } from '../user.entity.js';
import { Jwt } from '../jwt.entity.js';

@Module({
  imports: [
    TypeOrmModule.forFeature([User, Jwt]),

    JwtModule.register({
      secret: process.env.JWT_SECRET || 'storm-secret-key',
      signOptions: { expiresIn: '15m' },
    }),

    PassportModule,
  ],
  controllers: [AuthController],
  providers: [AuthService, JwtStrategy],
  exports: [AuthService],
})
export class AuthModule {}
