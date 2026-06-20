-- name: CountArticles :one
SELECT COUNT(*) FROM support_articles WHERE business_id = ?;

-- name: ListArticles :many
SELECT id, business_id, title, slug, body, category, is_published, views, created_at, updated_at
FROM support_articles WHERE business_id = ?
ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: GetArticle :one
SELECT id, business_id, title, slug, body, category, is_published, views, created_at, updated_at
FROM support_articles
WHERE business_id = sqlc.arg(business_id)
  AND (
    (sqlc.arg(article_id) > 0 AND id = sqlc.arg(article_id))
    OR (sqlc.arg(slug) <> '' AND slug = sqlc.arg(slug))
    OR (sqlc.arg(title) <> '' AND LOWER(title) = LOWER(sqlc.arg(title)))
  )
ORDER BY
  CASE
    WHEN sqlc.arg(article_id) > 0 AND id = sqlc.arg(article_id) THEN 0
    WHEN sqlc.arg(slug) <> '' AND slug = sqlc.arg(slug) THEN 1
    ELSE 2
  END,
  updated_at DESC
LIMIT 1;

-- name: CreateArticle :execresult
INSERT INTO support_articles (business_id, title, slug, body, category, is_published)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateArticle :exec
UPDATE support_articles SET title = ?, slug = ?, body = ?, category = ?, is_published = ?
WHERE business_id = ? AND id = ?;

-- name: DeleteArticle :exec
DELETE FROM support_articles WHERE business_id = ? AND id = ?;
