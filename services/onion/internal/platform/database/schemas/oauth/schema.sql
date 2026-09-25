CREATE TABLE IF NOT EXISTS oauth_clients (
    id VARCHAR(128) NOT NULL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    redirect_uris TEXT NOT NULL,
    scopes TEXT NOT NULL,
    is_public TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS oauth_auth_codes (
    code VARCHAR(128) NOT NULL PRIMARY KEY,
    client_id VARCHAR(128) NOT NULL,
    user_id BIGINT NOT NULL,
    redirect_uri TEXT NOT NULL,
    scope TEXT NOT NULL,
    code_challenge VARCHAR(256) NOT NULL,
    code_challenge_method VARCHAR(10) NOT NULL DEFAULT 'S256',
    expires_at TIMESTAMP NOT NULL,
    used TINYINT(1) NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_oauth_auth_codes_client (client_id),
    INDEX idx_oauth_auth_codes_user (user_id),
    INDEX idx_oauth_auth_codes_expires (expires_at)
);

CREATE TABLE IF NOT EXISTS oauth_tokens (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    access_token VARCHAR(512) NOT NULL,
    refresh_token VARCHAR(128) NOT NULL,
    client_id VARCHAR(128) NOT NULL,
    user_id BIGINT NOT NULL,
    scope TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked TINYINT(1) NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_oauth_tokens_access (access_token(64)),
    INDEX idx_oauth_tokens_refresh (refresh_token),
    INDEX idx_oauth_tokens_user (user_id),
    UNIQUE KEY uk_oauth_refresh_token (refresh_token)
);

INSERT IGNORE INTO oauth_clients (id, name, redirect_uris, scopes, is_public) VALUES
('rtylr-mcp', 'rtylr MCP Server', '["http://127.0.0.1"]', 'flow:read flow:write crm:read crm:write finance:read finance:write erp:read erp:write', 1);
