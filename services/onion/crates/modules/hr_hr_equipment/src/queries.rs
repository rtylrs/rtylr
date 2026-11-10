//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `AssignAsset` (Go `AssignAssetParams`).
pub struct AssignAssetParams {
    pub assigned_to: Option<i64>,
    pub business_id: i64,
    pub id: i64,
}

/// `AssignAsset`: `exec`.
pub async fn assign_asset<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AssignAssetParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_assets\nSET assigned_to = ?, status = 'assigned'\nWHERE business_id = ? AND id = ?")
        .bind(params.assigned_to)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `CountAssetsFiltered` (Go `CountAssetsFilteredParams`).
pub struct CountAssetsFilteredParams {
    pub business_id: i64,
    pub location_id: Option<i64>,
    pub status: Option<String>,
    pub condition: Option<String>,
    pub assigned_to: Option<i64>,
}

/// `CountAssetsFiltered`: `one`.
pub async fn count_assets_filtered<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountAssetsFilteredParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total\nFROM hr_assets\nWHERE business_id = ?\n  AND (? IS NULL OR location_id = ?)\n  AND (? IS NULL OR status = ?)\n  AND (? IS NULL OR `condition` = ?)\n  AND (? IS NULL OR assigned_to = ?)")
        .bind(params.business_id)
        .bind(params.location_id)
        .bind(params.location_id)
        .bind(&params.status)
        .bind(&params.status)
        .bind(&params.condition)
        .bind(&params.condition)
        .bind(params.assigned_to)
        .bind(params.assigned_to)
        .fetch_optional(exec)
        .await
}

/// Params for `CountEquipmentRequests` (Go `CountEquipmentRequestsParams`).
pub struct CountEquipmentRequestsParams {
    pub business_id: i64,
}

/// `CountEquipmentRequests`: `one`.
pub async fn count_equipment_requests<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountEquipmentRequestsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total FROM hr_equipment_requests WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountEquipmentRequestsByEmployee` (Go `CountEquipmentRequestsByEmployeeParams`).
pub struct CountEquipmentRequestsByEmployeeParams {
    pub business_id: i64,
    pub employee_id: i64,
}

/// `CountEquipmentRequestsByEmployee`: `one`.
pub async fn count_equipment_requests_by_employee<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountEquipmentRequestsByEmployeeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total FROM hr_equipment_requests WHERE business_id = ? AND employee_id = ?")
        .bind(params.business_id)
        .bind(params.employee_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountEquipmentRequestsByEmployeeAndStatus` (Go `CountEquipmentRequestsByEmployeeAndStatusParams`).
pub struct CountEquipmentRequestsByEmployeeAndStatusParams {
    pub business_id: i64,
    pub employee_id: i64,
    pub status: String,
}

/// `CountEquipmentRequestsByEmployeeAndStatus`: `one`.
pub async fn count_equipment_requests_by_employee_and_status<
    'e,
    E: sqlx::Executor<'e, Database = MySql>,
>(
    exec: E,
    params: &CountEquipmentRequestsByEmployeeAndStatusParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total FROM hr_equipment_requests WHERE business_id = ? AND employee_id = ? AND status = ?")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(&params.status)
        .fetch_optional(exec)
        .await
}

/// Params for `CountEquipmentRequestsByStatus` (Go `CountEquipmentRequestsByStatusParams`).
pub struct CountEquipmentRequestsByStatusParams {
    pub business_id: i64,
    pub status: String,
}

/// `CountEquipmentRequestsByStatus`: `one`.
pub async fn count_equipment_requests_by_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountEquipmentRequestsByStatusParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT COUNT(*) AS total FROM hr_equipment_requests WHERE business_id = ? AND status = ?",
    )
    .bind(params.business_id)
    .bind(&params.status)
    .fetch_optional(exec)
    .await
}

/// Params for `CreateAsset` (Go `CreateAssetParams`).
pub struct CreateAssetParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub category: String,
    pub serial_number: Option<String>,
    pub location_id: Option<i64>,
    pub status: String,
    pub condition: String,
    pub purchase_date: Option<chrono::DateTime<chrono::Utc>>,
    pub purchase_price: Option<String>,
    pub notes: Option<String>,
}

/// `CreateAsset`: `execresult`.
pub async fn create_asset<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateAssetParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_assets (owner_id, business_id, name, category, serial_number, location_id, status, `condition`, purchase_date, purchase_price, notes)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.category)
        .bind(&params.serial_number)
        .bind(params.location_id)
        .bind(&params.status)
        .bind(&params.condition)
        .bind(params.purchase_date)
        .bind(&params.purchase_price)
        .bind(&params.notes)
        .execute(exec)
        .await
}

/// Params for `CreateEquipmentRequest` (Go `CreateEquipmentRequestParams`).
pub struct CreateEquipmentRequestParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub employee_id: i64,
    pub item_name: String,
    pub quantity: i32,
    pub reason: Option<String>,
    pub needed_by_date: Option<chrono::DateTime<chrono::Utc>>,
}

/// `CreateEquipmentRequest`: `execresult`.
pub async fn create_equipment_request<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateEquipmentRequestParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_equipment_requests (owner_id, business_id, employee_id, item_name, quantity, reason, needed_by_date)\nVALUES (?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(&params.item_name)
        .bind(params.quantity)
        .bind(&params.reason)
        .bind(params.needed_by_date)
        .execute(exec)
        .await
}

/// Params for `GetAsset` (Go `GetAssetParams`).
pub struct GetAssetParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetAsset`: `one`.
pub async fn get_asset<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetAssetParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, category, serial_number, location_id, assigned_to, status, `condition`,\n       purchase_date, purchase_price, notes, created_at, updated_at\nFROM hr_assets\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetEquipmentRequest` (Go `GetEquipmentRequestParams`).
pub struct GetEquipmentRequestParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetEquipmentRequest`: `one`.
pub async fn get_equipment_request<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetEquipmentRequestParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, item_name, quantity, reason, status,\n       current_approval_level, approved_by, rejection_reason, needed_by_date,\n       created_at, updated_at\nFROM hr_equipment_requests\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListAssets` (Go `ListAssetsParams`).
pub struct ListAssetsParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListAssets`: `many`.
pub async fn list_assets<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListAssetsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, category, serial_number, location_id, assigned_to, status, `condition`,\n       purchase_date, purchase_price, notes, created_at, updated_at\nFROM hr_assets\nWHERE business_id = ?\nORDER BY name ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListAssetsFiltered` (Go `ListAssetsFilteredParams`).
pub struct ListAssetsFilteredParams {
    pub business_id: i64,
    pub location_id: Option<i64>,
    pub status: Option<String>,
    pub condition: Option<String>,
    pub assigned_to: Option<i64>,
    pub limit: i32,
    pub offset: i32,
}

/// `ListAssetsFiltered`: `many`.
pub async fn list_assets_filtered<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListAssetsFilteredParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, category, serial_number, location_id, assigned_to, status, `condition`,\n       purchase_date, purchase_price, notes, created_at, updated_at\nFROM hr_assets\nWHERE business_id = ?\n  AND (? IS NULL OR location_id = ?)\n  AND (? IS NULL OR status = ?)\n  AND (? IS NULL OR `condition` = ?)\n  AND (? IS NULL OR assigned_to = ?)\nORDER BY name ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.location_id)
        .bind(params.location_id)
        .bind(&params.status)
        .bind(&params.status)
        .bind(&params.condition)
        .bind(&params.condition)
        .bind(params.assigned_to)
        .bind(params.assigned_to)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListEquipmentRequests` (Go `ListEquipmentRequestsParams`).
pub struct ListEquipmentRequestsParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListEquipmentRequests`: `many`.
pub async fn list_equipment_requests<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListEquipmentRequestsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, item_name, quantity, reason, status,\n       current_approval_level, approved_by, rejection_reason, needed_by_date,\n       created_at, updated_at\nFROM hr_equipment_requests\nWHERE business_id = ?\nORDER BY created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListEquipmentRequestsByEmployee` (Go `ListEquipmentRequestsByEmployeeParams`).
pub struct ListEquipmentRequestsByEmployeeParams {
    pub business_id: i64,
    pub employee_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListEquipmentRequestsByEmployee`: `many`.
pub async fn list_equipment_requests_by_employee<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListEquipmentRequestsByEmployeeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, item_name, quantity, reason, status,\n       current_approval_level, approved_by, rejection_reason, needed_by_date,\n       created_at, updated_at\nFROM hr_equipment_requests\nWHERE business_id = ? AND employee_id = ?\nORDER BY created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListEquipmentRequestsByEmployeeAndStatus` (Go `ListEquipmentRequestsByEmployeeAndStatusParams`).
pub struct ListEquipmentRequestsByEmployeeAndStatusParams {
    pub business_id: i64,
    pub employee_id: i64,
    pub status: String,
    pub limit: i32,
    pub offset: i32,
}

/// `ListEquipmentRequestsByEmployeeAndStatus`: `many`.
pub async fn list_equipment_requests_by_employee_and_status<
    'e,
    E: sqlx::Executor<'e, Database = MySql>,
>(
    exec: E,
    params: &ListEquipmentRequestsByEmployeeAndStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, item_name, quantity, reason, status,\n       current_approval_level, approved_by, rejection_reason, needed_by_date,\n       created_at, updated_at\nFROM hr_equipment_requests\nWHERE business_id = ? AND employee_id = ? AND status = ?\nORDER BY created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(&params.status)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListEquipmentRequestsByStatus` (Go `ListEquipmentRequestsByStatusParams`).
pub struct ListEquipmentRequestsByStatusParams {
    pub business_id: i64,
    pub status: String,
    pub limit: i32,
    pub offset: i32,
}

/// `ListEquipmentRequestsByStatus`: `many`.
pub async fn list_equipment_requests_by_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListEquipmentRequestsByStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, item_name, quantity, reason, status,\n       current_approval_level, approved_by, rejection_reason, needed_by_date,\n       created_at, updated_at\nFROM hr_equipment_requests\nWHERE business_id = ? AND status = ?\nORDER BY created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(&params.status)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `UnassignAsset` (Go `UnassignAssetParams`).
pub struct UnassignAssetParams {
    pub business_id: i64,
    pub id: i64,
}

/// `UnassignAsset`: `exec`.
pub async fn unassign_asset<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UnassignAssetParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_assets\nSET assigned_to = NULL, status = 'available'\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpdateEquipmentRequestStatus` (Go `UpdateEquipmentRequestStatusParams`).
pub struct UpdateEquipmentRequestStatusParams {
    pub status: String,
    pub current_approval_level: i32,
    pub approved_by: Option<i64>,
    pub rejection_reason: Option<String>,
    pub business_id: i64,
    pub id: i64,
    pub current_approval_level_2: i32,
}

/// `UpdateEquipmentRequestStatus`: `execrows`.
pub async fn update_equipment_request_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateEquipmentRequestStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_equipment_requests\nSET status = ?, current_approval_level = ?, approved_by = ?, rejection_reason = ?\nWHERE business_id = ? AND id = ? AND status = 'pending' AND current_approval_level = ?")
        .bind(&params.status)
        .bind(params.current_approval_level)
        .bind(params.approved_by)
        .bind(&params.rejection_reason)
        .bind(params.business_id)
        .bind(params.id)
        .bind(params.current_approval_level_2)
        .execute(exec)
        .await
}
