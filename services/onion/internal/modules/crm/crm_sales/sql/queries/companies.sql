-- name: FindCompanyByName :one
-- Names compare case-insensitively under the column's collation.
SELECT * FROM crm_companies WHERE business_id = ? AND name = ? ORDER BY id LIMIT 1;

-- name: CreateCompany :execresult
INSERT INTO crm_companies (business_id, owner_id, name, website, industry, size, city, country, assigned_to)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
