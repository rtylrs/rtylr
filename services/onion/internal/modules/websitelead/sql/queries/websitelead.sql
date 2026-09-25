-- name: CreateWebsiteLead :execresult
INSERT INTO voxire_website_leads (name, email, phone, company, service, project, budget, source)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: CreateNewsletterSubscriber :execresult
INSERT INTO voxire_newsletter_subscribers (email, source, post_slug, post_title, post_category, language)
VALUES (?, ?, ?, ?, ?, ?);
