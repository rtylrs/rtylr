//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `AssignMemberToLocation` (Go `AssignMemberToLocationParams`).
pub struct AssignMemberToLocationParams {
    pub business_id: i64,
    pub user_id: i64,
    pub location_id: i64,
    pub can_manage_pos_users: Option<bool>,
    pub can_manage_terminals: Option<bool>,
    pub can_view_reports: Option<bool>,
}

/// `AssignMemberToLocation`: `execresult`.
pub async fn assign_member_to_location<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AssignMemberToLocationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_member_locations (business_id, user_id, location_id,\n    can_manage_pos_users, can_manage_terminals, can_view_reports)\nVALUES (?, ?, ?, ?, ?, ?)\nON DUPLICATE KEY UPDATE\n    id = LAST_INSERT_ID(id),\n    can_manage_pos_users = VALUES(can_manage_pos_users),\n    can_manage_terminals = VALUES(can_manage_terminals),\n    can_view_reports = VALUES(can_view_reports)")
        .bind(params.business_id)
        .bind(params.user_id)
        .bind(params.location_id)
        .bind(params.can_manage_pos_users)
        .bind(params.can_manage_terminals)
        .bind(params.can_view_reports)
        .execute(exec)
        .await
}

/// Params for `GetLocationRoster` (Go `GetLocationRosterParams`).
pub struct GetLocationRosterParams {
    pub business_id: i64,
    pub location_id: i64,
}

/// `GetLocationRoster`: `many`.
pub async fn get_location_roster<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetLocationRosterParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT ml.user_id,\n       ml.can_manage_pos_users,\n       (pu.id IS NOT NULL) AS has_pin,\n       COALESCE(pu.role, '') AS pos_role\nFROM pos_member_locations ml\nLEFT JOIN pos_users pu\n       ON pu.user_id = ml.user_id\n      AND pu.location_id = ml.location_id\n      AND pu.business_id = ml.business_id\n      AND pu.is_active = TRUE\nWHERE ml.business_id = ? AND ml.location_id = ?\nORDER BY ml.user_id ASC")
        .bind(params.business_id)
        .bind(params.location_id)
        .execute(exec)
        .await
}

/// Params for `GetMyLocations` (Go `GetMyLocationsParams`).
pub struct GetMyLocationsParams {
    pub business_id: i64,
    pub user_id: i64,
}

/// `GetMyLocations`: `many`.
pub async fn get_my_locations<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetMyLocationsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT ml.location_id, l.name, l.address, l.phone\nFROM pos_member_locations ml\nJOIN pos_locations l ON l.id = ml.location_id AND l.business_id = ml.business_id\nWHERE ml.business_id = ? AND ml.user_id = ?\nORDER BY l.name ASC")
        .bind(params.business_id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `ListBusinessLocations` (Go `ListBusinessLocationsParams`).
pub struct ListBusinessLocationsParams {
    pub business_id: i64,
}

/// `ListBusinessLocations`: `many`.
pub async fn list_business_locations<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListBusinessLocationsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT l.id AS location_id, l.name, l.address, l.phone\nFROM pos_locations l\nWHERE l.business_id = ? AND l.is_active = TRUE\nORDER BY l.name ASC")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListMemberLocations` (Go `ListMemberLocationsParams`).
pub struct ListMemberLocationsParams {
    pub business_id: i64,
    pub user_id: i64,
}

/// `ListMemberLocations`: `many`.
pub async fn list_member_locations<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListMemberLocationsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, business_id, user_id, location_id, can_manage_pos_users,\n       can_manage_terminals, can_view_reports, created_at\nFROM pos_member_locations WHERE business_id = ? AND user_id = ?")
        .bind(params.business_id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `LocationBelongsToBusiness` (Go `LocationBelongsToBusinessParams`).
pub struct LocationBelongsToBusinessParams {
    pub business_id: i64,
    pub id: i64,
}

/// `LocationBelongsToBusiness`: `one`.
pub async fn location_belongs_to_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LocationBelongsToBusinessParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT EXISTS (\n    SELECT 1\n    FROM pos_locations l\n    WHERE l.business_id = ? AND l.id = ?\n)")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `RemoveMemberFromLocation` (Go `RemoveMemberFromLocationParams`).
pub struct RemoveMemberFromLocationParams {
    pub business_id: i64,
    pub id: i64,
}

/// `RemoveMemberFromLocation`: `exec`.
pub async fn remove_member_from_location<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RemoveMemberFromLocationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM pos_member_locations WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
