-- Business config fixture so any handler that loads "settings for owner X"
-- has a row to return rather than the unconfigured-fallback path.

INSERT INTO config (id, owner_id, business_id, site_name, contact_email, contact_number, about_content, meta_tags)
VALUES
    (1, 1001, 1001, 'Test Site', 'info@test.local', '+961-1-000000', 'Test about content', '[]')
ON DUPLICATE KEY UPDATE site_name = VALUES(site_name), meta_tags = VALUES(meta_tags);
