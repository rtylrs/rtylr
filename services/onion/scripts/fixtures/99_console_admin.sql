-- Internal console admin account for console.voxire.com
-- Password omitted; fixture authentication is disabled.
INSERT INTO users (email, password_hash, full_name, onboarded, is_console_admin)
VALUES (
    'dev@voxire.com',
    'fixture-auth-disabled',
    'Voxire Dev',
    TRUE,
    TRUE
)
ON DUPLICATE KEY UPDATE
    is_console_admin = TRUE,
    password_hash    = VALUES(password_hash);
