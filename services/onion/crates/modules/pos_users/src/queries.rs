//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreatePOSUser` (Go `CreatePOSUserParams`).
pub struct CreatePOSUserParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub user_id: i64,
    pub location_id: i64,
    pub name: String,
    pub pin_hash: Option<String>,
    pub role: String,
    pub hourly_rate: Option<String>,
    pub max_discount_pct: Option<String>,
    pub can_void: Option<bool>,
    pub can_refund: Option<bool>,
    pub can_open_drawer: Option<bool>,
    pub can_apply_discount: Option<bool>,
    pub created_by: Option<i64>,
}

/// `CreatePOSUser`: `execresult`.
pub async fn create_pos_user<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreatePOSUserParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_users (owner_id, business_id, user_id, location_id, name, pin_hash, role,\n    hourly_rate, max_discount_pct, can_void, can_refund,\n    can_open_drawer, can_apply_discount, created_by)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.user_id)
        .bind(params.location_id)
        .bind(&params.name)
        .bind(&params.pin_hash)
        .bind(&params.role)
        .bind(&params.hourly_rate)
        .bind(&params.max_discount_pct)
        .bind(params.can_void)
        .bind(params.can_refund)
        .bind(params.can_open_drawer)
        .bind(params.can_apply_discount)
        .bind(params.created_by)
        .execute(exec)
        .await
}

/// Params for `DeactivatePOSUser` (Go `DeactivatePOSUserParams`).
pub struct DeactivatePOSUserParams {
    pub business_id: i64,
    pub id: i64,
}

/// `DeactivatePOSUser`: `exec`.
pub async fn deactivate_pos_user<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeactivatePOSUserParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_users SET is_active = FALSE WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `FindPOSUserByUserAndLocation` (Go `FindPOSUserByUserAndLocationParams`).
pub struct FindPOSUserByUserAndLocationParams {
    pub business_id: i64,
    pub user_id: i64,
    pub location_id: i64,
}

/// `FindPOSUserByUserAndLocation`: `one`.
pub async fn find_pos_user_by_user_and_location<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &FindPOSUserByUserAndLocationParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, user_id, location_id, name, role, is_active,\n       hourly_rate, max_discount_pct, can_void, can_refund,\n       can_open_drawer, can_apply_discount, created_by, created_at, updated_at\nFROM pos_users WHERE business_id = ? AND user_id = ? AND location_id = ? LIMIT 1")
        .bind(params.business_id)
        .bind(params.user_id)
        .bind(params.location_id)
        .fetch_optional(exec)
        .await
}

/// Params for `FindPOSUserForPINVerify` (Go `FindPOSUserForPINVerifyParams`).
pub struct FindPOSUserForPINVerifyParams {
    pub business_id: i64,
    pub location_id: i64,
    pub user_id: i64,
}

/// `FindPOSUserForPINVerify`: `one`.
pub async fn find_pos_user_for_pin_verify<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &FindPOSUserForPINVerifyParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, user_id, location_id, name, role, is_active,\n       hourly_rate, max_discount_pct, can_void, can_refund,\n       can_open_drawer, can_apply_discount, created_by, created_at, updated_at\nFROM pos_users WHERE business_id = ? AND location_id = ? AND user_id = ? AND is_active = TRUE LIMIT 1")
        .bind(params.business_id)
        .bind(params.location_id)
        .bind(params.user_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetMemberLocationAccess` (Go `GetMemberLocationAccessParams`).
pub struct GetMemberLocationAccessParams {
    pub business_id: i64,
    pub user_id: i64,
    pub location_id: i64,
}

/// `GetMemberLocationAccess`: `one`.
pub async fn get_member_location_access<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetMemberLocationAccessParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, user_id, location_id, can_manage_pos_users,\n       can_manage_terminals, can_view_reports\nFROM pos_member_locations WHERE business_id = ? AND user_id = ? AND location_id = ? LIMIT 1")
        .bind(params.business_id)
        .bind(params.user_id)
        .bind(params.location_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetPINHash` (Go `GetPINHashParams`).
pub struct GetPINHashParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetPINHash`: `one`.
pub async fn get_pin_hash<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPINHashParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT pin_hash\nFROM pos_users\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListPOSUsersByLocation` (Go `ListPOSUsersByLocationParams`).
pub struct ListPOSUsersByLocationParams {
    pub business_id: i64,
    pub location_id: i64,
}

/// `ListPOSUsersByLocation`: `many`.
pub async fn list_pos_users_by_location<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListPOSUsersByLocationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, user_id, location_id, name, role, is_active,\n       hourly_rate, max_discount_pct, can_void, can_refund,\n       can_open_drawer, can_apply_discount, created_by, created_at, updated_at\nFROM pos_users WHERE business_id = ? AND location_id = ? AND is_active = TRUE ORDER BY name ASC")
        .bind(params.business_id)
        .bind(params.location_id)
        .execute(exec)
        .await
}

/// Params for `ListPOSUsersByOwner` (Go `ListPOSUsersByOwnerParams`).
pub struct ListPOSUsersByOwnerParams {
    pub business_id: i64,
}

/// `ListPOSUsersByOwner`: `many`.
pub async fn list_pos_users_by_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListPOSUsersByOwnerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, user_id, location_id, name, role, is_active,\n       hourly_rate, max_discount_pct, can_void, can_refund,\n       can_open_drawer, can_apply_discount, created_by, created_at, updated_at\nFROM pos_users WHERE business_id = ? ORDER BY name ASC LIMIT 100")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdatePOSUser` (Go `UpdatePOSUserParams`).
pub struct UpdatePOSUserParams {
    pub name: String,
    pub is_active: Option<bool>,
    pub hourly_rate: Option<String>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdatePOSUser`: `exec`.
pub async fn update_pos_user<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdatePOSUserParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_users SET name = ?, is_active = ?, hourly_rate = ?\nWHERE business_id = ? AND id = ?")
        .bind(&params.name)
        .bind(params.is_active)
        .bind(&params.hourly_rate)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpdatePOSUserPINHash` (Go `UpdatePOSUserPINHashParams`).
pub struct UpdatePOSUserPINHashParams {
    pub pin_hash: Option<String>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdatePOSUserPINHash`: `exec`.
pub async fn update_pos_user_pin_hash<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdatePOSUserPINHashParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_users SET pin_hash = ? WHERE business_id = ? AND id = ?")
        .bind(&params.pin_hash)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
