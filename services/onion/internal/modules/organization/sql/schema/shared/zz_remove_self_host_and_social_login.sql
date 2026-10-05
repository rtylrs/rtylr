DROP TABLE IF EXISTS self_host_licenses;
DELETE FROM feature_flags WHERE feature_key = 'auth.social_login_enabled';
