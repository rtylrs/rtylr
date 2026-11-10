//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateRule` (Go `CreateRuleParams`).
pub struct CreateRuleParams {
    pub feature_key: String,
    pub priority: i32,
    pub description: String,
    pub variant: String,
    pub rollout_percentage: i32,
    pub enabled: bool,
}

/// `CreateRule`: `execresult`.
pub async fn create_rule<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateRuleParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO feature_rules (feature_key, priority, description, variant, rollout_percentage, enabled)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(&params.feature_key)
        .bind(params.priority)
        .bind(&params.description)
        .bind(&params.variant)
        .bind(params.rollout_percentage)
        .bind(params.enabled)
        .execute(exec)
        .await
}

/// Params for `DeleteConditionsByRule` (Go `DeleteConditionsByRuleParams`).
pub struct DeleteConditionsByRuleParams {
    pub rule_id: i64,
}

/// `DeleteConditionsByRule`: `exec`.
pub async fn delete_conditions_by_rule<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteConditionsByRuleParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM feature_rule_conditions WHERE rule_id = ?")
        .bind(params.rule_id)
        .execute(exec)
        .await
}

/// Params for `DeleteFeatureFlag` (Go `DeleteFeatureFlagParams`).
pub struct DeleteFeatureFlagParams {
    pub feature_key: String,
}

/// `DeleteFeatureFlag`: `exec`.
pub async fn delete_feature_flag<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteFeatureFlagParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM feature_flags WHERE feature_key = ?")
        .bind(&params.feature_key)
        .execute(exec)
        .await
}

/// Params for `DeleteFeatureTarget` (Go `DeleteFeatureTargetParams`).
pub struct DeleteFeatureTargetParams {
    pub feature_key: String,
    pub owner_id: i64,
}

/// `DeleteFeatureTarget`: `exec`.
pub async fn delete_feature_target<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteFeatureTargetParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM feature_targets WHERE feature_key = ? AND owner_id = ?")
        .bind(&params.feature_key)
        .bind(params.owner_id)
        .execute(exec)
        .await
}

/// Params for `DeleteRule` (Go `DeleteRuleParams`).
pub struct DeleteRuleParams {
    pub id: i64,
}

/// `DeleteRule`: `exec`.
pub async fn delete_rule<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteRuleParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM feature_rules WHERE id = ?")
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `GetFeatureFlag` (Go `GetFeatureFlagParams`).
pub struct GetFeatureFlagParams {
    pub feature_key: String,
}

/// `GetFeatureFlag`: `one`.
pub async fn get_feature_flag<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetFeatureFlagParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, feature_key, description, enabled, rollout_percentage, variant_on, variant_off, created_at, updated_at\nFROM feature_flags\nWHERE feature_key = ?")
        .bind(&params.feature_key)
        .fetch_optional(exec)
        .await
}

/// Params for `GetUserEmailByID` (Go `GetUserEmailByIDParams`).
pub struct GetUserEmailByIDParams {
    pub id: i64,
}

/// `GetUserEmailByID`: `one`.
pub async fn get_user_email_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetUserEmailByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT email FROM users WHERE id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// `ListAllRuleConditions`: `many`.
pub async fn list_all_rule_conditions<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "SELECT id, rule_id, attribute, operator, value, created_at\nFROM feature_rule_conditions",
    )
    .execute(exec)
    .await
}

/// `ListAllRules`: `many`.
pub async fn list_all_rules<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, feature_key, priority, description, variant, rollout_percentage, enabled, created_at, updated_at\nFROM feature_rules\nWHERE enabled = 1\nORDER BY feature_key ASC, priority ASC, id ASC")
        .execute(exec)
        .await
}

/// Params for `ListConditionsByRuleIDs` (Go `ListConditionsByRuleIDsParams`).
pub struct ListConditionsByRuleIDsParams {
    pub rule_ids: Vec<i64>,
}

/// `ListConditionsByRuleIDs`: `many`.
pub async fn list_conditions_by_rule_i_ds<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListConditionsByRuleIDsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT id, rule_id, attribute, operator, value, created_at\nFROM feature_rule_conditions\nWHERE rule_id IN ({SLICE:rule_ids})\nORDER BY rule_id ASC, id ASC".to_string();

    if params.rule_ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:rule_ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:rule_ids}",
            &std::iter::repeat_n("?", params.rule_ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query;
    for value in &params.rule_ids {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// `ListFeatureFlags`: `many`.
pub async fn list_feature_flags<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, feature_key, description, enabled, rollout_percentage, variant_on, variant_off, created_at, updated_at\nFROM feature_flags\nORDER BY feature_key ASC")
        .execute(exec)
        .await
}

/// Params for `ListFeatureTargetsByKey` (Go `ListFeatureTargetsByKeyParams`).
pub struct ListFeatureTargetsByKeyParams {
    pub feature_key: String,
}

/// `ListFeatureTargetsByKey`: `many`.
pub async fn list_feature_targets_by_key<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListFeatureTargetsByKeyParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, feature_key, owner_id, enabled, variant, created_at, updated_at\nFROM feature_targets\nWHERE feature_key = ?\nORDER BY owner_id ASC")
        .bind(&params.feature_key)
        .execute(exec)
        .await
}

/// Params for `ListFeatureTargetsByOwner` (Go `ListFeatureTargetsByOwnerParams`).
pub struct ListFeatureTargetsByOwnerParams {
    pub owner_id: i64,
}

/// `ListFeatureTargetsByOwner`: `many`.
pub async fn list_feature_targets_by_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListFeatureTargetsByOwnerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, feature_key, owner_id, enabled, variant, created_at, updated_at\nFROM feature_targets\nWHERE owner_id = ?")
        .bind(params.owner_id)
        .execute(exec)
        .await
}

/// Params for `ListRulesByKey` (Go `ListRulesByKeyParams`).
pub struct ListRulesByKeyParams {
    pub feature_key: String,
}

/// `ListRulesByKey`: `many`.
pub async fn list_rules_by_key<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListRulesByKeyParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, feature_key, priority, description, variant, rollout_percentage, enabled, created_at, updated_at\nFROM feature_rules\nWHERE feature_key = ?\nORDER BY priority ASC, id ASC")
        .bind(&params.feature_key)
        .execute(exec)
        .await
}

/// Params for `UpdateRule` (Go `UpdateRuleParams`).
pub struct UpdateRuleParams {
    pub priority: i32,
    pub description: String,
    pub variant: String,
    pub rollout_percentage: i32,
    pub enabled: bool,
    pub id: i64,
}

/// `UpdateRule`: `exec`.
pub async fn update_rule<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateRuleParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE feature_rules\nSET priority = ?, description = ?, variant = ?, rollout_percentage = ?, enabled = ?\nWHERE id = ?")
        .bind(params.priority)
        .bind(&params.description)
        .bind(&params.variant)
        .bind(params.rollout_percentage)
        .bind(params.enabled)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpsertFeatureFlag` (Go `UpsertFeatureFlagParams`).
pub struct UpsertFeatureFlagParams {
    pub feature_key: String,
    pub description: String,
    pub enabled: bool,
    pub rollout_percentage: i32,
    pub variant_on: String,
    pub variant_off: String,
}

/// `UpsertFeatureFlag`: `exec`.
pub async fn upsert_feature_flag<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertFeatureFlagParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO feature_flags (feature_key, description, enabled, rollout_percentage, variant_on, variant_off)\nVALUES (?, ?, ?, ?, ?, ?)\nON DUPLICATE KEY UPDATE\n  description = VALUES(description),\n  enabled = VALUES(enabled),\n  rollout_percentage = VALUES(rollout_percentage),\n  variant_on = VALUES(variant_on),\n  variant_off = VALUES(variant_off)")
        .bind(&params.feature_key)
        .bind(&params.description)
        .bind(params.enabled)
        .bind(params.rollout_percentage)
        .bind(&params.variant_on)
        .bind(&params.variant_off)
        .execute(exec)
        .await
}

/// Params for `UpsertFeatureTarget` (Go `UpsertFeatureTargetParams`).
pub struct UpsertFeatureTargetParams {
    pub feature_key: String,
    pub owner_id: i64,
    pub enabled: bool,
    pub variant: Option<String>,
}

/// `UpsertFeatureTarget`: `exec`.
pub async fn upsert_feature_target<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertFeatureTargetParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO feature_targets (feature_key, owner_id, enabled, variant)\nVALUES (?, ?, ?, ?)\nON DUPLICATE KEY UPDATE\n  enabled = VALUES(enabled),\n  variant = VALUES(variant)")
        .bind(&params.feature_key)
        .bind(params.owner_id)
        .bind(params.enabled)
        .bind(&params.variant)
        .execute(exec)
        .await
}
