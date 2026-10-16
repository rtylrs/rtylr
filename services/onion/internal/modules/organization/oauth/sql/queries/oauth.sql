-- name: GetOAuthClient :one
SELECT id, name, redirect_uris, scopes, is_public, created_at
FROM oauth_clients WHERE id = ?;

-- name: SaveAuthCode :exec
INSERT INTO oauth_auth_codes
(code, client_id, user_id, redirect_uri, scope, code_challenge, code_challenge_method, expires_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetAuthCode :one
SELECT code, client_id, user_id, redirect_uri, scope, code_challenge, code_challenge_method, expires_at, used
FROM oauth_auth_codes WHERE code = ?;

-- name: MarkAuthCodeUsed :exec
UPDATE oauth_auth_codes SET used = 1 WHERE code = ?;

-- name: ConsumeAuthCode :execresult
UPDATE oauth_auth_codes SET used = 1 WHERE code = ? AND used = 0;

-- name: SaveToken :exec
INSERT INTO oauth_tokens (access_token, refresh_token, client_id, user_id, scope, expires_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetTokenByRefresh :one
SELECT id, access_token, refresh_token, client_id, user_id, scope, expires_at, revoked
FROM oauth_tokens WHERE refresh_token = ?;

-- name: RevokeToken :exec
UPDATE oauth_tokens SET revoked = 1 WHERE refresh_token = ?;

-- name: ConsumeRefreshToken :execresult
UPDATE oauth_tokens SET revoked = 1 WHERE refresh_token = ? AND revoked = 0;

-- cross-schema read from shared users table — password hash for resource owner password grant
-- name: GetUserPasswordHash :one
SELECT password_hash FROM users WHERE id = ?;

-- cross-schema read from shared users table — resolve user id from email for password grant
-- name: GetUserIDByEmail :one
SELECT id FROM users WHERE email = ? ORDER BY id ASC LIMIT 1;

-- name: GetUserTokenEpoch :one
SELECT token_epoch FROM users WHERE id = ?;

-- name: GetUserBusinessID :one
SELECT business_id FROM users WHERE id = ?;
