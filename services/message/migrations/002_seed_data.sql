-- Seed DB Message (storm_message_db)
-- groups : user_id = UUID, group_id = int
-- messages : sender_id = UUID, group_id = int

-- Groups (memberships) - user_id en UUID de démo, group_id int
INSERT INTO groups (user_id, group_id, role) VALUES
  ('a0000001-0000-0000-0000-000000000001', 1, 2),
  ('a0000002-0000-0000-0000-000000000002', 1, 0),
  ('a0000003-0000-0000-0000-000000000003', 1, 0),
  ('a0000001-0000-0000-0000-000000000001', 2, 2),
  ('a0000002-0000-0000-0000-000000000002', 2, 0),
  ('a0000004-0000-0000-0000-000000000004', 3, 2),
  ('a0000005-0000-0000-0000-000000000005', 3, 0)
;

-- Messages de test : sender_id UUID, group_id int
INSERT INTO messages (sender_id, content, group_id) VALUES
  ('a0000001-0000-0000-0000-000000000001', 'Bienvenue dans le groupe 1 !', 1),
  ('a0000002-0000-0000-0000-000000000002', 'Merci alice !', 1),
  ('a0000003-0000-0000-0000-000000000003', 'Salut tout le monde', 1),
  ('a0000001-0000-0000-0000-000000000001', 'Conversation privée avec bob', 2),
  ('a0000002-0000-0000-0000-000000000002', 'Oui, juste nous deux ici', 2),
  ('a0000004-0000-0000-0000-000000000004', 'Groupe 3 - équipe projet', 3),
  ('a0000005-0000-0000-0000-000000000005', 'Parfait, on commence quand ?', 3);
