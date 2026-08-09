-- name: GetConfig :many
SELECT
    owner_id,
    business_id,
    site_name,
    logo_url,
    favicon_url,
    meta_tags,
    contact_email,
    contact_number,
    sender_email,
    website,
    maps_url,
    address,
    city,
    zipcode,
    country,
    instagram,
    maintenance_mode,
    maintenance_notice,
    addons_enabled,
    enable_reviews,
    enable_wishlist,
    enable_newsletter_signup,
    about_content,
    footer_sub_text
FROM config WHERE business_id = ?;

-- name: UpdateConfig :exec
UPDATE config SET
    site_name = ?,
    logo_url = ?,
    favicon_url = ?,
    meta_tags = ?,
    contact_email = ?,
    contact_number = ?,
    sender_email = ?,
    website = ?,
    maps_url = ?,
    address = ?,
    city = ?,
    zipcode = ?,
    country = ?,
    instagram = ?,
    maintenance_mode = ?,
    maintenance_notice = ?,
    addons_enabled = ?,
    enable_reviews = ?,
    enable_wishlist = ?,
    enable_newsletter_signup = ?,
    about_content = ?,
    footer_sub_text = ?
WHERE business_id = ?;

-- name: CreateConfig :exec
INSERT INTO config (
    owner_id,
    business_id,
    site_name,
    logo_url,
    favicon_url,
    meta_tags,
    contact_email,
    contact_number,
    sender_email,
    website,
    maps_url,
    address,
    city,
    zipcode,
    country,
    instagram,
    maintenance_mode,
    maintenance_notice,
    addons_enabled,
    enable_reviews,
    enable_wishlist,
    enable_newsletter_signup,
    about_content,
    footer_sub_text
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: GetBusinessSlugByID :one
SELECT slug FROM businesses WHERE id = ? LIMIT 1;
