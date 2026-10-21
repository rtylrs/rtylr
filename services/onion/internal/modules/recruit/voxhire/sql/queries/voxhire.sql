-- name: CreateVoxhireCompanyLead :execresult
INSERT INTO voxhire_company_leads (
    company_name,
    contact_name,
    contact_title,
    work_email,
    phone,
    role,
    headcount,
    timeline,
    service,
    message,
    agency_opt_in
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: CreateVoxhireTalentApplication :execresult
INSERT INTO voxhire_talent_applications (
    full_name,
    email,
    phone,
    country_of_residence,
    nationality,
    track,
    years_experience,
    notice_period,
    open_to,
    `current_role`,
    desired_roles,
    salary_expectation,
    work_authorization,
    work_history,
    education,
    linkedin_url,
    portfolio_url,
    pitch,
    consent,
    cv_key,
    cv_processing_status,
    cv_processing_error
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: FindRecentTalentApplicationID :one
SELECT id FROM voxhire_talent_applications
WHERE email = ? AND created_at >= NOW() - INTERVAL 10 MINUTE
ORDER BY id DESC
LIMIT 1;
