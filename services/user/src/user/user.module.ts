import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import Redis from 'ioredis';
import { UserController } from './user.controller.js';
import { UserService } from './user.service.js';
import { User } from '../user.entity.js';

export const REDIS_CLIENT = 'REDIS_CLIENT';

@Module({
  imports: [TypeOrmModule.forFeature([User])],
  controllers: [UserController],
  providers: [
    UserService,
    {
      provide: REDIS_CLIENT,
      useFactory: () =>
        new Redis({
          host: process.env.REDIS_HOST || 'localhost',
          port: parseInt(process.env.REDIS_PORT ?? '6379', 10),
          lazyConnect: true,
        }),
    },
  ],
  exports: [UserService],
})
export class UserModule {}
