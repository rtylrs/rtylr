//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateResource` (Go `CreateResourceParams`).
pub struct CreateResourceParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub r#type: String,
    pub capacity: Option<i32>,
    pub is_active: Option<bool>,
}

/// `CreateResource`: `execresult`.
pub async fn create_resource<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateResourceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO booking_resources (owner_id, business_id, name, type, capacity, is_active) VALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.r#type)
        .bind(params.capacity)
        .bind(params.is_active)
        .execute(exec)
        .await
}

/// Params for `DeleteResource` (Go `DeleteResourceParams`).
pub struct DeleteResourceParams {
    pub business_id: i64,
    pub id: i64,
}

/// `DeleteResource`: `exec`.
pub async fn delete_resource<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteResourceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM booking_resources WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `ListResources` (Go `ListResourcesParams`).
pub struct ListResourcesParams {
    pub business_id: i64,
}

/// `ListResources`: `many`.
pub async fn list_resources<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListResourcesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, type, capacity, is_active, created_at\nFROM booking_resources WHERE business_id = ? ORDER BY name ASC")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateResource` (Go `UpdateResourceParams`).
pub struct UpdateResourceParams {
    pub name: String,
    pub r#type: String,
    pub capacity: Option<i32>,
    pub is_active: Option<bool>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateResource`: `exec`.
pub async fn update_resource<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateResourceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE booking_resources SET name = ?, type = ?, capacity = ?, is_active = ? WHERE business_id = ? AND id = ?")
        .bind(&params.name)
        .bind(&params.r#type)
        .bind(params.capacity)
        .bind(params.is_active)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
