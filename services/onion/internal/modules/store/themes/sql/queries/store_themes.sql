-- name: ListThemes :many
SELECT id, owner_id, business_id, name, primary_color, font, header_style, footer_style, custom_css, created_at, updated_at
FROM store_themes WHERE business_id = ? ORDER BY name ASC
LIMIT 200;

-- name: CreateTheme :execresult
INSERT INTO store_themes (owner_id, business_id, name, primary_color, font, header_style, footer_style, custom_css) VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateTheme :exec
UPDATE store_themes SET name = ?, primary_color = ?, font = ?, header_style = ?, footer_style = ?, custom_css = ? WHERE business_id = ? AND id = ?;
