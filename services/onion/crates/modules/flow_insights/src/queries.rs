//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `GetAvgCompletionDays` (Go `GetAvgCompletionDaysParams`).
pub struct GetAvgCompletionDaysParams {
    pub business_id: i64,
}

/// `GetAvgCompletionDays`: `one`.
pub async fn get_avg_completion_days<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetAvgCompletionDaysParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(AVG(DATEDIFF(completed_at, created_at)), 0) AS avg_days\nFROM flow_tasks WHERE business_id = ? AND status = 'done' AND completed_at IS NOT NULL")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetMilestonesByProjectIDs` (Go `GetMilestonesByProjectIDsParams`).
pub struct GetMilestonesByProjectIDsParams {
    pub business_id: i64,
    pub project_ids: Vec<i64>,
}

/// `GetMilestonesByProjectIDs`: `many`.
pub async fn get_milestones_by_project_i_ds<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetMilestonesByProjectIDsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT\n    fm.project_id,\n    fm.id,\n    fm.name,\n    COALESCE(DATE_FORMAT(fm.target_date, '%Y-%m-%d'), '') AS due_date,\n    COUNT(ft.id) AS task_count,\n    COUNT(CASE WHEN ft.status = 'done' THEN 1 END) AS done_count\nFROM flow_milestones fm\nLEFT JOIN flow_milestone_tasks fmt ON fmt.milestone_id = fm.id\nLEFT JOIN flow_tasks ft ON ft.id = fmt.task_id AND ft.parent_task_id IS NULL\nWHERE fm.business_id = ? AND fm.project_id IN ({SLICE:project_ids})\nGROUP BY fm.project_id, fm.id, fm.name, fm.target_date\nORDER BY fm.target_date ASC".to_string();

    if params.project_ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:project_ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:project_ids}",
            &std::iter::repeat_n("?", params.project_ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query.bind(params.business_id);
    for value in &params.project_ids {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `GetRoadmapProjects` (Go `GetRoadmapProjectsParams`).
pub struct GetRoadmapProjectsParams {
    pub business_id: i64,
}

/// `GetRoadmapProjects`: `many`.
pub async fn get_roadmap_projects<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetRoadmapProjectsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    fp.id,\n    fp.name,\n    fp.status,\n    COALESCE(DATE_FORMAT(fp.start_date, '%Y-%m-%d'), '') AS start_date,\n    COALESCE(DATE_FORMAT(fp.due_date, '%Y-%m-%d'), '') AS due_date,\n    COUNT(ft.id) AS total_tasks,\n    COUNT(CASE WHEN ft.status = 'done' THEN 1 END) AS done_tasks\nFROM flow_projects fp\nLEFT JOIN flow_tasks ft ON ft.project_id = fp.id\n AND ft.parent_task_id IS NULL\n AND ft.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)\nWHERE fp.business_id = ? AND fp.status != 'archived'\nGROUP BY fp.id, fp.name, fp.status, fp.start_date, fp.due_date\nORDER BY fp.created_at ASC")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `GetTasksByPriority` (Go `GetTasksByPriorityParams`).
pub struct GetTasksByPriorityParams {
    pub business_id: i64,
}

/// `GetTasksByPriority`: `many`.
pub async fn get_tasks_by_priority<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTasksByPriorityParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT priority, COUNT(*) AS cnt\nFROM flow_tasks\nWHERE business_id = ?\n  AND created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)\nGROUP BY priority")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `GetTasksByStatus` (Go `GetTasksByStatusParams`).
pub struct GetTasksByStatusParams {
    pub business_id: i64,
}

/// `GetTasksByStatus`: `many`.
pub async fn get_tasks_by_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTasksByStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT status, COUNT(*) AS cnt\nFROM flow_tasks WHERE business_id = ?\n  AND created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)\nGROUP BY status")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `GetTeamWorkload` (Go `GetTeamWorkloadParams`).
pub struct GetTeamWorkloadParams {
    pub business_id: i64,
}

/// `GetTeamWorkload`: `many`.
pub async fn get_team_workload<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTeamWorkloadParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    ft.assignee_id AS user_id,\n    COUNT(*) AS total_tasks,\n    COUNT(CASE WHEN ft.status = 'done' THEN 1 END) AS completed_tasks,\n    COUNT(CASE WHEN ft.status = 'in_progress' THEN 1 END) AS in_progress_tasks,\n    COUNT(CASE WHEN ft.status NOT IN ('done','cancelled') AND ft.due_date < CURDATE() THEN 1 END) AS overdue_tasks\nFROM flow_tasks ft\nWHERE ft.business_id = ?\n  AND ft.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)\n  AND ft.assignee_id IS NOT NULL\nGROUP BY ft.assignee_id")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `GetTotalAndCompletedTasks` (Go `GetTotalAndCompletedTasksParams`).
pub struct GetTotalAndCompletedTasksParams {
    pub business_id: i64,
}

/// `GetTotalAndCompletedTasks`: `one`.
pub async fn get_total_and_completed_tasks<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTotalAndCompletedTasksParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    COUNT(*) AS total_tasks,\n    COUNT(CASE WHEN status = 'done' THEN 1 END) AS completed_tasks,\n    COUNT(CASE WHEN status NOT IN ('done','cancelled') AND due_date < CURDATE() THEN 1 END) AS overdue_count\nFROM flow_tasks WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetVelocityLastWeek` (Go `GetVelocityLastWeekParams`).
pub struct GetVelocityLastWeekParams {
    pub business_id: i64,
}

/// `GetVelocityLastWeek`: `one`.
pub async fn get_velocity_last_week<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetVelocityLastWeekParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS cnt\nFROM flow_tasks\nWHERE business_id = ? AND status = 'done'\n  AND completed_at >= DATE_SUB(DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY), INTERVAL 7 DAY)\n  AND completed_at < DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY)")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetVelocityThisWeek` (Go `GetVelocityThisWeekParams`).
pub struct GetVelocityThisWeekParams {
    pub business_id: i64,
}

/// `GetVelocityThisWeek`: `one`.
pub async fn get_velocity_this_week<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetVelocityThisWeekParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS cnt\nFROM flow_tasks\nWHERE business_id = ? AND status = 'done'\n  AND completed_at >= DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY)\n  AND completed_at < DATE_ADD(DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY), INTERVAL 7 DAY)")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}
