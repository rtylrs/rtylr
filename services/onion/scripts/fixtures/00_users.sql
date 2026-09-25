-- Baseline users used by integration-style tests. Keep tenant data minimal:
-- one owner, one staff user, one chat-only user.
INSERT INTO users (id, business_id, email, password_hash, full_name, onboarded)
VALUES
    (1001, 1, 'owner@test.local',  '$2a$10$abcdefghijklmnopqrstuv', 'Test Owner', TRUE),
    (1002, 1, 'staff@test.local',  '$2a$10$abcdefghijklmnopqrstuv', 'Test Staff', TRUE),
    (1003, 1, 'chat@test.local',   '$2a$10$abcdefghijklmnopqrstuv', 'Chat User',  TRUE),
    (5201, 5100, 'benefit-alice@test.local', '$2a$10$abcdefghijklmnopqrstuv', 'Benefit Alice', TRUE),
    (5202, 5100, 'benefit-bob@test.local', '$2a$10$abcdefghijklmnopqrstuv', 'Benefit Bob', TRUE)
ON DUPLICATE KEY UPDATE email = VALUES(email);
