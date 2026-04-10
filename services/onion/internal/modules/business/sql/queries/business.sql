-- name: GetBusinessProfile :one
SELECT COALESCE(site_name,'') AS site_name, COALESCE(logo_url,'') AS logo_url, COALESCE(contact_email,'') AS contact_email, COALESCE(contact_number,'') AS contact_number
FROM config
WHERE owner_id = ?
LIMIT 1;
