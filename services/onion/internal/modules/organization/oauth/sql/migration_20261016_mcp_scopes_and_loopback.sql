UPDATE oauth_clients
SET scopes = CONCAT(scopes, ' pos:read pos:write')
WHERE id = 'rtylr-mcp'
  AND scopes NOT LIKE '%pos:read%';

UPDATE oauth_tokens
SET scope = CONCAT(scope, ' pos:read pos:write')
WHERE client_id = 'rtylr-mcp'
  AND revoked = 0
  AND scope NOT LIKE '%pos:read%';

UPDATE oauth_clients
SET redirect_uris = '["http://127.0.0.1/callback"]'
WHERE id = 'rtylr-mcp'
  AND redirect_uris = '["http://127.0.0.1"]';
