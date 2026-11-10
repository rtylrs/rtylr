//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateConfig` (Go `CreateConfigParams`).
pub struct CreateConfigParams {
    pub owner_id: i32,
    pub business_id: i64,
    pub site_name: Option<String>,
    pub logo_url: Option<String>,
    pub favicon_url: Option<String>,
    pub meta_tags: String,
    pub contact_email: String,
    pub contact_number: String,
    pub sender_email: Option<String>,
    pub email_provider: String,
    pub sender_name: Option<String>,
    pub reply_to_email: Option<String>,
    pub email_notifications_enabled: bool,
    pub smtp_host: Option<String>,
    pub smtp_port: Option<i32>,
    pub smtp_username: Option<String>,
    pub smtp_password_hash: Option<String>,
    pub website: Option<String>,
    pub maps_url: Option<String>,
    pub address: Option<String>,
    pub city: Option<String>,
    pub zipcode: Option<String>,
    pub country: Option<String>,
    pub instagram: Option<String>,
    pub maintenance_mode: Option<bool>,
    pub maintenance_notice: Option<String>,
    pub addons_enabled: Option<bool>,
    pub enable_reviews: Option<bool>,
    pub enable_wishlist: Option<bool>,
    pub enable_newsletter_signup: Option<bool>,
    pub about_content: String,
    pub footer_sub_text: Option<String>,
}

/// `CreateConfig`: `exec`.
pub async fn create_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateConfigParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO config (\n    owner_id,\n    business_id,\n    site_name,\n    logo_url,\n    favicon_url,\n    meta_tags,\n    contact_email,\n    contact_number,\n    sender_email,\n    email_provider, sender_name, reply_to_email, email_notifications_enabled,\n    smtp_host, smtp_port, smtp_username, smtp_password_hash,\n    website,\n    maps_url,\n    address,\n    city,\n    zipcode,\n    country,\n    instagram,\n    maintenance_mode,\n    maintenance_notice,\n    addons_enabled,\n    enable_reviews,\n    enable_wishlist,\n    enable_newsletter_signup,\n    about_content,\n    footer_sub_text\n) VALUES (\n    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?\n)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.site_name)
        .bind(&params.logo_url)
        .bind(&params.favicon_url)
        .bind(&params.meta_tags)
        .bind(&params.contact_email)
        .bind(&params.contact_number)
        .bind(&params.sender_email)
        .bind(&params.email_provider)
        .bind(&params.sender_name)
        .bind(&params.reply_to_email)
        .bind(params.email_notifications_enabled)
        .bind(&params.smtp_host)
        .bind(params.smtp_port)
        .bind(&params.smtp_username)
        .bind(&params.smtp_password_hash)
        .bind(&params.website)
        .bind(&params.maps_url)
        .bind(&params.address)
        .bind(&params.city)
        .bind(&params.zipcode)
        .bind(&params.country)
        .bind(&params.instagram)
        .bind(params.maintenance_mode)
        .bind(&params.maintenance_notice)
        .bind(params.addons_enabled)
        .bind(params.enable_reviews)
        .bind(params.enable_wishlist)
        .bind(params.enable_newsletter_signup)
        .bind(&params.about_content)
        .bind(&params.footer_sub_text)
        .execute(exec)
        .await
}

/// Params for `GetBusinessSlugByID` (Go `GetBusinessSlugByIDParams`).
pub struct GetBusinessSlugByIDParams {
    pub id: i64,
}

/// `GetBusinessSlugByID`: `one`.
pub async fn get_business_slug_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetBusinessSlugByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT slug FROM businesses WHERE id = ? ORDER BY id ASC LIMIT 1")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetConfig` (Go `GetConfigParams`).
pub struct GetConfigParams {
    pub business_id: i64,
}

/// `GetConfig`: `many`.
pub async fn get_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetConfigParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    owner_id,\n    business_id,\n    site_name,\n    logo_url,\n    favicon_url,\n    meta_tags,\n    contact_email,\n    contact_number,\n    sender_email,\n    email_provider, sender_name, reply_to_email, email_notifications_enabled,\n    smtp_host, smtp_port, smtp_username, smtp_password_hash,\n    website,\n    maps_url,\n    address,\n    city,\n    zipcode,\n    country,\n    instagram,\n    maintenance_mode,\n    maintenance_notice,\n    addons_enabled,\n    enable_reviews,\n    enable_wishlist,\n    enable_newsletter_signup,\n    about_content,\n    footer_sub_text\nFROM config WHERE business_id = ?")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `GetThemeByOwnerID` (Go `GetThemeByOwnerIDParams`).
pub struct GetThemeByOwnerIDParams {
    pub business_id: i64,
}

/// `GetThemeByOwnerID`: `one`.
pub async fn get_theme_by_owner_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetThemeByOwnerIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    id,\n    owner_id,\n    business_id,\n    primary_color,\n    secondary_color,\n    accent_color,\n    dark_green_color,\n    background_color,\n    bg_color,\n    surface_color,\n    light_color,\n    dark_color,\n    item_background_color,\n    error_color,\n    success_color,\n    warning_color,\n    info_color,\n    text_primary_color,\n    text_secondary_color,\n    subtext_color,\n    light_gray_color,\n    active_color,\n    admin_background_color,\n    admin_primary_color,\n    vendor_background_color,\n    border_color,\n    font_family_primary,\n    font_family_secondary,\n    font_size_base,\n    font_size_heading,\n    mobile_breakpoint\nFROM themes WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `UpdateConfig` (Go `UpdateConfigParams`).
pub struct UpdateConfigParams {
    pub site_name: Option<String>,
    pub logo_url: Option<String>,
    pub favicon_url: Option<String>,
    pub meta_tags: String,
    pub contact_email: String,
    pub contact_number: String,
    pub sender_email: Option<String>,
    pub email_provider: String,
    pub sender_name: Option<String>,
    pub reply_to_email: Option<String>,
    pub email_notifications_enabled: bool,
    pub smtp_host: Option<String>,
    pub smtp_port: Option<i32>,
    pub smtp_username: Option<String>,
    pub website: Option<String>,
    pub maps_url: Option<String>,
    pub address: Option<String>,
    pub city: Option<String>,
    pub zipcode: Option<String>,
    pub country: Option<String>,
    pub instagram: Option<String>,
    pub maintenance_mode: Option<bool>,
    pub maintenance_notice: Option<String>,
    pub addons_enabled: Option<bool>,
    pub enable_reviews: Option<bool>,
    pub enable_wishlist: Option<bool>,
    pub enable_newsletter_signup: Option<bool>,
    pub about_content: String,
    pub footer_sub_text: Option<String>,
    pub business_id: i64,
}

/// `UpdateConfig`: `exec`.
pub async fn update_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateConfigParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE config SET\n    site_name = ?,\n    logo_url = ?,\n    favicon_url = ?,\n    meta_tags = ?,\n    contact_email = ?,\n    contact_number = ?,\n    sender_email = ?,\n    email_provider = ?, sender_name = ?, reply_to_email = ?, email_notifications_enabled = ?,\n    smtp_host = ?, smtp_port = ?, smtp_username = ?,\n    website = ?,\n    maps_url = ?,\n    address = ?,\n    city = ?,\n    zipcode = ?,\n    country = ?,\n    instagram = ?,\n    maintenance_mode = ?,\n    maintenance_notice = ?,\n    addons_enabled = ?,\n    enable_reviews = ?,\n    enable_wishlist = ?,\n    enable_newsletter_signup = ?,\n    about_content = ?,\n    footer_sub_text = ?\nWHERE business_id = ?")
        .bind(&params.site_name)
        .bind(&params.logo_url)
        .bind(&params.favicon_url)
        .bind(&params.meta_tags)
        .bind(&params.contact_email)
        .bind(&params.contact_number)
        .bind(&params.sender_email)
        .bind(&params.email_provider)
        .bind(&params.sender_name)
        .bind(&params.reply_to_email)
        .bind(params.email_notifications_enabled)
        .bind(&params.smtp_host)
        .bind(params.smtp_port)
        .bind(&params.smtp_username)
        .bind(&params.website)
        .bind(&params.maps_url)
        .bind(&params.address)
        .bind(&params.city)
        .bind(&params.zipcode)
        .bind(&params.country)
        .bind(&params.instagram)
        .bind(params.maintenance_mode)
        .bind(&params.maintenance_notice)
        .bind(params.addons_enabled)
        .bind(params.enable_reviews)
        .bind(params.enable_wishlist)
        .bind(params.enable_newsletter_signup)
        .bind(&params.about_content)
        .bind(&params.footer_sub_text)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateSMTPPassword` (Go `UpdateSMTPPasswordParams`).
pub struct UpdateSMTPPasswordParams {
    pub smtp_password_hash: Option<String>,
    pub business_id: i64,
}

/// `UpdateSMTPPassword`: `exec`.
pub async fn update_smtp_password<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateSMTPPasswordParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE config SET smtp_password_hash = ? WHERE business_id = ?")
        .bind(&params.smtp_password_hash)
        .bind(params.business_id)
        .execute(exec)
        .await
}
