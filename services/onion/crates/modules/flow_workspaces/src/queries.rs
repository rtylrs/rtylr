//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `AddProjectMember` (Go `AddProjectMemberParams`).
pub struct AddProjectMemberParams {
    pub project_id: i64,
    pub user_id: i64,
    pub role: String,
}

/// `AddProjectMember`: `exec`.
pub async fn add_project_member<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AddProjectMemberParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO flow_project_members (project_id, user_id, role)\nVALUES (?, ?, ?)\nON DUPLICATE KEY UPDATE role = VALUES(role)")
        .bind(params.project_id)
        .bind(params.user_id)
        .bind(&params.role)
        .execute(exec)
        .await
}

/// Params for `AddWorkspaceMember` (Go `AddWorkspaceMemberParams`).
pub struct AddWorkspaceMemberParams {
    pub workspace_id: i64,
    pub user_id: i64,
    pub role: String,
}

/// `AddWorkspaceMember`: `exec`.
pub async fn add_workspace_member<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AddWorkspaceMemberParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO flow_workspace_members (workspace_id, user_id, role) VALUES (?, ?, ?)\nON DUPLICATE KEY UPDATE role = VALUES(role)")
        .bind(params.workspace_id)
        .bind(params.user_id)
        .bind(&params.role)
        .execute(exec)
        .await
}

/// Params for `ArchiveProject` (Go `ArchiveProjectParams`).
pub struct ArchiveProjectParams {
    pub purge_at: Option<chrono::DateTime<chrono::Utc>>,
    pub id: i64,
    pub business_id: i64,
}

/// `ArchiveProject`: `exec`.
pub async fn archive_project<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ArchiveProjectParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE flow_projects SET status = 'archived', archived_at = NOW(), purge_at = ?\nWHERE id = ? AND business_id = ?")
        .bind(params.purge_at)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ClearProjectDefaultStatus` (Go `ClearProjectDefaultStatusParams`).
pub struct ClearProjectDefaultStatusParams {
    pub business_id: i64,
    pub project_id: i64,
}

/// `ClearProjectDefaultStatus`: `execresult`.
pub async fn clear_project_default_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ClearProjectDefaultStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE flow_statuses\nSET is_default = FALSE\nWHERE business_id = ? AND project_id = ?",
    )
    .bind(params.business_id)
    .bind(params.project_id)
    .execute(exec)
    .await
}

/// Params for `ClearTasksOfStatus` (Go `ClearTasksOfStatusParams`).
pub struct ClearTasksOfStatusParams {
    pub business_id: i64,
    pub status_id: Option<i64>,
}

/// `ClearTasksOfStatus`: `execresult`.
pub async fn clear_tasks_of_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ClearTasksOfStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE flow_tasks\nSET status_id = NULL\nWHERE business_id = ? AND status_id = ?")
        .bind(params.business_id)
        .bind(params.status_id)
        .execute(exec)
        .await
}

/// Params for `ClearTasksSectionID` (Go `ClearTasksSectionIDParams`).
pub struct ClearTasksSectionIDParams {
    pub section_id: Option<i64>,
    pub user_id: i64,
    pub owner_id: i64,
    pub created_by: i64,
    pub owner_id_2: i64,
    pub user_id_2: i64,
}

/// `ClearTasksSectionID`: `exec`.
pub async fn clear_tasks_section_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ClearTasksSectionIDParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE flow_tasks\nSET section_id = NULL\nWHERE section_id = ?\n  AND (\n    EXISTS (SELECT 1 FROM flow_project_members pm\n            WHERE pm.project_id = flow_tasks.project_id AND pm.user_id = ?)\n    OR EXISTS (SELECT 1 FROM flow_projects p\n               WHERE p.id = flow_tasks.project_id\n                 AND (p.owner_id = ? OR p.created_by = ?))\n    OR EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_tasks.workspace_id AND w.owner_id = ?)\n    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_tasks.workspace_id AND wm.user_id = ?)\n  )")
        .bind(params.section_id)
        .bind(params.user_id)
        .bind(params.owner_id)
        .bind(params.created_by)
        .bind(params.owner_id_2)
        .bind(params.user_id_2)
        .execute(exec)
        .await
}

/// Params for `CountArchivedProjectsV2` (Go `CountArchivedProjectsV2Params`).
pub struct CountArchivedProjectsV2Params {
    pub business_id: i64,
    pub column2: String,
    pub created_by: i64,
    pub user_id: i64,
}

/// `CountArchivedProjectsV2`: `one`.
pub async fn count_archived_projects_v2<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountArchivedProjectsV2Params,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM flow_projects p\nWHERE p.business_id = ?\n  AND p.archived_at IS NOT NULL\n  AND (? = 1 OR p.created_by = ? OR EXISTS (\n        SELECT 1 FROM flow_project_members pm\n        WHERE pm.project_id = p.id AND pm.user_id = ?))")
        .bind(params.business_id)
        .bind(&params.column2)
        .bind(params.created_by)
        .bind(params.user_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountCycles` (Go `CountCyclesParams`).
pub struct CountCyclesParams {
    pub workspace_id: i64,
}

/// `CountCycles`: `one`.
pub async fn count_cycles<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountCyclesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM flow_cycles WHERE workspace_id = ?")
        .bind(params.workspace_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountCyclesV2` (Go `CountCyclesV2Params`).
pub struct CountCyclesV2Params {
    pub project_id: Option<i64>,
    pub business_id: i64,
    pub column3: String,
    pub created_by: i64,
    pub owner_id: i64,
    pub user_id: i64,
}

/// `CountCyclesV2`: `one`.
pub async fn count_cycles_v2<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountCyclesV2Params,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM flow_cycles c\nWHERE c.project_id = ? AND c.business_id = ?\n  AND (? = 1 OR EXISTS (SELECT 1 FROM flow_projects p\n        WHERE p.id = c.project_id\n          AND (p.created_by = ? OR p.owner_id = ?\n               OR EXISTS (SELECT 1 FROM flow_project_members pm\n                          WHERE pm.project_id = p.id AND pm.user_id = ?))))")
        .bind(params.project_id)
        .bind(params.business_id)
        .bind(&params.column3)
        .bind(params.created_by)
        .bind(params.owner_id)
        .bind(params.user_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountLabelsByWorkspace` (Go `CountLabelsByWorkspaceParams`).
pub struct CountLabelsByWorkspaceParams {
    pub workspace_id: i64,
}

/// `CountLabelsByWorkspace`: `one`.
pub async fn count_labels_by_workspace<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountLabelsByWorkspaceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM flow_labels WHERE workspace_id = ?")
        .bind(params.workspace_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountLabelsV2` (Go `CountLabelsV2Params`).
pub struct CountLabelsV2Params {
    pub business_id: i64,
}

/// `CountLabelsV2`: `one`.
pub async fn count_labels_v2<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountLabelsV2Params,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM flow_labels WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountMilestonesByProject` (Go `CountMilestonesByProjectParams`).
pub struct CountMilestonesByProjectParams {
    pub workspace_id: Option<i64>,
    pub project_id: i64,
}

/// `CountMilestonesByProject`: `one`.
pub async fn count_milestones_by_project<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountMilestonesByProjectParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM flow_milestones WHERE workspace_id = ? AND project_id = ?")
        .bind(params.workspace_id)
        .bind(params.project_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountMilestonesByWorkspace` (Go `CountMilestonesByWorkspaceParams`).
pub struct CountMilestonesByWorkspaceParams {
    pub workspace_id: Option<i64>,
}

/// `CountMilestonesByWorkspace`: `one`.
pub async fn count_milestones_by_workspace<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountMilestonesByWorkspaceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM flow_milestones WHERE workspace_id = ?")
        .bind(params.workspace_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountNotifications` (Go `CountNotificationsParams`).
pub struct CountNotificationsParams {
    pub user_id: i64,
}

/// `CountNotifications`: `one`.
pub async fn count_notifications<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountNotificationsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM flow_notifications WHERE user_id = ?")
        .bind(params.user_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountProjectColumns` (Go `CountProjectColumnsParams`).
pub struct CountProjectColumnsParams {
    pub project_id: i64,
}

/// `CountProjectColumns`: `one`.
pub async fn count_project_columns<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountProjectColumnsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM flow_project_columns WHERE project_id = ?")
        .bind(params.project_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountProjectMembers` (Go `CountProjectMembersParams`).
pub struct CountProjectMembersParams {
    pub project_id: i64,
}

/// `CountProjectMembers`: `one`.
pub async fn count_project_members<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountProjectMembersParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM flow_project_members WHERE project_id = ?")
        .bind(params.project_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountProjectsByWorkspace` (Go `CountProjectsByWorkspaceParams`).
pub struct CountProjectsByWorkspaceParams {
    pub workspace_id: i64,
    pub business_id: i64,
}

/// `CountProjectsByWorkspace`: `one`.
pub async fn count_projects_by_workspace<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountProjectsByWorkspaceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM flow_projects WHERE workspace_id = ? AND business_id = ?")
        .bind(params.workspace_id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountProjectsV2` (Go `CountProjectsV2Params`).
pub struct CountProjectsV2Params {
    pub business_id: i64,
    pub column2: String,
    pub created_by: i64,
    pub user_id: i64,
    pub managed_creators: Vec<i64>,
    pub managed_owners: Vec<i64>,
    pub managed_members: Vec<i64>,
    pub column8: String,
    pub status: String,
    pub column10: String,
    pub visibility: String,
    pub column12: String,
    pub concat: String,
}

/// `CountProjectsV2`: `one`.
pub async fn count_projects_v2<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountProjectsV2Params,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    let mut sql_template: String = "SELECT COUNT(*) FROM flow_projects p\nWHERE p.business_id = ?\n  AND p.is_personal = FALSE\n  AND p.archived_at IS NULL\n  AND (? = 1 OR p.created_by = ? OR EXISTS (\n        SELECT 1 FROM flow_project_members pm\n        WHERE pm.project_id = p.id AND pm.user_id = ?)\n    -- Dept-manager visibility proxy (M3). A caller who manages nobody is passed\n    -- a single 0: sqlc expands an empty slice into an invalid IN (), and no user\n    -- has id 0, so IN (0) matches nothing.\n    --\n    -- The surrounding WHERE already has p.is_personal = FALSE, which is what\n    -- keeps a report's My Tasks bucket out, as Project.is_personal's contract\n    -- requires.\n    OR p.created_by IN ({SLICE:managed_creators})\n    OR p.owner_id IN ({SLICE:managed_owners})\n    OR EXISTS (\n        SELECT 1 FROM flow_project_members pmr\n        WHERE pmr.project_id = p.id AND pmr.user_id IN ({SLICE:managed_members})))\n  AND (? = '' OR p.status = ?)\n  AND (? = '' OR p.visibility = ?)\n  AND (? = '' OR p.name LIKE CONCAT('%', ?, '%'))".to_string();

    if params.managed_creators.is_empty() {
        sql_template = sql_template.replace("{SLICE:managed_creators}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:managed_creators}",
            &std::iter::repeat_n("?", params.managed_creators.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.managed_owners.is_empty() {
        sql_template = sql_template.replace("{SLICE:managed_owners}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:managed_owners}",
            &std::iter::repeat_n("?", params.managed_owners.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.managed_members.is_empty() {
        sql_template = sql_template.replace("{SLICE:managed_members}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:managed_members}",
            &std::iter::repeat_n("?", params.managed_members.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query
        .bind(params.business_id)
        .bind(&params.column2)
        .bind(params.created_by)
        .bind(params.user_id)
        .bind(&params.column8)
        .bind(&params.status)
        .bind(&params.column10)
        .bind(&params.visibility)
        .bind(&params.column12)
        .bind(&params.concat);
    for value in &params.managed_creators {
        query = query.bind(value);
    }
    for value in &params.managed_owners {
        query = query.bind(value);
    }
    for value in &params.managed_members {
        query = query.bind(value);
    }
    query.fetch_optional(exec).await
}

/// Params for `CountRoadmapProjects` (Go `CountRoadmapProjectsParams`).
pub struct CountRoadmapProjectsParams {
    pub workspace_id: i64,
}

/// `CountRoadmapProjects`: `one`.
pub async fn count_roadmap_projects<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountRoadmapProjectsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT COUNT(*) FROM flow_projects WHERE workspace_id = ? AND status != 'archived'",
    )
    .bind(params.workspace_id)
    .fetch_optional(exec)
    .await
}

/// Params for `CountTeamWorkDaysByRoster` (Go `CountTeamWorkDaysByRosterParams`).
pub struct CountTeamWorkDaysByRosterParams {
    pub business_id: i64,
    pub date: chrono::NaiveDate,
    pub date_2: chrono::NaiveDate,
    pub employee_ids: Vec<i64>,
    pub hidden_ids: Vec<i64>,
    pub user_ids: Vec<i64>,
}

/// `CountTeamWorkDaysByRoster`: `one`.
pub async fn count_team_work_days_by_roster<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountTeamWorkDaysByRosterParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    let mut sql_template: String = "SELECT COUNT(*) FROM flow_work_days\nWHERE business_id = ?\n  AND date >= ? AND date <= ?\n  AND ((employee_id IN ({SLICE:employee_ids})\n        AND (employee_id <> logged_by OR employee_id NOT IN ({SLICE:hidden_ids})))\n       OR (employee_id = logged_by AND employee_id IN ({SLICE:user_ids})))".to_string();

    if params.employee_ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:employee_ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:employee_ids}",
            &std::iter::repeat_n("?", params.employee_ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.hidden_ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:hidden_ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:hidden_ids}",
            &std::iter::repeat_n("?", params.hidden_ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
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
    let mut query = query
        .bind(params.business_id)
        .bind(params.date)
        .bind(params.date_2);
    for value in &params.employee_ids {
        query = query.bind(value);
    }
    for value in &params.hidden_ids {
        query = query.bind(value);
    }
    for value in &params.user_ids {
        query = query.bind(value);
    }
    query.fetch_optional(exec).await
}

/// Params for `CountViewsByWorkspace` (Go `CountViewsByWorkspaceParams`).
pub struct CountViewsByWorkspaceParams {
    pub business_id: i64,
    pub workspace_id: i64,
}

/// `CountViewsByWorkspace`: `one`.
pub async fn count_views_by_workspace<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountViewsByWorkspaceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM flow_views WHERE business_id = ? AND workspace_id = ?")
        .bind(params.business_id)
        .bind(params.workspace_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountWorkDays` (Go `CountWorkDaysParams`).
pub struct CountWorkDaysParams {
    pub business_id: i64,
    pub employee_id: i64,
    pub date: chrono::NaiveDate,
    pub date_2: chrono::NaiveDate,
}

/// `CountWorkDays`: `one`.
pub async fn count_work_days<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountWorkDaysParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM flow_work_days\nWHERE business_id = ? AND employee_id = ?\n  AND date >= ? AND date <= ?")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.date)
        .bind(params.date_2)
        .fetch_optional(exec)
        .await
}

/// Params for `CountWorkDaysByEmployeeIDs` (Go `CountWorkDaysByEmployeeIDsParams`).
pub struct CountWorkDaysByEmployeeIDsParams {
    pub business_id: i64,
    pub employee_ids: Vec<i64>,
    pub date: chrono::NaiveDate,
    pub date_2: chrono::NaiveDate,
}

/// `CountWorkDaysByEmployeeIDs`: `one`.
pub async fn count_work_days_by_employee_i_ds<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountWorkDaysByEmployeeIDsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    let mut sql_template: String = "SELECT COUNT(*) FROM flow_work_days\nWHERE business_id = ?\n  AND employee_id IN ({SLICE:employee_ids})\n  AND date >= ? AND date <= ?".to_string();

    if params.employee_ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:employee_ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:employee_ids}",
            &std::iter::repeat_n("?", params.employee_ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query
        .bind(params.business_id)
        .bind(params.date)
        .bind(params.date_2);
    for value in &params.employee_ids {
        query = query.bind(value);
    }
    query.fetch_optional(exec).await
}

/// Params for `CountWorkTypeChildren` (Go `CountWorkTypeChildrenParams`).
pub struct CountWorkTypeChildrenParams {
    pub business_id: i64,
    pub parent_id: Option<i64>,
}

/// `CountWorkTypeChildren`: `one`.
pub async fn count_work_type_children<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountWorkTypeChildrenParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM flow_work_types\nWHERE business_id = ? AND parent_id = ?")
        .bind(params.business_id)
        .bind(params.parent_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountWorkspaceMembers` (Go `CountWorkspaceMembersParams`).
pub struct CountWorkspaceMembersParams {
    pub workspace_id: i64,
}

/// `CountWorkspaceMembers`: `one`.
pub async fn count_workspace_members<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountWorkspaceMembersParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM flow_workspace_members WHERE workspace_id = ?")
        .bind(params.workspace_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountWorkspaceMembersForWorkload` (Go `CountWorkspaceMembersForWorkloadParams`).
pub struct CountWorkspaceMembersForWorkloadParams {
    pub workspace_id: i64,
}

/// `CountWorkspaceMembersForWorkload`: `one`.
pub async fn count_workspace_members_for_workload<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountWorkspaceMembersForWorkloadParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM flow_workspace_members WHERE workspace_id = ?")
        .bind(params.workspace_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountWorkspacesByOwner` (Go `CountWorkspacesByOwnerParams`).
pub struct CountWorkspacesByOwnerParams {
    pub business_id: i64,
}

/// `CountWorkspacesByOwner`: `one`.
pub async fn count_workspaces_by_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountWorkspacesByOwnerParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS cnt FROM flow_workspaces WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateAttachment` (Go `CreateAttachmentParams`).
pub struct CreateAttachmentParams {
    pub task_id: i64,
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub url: String,
    pub size: Option<i64>,
    pub uploaded_by: i64,
}

/// `CreateAttachment`: `execresult`.
pub async fn create_attachment<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateAttachmentParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO flow_task_attachments (task_id, owner_id, business_id, name, url, size, uploaded_by)\nVALUES (?, ?, ?, ?, ?, ?, ?)")
        .bind(params.task_id)
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.url)
        .bind(params.size)
        .bind(params.uploaded_by)
        .execute(exec)
        .await
}

/// Params for `CreateColumn` (Go `CreateColumnParams`).
pub struct CreateColumnParams {
    pub project_id: i64,
    pub workspace_id: i64,
    pub name: String,
    pub color: Option<String>,
    pub sort_order: Option<i32>,
    pub created_by: i64,
}

/// `CreateColumn`: `execresult`.
pub async fn create_column<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateColumnParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO flow_project_columns (project_id, workspace_id, name, color, sort_order, created_by)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.project_id)
        .bind(params.workspace_id)
        .bind(&params.name)
        .bind(&params.color)
        .bind(params.sort_order)
        .bind(params.created_by)
        .execute(exec)
        .await
}

/// Params for `CreateComment` (Go `CreateCommentParams`).
pub struct CreateCommentParams {
    pub task_id: i64,
    pub user_id: i64,
    pub content: String,
}

/// `CreateComment`: `execresult`.
pub async fn create_comment<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateCommentParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO flow_task_comments (task_id, user_id, content) VALUES (?, ?, ?)")
        .bind(params.task_id)
        .bind(params.user_id)
        .bind(&params.content)
        .execute(exec)
        .await
}

/// Params for `CreateCycle` (Go `CreateCycleParams`).
pub struct CreateCycleParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub workspace_id: i64,
    pub project_id: Option<i64>,
    pub name: String,
    pub description: Option<String>,
    pub start_date: Option<chrono::DateTime<chrono::Utc>>,
    pub end_date: Option<chrono::DateTime<chrono::Utc>>,
    pub created_by: Option<i64>,
}

/// `CreateCycle`: `execresult`.
pub async fn create_cycle<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateCycleParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO flow_cycles (owner_id, business_id, workspace_id, project_id, name, description, start_date, end_date, created_by)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.workspace_id)
        .bind(params.project_id)
        .bind(&params.name)
        .bind(&params.description)
        .bind(params.start_date)
        .bind(params.end_date)
        .bind(params.created_by)
        .execute(exec)
        .await
}

/// Params for `CreateCycleV2` (Go `CreateCycleV2Params`).
pub struct CreateCycleV2Params {
    pub business_id: i64,
    pub project_id: Option<i64>,
    pub name: String,
    pub start_date: Option<chrono::DateTime<chrono::Utc>>,
    pub end_date: Option<chrono::DateTime<chrono::Utc>>,
    pub status: String,
}

/// `CreateCycleV2`: `execresult`.
pub async fn create_cycle_v2<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateCycleV2Params,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO flow_cycles (business_id, workspace_id, project_id, name, start_date, end_date, status)\nVALUES (?, 0, ?, ?, ?, ?, ?)")
        .bind(params.business_id)
        .bind(params.project_id)
        .bind(&params.name)
        .bind(params.start_date)
        .bind(params.end_date)
        .bind(&params.status)
        .execute(exec)
        .await
}

/// Params for `CreateLabel` (Go `CreateLabelParams`).
pub struct CreateLabelParams {
    pub workspace_id: i64,
    pub business_id: i64,
    pub name: String,
    pub color: String,
}

/// `CreateLabel`: `execresult`.
pub async fn create_label<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateLabelParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "INSERT INTO flow_labels (workspace_id, business_id, name, color) VALUES (?, ?, ?, ?)",
    )
    .bind(params.workspace_id)
    .bind(params.business_id)
    .bind(&params.name)
    .bind(&params.color)
    .execute(exec)
    .await
}

/// Params for `CreateLabelV2` (Go `CreateLabelV2Params`).
pub struct CreateLabelV2Params {
    pub business_id: i64,
    pub name: String,
    pub color: String,
}

/// `CreateLabelV2`: `execresult`.
pub async fn create_label_v2<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateLabelV2Params,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "INSERT INTO flow_labels (workspace_id, business_id, name, color) VALUES (0, ?, ?, ?)",
    )
    .bind(params.business_id)
    .bind(&params.name)
    .bind(&params.color)
    .execute(exec)
    .await
}

/// Params for `CreateMilestone` (Go `CreateMilestoneParams`).
pub struct CreateMilestoneParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub project_id: i64,
    pub workspace_id: Option<i64>,
    pub name: String,
    pub description: Option<String>,
    pub due_date: Option<chrono::DateTime<chrono::Utc>>,
    pub created_by: Option<i64>,
}

/// `CreateMilestone`: `execresult`.
pub async fn create_milestone<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateMilestoneParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO flow_milestones (owner_id, business_id, project_id, workspace_id, name, description, due_date, created_by)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.project_id)
        .bind(params.workspace_id)
        .bind(&params.name)
        .bind(&params.description)
        .bind(params.due_date)
        .bind(params.created_by)
        .execute(exec)
        .await
}

/// Params for `CreateMilestoneV2` (Go `CreateMilestoneV2Params`).
pub struct CreateMilestoneV2Params {
    pub business_id: i64,
    pub project_id: i64,
    pub name: String,
    pub description: Option<String>,
    pub target_date: Option<chrono::DateTime<chrono::Utc>>,
}

/// `CreateMilestoneV2`: `execresult`.
pub async fn create_milestone_v2<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateMilestoneV2Params,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO flow_milestones (business_id, project_id, name, description, target_date, status)\nVALUES (?, ?, ?, ?, ?, 'planning')")
        .bind(params.business_id)
        .bind(params.project_id)
        .bind(&params.name)
        .bind(&params.description)
        .bind(params.target_date)
        .execute(exec)
        .await
}

/// Params for `CreateNotification` (Go `CreateNotificationParams`).
pub struct CreateNotificationParams {
    pub user_id: i64,
    pub workspace_id: Option<i64>,
    pub r#type: String,
    pub title: String,
    pub description: Option<String>,
    pub task_id: Option<i64>,
    pub task_title: Option<String>,
    pub actor_id: Option<i64>,
    pub actor_initials: Option<String>,
    pub project: Option<String>,
    pub project_color: Option<String>,
}

/// `CreateNotification`: `exec`.
pub async fn create_notification<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateNotificationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO flow_notifications\n    (user_id, workspace_id, type, title, description, task_id, task_title, actor_id, actor_initials, project, project_color)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.user_id)
        .bind(params.workspace_id)
        .bind(&params.r#type)
        .bind(&params.title)
        .bind(&params.description)
        .bind(params.task_id)
        .bind(&params.task_title)
        .bind(params.actor_id)
        .bind(&params.actor_initials)
        .bind(&params.project)
        .bind(&params.project_color)
        .execute(exec)
        .await
}

/// Params for `CreatePersonalWorkspace` (Go `CreatePersonalWorkspaceParams`).
pub struct CreatePersonalWorkspaceParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub created_by: i64,
}

/// `CreatePersonalWorkspace`: `execresult`.
pub async fn create_personal_workspace<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreatePersonalWorkspaceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO flow_workspaces (owner_id, business_id, name, icon, color, is_default, is_personal, created_by)\nVALUES (?, ?, 'Personal', '🏠', '#64C5B9', TRUE, TRUE, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.created_by)
        .execute(exec)
        .await
}

/// Params for `CreateProject` (Go `CreateProjectParams`).
pub struct CreateProjectParams {
    pub workspace_id: i64,
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub description: Option<String>,
    pub icon: Option<String>,
    pub color: Option<String>,
    pub visibility: String,
    pub is_personal: bool,
    pub start_date: Option<chrono::DateTime<chrono::Utc>>,
    pub due_date: Option<chrono::DateTime<chrono::Utc>>,
    pub created_by: i64,
}

/// `CreateProject`: `execresult`.
pub async fn create_project<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateProjectParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO flow_projects (workspace_id, owner_id, business_id, name, description, icon, color, visibility, is_personal, start_date, due_date, created_by)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.workspace_id)
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.description)
        .bind(&params.icon)
        .bind(&params.color)
        .bind(&params.visibility)
        .bind(params.is_personal)
        .bind(params.start_date)
        .bind(params.due_date)
        .bind(params.created_by)
        .execute(exec)
        .await
}

/// Params for `CreateStatus` (Go `CreateStatusParams`).
pub struct CreateStatusParams {
    pub business_id: i64,
    pub project_id: i64,
    pub name: String,
    pub bucket: String,
    pub legacy_status: String,
    pub color: String,
    pub sort_order: i32,
    pub is_default: bool,
    pub created_by: i64,
}

/// `CreateStatus`: `execresult`.
pub async fn create_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO flow_statuses\n  (business_id, project_id, name, bucket, legacy_status, color, sort_order, is_default, created_by)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.business_id)
        .bind(params.project_id)
        .bind(&params.name)
        .bind(&params.bucket)
        .bind(&params.legacy_status)
        .bind(&params.color)
        .bind(params.sort_order)
        .bind(params.is_default)
        .bind(params.created_by)
        .execute(exec)
        .await
}

/// Params for `CreateTask` (Go `CreateTaskParams`).
pub struct CreateTaskParams {
    pub project_id: i64,
    pub workspace_id: i64,
    pub owner_id: i64,
    pub business_id: i64,
    pub parent_task_id: Option<i64>,
    pub title: String,
    pub description: Option<String>,
    pub priority: String,
    pub assignee_id: Option<i64>,
    pub reporter_id: i64,
    pub due_date: Option<chrono::DateTime<chrono::Utc>>,
    pub start_date: Option<chrono::DateTime<chrono::Utc>>,
    pub estimated_hours: Option<String>,
    pub labels: String,
    pub section_id: Option<i64>,
    pub status: String,
    pub effort_minutes: Option<i32>,
    pub initial_status: String,
}

/// `CreateTask`: `execresult`.
pub async fn create_task<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateTaskParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO flow_tasks\n(project_id, workspace_id, owner_id, business_id, parent_task_id, title, description, priority,\n assignee_id, reporter_id, due_date, start_date, estimated_hours, labels, section_id, status, effort_minutes, completed_at)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CASE WHEN ? = 'done' THEN NOW() ELSE NULL END)")
        .bind(params.project_id)
        .bind(params.workspace_id)
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.parent_task_id)
        .bind(&params.title)
        .bind(&params.description)
        .bind(&params.priority)
        .bind(params.assignee_id)
        .bind(params.reporter_id)
        .bind(params.due_date)
        .bind(params.start_date)
        .bind(&params.estimated_hours)
        .bind(&params.labels)
        .bind(params.section_id)
        .bind(&params.status)
        .bind(params.effort_minutes)
        .bind(&params.initial_status)
        .execute(exec)
        .await
}

/// Params for `CreateTimeLog` (Go `CreateTimeLogParams`).
pub struct CreateTimeLogParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub task_id: i64,
    pub user_id: i64,
    pub duration_minutes: i32,
    pub description: Option<String>,
}

/// `CreateTimeLog`: `execresult`.
pub async fn create_time_log<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateTimeLogParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO flow_time_logs (owner_id, business_id, task_id, user_id, duration_minutes, description)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.task_id)
        .bind(params.user_id)
        .bind(params.duration_minutes)
        .bind(&params.description)
        .execute(exec)
        .await
}

/// Params for `CreateView` (Go `CreateViewParams`).
pub struct CreateViewParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub workspace_id: i64,
    pub name: String,
    pub icon: Option<String>,
    pub filters: String,
    pub group_by: Option<String>,
    pub sort_by: Option<String>,
    pub sort_dir: String,
    pub visibility: String,
    pub created_by: Option<i64>,
}

/// `CreateView`: `execresult`.
pub async fn create_view<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateViewParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO flow_views (owner_id, business_id, workspace_id, name, icon, filters, group_by, sort_by, sort_dir, visibility, created_by)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.workspace_id)
        .bind(&params.name)
        .bind(&params.icon)
        .bind(&params.filters)
        .bind(&params.group_by)
        .bind(&params.sort_by)
        .bind(&params.sort_dir)
        .bind(&params.visibility)
        .bind(params.created_by)
        .execute(exec)
        .await
}

/// Params for `CreateViewV2` (Go `CreateViewV2Params`).
pub struct CreateViewV2Params {
    pub business_id: i64,
    pub name: String,
    pub icon: Option<String>,
    pub filters: String,
    pub group_by: Option<String>,
    pub sort_by: Option<String>,
    pub sort_dir: String,
    pub visibility: String,
    pub created_by: Option<i64>,
}

/// `CreateViewV2`: `execresult`.
pub async fn create_view_v2<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateViewV2Params,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO flow_views (workspace_id, business_id, name, icon, filters, group_by, sort_by, sort_dir, visibility, created_by)\nVALUES (0, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.icon)
        .bind(&params.filters)
        .bind(&params.group_by)
        .bind(&params.sort_by)
        .bind(&params.sort_dir)
        .bind(&params.visibility)
        .bind(params.created_by)
        .execute(exec)
        .await
}

/// Params for `CreateWorkType` (Go `CreateWorkTypeParams`).
pub struct CreateWorkTypeParams {
    pub business_id: i64,
    pub parent_id: Option<i64>,
    pub name: String,
    pub description: Option<String>,
    pub hr_category_id: Option<i64>,
    pub sort_order: i32,
}

/// `CreateWorkType`: `execresult`.
pub async fn create_work_type<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateWorkTypeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO flow_work_types (\n    workspace_id, business_id, parent_id, name, description, hr_category_id, sort_order\n) VALUES (0, ?, ?, ?, ?, ?, ?)")
        .bind(params.business_id)
        .bind(params.parent_id)
        .bind(&params.name)
        .bind(&params.description)
        .bind(params.hr_category_id)
        .bind(params.sort_order)
        .execute(exec)
        .await
}

/// Params for `CreateWorkspace` (Go `CreateWorkspaceParams`).
pub struct CreateWorkspaceParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub description: Option<String>,
    pub icon: Option<String>,
    pub color: Option<String>,
    pub is_default: Option<bool>,
    pub created_by: i64,
}

/// `CreateWorkspace`: `execresult`.
pub async fn create_workspace<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateWorkspaceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO flow_workspaces (owner_id, business_id, name, description, icon, color, is_default, created_by)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.description)
        .bind(&params.icon)
        .bind(&params.color)
        .bind(params.is_default)
        .bind(params.created_by)
        .execute(exec)
        .await
}

/// Params for `DeleteAttachment` (Go `DeleteAttachmentParams`).
pub struct DeleteAttachmentParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteAttachment`: `exec`.
pub async fn delete_attachment<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteAttachmentParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM flow_task_attachments WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `DeleteColumn` (Go `DeleteColumnParams`).
pub struct DeleteColumnParams {
    pub id: i64,
    pub user_id: i64,
    pub owner_id: i64,
    pub created_by: i64,
    pub owner_id_2: i64,
    pub user_id_2: i64,
}

/// `DeleteColumn`: `execresult`.
pub async fn delete_column<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteColumnParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM flow_project_columns\nWHERE flow_project_columns.id = ?\n  AND (\n    EXISTS (SELECT 1 FROM flow_project_members pm\n            WHERE pm.project_id = flow_project_columns.project_id AND pm.user_id = ?)\n    OR EXISTS (SELECT 1 FROM flow_projects p\n               WHERE p.id = flow_project_columns.project_id\n                 AND (p.owner_id = ? OR p.created_by = ?))\n    OR EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_project_columns.workspace_id AND w.owner_id = ?)\n    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_project_columns.workspace_id AND wm.user_id = ?)\n  )")
        .bind(params.id)
        .bind(params.user_id)
        .bind(params.owner_id)
        .bind(params.created_by)
        .bind(params.owner_id_2)
        .bind(params.user_id_2)
        .execute(exec)
        .await
}

/// Params for `DeleteComment` (Go `DeleteCommentParams`).
pub struct DeleteCommentParams {
    pub id: i64,
    pub user_id: i64,
}

/// `DeleteComment`: `execresult`.
pub async fn delete_comment<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteCommentParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM flow_task_comments WHERE id = ? AND user_id = ?")
        .bind(params.id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `DeleteLabel` (Go `DeleteLabelParams`).
pub struct DeleteLabelParams {
    pub id: i64,
    pub business_id: i64,
    pub owner_id: i64,
    pub user_id: i64,
}

/// `DeleteLabel`: `execresult`.
pub async fn delete_label<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteLabelParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM flow_labels\nWHERE flow_labels.id = ?\n  AND (\n    (flow_labels.business_id <> 0 AND flow_labels.business_id = ?)\n    OR EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_labels.workspace_id AND w.owner_id = ?)\n    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_labels.workspace_id AND wm.user_id = ?)\n  )")
        .bind(params.id)
        .bind(params.business_id)
        .bind(params.owner_id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `DeleteStatus` (Go `DeleteStatusParams`).
pub struct DeleteStatusParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteStatus`: `execresult`.
pub async fn delete_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM flow_statuses\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `DeleteTask` (Go `DeleteTaskParams`).
pub struct DeleteTaskParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteTask`: `exec`.
pub async fn delete_task<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteTaskParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM flow_tasks WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `DeleteTimeLog` (Go `DeleteTimeLogParams`).
pub struct DeleteTimeLogParams {
    pub id: i64,
    pub user_id: i64,
}

/// `DeleteTimeLog`: `execresult`.
pub async fn delete_time_log<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteTimeLogParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM flow_time_logs\nWHERE id = ? AND user_id = ?")
        .bind(params.id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `DeleteView` (Go `DeleteViewParams`).
pub struct DeleteViewParams {
    pub id: i64,
    pub business_id: i64,
    pub created_by: Option<i64>,
    pub owner_id: i64,
    pub user_id: i64,
}

/// `DeleteView`: `execresult`.
pub async fn delete_view<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteViewParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM flow_views\nWHERE flow_views.id = ?\n  AND flow_views.business_id = ?\n  AND (\n    flow_views.created_by = ?\n    OR EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_views.workspace_id AND w.owner_id = ?)\n    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_views.workspace_id AND wm.user_id = ?)\n  )")
        .bind(params.id)
        .bind(params.business_id)
        .bind(params.created_by)
        .bind(params.owner_id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `DeleteWorkType` (Go `DeleteWorkTypeParams`).
pub struct DeleteWorkTypeParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteWorkType`: `execresult`.
pub async fn delete_work_type<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteWorkTypeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE flow_work_types\nSET archived = TRUE\nWHERE id = ? AND business_id = ? AND archived = FALSE")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `DeleteWorkspace` (Go `DeleteWorkspaceParams`).
pub struct DeleteWorkspaceParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteWorkspace`: `exec`.
pub async fn delete_workspace<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteWorkspaceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM flow_workspaces WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `GetActivityByTask` (Go `GetActivityByTaskParams`).
pub struct GetActivityByTaskParams {
    pub task_id: i64,
    pub limit: i32,
}

/// `GetActivityByTask`: `many`.
pub async fn get_activity_by_task<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetActivityByTaskParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT a.id, a.task_id, a.user_id, t.business_id,\n       a.action, a.old_value, a.new_value, a.created_at\nFROM flow_task_activity a\nJOIN flow_tasks t ON t.id = a.task_id\nWHERE a.task_id = ? ORDER BY a.created_at DESC LIMIT ?")
        .bind(params.task_id)
        .bind(params.limit)
        .execute(exec)
        .await
}

/// Params for `GetAttachmentTaskID` (Go `GetAttachmentTaskIDParams`).
pub struct GetAttachmentTaskIDParams {
    pub id: i64,
}

/// `GetAttachmentTaskID`: `one`.
pub async fn get_attachment_task_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetAttachmentTaskIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT task_id\nFROM flow_task_attachments\nWHERE id = ?\nLIMIT 1")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetAttachmentsByTask` (Go `GetAttachmentsByTaskParams`).
pub struct GetAttachmentsByTaskParams {
    pub task_id: i64,
}

/// `GetAttachmentsByTask`: `many`.
pub async fn get_attachments_by_task<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetAttachmentsByTaskParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, task_id, owner_id, business_id, name, url, size, uploaded_by, created_at\nFROM flow_task_attachments WHERE task_id = ? ORDER BY created_at DESC, id DESC LIMIT 500")
        .bind(params.task_id)
        .execute(exec)
        .await
}

/// Params for `GetColumnProject` (Go `GetColumnProjectParams`).
pub struct GetColumnProjectParams {
    pub id: i64,
}

/// `GetColumnProject`: `one`.
pub async fn get_column_project<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetColumnProjectParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT project_id\nFROM flow_project_columns\nWHERE id = ?\nLIMIT 1")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetColumnsByProject` (Go `GetColumnsByProjectParams`).
pub struct GetColumnsByProjectParams {
    pub project_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `GetColumnsByProject`: `many`.
pub async fn get_columns_by_project<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetColumnsByProjectParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, project_id, workspace_id, name, color, sort_order, created_by, created_at, updated_at\nFROM flow_project_columns\nWHERE project_id = ?\nORDER BY sort_order ASC\nLIMIT ? OFFSET ?")
        .bind(params.project_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `GetCommentByID` (Go `GetCommentByIDParams`).
pub struct GetCommentByIDParams {
    pub id: i64,
}

/// `GetCommentByID`: `one`.
pub async fn get_comment_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetCommentByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT c.id, c.task_id, c.user_id, t.business_id,\n       c.content, c.is_edited, c.created_at, c.updated_at\nFROM flow_task_comments c\nJOIN flow_tasks t ON t.id = c.task_id\nWHERE c.id = ?\nLIMIT 1")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetCommentTaskID` (Go `GetCommentTaskIDParams`).
pub struct GetCommentTaskIDParams {
    pub id: i64,
}

/// `GetCommentTaskID`: `one`.
pub async fn get_comment_task_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetCommentTaskIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT task_id\nFROM flow_task_comments\nWHERE id = ?\nLIMIT 1")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetCommentsByTask` (Go `GetCommentsByTaskParams`).
pub struct GetCommentsByTaskParams {
    pub task_id: i64,
}

/// `GetCommentsByTask`: `many`.
pub async fn get_comments_by_task<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetCommentsByTaskParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT c.id, c.task_id, c.user_id, t.business_id,\n       c.content, c.is_edited, c.created_at, c.updated_at\nFROM flow_task_comments c\nJOIN flow_tasks t ON t.id = c.task_id\nWHERE c.task_id = ? ORDER BY c.created_at ASC, c.id ASC LIMIT 500")
        .bind(params.task_id)
        .execute(exec)
        .await
}

/// Params for `GetDashboardStats` (Go `GetDashboardStatsParams`).
pub struct GetDashboardStatsParams {
    pub completed_at: Option<chrono::DateTime<chrono::Utc>>,
    pub completed_at_2: Option<chrono::DateTime<chrono::Utc>>,
    pub workspace_id: i64,
    pub business_id: i64,
}

/// `GetDashboardStats`: `one`.
pub async fn get_dashboard_stats<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetDashboardStatsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    COUNT(CASE WHEN status = 'todo' THEN 1 END) AS open_tasks,\n    COUNT(CASE WHEN status = 'in_progress' THEN 1 END) AS in_progress_tasks,\n    COUNT(CASE WHEN status = 'done' AND completed_at >= ? AND completed_at < ? THEN 1 END) AS completed_this_week,\n    COUNT(CASE WHEN status NOT IN ('done','cancelled') AND due_date < NOW() THEN 1 END) AS overdue_tasks\nFROM flow_tasks WHERE workspace_id = ? AND business_id = ?")
        .bind(params.completed_at)
        .bind(params.completed_at_2)
        .bind(params.workspace_id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetDashboardStatsV2` (Go `GetDashboardStatsV2Params`).
pub struct GetDashboardStatsV2Params {
    pub completed_at: Option<chrono::DateTime<chrono::Utc>>,
    pub completed_at_2: Option<chrono::DateTime<chrono::Utc>>,
    pub business_id: i64,
}

/// `GetDashboardStatsV2`: `one`.
pub async fn get_dashboard_stats_v2<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetDashboardStatsV2Params,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    COUNT(CASE WHEN status = 'todo' THEN 1 END) AS open_tasks,\n    COUNT(CASE WHEN status = 'in_progress' THEN 1 END) AS in_progress_tasks,\n    COUNT(CASE WHEN status = 'done' AND completed_at >= ? AND completed_at < ? THEN 1 END) AS completed_this_week,\n    COUNT(CASE WHEN status NOT IN ('done','cancelled') AND due_date < NOW() THEN 1 END) AS overdue_tasks\nFROM flow_tasks WHERE business_id = ?")
        .bind(params.completed_at)
        .bind(params.completed_at_2)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetInsightsBurndownBase` (Go `GetInsightsBurndownBaseParams`).
pub struct GetInsightsBurndownBaseParams {
    pub workspace_id: i64,
}

/// `GetInsightsBurndownBase`: `one`.
pub async fn get_insights_burndown_base<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetInsightsBurndownBaseParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS open_tasks\nFROM flow_tasks\nWHERE workspace_id = ?\n  AND parent_task_id IS NULL\n  AND status != 'cancelled'\n  AND created_at < DATE_SUB(CURDATE(), INTERVAL 14 DAY)\n  AND (completed_at IS NULL OR completed_at >= DATE_SUB(CURDATE(), INTERVAL 14 DAY))")
        .bind(params.workspace_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetInsightsBurndownDaily` (Go `GetInsightsBurndownDailyParams`).
pub struct GetInsightsBurndownDailyParams {
    pub workspace_id: i64,
    pub workspace_id_2: i64,
}

/// `GetInsightsBurndownDaily`: `many`.
pub async fn get_insights_burndown_daily<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetInsightsBurndownDailyParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    DATE(day) AS day,\n    COUNT(CASE WHEN kind = 1 THEN 1 END) AS created,\n    COUNT(CASE WHEN kind = 2 THEN 1 END) AS completed\nFROM (\n    SELECT created_at AS day, 1 AS kind\n    FROM flow_tasks\n    WHERE flow_tasks.workspace_id = ? AND flow_tasks.parent_task_id IS NULL AND flow_tasks.status != 'cancelled'\n      AND flow_tasks.created_at >= DATE_SUB(CURDATE(), INTERVAL 14 DAY)\n    UNION ALL\n    SELECT completed_at AS day, 2 AS kind\n    FROM flow_tasks\n    WHERE flow_tasks.workspace_id = ? AND flow_tasks.parent_task_id IS NULL AND flow_tasks.status = 'done'\n      AND flow_tasks.completed_at IS NOT NULL\n      AND flow_tasks.completed_at >= DATE_SUB(CURDATE(), INTERVAL 14 DAY)\n) events\nGROUP BY DATE(day)\nORDER BY day ASC")
        .bind(params.workspace_id)
        .bind(params.workspace_id_2)
        .execute(exec)
        .await
}

/// Params for `GetInsightsCycleTime` (Go `GetInsightsCycleTimeParams`).
pub struct GetInsightsCycleTimeParams {
    pub workspace_id: i64,
}

/// `GetInsightsCycleTime`: `one`.
pub async fn get_insights_cycle_time<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetInsightsCycleTimeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT AVG(DATEDIFF(completed_at, created_at)) AS avg_days\nFROM flow_tasks\nWHERE workspace_id = ? AND status = 'done' AND completed_at IS NOT NULL")
        .bind(params.workspace_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetInsightsTaskCounts` (Go `GetInsightsTaskCountsParams`).
pub struct GetInsightsTaskCountsParams {
    pub workspace_id: i64,
}

/// `GetInsightsTaskCounts`: `one`.
pub async fn get_insights_task_counts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetInsightsTaskCountsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    COUNT(*) AS total_tasks,\n    COALESCE(SUM(status = 'done'), 0) AS done_tasks\nFROM flow_tasks\nWHERE workspace_id = ? AND parent_task_id IS NULL")
        .bind(params.workspace_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetInsightsTimeByMember` (Go `GetInsightsTimeByMemberParams`).
pub struct GetInsightsTimeByMemberParams {
    pub workspace_id: i64,
}

/// `GetInsightsTimeByMember`: `many`.
pub async fn get_insights_time_by_member<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetInsightsTimeByMemberParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    tl.user_id AS member_id,\n    CAST(COALESCE(SUM(tl.duration_minutes), 0) AS SIGNED) AS minutes\nFROM flow_time_logs tl\nJOIN flow_tasks t ON t.id = tl.task_id\nWHERE t.workspace_id = ? AND tl.time_entry_id IS NULL\nGROUP BY tl.user_id")
        .bind(params.workspace_id)
        .execute(exec)
        .await
}

/// Params for `GetInsightsVelocity` (Go `GetInsightsVelocityParams`).
pub struct GetInsightsVelocityParams {
    pub workspace_id: i64,
    pub workspace_id_2: i64,
}

/// `GetInsightsVelocity`: `many`.
pub async fn get_insights_velocity<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetInsightsVelocityParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    DATE_FORMAT(DATE_SUB(MIN(day), INTERVAL WEEKDAY(MIN(day)) DAY), '%Y-%m-%d') AS week_start,\n    COUNT(CASE WHEN kind = 2 THEN 1 END) AS completed,\n    COUNT(*) AS total\nFROM (\n    SELECT created_at AS day, 1 AS kind\n    FROM flow_tasks\n    WHERE flow_tasks.workspace_id = ? AND flow_tasks.created_at >= DATE_SUB(NOW(), INTERVAL 8 WEEK)\n    UNION ALL\n    SELECT completed_at AS day, 2 AS kind\n    FROM flow_tasks\n    WHERE flow_tasks.workspace_id = ? AND flow_tasks.status = 'done' AND flow_tasks.completed_at IS NOT NULL\n      AND completed_at >= DATE_SUB(NOW(), INTERVAL 8 WEEK)\n) events\nGROUP BY YEARWEEK(day, 1)\nORDER BY week_start ASC")
        .bind(params.workspace_id)
        .bind(params.workspace_id_2)
        .execute(exec)
        .await
}

/// Params for `GetLabelsByWorkspace` (Go `GetLabelsByWorkspaceParams`).
pub struct GetLabelsByWorkspaceParams {
    pub workspace_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `GetLabelsByWorkspace`: `many`.
pub async fn get_labels_by_workspace<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetLabelsByWorkspaceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, workspace_id, name, color, created_at\nFROM flow_labels WHERE workspace_id = ? ORDER BY name ASC\nLIMIT ? OFFSET ?")
        .bind(params.workspace_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `GetMyTasksProject` (Go `GetMyTasksProjectParams`).
pub struct GetMyTasksProjectParams {
    pub owner_id: i64,
    pub workspace_id: i64,
}

/// `GetMyTasksProject`: `one`.
pub async fn get_my_tasks_project<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetMyTasksProjectParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, name\nFROM flow_projects\nWHERE owner_id = ? AND workspace_id = ? AND name IN ('My Tasks', 'Inbox')\nORDER BY (name = 'My Tasks') DESC, created_at ASC\nLIMIT 1")
        .bind(params.owner_id)
        .bind(params.workspace_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetPersonalWorkspaceByOwner` (Go `GetPersonalWorkspaceByOwnerParams`).
pub struct GetPersonalWorkspaceByOwnerParams {
    pub owner_id: i64,
    pub business_id: i64,
}

/// `GetPersonalWorkspaceByOwner`: `one`.
pub async fn get_personal_workspace_by_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPersonalWorkspaceByOwnerParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id\nFROM flow_workspaces\nWHERE owner_id = ? AND business_id = ? AND (is_personal = TRUE OR is_default = TRUE)\nORDER BY is_personal DESC, is_default DESC, id ASC\nLIMIT 1")
        .bind(params.owner_id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetProjectMemberRole` (Go `GetProjectMemberRoleParams`).
pub struct GetProjectMemberRoleParams {
    pub project_id: i64,
    pub user_id: i64,
}

/// `GetProjectMemberRole`: `one`.
pub async fn get_project_member_role<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetProjectMemberRoleParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT role\nFROM flow_project_members\nWHERE project_id = ? AND user_id = ?\nLIMIT 1",
    )
    .bind(params.project_id)
    .bind(params.user_id)
    .fetch_optional(exec)
    .await
}

/// Params for `GetProjectOwnerWorkspace` (Go `GetProjectOwnerWorkspaceParams`).
pub struct GetProjectOwnerWorkspaceParams {
    pub id: i64,
}

/// `GetProjectOwnerWorkspace`: `one`.
pub async fn get_project_owner_workspace<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetProjectOwnerWorkspaceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT owner_id, workspace_id\nFROM flow_projects\nWHERE id = ?\nLIMIT 1")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetProjectProxyFacts` (Go `GetProjectProxyFactsParams`).
pub struct GetProjectProxyFactsParams {
    pub id: i64,
}

/// `GetProjectProxyFacts`: `one`.
pub async fn get_project_proxy_facts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetProjectProxyFactsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT p.business_id, p.owner_id, p.created_by, p.is_personal\nFROM flow_projects p\nWHERE p.id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetProjectTaskCounts` (Go `GetProjectTaskCountsParams`).
pub struct GetProjectTaskCountsParams {
    pub project_id: i64,
}

/// `GetProjectTaskCounts`: `one`.
pub async fn get_project_task_counts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetProjectTaskCountsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    COUNT(*) AS total,\n    COALESCE(SUM(status = 'done'), 0) AS done\nFROM flow_tasks\nWHERE project_id = ? AND parent_task_id IS NULL")
        .bind(params.project_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetProjectTaskCountsByProjectIDs` (Go `GetProjectTaskCountsByProjectIDsParams`).
pub struct GetProjectTaskCountsByProjectIDsParams {
    pub project_ids: Vec<i64>,
}

/// `GetProjectTaskCountsByProjectIDs`: `many`.
pub async fn get_project_task_counts_by_project_i_ds<
    'e,
    E: sqlx::Executor<'e, Database = MySql>,
>(
    exec: E,
    params: &GetProjectTaskCountsByProjectIDsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT\n    project_id,\n    COUNT(*) AS total,\n    COALESCE(SUM(status = 'done'), 0) AS done\nFROM flow_tasks\nWHERE project_id IN ({SLICE:project_ids})\n  AND parent_task_id IS NULL\nGROUP BY project_id".to_string();

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
    let mut query = query;
    for value in &params.project_ids {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `GetProjectV2ByID` (Go `GetProjectV2ByIDParams`).
pub struct GetProjectV2ByIDParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetProjectV2ByID`: `one`.
pub async fn get_project_v2by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetProjectV2ByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    p.id, p.workspace_id, p.owner_id, p.business_id, p.name,\n    COALESCE(p.description, '') AS description,\n    COALESCE(p.icon, '') AS icon,\n    COALESCE(p.color, '') AS color,\n    COALESCE(p.status, 'active') AS status,\n    COALESCE(p.visibility, 'private') AS visibility,\n    p.is_personal,\n    p.start_date, p.due_date,\n    p.created_by, p.created_at, p.updated_at,\n    p.archived_at, p.purge_at,\n    (SELECT COUNT(*) FROM flow_tasks t WHERE t.project_id = p.id) AS task_count,\n    (SELECT COUNT(*) FROM flow_tasks t WHERE t.project_id = p.id AND t.status = 'done') AS completed_count,\n    (SELECT COUNT(*) FROM flow_project_members pm WHERE pm.project_id = p.id) AS member_count\nFROM flow_projects p\nWHERE p.id = ? AND p.business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetProjectWorkspace` (Go `GetProjectWorkspaceParams`).
pub struct GetProjectWorkspaceParams {
    pub id: i64,
}

/// `GetProjectWorkspace`: `one`.
pub async fn get_project_workspace<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetProjectWorkspaceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT workspace_id\nFROM flow_projects\nWHERE id = ?\nLIMIT 1")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetProjectsByWorkspace` (Go `GetProjectsByWorkspaceParams`).
pub struct GetProjectsByWorkspaceParams {
    pub workspace_id: i64,
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `GetProjectsByWorkspace`: `many`.
pub async fn get_projects_by_workspace<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetProjectsByWorkspaceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("WITH project_page AS (\n    SELECT p.id, p.workspace_id, p.owner_id, p.business_id, p.name, p.description, p.icon, p.color, p.status,\n           p.start_date, p.due_date, p.created_by, p.created_at, p.updated_at\n    FROM flow_projects p\n    WHERE p.workspace_id = ? AND p.business_id = ?\n    ORDER BY p.created_at DESC, p.id DESC\n    LIMIT ? OFFSET ?\n),\ntask_counts AS (\n    SELECT t.project_id,\n           COUNT(*) AS task_count,\n           SUM(t.status = 'done') AS completed_count\n    FROM flow_tasks t\n    JOIN project_page p ON p.id = t.project_id\n    GROUP BY t.project_id\n),\nmember_counts AS (\n    SELECT pm.project_id, COUNT(*) AS member_count\n    FROM flow_project_members pm\n    JOIN project_page p ON p.id = pm.project_id\n    GROUP BY pm.project_id\n)\nSELECT\n    p.id, p.workspace_id, p.owner_id, p.business_id, p.name,\n    COALESCE(p.description, '') AS description,\n    COALESCE(p.icon, '') AS icon,\n    COALESCE(p.color, '') AS color,\n    COALESCE(p.status, 'active') AS status,\n    p.start_date, p.due_date,\n    p.created_by, p.created_at, p.updated_at,\n    COALESCE(tc.task_count, 0) AS task_count,\n    COALESCE(tc.completed_count, 0) AS completed_count,\n    COALESCE(mc.member_count, 0) AS member_count\nFROM project_page p\nLEFT JOIN task_counts tc ON tc.project_id = p.id\nLEFT JOIN member_counts mc ON mc.project_id = p.id\nORDER BY p.created_at DESC, p.id DESC")
        .bind(params.workspace_id)
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `GetRoadmap` (Go `GetRoadmapParams`).
pub struct GetRoadmapParams {
    pub workspace_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `GetRoadmap`: `many`.
pub async fn get_roadmap<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetRoadmapParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    p.id, p.workspace_id, p.owner_id, p.name, p.description, p.icon, p.color,\n    p.status, p.start_date, p.due_date, p.created_by, p.created_at, p.updated_at\nFROM flow_projects p\nWHERE p.workspace_id = ? AND p.status != 'archived'\nORDER BY p.start_date ASC, p.created_at ASC\nLIMIT ? OFFSET ?")
        .bind(params.workspace_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `GetStatusByID` (Go `GetStatusByIDParams`).
pub struct GetStatusByIDParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetStatusByID`: `one`.
pub async fn get_status_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetStatusByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, project_id, name, bucket, legacy_status, color, sort_order, is_default, created_by, created_at\nFROM flow_statuses\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetTaskBusinessID` (Go `GetTaskBusinessIDParams`).
pub struct GetTaskBusinessIDParams {
    pub id: i64,
}

/// `GetTaskBusinessID`: `one`.
pub async fn get_task_business_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTaskBusinessIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT w.business_id\nFROM flow_tasks t\nJOIN flow_workspaces w ON w.id = t.workspace_id\nWHERE t.id = ?\nLIMIT 1")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetTaskByID` (Go `GetTaskByIDParams`).
pub struct GetTaskByIDParams {
    pub id: i64,
}

/// `GetTaskByID`: `one`.
pub async fn get_task_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTaskByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, project_id, workspace_id, owner_id, parent_task_id, title, description,\n       status, priority, assignee_id, reporter_id, due_date, start_date, completed_at,\n       estimated_hours, actual_hours, sort_order, labels, created_at, updated_at\nFROM flow_tasks WHERE id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetTeamWorkload` (Go `GetTeamWorkloadParams`).
pub struct GetTeamWorkloadParams {
    pub workspace_id: i64,
    pub workspace_id_2: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `GetTeamWorkload`: `many`.
pub async fn get_team_workload<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTeamWorkloadParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    wm.user_id AS member_id, w.business_id,\n    COUNT(t.id) AS task_count\nFROM flow_workspace_members wm\nJOIN flow_workspaces w ON w.id = wm.workspace_id\nLEFT JOIN flow_tasks t ON t.assignee_id = wm.user_id\n    AND t.workspace_id = ?\n    AND t.status NOT IN ('done', 'cancelled')\n    AND t.parent_task_id IS NULL\nWHERE wm.workspace_id = ?\nGROUP BY wm.user_id, w.business_id\nORDER BY task_count DESC\nLIMIT ? OFFSET ?")
        .bind(params.workspace_id)
        .bind(params.workspace_id_2)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `GetTeamWorkloadMemberTasks` (Go `GetTeamWorkloadMemberTasksParams`).
pub struct GetTeamWorkloadMemberTasksParams {
    pub workspace_id: i64,
    pub assignee_id: Option<i64>,
}

/// `GetTeamWorkloadMemberTasks`: `many`.
pub async fn get_team_workload_member_tasks<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTeamWorkloadMemberTasksParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    id, project_id, workspace_id, owner_id, parent_task_id, section_id,\n    title, description, status, priority,\n    assignee_id, reporter_id, due_date, start_date, completed_at,\n    estimated_hours, actual_hours, sort_order, labels, archived,\n    created_at, updated_at\nFROM flow_tasks\nWHERE workspace_id = ? AND assignee_id = ? AND status NOT IN ('done', 'cancelled') AND parent_task_id IS NULL\nORDER BY priority DESC, due_date ASC\nLIMIT 10")
        .bind(params.workspace_id)
        .bind(params.assignee_id)
        .execute(exec)
        .await
}

/// Params for `GetTimeLogTaskID` (Go `GetTimeLogTaskIDParams`).
pub struct GetTimeLogTaskIDParams {
    pub id: i64,
}

/// `GetTimeLogTaskID`: `one`.
pub async fn get_time_log_task_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTimeLogTaskIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT task_id\nFROM flow_time_logs\nWHERE id = ?\nLIMIT 1")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetTimeLogsByTask` (Go `GetTimeLogsByTaskParams`).
pub struct GetTimeLogsByTaskParams {
    pub task_id: i64,
}

/// `GetTimeLogsByTask`: `many`.
pub async fn get_time_logs_by_task<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTimeLogsByTaskParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, task_id, user_id, duration_minutes, description, logged_at, created_at\nFROM flow_time_logs WHERE task_id = ? ORDER BY logged_at DESC, id DESC LIMIT 500")
        .bind(params.task_id)
        .execute(exec)
        .await
}

/// Params for `GetUserSettings` (Go `GetUserSettingsParams`).
pub struct GetUserSettingsParams {
    pub user_id: i64,
}

/// `GetUserSettings`: `one`.
pub async fn get_user_settings<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetUserSettingsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT task_assigned, task_comments, mentions, status_changes, weekly_digest, email_notifications, sidebar_preferences_json\nFROM flow_user_settings\nWHERE user_id = ?")
        .bind(params.user_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetUserWorkspaceRole` (Go `GetUserWorkspaceRoleParams`).
pub struct GetUserWorkspaceRoleParams {
    pub workspace_id: i64,
    pub user_id: i64,
}

/// `GetUserWorkspaceRole`: `one`.
pub async fn get_user_workspace_role<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetUserWorkspaceRoleParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT role\nFROM flow_workspace_members\nWHERE workspace_id = ? AND user_id = ?\nLIMIT 1",
    )
    .bind(params.workspace_id)
    .bind(params.user_id)
    .fetch_optional(exec)
    .await
}

/// Params for `GetWorkDay` (Go `GetWorkDayParams`).
pub struct GetWorkDayParams {
    pub business_id: i64,
    pub employee_id: i64,
    pub date: chrono::NaiveDate,
}

/// `GetWorkDay`: `one`.
pub async fn get_work_day<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetWorkDayParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, employee_id, date, timezone, work_mode, start_time,\n         end_time, break_minutes, status, logged_by, approved_by,\n         approved_at, reopened_by, reopened_at, created_at, updated_at\nFROM flow_work_days\nWHERE business_id = ? AND employee_id = ? AND date = ?\nLIMIT 1")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.date)
        .fetch_optional(exec)
        .await
}

/// Params for `GetWorkDayByID` (Go `GetWorkDayByIDParams`).
pub struct GetWorkDayByIDParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetWorkDayByID`: `one`.
pub async fn get_work_day_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetWorkDayByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, employee_id, date, timezone, work_mode, start_time,\n         end_time, break_minutes, status, logged_by, approved_by,\n         approved_at, reopened_by, reopened_at, created_at, updated_at\nFROM flow_work_days\nWHERE id = ? AND business_id = ?\nLIMIT 1")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetWorkType` (Go `GetWorkTypeParams`).
pub struct GetWorkTypeParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetWorkType`: `one`.
pub async fn get_work_type<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetWorkTypeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, parent_id, name, description, hr_category_id,\n         sort_order, archived\nFROM flow_work_types\nWHERE id = ? AND business_id = ?\nLIMIT 1")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetWorkspaceOwner` (Go `GetWorkspaceOwnerParams`).
pub struct GetWorkspaceOwnerParams {
    pub id: i64,
}

/// `GetWorkspaceOwner`: `one`.
pub async fn get_workspace_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetWorkspaceOwnerParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT owner_id\nFROM flow_workspaces\nWHERE id = ?\nLIMIT 1")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetWorkspacesByOwner` (Go `GetWorkspacesByOwnerParams`).
pub struct GetWorkspacesByOwnerParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `GetWorkspacesByOwner`: `many`.
pub async fn get_workspaces_by_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetWorkspacesByOwnerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, description, icon, color, is_default, created_by, created_at, updated_at\nFROM flow_workspaces\nWHERE business_id = ?\nORDER BY is_default DESC, created_at ASC, id ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `IsProjectMember` (Go `IsProjectMemberParams`).
pub struct IsProjectMemberParams {
    pub project_id: i64,
    pub user_id: i64,
    pub user_id_2: i64,
    pub id: i64,
    pub owner_id: i64,
}

/// `IsProjectMember`: `one`.
pub async fn is_project_member<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &IsProjectMemberParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT EXISTS(\n    SELECT 1 FROM flow_project_members pm\n    WHERE pm.project_id = ? AND pm.user_id = ?\n) OR EXISTS(\n    SELECT 1\n    FROM flow_projects p\n    LEFT JOIN flow_workspace_members wm\n      ON wm.workspace_id = p.workspace_id AND wm.user_id = ?\n    WHERE p.id = ? AND (p.owner_id = ? OR wm.user_id IS NOT NULL)\n) AS is_member")
        .bind(params.project_id)
        .bind(params.user_id)
        .bind(params.user_id_2)
        .bind(params.id)
        .bind(params.owner_id)
        .fetch_optional(exec)
        .await
}

/// Params for `IsWorkspaceMember` (Go `IsWorkspaceMemberParams`).
pub struct IsWorkspaceMemberParams {
    pub id: i64,
    pub owner_id: i64,
    pub workspace_id: i64,
    pub user_id: i64,
}

/// `IsWorkspaceMember`: `one`.
pub async fn is_workspace_member<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &IsWorkspaceMemberParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT EXISTS(\n    SELECT 1 FROM flow_workspaces w\n    WHERE w.id = ? AND w.owner_id = ?\n) OR EXISTS(\n    SELECT 1 FROM flow_workspace_members wm\n    WHERE wm.workspace_id = ? AND wm.user_id = ?\n) AS is_member")
        .bind(params.id)
        .bind(params.owner_id)
        .bind(params.workspace_id)
        .bind(params.user_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListArchivedProjectsV2` (Go `ListArchivedProjectsV2Params`).
pub struct ListArchivedProjectsV2Params {
    pub business_id: i64,
    pub column2: String,
    pub created_by: i64,
    pub user_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListArchivedProjectsV2`: `many`.
pub async fn list_archived_projects_v2<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListArchivedProjectsV2Params,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    p.id, p.workspace_id, p.owner_id, p.business_id, p.name,\n    COALESCE(p.description, '') AS description,\n    COALESCE(p.icon, '') AS icon,\n    COALESCE(p.color, '') AS color,\n    COALESCE(p.status, 'active') AS status,\n    COALESCE(p.visibility, 'private') AS visibility,\n    p.is_personal,\n    p.start_date, p.due_date,\n    p.created_by, p.created_at, p.updated_at,\n    p.archived_at, p.purge_at,\n    (SELECT COUNT(*) FROM flow_tasks t WHERE t.project_id = p.id) AS task_count,\n    (SELECT COUNT(*) FROM flow_tasks t WHERE t.project_id = p.id AND t.status = 'done') AS completed_count,\n    (SELECT COUNT(*) FROM flow_project_members pm WHERE pm.project_id = p.id) AS member_count\nFROM flow_projects p\nWHERE p.business_id = ?\n  AND p.archived_at IS NOT NULL\n  AND (? = 1 OR p.created_by = ? OR EXISTS (\n        SELECT 1 FROM flow_project_members pm\n        WHERE pm.project_id = p.id AND pm.user_id = ?))\nORDER BY p.purge_at ASC, p.id DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(&params.column2)
        .bind(params.created_by)
        .bind(params.user_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListCycles` (Go `ListCyclesParams`).
pub struct ListCyclesParams {
    pub workspace_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListCycles`: `many`.
pub async fn list_cycles<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListCyclesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    c.id, c.workspace_id, c.project_id, c.name, c.description, c.status,\n    c.start_date, c.end_date, c.created_by, c.created_at,\n    COUNT(ct.task_id) AS scope,\n    COUNT(CASE WHEN t.status = 'done' THEN 1 END) AS completed\nFROM flow_cycles c\nLEFT JOIN flow_cycle_tasks ct ON ct.cycle_id = c.id\nLEFT JOIN flow_tasks t ON t.id = ct.task_id\nWHERE c.workspace_id = ?\nGROUP BY c.id\nORDER BY c.start_date DESC\nLIMIT ? OFFSET ?")
        .bind(params.workspace_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListCyclesV2` (Go `ListCyclesV2Params`).
pub struct ListCyclesV2Params {
    pub project_id: Option<i64>,
    pub business_id: i64,
    pub column3: String,
    pub created_by: i64,
    pub owner_id: i64,
    pub user_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListCyclesV2`: `many`.
pub async fn list_cycles_v2<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListCyclesV2Params,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT c.id, c.project_id, c.name, c.description, c.status,\n       c.start_date, c.end_date, c.created_by, c.created_at,\n       (SELECT COUNT(*) FROM flow_cycle_tasks ct WHERE ct.cycle_id = c.id) AS scope,\n       (SELECT COUNT(*) FROM flow_cycle_tasks ct JOIN flow_tasks t ON t.id = ct.task_id\n        WHERE ct.cycle_id = c.id AND t.status = 'done') AS completed\nFROM flow_cycles c\nWHERE c.project_id = ? AND c.business_id = ?\n  AND (? = 1 OR EXISTS (SELECT 1 FROM flow_projects p\n        WHERE p.id = c.project_id\n          AND (p.created_by = ? OR p.owner_id = ?\n               OR EXISTS (SELECT 1 FROM flow_project_members pm\n                          WHERE pm.project_id = p.id AND pm.user_id = ?))))\nORDER BY c.start_date DESC\nLIMIT ? OFFSET ?")
        .bind(params.project_id)
        .bind(params.business_id)
        .bind(&params.column3)
        .bind(params.created_by)
        .bind(params.owner_id)
        .bind(params.user_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListInsightsTaskIDsByWorkspace` (Go `ListInsightsTaskIDsByWorkspaceParams`).
pub struct ListInsightsTaskIDsByWorkspaceParams {
    pub workspace_id: i64,
    pub id: i64,
    pub limit: i32,
}

/// `ListInsightsTaskIDsByWorkspace`: `many`.
pub async fn list_insights_task_i_ds_by_workspace<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListInsightsTaskIDsByWorkspaceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "SELECT id FROM flow_tasks\nWHERE workspace_id = ? AND id > ?\nORDER BY id ASC\nLIMIT ?",
    )
    .bind(params.workspace_id)
    .bind(params.id)
    .bind(params.limit)
    .execute(exec)
    .await
}

/// Params for `ListLabelsV2` (Go `ListLabelsV2Params`).
pub struct ListLabelsV2Params {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListLabelsV2`: `many`.
pub async fn list_labels_v2<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLabelsV2Params,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, business_id, name, color, created_at\nFROM flow_labels WHERE business_id = ?\nORDER BY name ASC LIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListMilestonesByBusiness` (Go `ListMilestonesByBusinessParams`).
pub struct ListMilestonesByBusinessParams {
    pub business_id: i64,
    pub column2: String,
    pub project_id: i64,
    pub column4: String,
    pub created_by: i64,
    pub owner_id: i64,
    pub user_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListMilestonesByBusiness`: `many`.
pub async fn list_milestones_by_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListMilestonesByBusinessParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT m.id, m.project_id, m.name,\n       m.target_date,\n       COALESCE(m.status, 'upcoming') AS status,\n       COALESCE(m.description, '') AS description,\n       p.name AS project_name, p.color AS project_color,\n       (SELECT COUNT(*) FROM flow_milestone_tasks mt WHERE mt.milestone_id = m.id) AS task_count,\n       (SELECT COUNT(*) FROM flow_milestone_tasks mt JOIN flow_tasks t ON t.id = mt.task_id\n        WHERE mt.milestone_id = m.id AND t.status = 'done') AS done_count\nFROM flow_milestones m\nLEFT JOIN flow_projects p ON p.id = m.project_id\nWHERE m.business_id = ?\n  AND (? = 0 OR m.project_id = ?)\n  AND (? = 1 OR EXISTS (SELECT 1 FROM flow_projects p\n        WHERE p.id = m.project_id\n          AND (p.created_by = ? OR p.owner_id = ?\n               OR EXISTS (SELECT 1 FROM flow_project_members pm\n                          WHERE pm.project_id = p.id AND pm.user_id = ?))))\nORDER BY m.created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(&params.column2)
        .bind(params.project_id)
        .bind(&params.column4)
        .bind(params.created_by)
        .bind(params.owner_id)
        .bind(params.user_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListMilestonesByProject` (Go `ListMilestonesByProjectParams`).
pub struct ListMilestonesByProjectParams {
    pub workspace_id: Option<i64>,
    pub project_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListMilestonesByProject`: `many`.
pub async fn list_milestones_by_project<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListMilestonesByProjectParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    m.id, m.project_id, m.workspace_id, m.name, m.description, m.status,\n    m.due_date, m.completed_at, m.created_by, m.created_at,\n    p.name AS project_name, p.color AS project_color,\n    COUNT(mt.task_id) AS task_count,\n    COUNT(CASE WHEN t.status = 'done' THEN 1 END) AS done_count\nFROM flow_milestones m\nLEFT JOIN flow_projects p ON p.id = m.project_id\nLEFT JOIN flow_milestone_tasks mt ON mt.milestone_id = m.id\nLEFT JOIN flow_tasks t ON t.id = mt.task_id\nWHERE m.workspace_id = ? AND m.project_id = ?\nGROUP BY m.id\nORDER BY m.due_date ASC\nLIMIT ? OFFSET ?")
        .bind(params.workspace_id)
        .bind(params.project_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListMilestonesByWorkspace` (Go `ListMilestonesByWorkspaceParams`).
pub struct ListMilestonesByWorkspaceParams {
    pub workspace_id: Option<i64>,
    pub limit: i32,
    pub offset: i32,
}

/// `ListMilestonesByWorkspace`: `many`.
pub async fn list_milestones_by_workspace<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListMilestonesByWorkspaceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    m.id, m.project_id, m.workspace_id, m.name, m.description, m.status,\n    m.due_date, m.completed_at, m.created_by, m.created_at,\n    p.name AS project_name, p.color AS project_color,\n    COUNT(mt.task_id) AS task_count,\n    COUNT(CASE WHEN t.status = 'done' THEN 1 END) AS done_count\nFROM flow_milestones m\nLEFT JOIN flow_projects p ON p.id = m.project_id\nLEFT JOIN flow_milestone_tasks mt ON mt.milestone_id = m.id\nLEFT JOIN flow_tasks t ON t.id = mt.task_id\nWHERE m.workspace_id = ?\nGROUP BY m.id\nORDER BY m.due_date ASC\nLIMIT ? OFFSET ?")
        .bind(params.workspace_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListNotifications` (Go `ListNotificationsParams`).
pub struct ListNotificationsParams {
    pub user_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListNotifications`: `many`.
pub async fn list_notifications<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListNotificationsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    n.id, n.user_id, n.workspace_id, n.type, n.title, COALESCE(n.description, '') AS description,\n    n.task_id, COALESCE(n.task_title, '') AS task_title, n.actor_id,\n    COALESCE(w.business_id, t.business_id, 0) AS business_id, COALESCE(n.actor_initials, '') AS actor_initials,\n    COALESCE(n.project, '') AS project, COALESCE(n.project_color, '') AS project_color,\n    n.is_read, n.created_at\nFROM flow_notifications n\nLEFT JOIN flow_workspaces w ON w.id = n.workspace_id\nLEFT JOIN flow_tasks t ON t.id = n.task_id\nWHERE n.user_id = ?\nORDER BY n.created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.user_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListProjectMembers` (Go `ListProjectMembersParams`).
pub struct ListProjectMembersParams {
    pub project_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListProjectMembers`: `many`.
pub async fn list_project_members<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListProjectMembersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    pm.id, pm.project_id, pm.user_id, pm.role, pm.joined_at,\n    w.business_id\nFROM flow_project_members pm\nJOIN flow_projects w ON w.id = pm.project_id\nWHERE pm.project_id = ?\nORDER BY pm.joined_at ASC\nLIMIT ? OFFSET ?")
        .bind(params.project_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListProjectsV2` (Go `ListProjectsV2Params`).
pub struct ListProjectsV2Params {
    pub business_id: i64,
    pub column2: String,
    pub created_by: i64,
    pub user_id: i64,
    pub managed_creators: Vec<i64>,
    pub managed_owners: Vec<i64>,
    pub managed_members: Vec<i64>,
    pub column8: String,
    pub status: String,
    pub column10: String,
    pub visibility: String,
    pub column12: String,
    pub concat: String,
    pub limit: i32,
    pub offset: i32,
}

/// `ListProjectsV2`: `many`.
pub async fn list_projects_v2<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListProjectsV2Params,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT\n    p.id, p.workspace_id, p.owner_id, p.business_id, p.name,\n    COALESCE(p.description, '') AS description,\n    COALESCE(p.icon, '') AS icon,\n    COALESCE(p.color, '') AS color,\n    COALESCE(p.status, 'active') AS status,\n    COALESCE(p.visibility, 'private') AS visibility,\n    p.is_personal,\n    p.start_date, p.due_date,\n    p.created_by, p.created_at, p.updated_at,\n    p.archived_at, p.purge_at,\n    (SELECT COUNT(*) FROM flow_tasks t WHERE t.project_id = p.id) AS task_count,\n    (SELECT COUNT(*) FROM flow_tasks t WHERE t.project_id = p.id AND t.status = 'done') AS completed_count,\n    (SELECT COUNT(*) FROM flow_project_members pm WHERE pm.project_id = p.id) AS member_count\nFROM flow_projects p\nWHERE p.business_id = ?\n  AND p.is_personal = FALSE\n  AND p.archived_at IS NULL\n  AND (? = 1 OR p.created_by = ? OR EXISTS (\n        SELECT 1 FROM flow_project_members pm\n        WHERE pm.project_id = p.id AND pm.user_id = ?)\n    -- Dept-manager visibility proxy (M3). A caller who manages nobody is passed\n    -- a single 0: sqlc expands an empty slice into an invalid IN (), and no user\n    -- has id 0, so IN (0) matches nothing.\n    --\n    -- The surrounding WHERE already has p.is_personal = FALSE, which is what\n    -- keeps a report's My Tasks bucket out, as Project.is_personal's contract\n    -- requires.\n    OR p.created_by IN ({SLICE:managed_creators})\n    OR p.owner_id IN ({SLICE:managed_owners})\n    OR EXISTS (\n        SELECT 1 FROM flow_project_members pmr\n        WHERE pmr.project_id = p.id AND pmr.user_id IN ({SLICE:managed_members})))\n  AND (? = '' OR p.status = ?)\n  AND (? = '' OR p.visibility = ?)\n  AND (? = '' OR p.name LIKE CONCAT('%', ?, '%'))\nORDER BY p.created_at DESC, p.id DESC\nLIMIT ? OFFSET ?".to_string();

    if params.managed_creators.is_empty() {
        sql_template = sql_template.replace("{SLICE:managed_creators}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:managed_creators}",
            &std::iter::repeat_n("?", params.managed_creators.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.managed_owners.is_empty() {
        sql_template = sql_template.replace("{SLICE:managed_owners}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:managed_owners}",
            &std::iter::repeat_n("?", params.managed_owners.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.managed_members.is_empty() {
        sql_template = sql_template.replace("{SLICE:managed_members}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:managed_members}",
            &std::iter::repeat_n("?", params.managed_members.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query
        .bind(params.business_id)
        .bind(&params.column2)
        .bind(params.created_by)
        .bind(params.user_id)
        .bind(&params.column8)
        .bind(&params.status)
        .bind(&params.column10)
        .bind(&params.visibility)
        .bind(&params.column12)
        .bind(&params.concat)
        .bind(params.limit)
        .bind(params.offset);
    for value in &params.managed_creators {
        query = query.bind(value);
    }
    for value in &params.managed_owners {
        query = query.bind(value);
    }
    for value in &params.managed_members {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `ListPublicProjectsByOwnerV2` (Go `ListPublicProjectsByOwnerV2Params`).
pub struct ListPublicProjectsByOwnerV2Params {
    pub business_id: i64,
    pub created_by: i64,
    pub limit: i32,
}

/// `ListPublicProjectsByOwnerV2`: `many`.
pub async fn list_public_projects_by_owner_v2<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListPublicProjectsByOwnerV2Params,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    p.id, p.name,\n    COALESCE(p.icon, '') AS icon,\n    COALESCE(p.color, '') AS color,\n    COALESCE(p.description, '') AS description,\n    COALESCE(p.status, 'active') AS status,\n    (SELECT COUNT(*) FROM flow_tasks t WHERE t.project_id = p.id) AS task_count,\n    (SELECT COUNT(*) FROM flow_tasks t WHERE t.project_id = p.id AND t.status = 'done') AS completed_count\nFROM flow_projects p\nWHERE p.business_id = ?\n  AND p.created_by = ?\n  AND p.visibility = 'public'\n  AND p.is_personal = FALSE\n  AND p.archived_at IS NULL\nORDER BY p.created_at DESC, p.id DESC\nLIMIT ?")
        .bind(params.business_id)
        .bind(params.created_by)
        .bind(params.limit)
        .execute(exec)
        .await
}

/// Params for `ListStatusesByProject` (Go `ListStatusesByProjectParams`).
pub struct ListStatusesByProjectParams {
    pub business_id: i64,
    pub project_id: i64,
}

/// `ListStatusesByProject`: `many`.
pub async fn list_statuses_by_project<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListStatusesByProjectParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, business_id, project_id, name, bucket, legacy_status, color, sort_order, is_default, created_by, created_at\nFROM flow_statuses\nWHERE business_id = ? AND project_id = ?\nORDER BY sort_order ASC, id ASC")
        .bind(params.business_id)
        .bind(params.project_id)
        .execute(exec)
        .await
}

/// Params for `ListTeamWorkDaysByRoster` (Go `ListTeamWorkDaysByRosterParams`).
pub struct ListTeamWorkDaysByRosterParams {
    pub business_id: i64,
    pub date: chrono::NaiveDate,
    pub date_2: chrono::NaiveDate,
    pub employee_ids: Vec<i64>,
    pub hidden_ids: Vec<i64>,
    pub user_ids: Vec<i64>,
    pub limit: i32,
    pub offset: i32,
}

/// `ListTeamWorkDaysByRoster`: `many`.
pub async fn list_team_work_days_by_roster<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListTeamWorkDaysByRosterParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT id, employee_id, date, timezone, work_mode, start_time,\n       end_time, break_minutes, status, logged_by, approved_by,\n       approved_at, reopened_by, reopened_at, created_at, updated_at\nFROM flow_work_days\nWHERE business_id = ?\n  AND date >= ? AND date <= ?\n  AND ((employee_id IN ({SLICE:employee_ids})\n        AND (employee_id <> logged_by OR employee_id NOT IN ({SLICE:hidden_ids})))\n       OR (employee_id = logged_by AND employee_id IN ({SLICE:user_ids})))\nORDER BY employee_id, date DESC\nLIMIT ? OFFSET ?".to_string();

    if params.employee_ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:employee_ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:employee_ids}",
            &std::iter::repeat_n("?", params.employee_ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.hidden_ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:hidden_ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:hidden_ids}",
            &std::iter::repeat_n("?", params.hidden_ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
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
    let mut query = query
        .bind(params.business_id)
        .bind(params.date)
        .bind(params.date_2)
        .bind(params.limit)
        .bind(params.offset);
    for value in &params.employee_ids {
        query = query.bind(value);
    }
    for value in &params.hidden_ids {
        query = query.bind(value);
    }
    for value in &params.user_ids {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `ListViews` (Go `ListViewsParams`).
pub struct ListViewsParams {
    pub business_id: i64,
    pub workspace_id: i64,
    pub created_by: Option<i64>,
    pub limit: i32,
    pub offset: i32,
}

/// `ListViews`: `many`.
pub async fn list_views<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListViewsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, workspace_id, name, icon, filters, group_by, sort_by, sort_dir,\n       COALESCE(visibility, 'shared') AS visibility, created_by, created_at\nFROM flow_views\nWHERE business_id = ? AND workspace_id = ?\n  AND (COALESCE(visibility, 'shared') <> 'private' OR created_by = ?)\nORDER BY created_at ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.workspace_id)
        .bind(params.created_by)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListViewsV2` (Go `ListViewsV2Params`).
pub struct ListViewsV2Params {
    pub business_id: i64,
    pub created_by: Option<i64>,
    pub limit: i32,
    pub offset: i32,
}

/// `ListViewsV2`: `many`.
pub async fn list_views_v2<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListViewsV2Params,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, business_id, name, COALESCE(icon,'') AS icon,\n       COALESCE(filters,'{}') AS filters,\n       COALESCE(group_by,'') AS group_by,\n       COALESCE(sort_by,'') AS sort_by,\n       COALESCE(sort_dir,'asc') AS sort_dir,\n       COALESCE(visibility,'shared') AS visibility\nFROM flow_views\nWHERE business_id = ?\n  -- A private view is visible only to its creator (S2). Everything else is\n  -- shared business-wide, which is how every view behaved before the column.\n  AND (COALESCE(visibility,'shared') <> 'private' OR created_by = ?)\nORDER BY created_at DESC LIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.created_by)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListWorkDays` (Go `ListWorkDaysParams`).
pub struct ListWorkDaysParams {
    pub business_id: i64,
    pub employee_id: i64,
    pub date: chrono::NaiveDate,
    pub date_2: chrono::NaiveDate,
    pub limit: i32,
    pub offset: i32,
}

/// `ListWorkDays`: `many`.
pub async fn list_work_days<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListWorkDaysParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, employee_id, date, timezone, work_mode, start_time,\n         end_time, break_minutes, status, logged_by, approved_by,\n         approved_at, reopened_by, reopened_at, created_at, updated_at\nFROM flow_work_days\nWHERE business_id = ? AND employee_id = ?\n  AND date >= ? AND date <= ?\nORDER BY date DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.date)
        .bind(params.date_2)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListWorkDaysByEmployeeIDs` (Go `ListWorkDaysByEmployeeIDsParams`).
pub struct ListWorkDaysByEmployeeIDsParams {
    pub business_id: i64,
    pub employee_ids: Vec<i64>,
    pub date: chrono::NaiveDate,
    pub date_2: chrono::NaiveDate,
    pub limit: i32,
    pub offset: i32,
}

/// `ListWorkDaysByEmployeeIDs`: `many`.
pub async fn list_work_days_by_employee_i_ds<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListWorkDaysByEmployeeIDsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT id, employee_id, date, timezone, work_mode, start_time,\n         end_time, break_minutes, status, logged_by, approved_by,\n         approved_at, reopened_by, reopened_at, created_at, updated_at\nFROM flow_work_days\nWHERE business_id = ?\n  AND employee_id IN ({SLICE:employee_ids})\n  AND date >= ? AND date <= ?\nORDER BY employee_id, date DESC\nLIMIT ? OFFSET ?".to_string();

    if params.employee_ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:employee_ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:employee_ids}",
            &std::iter::repeat_n("?", params.employee_ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query
        .bind(params.business_id)
        .bind(params.date)
        .bind(params.date_2)
        .bind(params.limit)
        .bind(params.offset);
    for value in &params.employee_ids {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `ListWorkTypes` (Go `ListWorkTypesParams`).
pub struct ListWorkTypesParams {
    pub business_id: i64,
    pub column2: String,
}

/// `ListWorkTypes`: `many`.
pub async fn list_work_types<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListWorkTypesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, parent_id, name, description, hr_category_id,\n         sort_order, archived\nFROM flow_work_types\nWHERE business_id = ?\n  AND (archived = FALSE OR ? = TRUE)\nORDER BY sort_order ASC, name ASC")
        .bind(params.business_id)
        .bind(&params.column2)
        .execute(exec)
        .await
}

/// Params for `ListWorkspaceMembers` (Go `ListWorkspaceMembersParams`).
pub struct ListWorkspaceMembersParams {
    pub workspace_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListWorkspaceMembers`: `many`.
pub async fn list_workspace_members<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListWorkspaceMembersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    wm.id, wm.workspace_id, wm.user_id, wm.role, wm.joined_at,\n    w.business_id\nFROM flow_workspace_members wm\nJOIN flow_workspaces w ON w.id = wm.workspace_id\nWHERE wm.workspace_id = ?\nORDER BY wm.joined_at ASC\nLIMIT ? OFFSET ?")
        .bind(params.workspace_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `LogActivity` (Go `LogActivityParams`).
pub struct LogActivityParams {
    pub task_id: i64,
    pub user_id: i64,
    pub action: String,
    pub old_value: Option<String>,
    pub new_value: Option<String>,
}

/// `LogActivity`: `exec`.
pub async fn log_activity<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LogActivityParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO flow_task_activity (task_id, user_id, action, old_value, new_value) VALUES (?, ?, ?, ?, ?)")
        .bind(params.task_id)
        .bind(params.user_id)
        .bind(&params.action)
        .bind(&params.old_value)
        .bind(&params.new_value)
        .execute(exec)
        .await
}

/// Params for `MarkAllNotificationsRead` (Go `MarkAllNotificationsReadParams`).
pub struct MarkAllNotificationsReadParams {
    pub user_id: i64,
}

/// `MarkAllNotificationsRead`: `exec`.
pub async fn mark_all_notifications_read<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MarkAllNotificationsReadParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE flow_notifications\nSET is_read = TRUE\nWHERE user_id = ?")
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `MarkNotificationRead` (Go `MarkNotificationReadParams`).
pub struct MarkNotificationReadParams {
    pub id: i64,
    pub user_id: i64,
}

/// `MarkNotificationRead`: `exec`.
pub async fn mark_notification_read<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MarkNotificationReadParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE flow_notifications\nSET is_read = TRUE\nWHERE id = ? AND user_id = ?")
        .bind(params.id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `MoveProject` (Go `MoveProjectParams`).
pub struct MoveProjectParams {
    pub workspace_id: i64,
    pub id: i64,
    pub owner_id: i64,
}

/// `MoveProject`: `exec`.
pub async fn move_project<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MoveProjectParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE flow_projects\nSET workspace_id = ?\nWHERE id = ? AND owner_id = ?")
        .bind(params.workspace_id)
        .bind(params.id)
        .bind(params.owner_id)
        .execute(exec)
        .await
}

/// Params for `MoveTaskToColumn` (Go `MoveTaskToColumnParams`).
pub struct MoveTaskToColumnParams {
    pub section_id: Option<i64>,
    pub id: i64,
    pub user_id: i64,
    pub owner_id: i64,
    pub created_by: i64,
    pub owner_id_2: i64,
    pub user_id_2: i64,
    pub column_id: i64,
}

/// `MoveTaskToColumn`: `execresult`.
pub async fn move_task_to_column<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MoveTaskToColumnParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE flow_tasks\nSET section_id = ?, updated_at = NOW()\nWHERE flow_tasks.id = ?\n  AND (\n    EXISTS (SELECT 1 FROM flow_project_members pm\n            WHERE pm.project_id = flow_tasks.project_id AND pm.user_id = ?)\n    OR EXISTS (SELECT 1 FROM flow_projects p\n               WHERE p.id = flow_tasks.project_id\n                 AND (p.owner_id = ? OR p.created_by = ?))\n    OR EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_tasks.workspace_id AND w.owner_id = ?)\n    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_tasks.workspace_id AND wm.user_id = ?)\n  )\n  AND (\n    ? = 0\n    OR EXISTS (\n      SELECT 1 FROM flow_project_columns c\n      WHERE c.id = ?\n        AND (\n          c.project_id = flow_tasks.project_id\n          -- Legacy workspace-era rows have no reliable project boundary. Keep\n          -- that compatibility path, but only for a real non-zero workspace;\n          -- v2 rows use workspace_id=0 and must take the project branch above.\n          OR (\n            flow_tasks.workspace_id <> 0\n            AND c.workspace_id = flow_tasks.workspace_id\n          )\n        )\n    )\n  )")
        .bind(params.section_id)
        .bind(params.id)
        .bind(params.user_id)
        .bind(params.owner_id)
        .bind(params.created_by)
        .bind(params.owner_id_2)
        .bind(params.user_id_2)
        .bind(params.column_id)
        .bind(params.column_id)
        .execute(exec)
        .await
}

/// Params for `NextStatusSortOrder` (Go `NextStatusSortOrderParams`).
pub struct NextStatusSortOrderParams {
    pub business_id: i64,
    pub project_id: i64,
}

/// `NextStatusSortOrder`: `one`.
pub async fn next_status_sort_order<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &NextStatusSortOrderParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT CAST(COALESCE(MAX(sort_order) + 1, 0) AS SIGNED) FROM flow_statuses\nWHERE business_id = ? AND project_id = ?")
        .bind(params.business_id)
        .bind(params.project_id)
        .fetch_optional(exec)
        .await
}

/// Params for `RecalcActualHours` (Go `RecalcActualHoursParams`).
pub struct RecalcActualHoursParams {
    pub actual_hours: Option<String>,
    pub id: i64,
    pub user_id: i64,
    pub owner_id: i64,
    pub created_by: i64,
    pub owner_id_2: i64,
    pub user_id_2: i64,
}

/// `RecalcActualHours`: `exec`.
pub async fn recalc_actual_hours<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RecalcActualHoursParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE flow_tasks\nSET actual_hours = ?\nWHERE flow_tasks.id = ?\n  AND (\n    EXISTS (SELECT 1 FROM flow_project_members pm\n            WHERE pm.project_id = flow_tasks.project_id AND pm.user_id = ?)\n    OR EXISTS (SELECT 1 FROM flow_projects p\n               WHERE p.id = flow_tasks.project_id\n                 AND (p.owner_id = ? OR p.created_by = ?))\n    OR EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_tasks.workspace_id AND w.owner_id = ?)\n    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_tasks.workspace_id AND wm.user_id = ?)\n  )")
        .bind(&params.actual_hours)
        .bind(params.id)
        .bind(params.user_id)
        .bind(params.owner_id)
        .bind(params.created_by)
        .bind(params.owner_id_2)
        .bind(params.user_id_2)
        .execute(exec)
        .await
}

/// Params for `RemoveProjectMember` (Go `RemoveProjectMemberParams`).
pub struct RemoveProjectMemberParams {
    pub project_id: i64,
    pub user_id: i64,
}

/// `RemoveProjectMember`: `exec`.
pub async fn remove_project_member<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RemoveProjectMemberParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM flow_project_members\nWHERE project_id = ? AND user_id = ?")
        .bind(params.project_id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `RemoveWorkspaceMember` (Go `RemoveWorkspaceMemberParams`).
pub struct RemoveWorkspaceMemberParams {
    pub workspace_id: i64,
    pub user_id: i64,
}

/// `RemoveWorkspaceMember`: `exec`.
pub async fn remove_workspace_member<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RemoveWorkspaceMemberParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM flow_workspace_members WHERE workspace_id = ? AND user_id = ?")
        .bind(params.workspace_id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `RenameProject` (Go `RenameProjectParams`).
pub struct RenameProjectParams {
    pub name: String,
    pub id: i64,
    pub owner_id: i64,
    pub workspace_id: i64,
}

/// `RenameProject`: `exec`.
pub async fn rename_project<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RenameProjectParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE flow_projects\nSET name = ?\nWHERE id = ? AND owner_id = ? AND workspace_id = ?",
    )
    .bind(&params.name)
    .bind(params.id)
    .bind(params.owner_id)
    .bind(params.workspace_id)
    .execute(exec)
    .await
}

/// Params for `RestoreProjectV2` (Go `RestoreProjectV2Params`).
pub struct RestoreProjectV2Params {
    pub id: i64,
    pub business_id: i64,
}

/// `RestoreProjectV2`: `execresult`.
pub async fn restore_project_v2<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RestoreProjectV2Params,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE flow_projects SET status = 'active', archived_at = NULL, purge_at = NULL, updated_at = NOW()\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `SeedDefaultColumn` (Go `SeedDefaultColumnParams`).
pub struct SeedDefaultColumnParams {
    pub project_id: i64,
    pub workspace_id: i64,
    pub name: String,
    pub color: Option<String>,
    pub sort_order: Option<i32>,
    pub created_by: i64,
}

/// `SeedDefaultColumn`: `execresult`.
pub async fn seed_default_column<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SeedDefaultColumnParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO flow_project_columns (project_id, workspace_id, name, color, sort_order, created_by)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.project_id)
        .bind(params.workspace_id)
        .bind(&params.name)
        .bind(&params.color)
        .bind(params.sort_order)
        .bind(params.created_by)
        .execute(exec)
        .await
}

/// Params for `SetTaskLabels` (Go `SetTaskLabelsParams`).
pub struct SetTaskLabelsParams {
    pub labels: String,
    pub id: i64,
    pub user_id: i64,
    pub owner_id: i64,
    pub created_by: i64,
    pub owner_id_2: i64,
    pub user_id_2: i64,
}

/// `SetTaskLabels`: `execresult`.
pub async fn set_task_labels<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SetTaskLabelsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE flow_tasks\nSET labels = ?\nWHERE flow_tasks.id = ?\n  AND (\n    EXISTS (SELECT 1 FROM flow_project_members pm\n            WHERE pm.project_id = flow_tasks.project_id AND pm.user_id = ?)\n    OR EXISTS (SELECT 1 FROM flow_projects p\n               WHERE p.id = flow_tasks.project_id\n                 AND (p.owner_id = ? OR p.created_by = ?))\n    OR EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_tasks.workspace_id AND w.owner_id = ?)\n    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_tasks.workspace_id AND wm.user_id = ?)\n  )")
        .bind(&params.labels)
        .bind(params.id)
        .bind(params.user_id)
        .bind(params.owner_id)
        .bind(params.created_by)
        .bind(params.owner_id_2)
        .bind(params.user_id_2)
        .execute(exec)
        .await
}

/// Params for `SetTaskStatusID` (Go `SetTaskStatusIDParams`).
pub struct SetTaskStatusIDParams {
    pub status_id: Option<i64>,
    pub status: String,
    pub id: i64,
    pub business_id: i64,
}

/// `SetTaskStatusID`: `execresult`.
pub async fn set_task_status_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SetTaskStatusIDParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE flow_tasks\nSET status_id = ?, status = ?\nWHERE id = ? AND business_id = ?",
    )
    .bind(params.status_id)
    .bind(&params.status)
    .bind(params.id)
    .bind(params.business_id)
    .execute(exec)
    .await
}

/// Params for `SumTaskTimeLogMinutes` (Go `SumTaskTimeLogMinutesParams`).
pub struct SumTaskTimeLogMinutesParams {
    pub task_id: i64,
}

/// `SumTaskTimeLogMinutes`: `one`.
pub async fn sum_task_time_log_minutes<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumTaskTimeLogMinutesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT CAST(COALESCE(SUM(duration_minutes), 0) AS SIGNED)\nFROM flow_time_logs\nWHERE task_id = ? AND time_entry_id IS NULL")
        .bind(params.task_id)
        .fetch_optional(exec)
        .await
}

/// Params for `TransferProjectOwnership` (Go `TransferProjectOwnershipParams`).
pub struct TransferProjectOwnershipParams {
    pub owner_id: i64,
    pub id: i64,
    pub owner_id_2: i64,
}

/// `TransferProjectOwnership`: `exec`.
pub async fn transfer_project_ownership<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &TransferProjectOwnershipParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE flow_projects\nSET owner_id = ?\nWHERE id = ? AND owner_id = ?")
        .bind(params.owner_id)
        .bind(params.id)
        .bind(params.owner_id_2)
        .execute(exec)
        .await
}

/// Params for `UpdateColumn` (Go `UpdateColumnParams`).
pub struct UpdateColumnParams {
    pub name: String,
    pub color: Option<String>,
    pub sort_order: Option<i32>,
    pub id: i64,
    pub user_id: i64,
    pub owner_id: i64,
    pub created_by: i64,
    pub owner_id_2: i64,
    pub user_id_2: i64,
}

/// `UpdateColumn`: `execresult`.
pub async fn update_column<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateColumnParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE flow_project_columns\nSET name = ?, color = ?, sort_order = ?, updated_at = NOW()\nWHERE flow_project_columns.id = ?\n  AND (\n    EXISTS (SELECT 1 FROM flow_project_members pm\n            WHERE pm.project_id = flow_project_columns.project_id AND pm.user_id = ?)\n    OR EXISTS (SELECT 1 FROM flow_projects p\n               WHERE p.id = flow_project_columns.project_id\n                 AND (p.owner_id = ? OR p.created_by = ?))\n    OR EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_project_columns.workspace_id AND w.owner_id = ?)\n    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_project_columns.workspace_id AND wm.user_id = ?)\n  )")
        .bind(&params.name)
        .bind(&params.color)
        .bind(params.sort_order)
        .bind(params.id)
        .bind(params.user_id)
        .bind(params.owner_id)
        .bind(params.created_by)
        .bind(params.owner_id_2)
        .bind(params.user_id_2)
        .execute(exec)
        .await
}

/// Params for `UpdateLabel` (Go `UpdateLabelParams`).
pub struct UpdateLabelParams {
    pub name: String,
    pub color: String,
    pub id: i64,
    pub business_id: i64,
    pub owner_id: i64,
    pub user_id: i64,
}

/// `UpdateLabel`: `execresult`.
pub async fn update_label<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateLabelParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE flow_labels\nSET name = ?, color = ?\nWHERE flow_labels.id = ?\n  AND (\n    (flow_labels.business_id <> 0 AND flow_labels.business_id = ?)\n    OR EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_labels.workspace_id AND w.owner_id = ?)\n    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_labels.workspace_id AND wm.user_id = ?)\n  )")
        .bind(&params.name)
        .bind(&params.color)
        .bind(params.id)
        .bind(params.business_id)
        .bind(params.owner_id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `UpdateProject` (Go `UpdateProjectParams`).
pub struct UpdateProjectParams {
    pub name: String,
    pub description: Option<String>,
    pub icon: Option<String>,
    pub color: Option<String>,
    pub status: String,
    pub start_date: Option<chrono::DateTime<chrono::Utc>>,
    pub due_date: Option<chrono::DateTime<chrono::Utc>>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateProject`: `exec`.
pub async fn update_project<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateProjectParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE flow_projects SET name = ?, description = ?, icon = ?, color = ?, status = ?, start_date = ?, due_date = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(&params.description)
        .bind(&params.icon)
        .bind(&params.color)
        .bind(&params.status)
        .bind(params.start_date)
        .bind(params.due_date)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateProjectMemberRole` (Go `UpdateProjectMemberRoleParams`).
pub struct UpdateProjectMemberRoleParams {
    pub role: String,
    pub project_id: i64,
    pub user_id: i64,
}

/// `UpdateProjectMemberRole`: `exec`.
pub async fn update_project_member_role<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateProjectMemberRoleParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE flow_project_members\nSET role = ?\nWHERE project_id = ? AND user_id = ?")
        .bind(&params.role)
        .bind(params.project_id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `UpdateProjectVisibility` (Go `UpdateProjectVisibilityParams`).
pub struct UpdateProjectVisibilityParams {
    pub visibility: String,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateProjectVisibility`: `execresult`.
pub async fn update_project_visibility<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateProjectVisibilityParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE flow_projects SET visibility = ?, updated_at = NOW()\nWHERE id = ? AND business_id = ?")
        .bind(&params.visibility)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateStatus` (Go `UpdateStatusParams`).
pub struct UpdateStatusParams {
    pub name: String,
    pub bucket: String,
    pub legacy_status: String,
    pub color: String,
    pub sort_order: i32,
    pub is_default: bool,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateStatus`: `execresult`.
pub async fn update_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE flow_statuses\nSET name = ?, bucket = ?, legacy_status = ?, color = ?, sort_order = ?, is_default = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(&params.bucket)
        .bind(&params.legacy_status)
        .bind(&params.color)
        .bind(params.sort_order)
        .bind(params.is_default)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateTask` (Go `UpdateTaskParams`).
pub struct UpdateTaskParams {
    pub title: String,
    pub description: Option<String>,
    pub status: String,
    pub priority: String,
    pub assignee_id: Option<i64>,
    pub due_date: Option<chrono::DateTime<chrono::Utc>>,
    pub start_date: Option<chrono::DateTime<chrono::Utc>>,
    pub labels: String,
    pub column9: String,
    pub column10: String,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateTask`: `exec`.
pub async fn update_task<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateTaskParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE flow_tasks\nSET title = ?, description = ?, status = ?, priority = ?, assignee_id = ?,\n    due_date = ?, start_date = ?, labels = ?,\n    completed_at = CASE WHEN ? = 'done' AND completed_at IS NULL THEN NOW() WHEN ? != 'done' THEN NULL ELSE completed_at END\nWHERE id = ? AND business_id = ?")
        .bind(&params.title)
        .bind(&params.description)
        .bind(&params.status)
        .bind(&params.priority)
        .bind(params.assignee_id)
        .bind(params.due_date)
        .bind(params.start_date)
        .bind(&params.labels)
        .bind(&params.column9)
        .bind(&params.column10)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateTaskStatus` (Go `UpdateTaskStatusParams`).
pub struct UpdateTaskStatusParams {
    pub status: String,
    pub column2: String,
    pub column3: String,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateTaskStatus`: `exec`.
pub async fn update_task_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateTaskStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE flow_tasks\nSET status = ?,\n    completed_at = CASE WHEN ? = 'done' AND completed_at IS NULL THEN NOW() WHEN ? != 'done' THEN NULL ELSE completed_at END\nWHERE id = ? AND business_id = ?")
        .bind(&params.status)
        .bind(&params.column2)
        .bind(&params.column3)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateView` (Go `UpdateViewParams`).
pub struct UpdateViewParams {
    pub name: String,
    pub icon: Option<String>,
    pub filters: String,
    pub group_by: Option<String>,
    pub sort_by: Option<String>,
    pub sort_dir: String,
    pub visibility: String,
    pub id: i64,
    pub business_id: i64,
    pub created_by: Option<i64>,
    pub owner_id: i64,
    pub user_id: i64,
}

/// `UpdateView`: `execresult`.
pub async fn update_view<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateViewParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE flow_views\nSET name = ?, icon = ?, filters = ?, group_by = ?, sort_by = ?, sort_dir = ?, visibility = ?\nWHERE flow_views.id = ?\n  AND flow_views.business_id = ?\n  AND (\n    flow_views.created_by = ?\n    OR EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_views.workspace_id AND w.owner_id = ?)\n    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_views.workspace_id AND wm.user_id = ?)\n  )")
        .bind(&params.name)
        .bind(&params.icon)
        .bind(&params.filters)
        .bind(&params.group_by)
        .bind(&params.sort_by)
        .bind(&params.sort_dir)
        .bind(&params.visibility)
        .bind(params.id)
        .bind(params.business_id)
        .bind(params.created_by)
        .bind(params.owner_id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `UpdateWorkType` (Go `UpdateWorkTypeParams`).
pub struct UpdateWorkTypeParams {
    pub name: String,
    pub description: Option<String>,
    pub parent_id: Option<i64>,
    pub hr_category_id: Option<i64>,
    pub sort_order: i32,
    pub archived: bool,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateWorkType`: `execresult`.
pub async fn update_work_type<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateWorkTypeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE flow_work_types\nSET name = ?, description = ?, parent_id = ?, hr_category_id = ?, sort_order = ?, archived = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(&params.description)
        .bind(params.parent_id)
        .bind(params.hr_category_id)
        .bind(params.sort_order)
        .bind(params.archived)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateWorkspace` (Go `UpdateWorkspaceParams`).
pub struct UpdateWorkspaceParams {
    pub name: String,
    pub description: Option<String>,
    pub icon: Option<String>,
    pub color: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateWorkspace`: `exec`.
pub async fn update_workspace<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateWorkspaceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE flow_workspaces SET name = ?, description = ?, icon = ?, color = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(&params.description)
        .bind(&params.icon)
        .bind(&params.color)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateWorkspaceMemberRole` (Go `UpdateWorkspaceMemberRoleParams`).
pub struct UpdateWorkspaceMemberRoleParams {
    pub role: String,
    pub workspace_id: i64,
    pub user_id: i64,
}

/// `UpdateWorkspaceMemberRole`: `exec`.
pub async fn update_workspace_member_role<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateWorkspaceMemberRoleParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE flow_workspace_members\nSET role = ?\nWHERE workspace_id = ? AND user_id = ?",
    )
    .bind(&params.role)
    .bind(params.workspace_id)
    .bind(params.user_id)
    .execute(exec)
    .await
}

/// Params for `UpsertUserSettings` (Go `UpsertUserSettingsParams`).
pub struct UpsertUserSettingsParams {
    pub user_id: i64,
    pub task_assigned: Option<bool>,
    pub task_comments: Option<bool>,
    pub mentions: Option<bool>,
    pub status_changes: Option<bool>,
    pub weekly_digest: Option<bool>,
    pub email_notifications: Option<bool>,
    pub sidebar_preferences_json: Option<String>,
}

/// `UpsertUserSettings`: `exec`.
pub async fn upsert_user_settings<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertUserSettingsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO flow_user_settings (user_id, task_assigned, task_comments, mentions, status_changes, weekly_digest, email_notifications, sidebar_preferences_json)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?)\nON DUPLICATE KEY UPDATE\n    task_assigned = VALUES(task_assigned),\n    task_comments = VALUES(task_comments),\n    mentions = VALUES(mentions),\n    status_changes = VALUES(status_changes),\n    weekly_digest = VALUES(weekly_digest),\n    email_notifications = VALUES(email_notifications),\n    sidebar_preferences_json = VALUES(sidebar_preferences_json)")
        .bind(params.user_id)
        .bind(params.task_assigned)
        .bind(params.task_comments)
        .bind(params.mentions)
        .bind(params.status_changes)
        .bind(params.weekly_digest)
        .bind(params.email_notifications)
        .bind(&params.sidebar_preferences_json)
        .execute(exec)
        .await
}

/// Params for `UpsertWorkDay` (Go `UpsertWorkDayParams`).
pub struct UpsertWorkDayParams {
    pub business_id: i64,
    pub employee_id: i64,
    pub date: chrono::NaiveDate,
    pub timezone: String,
    pub work_mode: String,
    pub start_time: Option<String>,
    pub end_time: Option<String>,
    pub break_minutes: i32,
    pub logged_by: Option<i64>,
}

/// `UpsertWorkDay`: `execresult`.
pub async fn upsert_work_day<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertWorkDayParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO flow_work_days (\n    business_id, employee_id, date, timezone,\n    work_mode, start_time, end_time, break_minutes, logged_by\n) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)\nON DUPLICATE KEY UPDATE\n    timezone = VALUES(timezone),\n    work_mode = VALUES(work_mode),\n    start_time = VALUES(start_time),\n    end_time = VALUES(end_time),\n    break_minutes = VALUES(break_minutes),\n    updated_at = CURRENT_TIMESTAMP")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.date)
        .bind(&params.timezone)
        .bind(&params.work_mode)
        .bind(&params.start_time)
        .bind(&params.end_time)
        .bind(params.break_minutes)
        .bind(params.logged_by)
        .execute(exec)
        .await
}
