-- name: GetConfig :many
SELECT 
    owner_id,
    site_name,
    logo_url,
    favicon_url,
    meta_tags,
    contact_email,
    contact_number,
    instagram,
    maintenance_mode,
    maintenance_notice,
    addons_enabled,
    enable_reviews,
    enable_wishlist,
    enable_newsletter_signup,
    about_content,
    footer_sub_text
FROM config WHERE owner_id = $1;

-- name: CreateConfig :one
INSERT INTO config (
    owner_id,
    site_name,
    logo_url,
    favicon_url,
    meta_tags,
    contact_email,
    contact_number,
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
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
) RETURNING 
    owner_id,
    site_name,
    logo_url,
    favicon_url,
    meta_tags,
    contact_email,
    contact_number,
    instagram,
    maintenance_mode,
    maintenance_notice,
    addons_enabled,
    enable_reviews,
    enable_wishlist,
    enable_newsletter_signup,
    about_content,
    footer_sub_text; 