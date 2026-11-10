//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountDocuments` (Go `CountDocumentsParams`).
pub struct CountDocumentsParams {
    pub business_id: i64,
}

/// `CountDocuments`: `one`.
pub async fn count_documents<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountDocumentsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM hr_documents WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateDocument` (Go `CreateDocumentParams`).
pub struct CreateDocumentParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub r#type: Option<String>,
    pub file_url: Option<String>,
    pub expiry_date: Option<chrono::DateTime<chrono::Utc>>,
}

/// `CreateDocument`: `execresult`.
pub async fn create_document<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateDocumentParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_documents (owner_id, business_id, name, type, file_url, expiry_date)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.r#type)
        .bind(&params.file_url)
        .bind(params.expiry_date)
        .execute(exec)
        .await
}

/// Params for `CreateDocumentForStaff` (Go `CreateDocumentForStaffParams`).
pub struct CreateDocumentForStaffParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub r#type: Option<String>,
    pub file_url: Option<String>,
    pub expiry_date: Option<chrono::DateTime<chrono::Utc>>,
    pub staff_id: i64,
    pub staff_business_id: i64,
}

/// `CreateDocumentForStaff`: `execresult`.
pub async fn create_document_for_staff<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateDocumentForStaffParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_documents (owner_id, business_id, staff_id, name, type, file_url, expiry_date)\nSELECT ?, ?, staff.id,\n       ?, ?, ?, ?\nFROM hr_staff AS staff\nWHERE staff.id = ?\n  AND staff.business_id = ?\nLIMIT 1")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.r#type)
        .bind(&params.file_url)
        .bind(params.expiry_date)
        .bind(params.staff_id)
        .bind(params.staff_business_id)
        .execute(exec)
        .await
}

/// Params for `DeleteDocument` (Go `DeleteDocumentParams`).
pub struct DeleteDocumentParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteDocument`: `execresult`.
pub async fn delete_document<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteDocumentParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM hr_documents WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListDocuments` (Go `ListDocumentsParams`).
pub struct ListDocumentsParams {
    pub business_id: i64,
}

/// `ListDocuments`: `many`.
pub async fn list_documents<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListDocumentsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, staff_id, name, type, file_url, expiry_date, created_at\nFROM hr_documents\nWHERE business_id = ?\nORDER BY created_at DESC LIMIT 100")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListDocumentsPaginated` (Go `ListDocumentsPaginatedParams`).
pub struct ListDocumentsPaginatedParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListDocumentsPaginated`: `many`.
pub async fn list_documents_paginated<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListDocumentsPaginatedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, staff_id, name, type, file_url, expiry_date, created_at\nFROM hr_documents\nWHERE business_id = ?\nORDER BY created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}
