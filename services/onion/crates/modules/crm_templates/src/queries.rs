//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateTemplate` (Go `CreateTemplateParams`).
pub struct CreateTemplateParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub r#type: String,
    pub subject: Option<String>,
    pub body: String,
    pub variables: String,
    pub is_active: Option<bool>,
}

/// `CreateTemplate`: `execresult`.
pub async fn create_template<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateTemplateParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_templates\n  (owner_id, business_id, name, type, subject, body, variables, is_active)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.r#type)
        .bind(&params.subject)
        .bind(&params.body)
        .bind(&params.variables)
        .bind(params.is_active)
        .execute(exec)
        .await
}

/// Params for `DeleteTemplate` (Go `DeleteTemplateParams`).
pub struct DeleteTemplateParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteTemplate`: `execresult`.
pub async fn delete_template<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteTemplateParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM crm_templates WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListTemplates` (Go `ListTemplatesParams`).
pub struct ListTemplatesParams {
    pub business_id: i64,
}

/// `ListTemplates`: `many`.
pub async fn list_templates<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListTemplatesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, type, subject, body, variables, is_active, created_at, updated_at\nFROM crm_templates WHERE business_id = ? ORDER BY created_at DESC, id DESC")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateTemplate` (Go `UpdateTemplateParams`).
pub struct UpdateTemplateParams {
    pub name: String,
    pub r#type: String,
    pub subject: Option<String>,
    pub body: String,
    pub variables: String,
    pub is_active: Option<bool>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateTemplate`: `execresult`.
pub async fn update_template<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateTemplateParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_templates\nSET name = ?, type = ?, subject = ?, body = ?, variables = ?, is_active = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(&params.r#type)
        .bind(&params.subject)
        .bind(&params.body)
        .bind(&params.variables)
        .bind(params.is_active)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
