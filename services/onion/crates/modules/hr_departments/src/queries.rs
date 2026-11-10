//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateDepartment` (Go `CreateDepartmentParams`).
pub struct CreateDepartmentParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub description: Option<String>,
}

/// `CreateDepartment`: `execresult`.
pub async fn create_department<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateDepartmentParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_departments (owner_id, business_id, name, description)\nVALUES (?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.description)
        .execute(exec)
        .await
}

/// Params for `DeleteDepartment` (Go `DeleteDepartmentParams`).
pub struct DeleteDepartmentParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteDepartment`: `exec`.
pub async fn delete_department<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteDepartmentParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM hr_departments WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListDepartments` (Go `ListDepartmentsParams`).
pub struct ListDepartmentsParams {
    pub business_id: i64,
}

/// `ListDepartments`: `many`.
pub async fn list_departments<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListDepartmentsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, description, manager_staff_id, created_at, updated_at\nFROM hr_departments\nWHERE business_id = ?\nORDER BY name ASC\nLIMIT 100")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateDepartment` (Go `UpdateDepartmentParams`).
pub struct UpdateDepartmentParams {
    pub name: String,
    pub description: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateDepartment`: `exec`.
pub async fn update_department<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateDepartmentParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_departments\nSET name = ?, description = ?, updated_at = NOW()\nWHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(&params.description)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
