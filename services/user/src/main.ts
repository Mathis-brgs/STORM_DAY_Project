import { NestFactory } from '@nestjs/core';
import { Transport } from '@nestjs/microservices';
import { ValidationPipe } from '@nestjs/common';
import { AppModule } from './app.module.js';

async function bootstrap() {
  // 1. Créer l'app HTTP (endpoints REST classiques)
  const app = await NestFactory.create(AppModule);

  // 2. Activer la validation des DTOs
  app.useGlobalPipes(new ValidationPipe({ whitelist: true }));

  // 3. Connecter le transport NATS pour écouter les messages du Gateway
  app.connectMicroservice({
    transport: Transport.NATS,
    options: {
      servers: [process.env.NATS_URL || 'nats://localhost:4222'],
    },
  });

  // 4. Démarrer les deux : NATS listener + HTTP server
  await app.startAllMicroservices();
  await app.listen(process.env.PORT ?? 3000);

  console.log(`User service HTTP on port ${process.env.PORT ?? 3000}`);
  console.log(
    `User service NATS connected to ${process.env.NATS_URL || 'nats://localhost:4222'}`,
  );
}
bootstrap();
