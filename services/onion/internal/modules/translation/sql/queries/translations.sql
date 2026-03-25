-- name: GetTranslation :one
SELECT key_text, en, ar, fr
FROM translations
WHERE key_text = $1;

-- name: AddTranslation :exec
INSERT INTO translations (key_text, en, ar, fr, key_group)
VALUES ($1, $2, $3, $4, '')
ON CONFLICT (key_text, key_group)
DO UPDATE SET
    en = EXCLUDED.en,
    ar = EXCLUDED.ar,
    fr = EXCLUDED.fr;
