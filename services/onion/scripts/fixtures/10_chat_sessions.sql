-- Two chat sessions for user 1003, plus one message each, so that
-- ListUserSessions and GetSessionMessages have something to return.
INSERT INTO chat_sessions (id, user_id, session_id, title)
VALUES
    (5001, 1003, 'fixt-session-recent', 'Recent session'),
    (5002, 1003, 'fixt-session-older',  'Older session')
ON DUPLICATE KEY UPDATE title = VALUES(title);

UPDATE chat_sessions SET last_activity = NOW()                       WHERE id = 5001;
UPDATE chat_sessions SET last_activity = DATE_SUB(NOW(), INTERVAL 7 DAY) WHERE id = 5002;

INSERT INTO conversation_messages (id, session_id, user_id, role, content, tokens_used)
VALUES
    (6001, 'fixt-session-recent', 1003, 'user',      'hello',          1),
    (6002, 'fixt-session-recent', 1003, 'assistant', 'hi there',       2),
    (6003, 'fixt-session-older',  1003, 'user',      'old message',    2)
ON DUPLICATE KEY UPDATE content = VALUES(content);
