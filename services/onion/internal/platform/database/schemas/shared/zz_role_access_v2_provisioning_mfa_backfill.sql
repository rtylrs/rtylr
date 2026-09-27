INSERT INTO user_provisioning_states
    (user_id, business_id, mode, state, email_verified, password_change_required, mfa_required, mfa_enrolled)
SELECT account.id, business.id, 'direct_credentials', 'active', TRUE, FALSE,
    COALESCE(account.two_factor_enabled, FALSE),
    EXISTS (
        SELECT 1
        FROM user_mfa_enrollments AS enrollment
        WHERE enrollment.user_id = account.id
          AND enrollment.method = 'totp'
          AND enrollment.is_active = TRUE
    )
FROM users AS account
LEFT JOIN (
    SELECT access.user_id, MIN(access.business_id) AS business_id
    FROM user_business_access AS access
    JOIN businesses AS member_business ON member_business.id = access.business_id
    WHERE access.business_id > 0
    GROUP BY access.user_id
) AS legacy_membership ON legacy_membership.user_id = account.id
LEFT JOIN (
    SELECT member.user_id, MIN(member.business_id) AS business_id
    FROM authorization_business_members AS member
    JOIN businesses AS member_business ON member_business.id = member.business_id
    WHERE member.status = 'active'
    GROUP BY member.user_id
) AS active_membership ON active_membership.user_id = account.id
LEFT JOIN businesses AS account_business ON account_business.id = account.business_id
JOIN businesses AS business ON business.id = COALESCE(
    account_business.id,
    legacy_membership.business_id,
    active_membership.business_id
)
LEFT JOIN user_provisioning_states AS provisioning ON provisioning.user_id = account.id
WHERE provisioning.user_id IS NULL;
