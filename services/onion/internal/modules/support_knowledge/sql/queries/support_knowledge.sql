-- name: CountArticles :one
SELECT COUNT(*) FROM support_articles WHERE business_id = ?;

-- name: ListArticles :many
SELECT id, business_id, title, slug, body, category, is_published, views, created_at, updated_at
FROM support_articles WHERE business_id = ?
ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: CreateArticle :execresult
INSERT INTO support_articles (business_id, title, slug, body, category, is_published)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateArticle :exec
UPDATE support_articles SET title = ?, slug = ?, body = ?, category = ?, is_published = ?
WHERE business_id = ? AND id = ?;

-- name: DeleteArticle :exec
DELETE FROM support_articles WHERE business_id = ? AND id = ?;
