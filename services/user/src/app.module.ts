import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { AppController } from './app.controller';
import { AppService } from './app.service';
import { User } from './user.entity';
import { Jwt } from './jwt.entity';

@Module({
  imports: [
    TypeOrmModule.forRoot({
      type: 'postgres',
      host: process.env.DB_HOST || 'localhost',
      port: parseInt(process.env.DB_PORT ?? '5432', 10),
      username: process.env.DB_USER || 'storm',
      password: process.env.DB_PASSWORD || 'password',
      database: process.env.DB_NAME || 'storm_user_db',
      entities: [User, Jwt],
      synchronize: true, // DEV only - creates tables automatically
    }),
  ],
  controllers: [AppController],
  providers: [AppService],
})
export class AppModule {}
