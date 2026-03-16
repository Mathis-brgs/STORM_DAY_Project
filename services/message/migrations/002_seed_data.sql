-- Seed DB Message (storm_message_db)
-- Schéma cible:
--   - conversations
--   - conversations_users (table de liaison users <-> conversation)
--   - messages (conversation_id)

-- Conversations de démo
INSERT INTO conversations (id, name, avatar_url, created_by)
VALUES
  (1, 'Groupe 1', 'https://cdn.example.com/conversations/1.png', 'a0000001-0000-0000-0000-000000000001'),
  (2, 'Discussion privée Alice/Bob', 'https://cdn.example.com/conversations/2.png', 'a0000001-0000-0000-0000-000000000001'),
  (3, 'Équipe projet', 'https://cdn.example.com/conversations/3.png', 'a0000004-0000-0000-0000-000000000004')
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

-- Messages de test
INSERT INTO messages (sender_id, content, conversation_id)
VALUES
  ('a0000001-0000-0000-0000-000000000001', 'Bienvenue dans la conversation 1 !', 1),
  ('a0000002-0000-0000-0000-000000000002', 'Merci alice !', 1),
  ('a0000003-0000-0000-0000-000000000003', 'Salut tout le monde', 1),
  ('a0000001-0000-0000-0000-000000000001', 'Conversation privée avec bob', 2),
  ('a0000002-0000-0000-0000-000000000002', 'Oui, juste nous deux ici', 2),
  ('a0000004-0000-0000-0000-000000000004', 'Conversation équipe projet', 3),
  ('a0000005-0000-0000-0000-000000000005', 'Parfait, on commence quand ?', 3);
