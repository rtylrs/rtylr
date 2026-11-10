//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `AddGroupMember` (Go `AddGroupMemberParams`).
pub struct AddGroupMemberParams {
    pub group_id: i64,
    pub user_id: i64,
    pub business_id: i64,
    pub added_by: Option<i64>,
    pub user_id_2: i64,
    pub id: i64,
    pub business_id_2: i64,
}

/// `AddGroupMember`: `exec`.
pub async fn add_group_member<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AddGroupMemberParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO user_group_members (group_id, user_id, business_id, added_by)\nSELECT ?, ?, ?, ?\nFROM user_groups g\nJOIN authorization_business_members m ON m.business_id = g.business_id AND m.user_id = ? AND m.status = 'active'\nWHERE g.id = ? AND g.business_id = ?\nON DUPLICATE KEY UPDATE business_id = VALUES(business_id), added_by = VALUES(added_by)")
        .bind(params.group_id)
        .bind(params.user_id)
        .bind(params.business_id)
        .bind(params.added_by)
        .bind(params.user_id_2)
        .bind(params.id)
        .bind(params.business_id_2)
        .execute(exec)
        .await
}

/// Params for `CloseActiveSeatPeriod` (Go `CloseActiveSeatPeriodParams`).
pub struct CloseActiveSeatPeriodParams {
    pub period_end: Option<chrono::DateTime<chrono::Utc>>,
    pub subscription_id: i64,
}

/// `CloseActiveSeatPeriod`: `exec`.
pub async fn close_active_seat_period<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CloseActiveSeatPeriodParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE seat_periods SET period_end = ?\nWHERE subscription_id = ? AND period_end IS NULL",
    )
    .bind(params.period_end)
    .bind(params.subscription_id)
    .execute(exec)
    .await
}

/// Params for `ConsumeInviteToken` (Go `ConsumeInviteTokenParams`).
pub struct ConsumeInviteTokenParams {
    pub token: String,
}

/// `ConsumeInviteToken`: `execresult`.
pub async fn consume_invite_token<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ConsumeInviteTokenParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE invite_tokens\nSET used_at = NOW()\nWHERE token = ? AND used_at IS NULL AND expires_at > NOW()")
        .bind(&params.token)
        .execute(exec)
        .await
}

/// Params for `CountActiveMembers` (Go `CountActiveMembersParams`).
pub struct CountActiveMembersParams {
    pub business_id: i64,
}

/// `CountActiveMembers`: `one`.
pub async fn count_active_members<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountActiveMembersParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*)\nFROM (\n    SELECT b.owner_id AS user_id\n    FROM businesses b\n    JOIN users u ON u.id = b.owner_id\n    WHERE b.id = ? AND COALESCE(b.is_active, TRUE) = TRUE AND COALESCE(u.is_banned, FALSE) = FALSE\n    UNION\n    SELECT member.user_id\n    FROM authorization_business_members member\n    JOIN businesses b ON b.id = member.business_id\n    JOIN users u ON u.id = member.user_id\n    WHERE member.business_id = ? AND COALESCE(b.is_active, TRUE) = TRUE\n      AND member.status = 'active' AND COALESCE(u.is_banned, FALSE) = FALSE\n) active_users")
        .bind(params.business_id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountStaffUsersByBusinessID` (Go `CountStaffUsersByBusinessIDParams`).
pub struct CountStaffUsersByBusinessIDParams {
    pub business_id: i64,
}

/// `CountStaffUsersByBusinessID`: `one`.
pub async fn count_staff_users_by_business_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountStaffUsersByBusinessIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*)\nFROM authorization_business_members\nWHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateAccessRequest` (Go `CreateAccessRequestParams`).
pub struct CreateAccessRequestParams {
    pub business_id: i64,
    pub user_id: i64,
    pub app: String,
}

/// `CreateAccessRequest`: `exec`.
pub async fn create_access_request<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateAccessRequestParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO access_requests (business_id, user_id, app)\nVALUES (?, ?, ?)\nON DUPLICATE KEY UPDATE status = 'pending', created_at = NOW()")
        .bind(params.business_id)
        .bind(params.user_id)
        .bind(&params.app)
        .execute(exec)
        .await
}

/// Params for `CreateGroup` (Go `CreateGroupParams`).
pub struct CreateGroupParams {
    pub business_id: i64,
    pub name: String,
    pub description: Option<String>,
    pub color: Option<String>,
}

/// `CreateGroup`: `execresult`.
pub async fn create_group<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateGroupParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "INSERT INTO user_groups (business_id, name, description, color)\nVALUES (?, ?, ?, ?)",
    )
    .bind(params.business_id)
    .bind(&params.name)
    .bind(&params.description)
    .bind(&params.color)
    .execute(exec)
    .await
}

/// Params for `CreateInviteToken` (Go `CreateInviteTokenParams`).
pub struct CreateInviteTokenParams {
    pub user_id: i64,
    pub business_id: i64,
    pub token: String,
    pub expires_at: chrono::NaiveDate,
}

/// `CreateInviteToken`: `exec`.
pub async fn create_invite_token<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateInviteTokenParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "INSERT INTO invite_tokens (user_id, business_id, token, expires_at) VALUES (?, ?, ?, ?)",
    )
    .bind(params.user_id)
    .bind(params.business_id)
    .bind(&params.token)
    .bind(params.expires_at)
    .execute(exec)
    .await
}

/// Params for `CreateUser` (Go `CreateUserParams`).
pub struct CreateUserParams {
    pub email: String,
    pub username: Option<String>,
    pub password_hash: String,
    pub full_name: Option<String>,
}

/// `CreateUser`: `execresult`.
pub async fn create_user<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateUserParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO users (email, username, password_hash, full_name) VALUES (?, ?, ?, ?)")
        .bind(&params.email)
        .bind(&params.username)
        .bind(&params.password_hash)
        .bind(&params.full_name)
        .execute(exec)
        .await
}

/// Params for `DeleteGroup` (Go `DeleteGroupParams`).
pub struct DeleteGroupParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteGroup`: `execresult`.
pub async fn delete_group<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteGroupParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM user_groups WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `GetActiveMemberContact` (Go `GetActiveMemberContactParams`).
pub struct GetActiveMemberContactParams {
    pub business_id: i64,
    pub user_id: i64,
}

/// `GetActiveMemberContact`: `one`.
pub async fn get_active_member_contact<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetActiveMemberContactParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT u.email, COALESCE(u.full_name, '') AS full_name\nFROM authorization_business_members AS member\nJOIN users AS u ON u.id = member.user_id\nWHERE member.business_id = ? AND member.user_id = ? AND member.status = 'active'\nLIMIT 1")
        .bind(params.business_id)
        .bind(params.user_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetBusinessApps` (Go `GetBusinessAppsParams`).
pub struct GetBusinessAppsParams {
    pub business_id: i64,
}

/// `GetBusinessApps`: `one`.
pub async fn get_business_apps<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetBusinessAppsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT p.apps\nFROM subscriptions s\nJOIN plans p ON p.id = s.plan_id\nWHERE s.business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetBusinessByID` (Go `GetBusinessByIDParams`).
pub struct GetBusinessByIDParams {
    pub id: i64,
}

/// `GetBusinessByID`: `one`.
pub async fn get_business_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetBusinessByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT name, owner_id FROM businesses WHERE id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetBusinessSubscriptionInfo` (Go `GetBusinessSubscriptionInfoParams`).
pub struct GetBusinessSubscriptionInfoParams {
    pub business_id: i64,
}

/// `GetBusinessSubscriptionInfo`: `one`.
pub async fn get_business_subscription_info<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetBusinessSubscriptionInfoParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT p.slug, s.status\nFROM subscriptions s\nJOIN plans p ON p.id = s.plan_id\nWHERE s.business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// `GetDefaultPlanApps`: `one`.
pub async fn get_default_plan_apps<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT apps FROM plans WHERE slug = 'free'")
        .fetch_optional(exec)
        .await
}

/// Params for `GetGroup` (Go `GetGroupParams`).
pub struct GetGroupParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetGroup`: `one`.
pub async fn get_group<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetGroupParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, name, COALESCE(description,'') AS description, color, created_at, updated_at\nFROM user_groups WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetInviteToken` (Go `GetInviteTokenParams`).
pub struct GetInviteTokenParams {
    pub token: String,
}

/// `GetInviteToken`: `one`.
pub async fn get_invite_token<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetInviteTokenParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT user_id, business_id, expires_at, used_at FROM invite_tokens WHERE token = ?",
    )
    .bind(&params.token)
    .fetch_optional(exec)
    .await
}

/// Params for `GetStaffUserByID` (Go `GetStaffUserByIDParams`).
pub struct GetStaffUserByIDParams {
    pub business_id: i64,
    pub user_id: i64,
}

/// `GetStaffUserByID`: `one`.
pub async fn get_staff_user_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetStaffUserByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT u.id, u.email, COALESCE(u.full_name,'') AS full_name, u.created_at,\n       member.member_type, member.status AS member_status\nFROM authorization_business_members AS member\nJOIN users AS u ON u.id = member.user_id\nWHERE member.business_id = ? AND member.user_id = ?\nLIMIT 1")
        .bind(params.business_id)
        .bind(params.user_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetUserBusinessID` (Go `GetUserBusinessIDParams`).
pub struct GetUserBusinessIDParams {
    pub id: i64,
}

/// `GetUserBusinessID`: `one`.
pub async fn get_user_business_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetUserBusinessIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT business_id FROM users WHERE id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetUserByEmail` (Go `GetUserByEmailParams`).
pub struct GetUserByEmailParams {
    pub email: String,
}

/// `GetUserByEmail`: `one`.
pub async fn get_user_by_email<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetUserByEmailParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT id, email, COALESCE(full_name,'') AS full_name FROM users WHERE email = ? LIMIT 1",
    )
    .bind(&params.email)
    .fetch_optional(exec)
    .await
}

/// Params for `GetUserByID` (Go `GetUserByIDParams`).
pub struct GetUserByIDParams {
    pub id: i64,
}

/// `GetUserByID`: `one`.
pub async fn get_user_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetUserByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT email, COALESCE(full_name,'') AS full_name, is_console_admin FROM users WHERE id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `IsOnboarded` (Go `IsOnboardedParams`).
pub struct IsOnboardedParams {
    pub id: i64,
}

/// `IsOnboarded`: `one`.
pub async fn is_onboarded<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &IsOnboardedParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(onboarded, FALSE) FROM users WHERE id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListAccessRequestsByBusiness` (Go `ListAccessRequestsByBusinessParams`).
pub struct ListAccessRequestsByBusinessParams {
    pub business_id: i64,
}

/// `ListAccessRequestsByBusiness`: `many`.
pub async fn list_access_requests_by_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListAccessRequestsByBusinessParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT ar.id, ar.user_id, ar.app, ar.status, ar.created_at,\n       COALESCE(u.full_name,'') AS requester_name, COALESCE(u.email,'') AS requester_email\nFROM access_requests ar\nJOIN users u ON u.id = ar.user_id\nWHERE ar.business_id = ?\n  AND ar.status = 'pending'\nORDER BY ar.created_at DESC")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListActiveMemberUserIDsByBusinessID` (Go `ListActiveMemberUserIDsByBusinessIDParams`).
pub struct ListActiveMemberUserIDsByBusinessIDParams {
    pub id: i64,
    pub business_id: i64,
}

/// `ListActiveMemberUserIDsByBusinessID`: `many`.
pub async fn list_active_member_user_i_ds_by_business_id<
    'e,
    E: sqlx::Executor<'e, Database = MySql>,
>(
    exec: E,
    params: &ListActiveMemberUserIDsByBusinessIDParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT business.owner_id AS user_id\nFROM businesses AS business\nWHERE business.id = ?\nUNION\nSELECT user_id\nFROM authorization_business_members\nWHERE business_id = ? AND status = 'active'")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListAdminsAndOwnerUserIDs` (Go `ListAdminsAndOwnerUserIDsParams`).
pub struct ListAdminsAndOwnerUserIDsParams {
    pub id: i64,
    pub business_id: i64,
}

/// `ListAdminsAndOwnerUserIDs`: `many`.
pub async fn list_admins_and_owner_user_i_ds<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListAdminsAndOwnerUserIDsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT b.owner_id AS user_id\nFROM businesses b\nWHERE b.id = ?\nUNION\nSELECT member.user_id\nFROM authorization_business_members member\nWHERE member.business_id = ? AND member.member_type = 'admin' AND member.status = 'active'")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListGroupMembers` (Go `ListGroupMembersParams`).
pub struct ListGroupMembersParams {
    pub group_id: i64,
    pub business_id: i64,
}

/// `ListGroupMembers`: `many`.
pub async fn list_group_members<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListGroupMembersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT m.user_id, u.email, COALESCE(u.full_name,'') AS full_name, m.added_at\nFROM user_group_members m\nJOIN users u ON u.id = m.user_id\nWHERE m.group_id = ? AND m.business_id = ?\nORDER BY m.added_at DESC\nLIMIT 500")
        .bind(params.group_id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListGroups` (Go `ListGroupsParams`).
pub struct ListGroupsParams {
    pub business_id: i64,
}

/// `ListGroups`: `many`.
pub async fn list_groups<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListGroupsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT g.id, g.business_id, g.name, COALESCE(g.description,'') AS description, g.color,\n       g.created_at, g.updated_at, COUNT(m.user_id) AS member_count\nFROM user_groups g\nLEFT JOIN user_group_members m ON m.group_id = g.id AND m.business_id = g.business_id\nWHERE g.business_id = ?\nGROUP BY g.id\nORDER BY g.created_at DESC\nLIMIT 500")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListStaffUsersByBusinessID` (Go `ListStaffUsersByBusinessIDParams`).
pub struct ListStaffUsersByBusinessIDParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListStaffUsersByBusinessID`: `many`.
pub async fn list_staff_users_by_business_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListStaffUsersByBusinessIDParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT u.id, u.email, COALESCE(u.full_name,'') AS full_name, u.created_at,\n       member.member_type, member.status AS member_status\nFROM authorization_business_members AS member\nJOIN users AS u ON u.id = member.user_id\nWHERE member.business_id = ?\nORDER BY u.created_at DESC, u.id DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListUserGroups` (Go `ListUserGroupsParams`).
pub struct ListUserGroupsParams {
    pub user_id: i64,
    pub business_id: i64,
}

/// `ListUserGroups`: `many`.
pub async fn list_user_groups<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListUserGroupsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT g.id, g.name, g.color\nFROM user_group_members m\nJOIN user_groups g ON g.id = m.group_id AND g.business_id = m.business_id\nWHERE m.user_id = ? AND m.business_id = ?\nORDER BY g.name\nLIMIT 500")
        .bind(params.user_id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListUserGroupsForUsers` (Go `ListUserGroupsForUsersParams`).
pub struct ListUserGroupsForUsersParams {
    pub user_ids: Vec<i64>,
    pub business_id: i64,
}

/// `ListUserGroupsForUsers`: `many`.
pub async fn list_user_groups_for_users<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListUserGroupsForUsersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT m.user_id, g.id, g.name, g.color\nFROM user_group_members m\nJOIN user_groups g ON g.id = m.group_id AND g.business_id = m.business_id\nWHERE m.user_id IN ({SLICE:user_ids}) AND m.business_id = ?\nORDER BY g.name".to_string();

    if params.user_ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:user_ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:user_ids}",
            &std::iter::repeat_n("?", params.user_ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query.bind(params.business_id);
    for value in &params.user_ids {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `LockSubscriptionForSeatChange` (Go `LockSubscriptionForSeatChangeParams`).
pub struct LockSubscriptionForSeatChangeParams {
    pub business_id: i64,
}

/// `LockSubscriptionForSeatChange`: `one`.
pub async fn lock_subscription_for_seat_change<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LockSubscriptionForSeatChangeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT s.id, p.price_per_seat_cents, p.max_seats\nFROM subscriptions s\nJOIN plans p ON p.id = s.plan_id\nWHERE s.business_id = ?\nFOR UPDATE")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `MarkInviteTokenUsed` (Go `MarkInviteTokenUsedParams`).
pub struct MarkInviteTokenUsedParams {
    pub token: String,
}

/// `MarkInviteTokenUsed`: `exec`.
pub async fn mark_invite_token_used<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MarkInviteTokenUsedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE invite_tokens SET used_at = NOW() WHERE token = ?")
        .bind(&params.token)
        .execute(exec)
        .await
}

/// Params for `MarkOnboarded` (Go `MarkOnboardedParams`).
pub struct MarkOnboardedParams {
    pub id: i64,
}

/// `MarkOnboarded`: `exec`.
pub async fn mark_onboarded<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MarkOnboardedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE users SET onboarded = TRUE WHERE id = ?")
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `OpenSeatPeriod` (Go `OpenSeatPeriodParams`).
pub struct OpenSeatPeriodParams {
    pub subscription_id: i64,
    pub business_id: i64,
    pub seat_count: i32,
    pub price_per_seat_cents: i32,
    pub period_start: chrono::NaiveDate,
    pub billing_month: chrono::NaiveDate,
}

/// `OpenSeatPeriod`: `exec`.
pub async fn open_seat_period<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &OpenSeatPeriodParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO seat_periods (subscription_id, business_id, seat_count, price_per_seat_cents, period_start, billing_month)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.subscription_id)
        .bind(params.business_id)
        .bind(params.seat_count)
        .bind(params.price_per_seat_cents)
        .bind(params.period_start)
        .bind(params.billing_month)
        .execute(exec)
        .await
}

/// Params for `ReleaseInviteToken` (Go `ReleaseInviteTokenParams`).
pub struct ReleaseInviteTokenParams {
    pub token: String,
}

/// `ReleaseInviteToken`: `exec`.
pub async fn release_invite_token<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ReleaseInviteTokenParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE invite_tokens SET used_at = NULL WHERE token = ?")
        .bind(&params.token)
        .execute(exec)
        .await
}

/// Params for `RemoveGroupMember` (Go `RemoveGroupMemberParams`).
pub struct RemoveGroupMemberParams {
    pub group_id: i64,
    pub user_id: i64,
    pub business_id: i64,
}

/// `RemoveGroupMember`: `exec`.
pub async fn remove_group_member<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RemoveGroupMemberParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "DELETE FROM user_group_members WHERE group_id = ? AND user_id = ? AND business_id = ?",
    )
    .bind(params.group_id)
    .bind(params.user_id)
    .bind(params.business_id)
    .execute(exec)
    .await
}

/// Params for `ResolveAccessRequest` (Go `ResolveAccessRequestParams`).
pub struct ResolveAccessRequestParams {
    pub status: String,
    pub id: i64,
    pub business_id: i64,
}

/// `ResolveAccessRequest`: `exec`.
pub async fn resolve_access_request<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ResolveAccessRequestParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE access_requests SET status = ? WHERE id = ? AND business_id = ?")
        .bind(&params.status)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `SetPassword` (Go `SetPasswordParams`).
pub struct SetPasswordParams {
    pub password_hash: String,
    pub id: i64,
}

/// `SetPassword`: `exec`.
pub async fn set_password<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SetPasswordParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE users SET password_hash = ? WHERE id = ?")
        .bind(&params.password_hash)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `SetUserBusinessIDIfUnset` (Go `SetUserBusinessIDIfUnsetParams`).
pub struct SetUserBusinessIDIfUnsetParams {
    pub business_id: Option<i64>,
    pub id: i64,
}

/// `SetUserBusinessIDIfUnset`: `execrows`.
pub async fn set_user_business_id_if_unset<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SetUserBusinessIDIfUnsetParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE users\nSET business_id = ?\nWHERE id = ?\n  AND (business_id IS NULL OR business_id = 0)")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpdateGroup` (Go `UpdateGroupParams`).
pub struct UpdateGroupParams {
    pub name: String,
    pub description: Option<String>,
    pub color: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateGroup`: `execresult`.
pub async fn update_group<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateGroupParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE user_groups SET name = ?, description = ?, color = ?, updated_at = NOW()\nWHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(&params.description)
        .bind(&params.color)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateSubscriptionSeatCount` (Go `UpdateSubscriptionSeatCountParams`).
pub struct UpdateSubscriptionSeatCountParams {
    pub seat_count: i32,
    pub business_id: i64,
}

/// `UpdateSubscriptionSeatCount`: `exec`.
pub async fn update_subscription_seat_count<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateSubscriptionSeatCountParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE subscriptions SET seat_count = ?, updated_at = CURRENT_TIMESTAMP WHERE business_id = ?")
        .bind(params.seat_count)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateUserProfile` (Go `UpdateUserProfileParams`).
pub struct UpdateUserProfileParams {
    pub full_name: Option<String>,
    pub id: i64,
}

/// `UpdateUserProfile`: `exec`.
pub async fn update_user_profile<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateUserProfileParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE users SET full_name = ?, updated_at = NOW() WHERE id = ?")
        .bind(&params.full_name)
        .bind(params.id)
        .execute(exec)
        .await
}
