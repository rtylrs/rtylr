//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountPerformanceReviewsByCycle` (Go `CountPerformanceReviewsByCycleParams`).
pub struct CountPerformanceReviewsByCycleParams {
    pub business_id: i64,
    pub review_cycle_id: i64,
}

/// `CountPerformanceReviewsByCycle`: `one`.
pub async fn count_performance_reviews_by_cycle<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountPerformanceReviewsByCycleParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total\nFROM hr_performance_reviews\nWHERE business_id = ? AND review_cycle_id = ?")
        .bind(params.business_id)
        .bind(params.review_cycle_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountPerformanceReviewsByCycleAndEmployee` (Go `CountPerformanceReviewsByCycleAndEmployeeParams`).
pub struct CountPerformanceReviewsByCycleAndEmployeeParams {
    pub business_id: i64,
    pub review_cycle_id: i64,
    pub employee_id: i64,
}

/// `CountPerformanceReviewsByCycleAndEmployee`: `one`.
pub async fn count_performance_reviews_by_cycle_and_employee<
    'e,
    E: sqlx::Executor<'e, Database = MySql>,
>(
    exec: E,
    params: &CountPerformanceReviewsByCycleAndEmployeeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total\nFROM hr_performance_reviews\nWHERE business_id = ? AND review_cycle_id = ? AND employee_id = ?")
        .bind(params.business_id)
        .bind(params.review_cycle_id)
        .bind(params.employee_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountReviewCycles` (Go `CountReviewCyclesParams`).
pub struct CountReviewCyclesParams {
    pub business_id: i64,
}

/// `CountReviewCycles`: `one`.
pub async fn count_review_cycles<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountReviewCyclesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT COUNT(*) AS total\nFROM hr_performance_review_cycles\nWHERE business_id = ?",
    )
    .bind(params.business_id)
    .fetch_optional(exec)
    .await
}

/// Params for `CountReviewGoals` (Go `CountReviewGoalsParams`).
pub struct CountReviewGoalsParams {
    pub business_id: i64,
    pub employee_id: i64,
    pub review_cycle_id: i64,
}

/// `CountReviewGoals`: `one`.
pub async fn count_review_goals<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountReviewGoalsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total\nFROM hr_performance_review_goals\nWHERE business_id = ?\n  AND (? = 0 OR employee_id = ?)\n  AND (? = 0 OR review_cycle_id = ?)")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.employee_id)
        .bind(params.review_cycle_id)
        .bind(params.review_cycle_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CreatePerformanceReview` (Go `CreatePerformanceReviewParams`).
pub struct CreatePerformanceReviewParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub employee_id: i64,
    pub review_cycle_id: i64,
    pub reviewer_employee_id: i64,
    pub review_type: String,
}

/// `CreatePerformanceReview`: `exec`.
pub async fn create_performance_review<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreatePerformanceReviewParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_performance_reviews (owner_id, business_id, employee_id, review_cycle_id, reviewer_employee_id, review_type, status)\nVALUES (?, ?, ?, ?, ?, ?, 'draft')")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.review_cycle_id)
        .bind(params.reviewer_employee_id)
        .bind(&params.review_type)
        .execute(exec)
        .await
}

/// Params for `CreateReviewCycle` (Go `CreateReviewCycleParams`).
pub struct CreateReviewCycleParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub cycle_type: String,
    pub period_start: chrono::NaiveDate,
    pub period_end: chrono::NaiveDate,
}

/// `CreateReviewCycle`: `execresult`.
pub async fn create_review_cycle<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateReviewCycleParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_performance_review_cycles (owner_id, business_id, name, cycle_type, period_start, period_end, status)\nVALUES (?, ?, ?, ?, ?, ?, 'draft')")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.cycle_type)
        .bind(params.period_start)
        .bind(params.period_end)
        .execute(exec)
        .await
}

/// Params for `CreateReviewGoal` (Go `CreateReviewGoalParams`).
pub struct CreateReviewGoalParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub employee_id: i64,
    pub review_cycle_id: i64,
    pub title: String,
    pub description: Option<String>,
    pub weight: Option<String>,
    pub success_criteria: Option<String>,
}

/// `CreateReviewGoal`: `execresult`.
pub async fn create_review_goal<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateReviewGoalParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_performance_review_goals (owner_id, business_id, employee_id, review_cycle_id, title, description, weight, success_criteria, status)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, 'active')")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.review_cycle_id)
        .bind(&params.title)
        .bind(&params.description)
        .bind(&params.weight)
        .bind(&params.success_criteria)
        .execute(exec)
        .await
}

/// Params for `GetPerformanceConfig` (Go `GetPerformanceConfigParams`).
pub struct GetPerformanceConfigParams {
    pub business_id: i64,
}

/// `GetPerformanceConfig`: `one`.
pub async fn get_performance_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPerformanceConfigParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, rating_scale_type, rating_min, rating_max, COALESCE(rating_labels, '[]') AS rating_labels, peer_feedback_enabled, created_at, updated_at\nFROM hr_performance_config\nWHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetPerformanceReview` (Go `GetPerformanceReviewParams`).
pub struct GetPerformanceReviewParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetPerformanceReview`: `one`.
pub async fn get_performance_review<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPerformanceReviewParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, review_cycle_id, reviewer_employee_id, review_type, overall_rating, narrative, status, shared_with_employee_at, created_at, updated_at\nFROM hr_performance_reviews\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetReviewCycle` (Go `GetReviewCycleParams`).
pub struct GetReviewCycleParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetReviewCycle`: `one`.
pub async fn get_review_cycle<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetReviewCycleParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, cycle_type, period_start, period_end, status, created_at, updated_at\nFROM hr_performance_review_cycles\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetReviewGoal` (Go `GetReviewGoalParams`).
pub struct GetReviewGoalParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetReviewGoal`: `one`.
pub async fn get_review_goal<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetReviewGoalParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, review_cycle_id, title, description, weight, success_criteria, status, created_at, updated_at\nFROM hr_performance_review_goals\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListActiveEmployeeIDs` (Go `ListActiveEmployeeIDsParams`).
pub struct ListActiveEmployeeIDsParams {
    pub business_id: i64,
}

/// `ListActiveEmployeeIDs`: `many`.
pub async fn list_active_employee_i_ds<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListActiveEmployeeIDsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id FROM hr_employees WHERE business_id = ? AND status = 'active' ORDER BY id LIMIT 500")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListPerformanceReviewsByCycle` (Go `ListPerformanceReviewsByCycleParams`).
pub struct ListPerformanceReviewsByCycleParams {
    pub business_id: i64,
    pub review_cycle_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListPerformanceReviewsByCycle`: `many`.
pub async fn list_performance_reviews_by_cycle<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListPerformanceReviewsByCycleParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, review_cycle_id, reviewer_employee_id, review_type, overall_rating, narrative, status, shared_with_employee_at, created_at, updated_at\nFROM hr_performance_reviews\nWHERE business_id = ? AND review_cycle_id = ?\nORDER BY created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.review_cycle_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListPerformanceReviewsByCycleAndEmployee` (Go `ListPerformanceReviewsByCycleAndEmployeeParams`).
pub struct ListPerformanceReviewsByCycleAndEmployeeParams {
    pub business_id: i64,
    pub review_cycle_id: i64,
    pub employee_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListPerformanceReviewsByCycleAndEmployee`: `many`.
pub async fn list_performance_reviews_by_cycle_and_employee<
    'e,
    E: sqlx::Executor<'e, Database = MySql>,
>(
    exec: E,
    params: &ListPerformanceReviewsByCycleAndEmployeeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, review_cycle_id, reviewer_employee_id, review_type, overall_rating, narrative, status, shared_with_employee_at, created_at, updated_at\nFROM hr_performance_reviews\nWHERE business_id = ? AND review_cycle_id = ? AND employee_id = ?\nORDER BY created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.review_cycle_id)
        .bind(params.employee_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListReviewCycles` (Go `ListReviewCyclesParams`).
pub struct ListReviewCyclesParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListReviewCycles`: `many`.
pub async fn list_review_cycles<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListReviewCyclesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, cycle_type, period_start, period_end, status, created_at, updated_at\nFROM hr_performance_review_cycles\nWHERE business_id = ?\nORDER BY created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListReviewGoals` (Go `ListReviewGoalsParams`).
pub struct ListReviewGoalsParams {
    pub business_id: i64,
    pub employee_id: i64,
    pub review_cycle_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListReviewGoals`: `many`.
pub async fn list_review_goals<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListReviewGoalsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, review_cycle_id, title, description, weight, success_criteria, status, created_at, updated_at\nFROM hr_performance_review_goals\nWHERE business_id = ?\n  AND (? = 0 OR employee_id = ?)\n  AND (? = 0 OR review_cycle_id = ?)\nORDER BY created_at ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.employee_id)
        .bind(params.review_cycle_id)
        .bind(params.review_cycle_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ShareReview` (Go `ShareReviewParams`).
pub struct ShareReviewParams {
    pub business_id: i64,
    pub id: i64,
}

/// `ShareReview`: `exec`.
pub async fn share_review<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ShareReviewParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_performance_reviews\nSET status = 'shared', shared_with_employee_at = NOW()\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpdateReviewCycleStatus` (Go `UpdateReviewCycleStatusParams`).
pub struct UpdateReviewCycleStatusParams {
    pub status: String,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateReviewCycleStatus`: `exec`.
pub async fn update_review_cycle_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateReviewCycleStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE hr_performance_review_cycles\nSET status = ?\nWHERE business_id = ? AND id = ?",
    )
    .bind(&params.status)
    .bind(params.business_id)
    .bind(params.id)
    .execute(exec)
    .await
}

/// Params for `UpdateReviewGoal` (Go `UpdateReviewGoalParams`).
pub struct UpdateReviewGoalParams {
    pub title: String,
    pub description: Option<String>,
    pub weight: Option<String>,
    pub success_criteria: Option<String>,
    pub status: String,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateReviewGoal`: `exec`.
pub async fn update_review_goal<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateReviewGoalParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_performance_review_goals\nSET title = ?, description = ?, weight = ?, success_criteria = ?, status = ?\nWHERE business_id = ? AND id = ?")
        .bind(&params.title)
        .bind(&params.description)
        .bind(&params.weight)
        .bind(&params.success_criteria)
        .bind(&params.status)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpdateReviewStatus` (Go `UpdateReviewStatusParams`).
pub struct UpdateReviewStatusParams {
    pub status: String,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateReviewStatus`: `exec`.
pub async fn update_review_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateReviewStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_performance_reviews\nSET status = ?\nWHERE business_id = ? AND id = ?")
        .bind(&params.status)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpdateReviewSubmission` (Go `UpdateReviewSubmissionParams`).
pub struct UpdateReviewSubmissionParams {
    pub overall_rating: Option<String>,
    pub narrative: Option<String>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateReviewSubmission`: `exec`.
pub async fn update_review_submission<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateReviewSubmissionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_performance_reviews\nSET overall_rating = ?, narrative = ?, status = 'submitted'\nWHERE business_id = ? AND id = ?")
        .bind(&params.overall_rating)
        .bind(&params.narrative)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpsertPerformanceConfig` (Go `UpsertPerformanceConfigParams`).
pub struct UpsertPerformanceConfigParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub rating_scale_type: String,
    pub rating_min: i32,
    pub rating_max: i32,
    pub rating_labels: String,
    pub peer_feedback_enabled: bool,
}

/// `UpsertPerformanceConfig`: `execresult`.
pub async fn upsert_performance_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertPerformanceConfigParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_performance_config (owner_id, business_id, rating_scale_type, rating_min, rating_max, rating_labels, peer_feedback_enabled)\nVALUES (?, ?, ?, ?, ?, ?, ?)\nON DUPLICATE KEY UPDATE\n    rating_scale_type = VALUES(rating_scale_type),\n    rating_min = VALUES(rating_min),\n    rating_max = VALUES(rating_max),\n    rating_labels = VALUES(rating_labels),\n    peer_feedback_enabled = VALUES(peer_feedback_enabled)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.rating_scale_type)
        .bind(params.rating_min)
        .bind(params.rating_max)
        .bind(&params.rating_labels)
        .bind(params.peer_feedback_enabled)
        .execute(exec)
        .await
}
