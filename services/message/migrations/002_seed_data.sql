-- Seed DB Message (storm_message_db)
-- Users 1-5 : certains assignés à des groupes, d'autres non
-- Role: 0=member, 1=admin, 2=owner

-- Groups (memberships) - qui est dans quel groupe
-- Groupe 1 : alice (1), bob (2), charlie (3)
-- Groupe 2 : alice (1), bob (2)
-- Groupe 3 : diana (4), eve (5)
-- User 6 : aucun groupe (pour tester un user non assigné)
INSERT INTO groups (user_id, group_id, role) VALUES
  (1, 1, 2),  -- alice owner groupe 1
  (2, 1, 0),  -- bob member groupe 1
  (3, 1, 0),  -- charlie member groupe 1
  (1, 2, 2),  -- alice owner groupe 2
  (2, 2, 0),  -- bob member groupe 2
  (4, 3, 2),  -- diana owner groupe 3
  (5, 3, 0)   -- eve member groupe 3
;

-- Messages de test
INSERT INTO messages (sender_id, content, group_id) VALUES
  (1, 'Bienvenue dans le groupe 1 !', 1),
  (2, 'Merci alice !', 1),
  (3, 'Salut tout le monde', 1),
  (1, 'Conversation privée avec bob', 2),
  (2, 'Oui, juste nous deux ici', 2),
  (4, 'Groupe 3 - équipe projet', 3),
  (5, 'Parfait, on commence quand ?', 3);
