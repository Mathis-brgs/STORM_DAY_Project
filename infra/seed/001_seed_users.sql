-- Seed DB User (storm_user_db)
-- Crée des utilisateurs de test (le user-service utilise UUID pour id)
-- À exécuter après que le user-service ait créé les tables (TypeORM)

-- Activer l'extension uuid si nécessaire
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users de test (éviter les doublons avec ON CONFLICT)
INSERT INTO users (id, username, email, password_hash, avatar_url, created_at)
VALUES
  (gen_random_uuid(), 'alice', 'alice@test.com', '$2b$10$dummyhash1', NULL, NOW()),
  (gen_random_uuid(), 'bob', 'bob@test.com', '$2b$10$dummyhash2', NULL, NOW()),
  (gen_random_uuid(), 'charlie', 'charlie@test.com', '$2b$10$dummyhash3', NULL, NOW()),
  (gen_random_uuid(), 'diana', 'diana@test.com', '$2b$10$dummyhash4', NULL, NOW()),
  (gen_random_uuid(), 'eve', 'eve@test.com', '$2b$10$dummyhash5', NULL, NOW())
ON CONFLICT (email) DO NOTHING;
