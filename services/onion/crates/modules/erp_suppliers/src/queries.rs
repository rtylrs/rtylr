//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountActiveSuppliers` (Go `CountActiveSuppliersParams`).
pub struct CountActiveSuppliersParams {
    pub business_id: i64,
}

/// `CountActiveSuppliers`: `one`.
pub async fn count_active_suppliers<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountActiveSuppliersParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM erp_suppliers WHERE business_id = ? AND is_active = 1")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountSuppliers` (Go `CountSuppliersParams`).
pub struct CountSuppliersParams {
    pub business_id: i64,
}

/// `CountSuppliers`: `one`.
pub async fn count_suppliers<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountSuppliersParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM erp_suppliers WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateSupplier` (Go `CreateSupplierParams`).
pub struct CreateSupplierParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub email: Option<String>,
    pub phone: Option<String>,
    pub address: Option<String>,
    pub city: Option<String>,
    pub country: Option<String>,
    pub tax_id: Option<String>,
    pub payment_terms: Option<String>,
    pub notes: Option<String>,
    pub is_active: Option<bool>,
}

/// `CreateSupplier`: `execresult`.
pub async fn create_supplier<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateSupplierParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO erp_suppliers\n(owner_id, business_id, name, email, phone, address, city, country, tax_id, payment_terms, notes, is_active)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.email)
        .bind(&params.phone)
        .bind(&params.address)
        .bind(&params.city)
        .bind(&params.country)
        .bind(&params.tax_id)
        .bind(&params.payment_terms)
        .bind(&params.notes)
        .bind(params.is_active)
        .execute(exec)
        .await
}

/// Params for `DeleteSupplier` (Go `DeleteSupplierParams`).
pub struct DeleteSupplierParams {
    pub business_id: i64,
    pub id: i64,
}

/// `DeleteSupplier`: `exec`.
pub async fn delete_supplier<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteSupplierParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM erp_suppliers WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `GetSupplierByID` (Go `GetSupplierByIDParams`).
pub struct GetSupplierByIDParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetSupplierByID`: `one`.
pub async fn get_supplier_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetSupplierByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, email, phone, address, city, country,\n       tax_id, payment_terms, notes, is_active, created_at, updated_at\nFROM erp_suppliers\nWHERE business_id = ? AND id = ?\nORDER BY id ASC\nLIMIT 1")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListSuppliers` (Go `ListSuppliersParams`).
pub struct ListSuppliersParams {
    pub business_id: i64,
}

/// `ListSuppliers`: `many`.
pub async fn list_suppliers<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListSuppliersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, email, phone, address, city, country,\n       tax_id, payment_terms, notes, is_active, created_at, updated_at\nFROM erp_suppliers\nWHERE business_id = ?\nORDER BY name ASC\nLIMIT 200")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListSuppliersPaginated` (Go `ListSuppliersPaginatedParams`).
pub struct ListSuppliersPaginatedParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListSuppliersPaginated`: `many`.
pub async fn list_suppliers_paginated<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListSuppliersPaginatedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, email, phone, address, city, country,\n       tax_id, payment_terms, notes, is_active, created_at, updated_at\nFROM erp_suppliers\nWHERE business_id = ?\nORDER BY name ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `UpdateSupplier` (Go `UpdateSupplierParams`).
pub struct UpdateSupplierParams {
    pub name: String,
    pub email: Option<String>,
    pub phone: Option<String>,
    pub address: Option<String>,
    pub city: Option<String>,
    pub country: Option<String>,
    pub tax_id: Option<String>,
    pub payment_terms: Option<String>,
    pub notes: Option<String>,
    pub is_active: Option<bool>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateSupplier`: `exec`.
pub async fn update_supplier<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateSupplierParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE erp_suppliers\nSET name = ?, email = ?, phone = ?, address = ?, city = ?, country = ?,\n    tax_id = ?, payment_terms = ?, notes = ?, is_active = ?\nWHERE business_id = ? AND id = ?")
        .bind(&params.name)
        .bind(&params.email)
        .bind(&params.phone)
        .bind(&params.address)
        .bind(&params.city)
        .bind(&params.country)
        .bind(&params.tax_id)
        .bind(&params.payment_terms)
        .bind(&params.notes)
        .bind(params.is_active)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
