//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountActiveWarehouses` (Go `CountActiveWarehousesParams`).
pub struct CountActiveWarehousesParams {
    pub business_id: i64,
}

/// `CountActiveWarehouses`: `one`.
pub async fn count_active_warehouses<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountActiveWarehousesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM erp_warehouses WHERE business_id = ? AND is_active = 1")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountWarehouses` (Go `CountWarehousesParams`).
pub struct CountWarehousesParams {
    pub business_id: i64,
}

/// `CountWarehouses`: `one`.
pub async fn count_warehouses<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountWarehousesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM erp_warehouses WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateWarehouse` (Go `CreateWarehouseParams`).
pub struct CreateWarehouseParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub address: Option<String>,
    pub city: Option<String>,
    pub country: Option<String>,
    pub manager_name: Option<String>,
    pub capacity: Option<i64>,
    pub is_default: Option<bool>,
    pub is_active: Option<bool>,
}

/// `CreateWarehouse`: `execresult`.
pub async fn create_warehouse<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateWarehouseParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO erp_warehouses (owner_id, business_id, name, address, city, country, manager_name, capacity, is_default, is_active)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.address)
        .bind(&params.city)
        .bind(&params.country)
        .bind(&params.manager_name)
        .bind(params.capacity)
        .bind(params.is_default)
        .bind(params.is_active)
        .execute(exec)
        .await
}

/// Params for `DeleteWarehouse` (Go `DeleteWarehouseParams`).
pub struct DeleteWarehouseParams {
    pub business_id: i64,
    pub id: i64,
}

/// `DeleteWarehouse`: `exec`.
pub async fn delete_warehouse<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteWarehouseParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM erp_warehouses WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `GetWarehouseByID` (Go `GetWarehouseByIDParams`).
pub struct GetWarehouseByIDParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetWarehouseByID`: `one`.
pub async fn get_warehouse_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetWarehouseByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, address, city, country, manager_name,\n       capacity, is_default, is_active, created_at, updated_at\nFROM erp_warehouses\nWHERE business_id = ? AND id = ?\nLIMIT 1")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetWarehouseForUpdate` (Go `GetWarehouseForUpdateParams`).
pub struct GetWarehouseForUpdateParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetWarehouseForUpdate`: `one`.
pub async fn get_warehouse_for_update<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetWarehouseForUpdateParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, address, city, country, manager_name,\n       capacity, is_default, is_active, created_at, updated_at\nFROM erp_warehouses\nWHERE business_id = ? AND id = ?\nLIMIT 1 FOR UPDATE")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListWarehouses` (Go `ListWarehousesParams`).
pub struct ListWarehousesParams {
    pub business_id: i64,
}

/// `ListWarehouses`: `many`.
pub async fn list_warehouses<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListWarehousesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, address, city, country, manager_name,\n       capacity, is_default, is_active, created_at, updated_at\nFROM erp_warehouses\nWHERE business_id = ?\nORDER BY name ASC")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateWarehouse` (Go `UpdateWarehouseParams`).
pub struct UpdateWarehouseParams {
    pub name: String,
    pub address: Option<String>,
    pub city: Option<String>,
    pub country: Option<String>,
    pub manager_name: Option<String>,
    pub capacity: Option<i64>,
    pub is_default: Option<bool>,
    pub is_active: Option<bool>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateWarehouse`: `exec`.
pub async fn update_warehouse<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateWarehouseParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE erp_warehouses\nSET name = ?, address = ?, city = ?, country = ?, manager_name = ?,\n    capacity = ?, is_default = ?, is_active = ?\nWHERE business_id = ? AND id = ?")
        .bind(&params.name)
        .bind(&params.address)
        .bind(&params.city)
        .bind(&params.country)
        .bind(&params.manager_name)
        .bind(params.capacity)
        .bind(params.is_default)
        .bind(params.is_active)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
