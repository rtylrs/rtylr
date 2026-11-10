//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateTheme` (Go `CreateThemeParams`).
pub struct CreateThemeParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub primary_color: Option<String>,
    pub font: Option<String>,
    pub header_style: Option<String>,
    pub footer_style: Option<String>,
    pub custom_css: Option<String>,
}

/// `CreateTheme`: `execresult`.
pub async fn create_theme<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateThemeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO store_themes (owner_id, business_id, name, primary_color, font, header_style, footer_style, custom_css) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.primary_color)
        .bind(&params.font)
        .bind(&params.header_style)
        .bind(&params.footer_style)
        .bind(&params.custom_css)
        .execute(exec)
        .await
}

/// Params for `ListThemes` (Go `ListThemesParams`).
pub struct ListThemesParams {
    pub business_id: i64,
}

/// `ListThemes`: `many`.
pub async fn list_themes<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListThemesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, primary_color, font, header_style, footer_style, custom_css, created_at, updated_at\nFROM store_themes WHERE business_id = ? ORDER BY name ASC\nLIMIT 200")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateTheme` (Go `UpdateThemeParams`).
pub struct UpdateThemeParams {
    pub name: String,
    pub primary_color: Option<String>,
    pub font: Option<String>,
    pub header_style: Option<String>,
    pub footer_style: Option<String>,
    pub custom_css: Option<String>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateTheme`: `exec`.
pub async fn update_theme<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateThemeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE store_themes SET name = ?, primary_color = ?, font = ?, header_style = ?, footer_style = ?, custom_css = ? WHERE business_id = ? AND id = ?")
        .bind(&params.name)
        .bind(&params.primary_color)
        .bind(&params.font)
        .bind(&params.header_style)
        .bind(&params.footer_style)
        .bind(&params.custom_css)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
