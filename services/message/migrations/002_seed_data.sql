-- Seed DB Message (storm_message_db)
-- Schéma cible:
--   - conversations
--   - conversations_users (table de liaison users <-> conversation)
--
-- Pas de messages à la création : les conversations démarrent vides.
-- Les noms de conversations sont vides par défaut (résolus dynamiquement par le gateway).

-- Conversations de démo (name vide = résolu dynamiquement depuis les membres)
INSERT INTO conversations (id, name, avatar_url, created_by)
VALUES
  (1, '', NULL, 'a0000001-0000-0000-0000-000000000001'),
  (2, '', NULL, 'a0000001-0000-0000-0000-000000000001'),
  (3, '', NULL, 'a0000004-0000-0000-0000-000000000004')
ON CONFLICT (id) DO NOTHING;

-- Memberships de démo (0=member, 1=admin, 2=owner)
INSERT INTO conversations_users (user_id, conversation_id, role)
VALUES
  ('a0000001-0000-0000-0000-000000000001', 1, 2),
  ('a0000002-0000-0000-0000-000000000002', 1, 0),
  ('a0000003-0000-0000-0000-000000000003', 1, 0),
  ('a0000001-0000-0000-0000-000000000001', 2, 2),
  ('a0000002-0000-0000-0000-000000000002', 2, 0),
  ('a0000004-0000-0000-0000-000000000004', 3, 2),
  ('a0000005-0000-0000-0000-000000000005', 3, 0)
ON CONFLICT DO NOTHING;
