-- name: CreateWebsiteLead :execresult
INSERT INTO voxire_website_leads (name, email, phone, company, service, project, budget, source)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);
