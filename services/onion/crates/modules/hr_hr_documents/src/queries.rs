//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountDocuments` (Go `CountDocumentsParams`).
pub struct CountDocumentsParams {
    pub business_id: i64,
    pub employee_id: i64,
}

/// `CountDocuments`: `one`.
pub async fn count_documents<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountDocumentsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total\nFROM hr_employee_documents\nWHERE business_id = ?\n  AND (? = 0 OR employee_id = ?)")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.employee_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountExpiringDocuments` (Go `CountExpiringDocumentsParams`).
pub struct CountExpiringDocumentsParams {
    pub business_id: i64,
    pub expiry_date: Option<chrono::DateTime<chrono::Utc>>,
}

/// `CountExpiringDocuments`: `one`.
pub async fn count_expiring_documents<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountExpiringDocumentsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total\nFROM hr_employee_documents\nWHERE business_id = ? AND expiry_date IS NOT NULL AND expiry_date <= ?")
        .bind(params.business_id)
        .bind(params.expiry_date)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateDocument` (Go `CreateDocumentParams`).
pub struct CreateDocumentParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub employee_id: i64,
    pub document_type: String,
    pub title: String,
    pub storage_mode: String,
    pub file_url: Option<String>,
    pub external_url: Option<String>,
    pub expiry_date: Option<chrono::DateTime<chrono::Utc>>,
    pub uploaded_by: i64,
}

/// `CreateDocument`: `execresult`.
pub async fn create_document<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateDocumentParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_employee_documents (owner_id, business_id, employee_id, document_type, title, storage_mode, file_url, external_url, expiry_date, uploaded_by)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(&params.document_type)
        .bind(&params.title)
        .bind(&params.storage_mode)
        .bind(&params.file_url)
        .bind(&params.external_url)
        .bind(params.expiry_date)
        .bind(params.uploaded_by)
        .execute(exec)
        .await
}

/// Params for `DeleteDocument` (Go `DeleteDocumentParams`).
pub struct DeleteDocumentParams {
    pub business_id: i64,
    pub id: i64,
}

/// `DeleteDocument`: `execresult`.
pub async fn delete_document<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteDocumentParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM hr_employee_documents\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `GetDocument` (Go `GetDocumentParams`).
pub struct GetDocumentParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetDocument`: `one`.
pub async fn get_document<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetDocumentParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, document_type, title, storage_mode, file_url, external_url, expiry_date, uploaded_by, created_at, updated_at\nFROM hr_employee_documents\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListDocuments` (Go `ListDocumentsParams`).
pub struct ListDocumentsParams {
    pub business_id: i64,
    pub employee_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListDocuments`: `many`.
pub async fn list_documents<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListDocumentsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, document_type, title, storage_mode, file_url, external_url, expiry_date, uploaded_by, created_at, updated_at\nFROM hr_employee_documents\nWHERE business_id = ?\n  AND (? = 0 OR employee_id = ?)\nORDER BY created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.employee_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListExpiringDocuments` (Go `ListExpiringDocumentsParams`).
pub struct ListExpiringDocumentsParams {
    pub business_id: i64,
    pub expiry_date: Option<chrono::DateTime<chrono::Utc>>,
    pub limit: i32,
    pub offset: i32,
}

/// `ListExpiringDocuments`: `many`.
pub async fn list_expiring_documents<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListExpiringDocumentsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, document_type, title, storage_mode, file_url, external_url, expiry_date, uploaded_by, created_at, updated_at\nFROM hr_employee_documents\nWHERE business_id = ? AND expiry_date IS NOT NULL AND expiry_date <= ?\nORDER BY expiry_date ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.expiry_date)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}
