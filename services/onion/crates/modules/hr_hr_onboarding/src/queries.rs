//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CompleteEmployeeChecklistItem` (Go `CompleteEmployeeChecklistItemParams`).
pub struct CompleteEmployeeChecklistItemParams {
    pub completed_by: Option<i64>,
    pub attachment_url: Option<String>,
    pub business_id: i64,
    pub id: i64,
}

/// `CompleteEmployeeChecklistItem`: `execresult`.
pub async fn complete_employee_checklist_item<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CompleteEmployeeChecklistItemParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_onboarding_employee_checklist_items\nSET status = 'complete',\n    completed_at = NOW(),\n    completed_by = ?,\n    attachment_url = COALESCE(?, attachment_url)\nWHERE business_id = ?\n  AND id = ?\n  AND status NOT IN ('complete', 'completed')")
        .bind(params.completed_by)
        .bind(&params.attachment_url)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `CountChecklistItemsCompleted` (Go `CountChecklistItemsCompletedParams`).
pub struct CountChecklistItemsCompletedParams {
    pub business_id: i64,
    pub employee_checklist_id: i64,
}

/// `CountChecklistItemsCompleted`: `one`.
pub async fn count_checklist_items_completed<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountChecklistItemsCompletedParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total\nFROM hr_onboarding_employee_checklist_items\nWHERE business_id = ? AND employee_checklist_id = ? AND status = 'complete'")
        .bind(params.business_id)
        .bind(params.employee_checklist_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountChecklistItemsTotal` (Go `CountChecklistItemsTotalParams`).
pub struct CountChecklistItemsTotalParams {
    pub business_id: i64,
    pub employee_checklist_id: i64,
}

/// `CountChecklistItemsTotal`: `one`.
pub async fn count_checklist_items_total<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountChecklistItemsTotalParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total\nFROM hr_onboarding_employee_checklist_items\nWHERE business_id = ? AND employee_checklist_id = ?")
        .bind(params.business_id)
        .bind(params.employee_checklist_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountChecklistTemplates` (Go `CountChecklistTemplatesParams`).
pub struct CountChecklistTemplatesParams {
    pub business_id: i64,
    pub trigger_type: Option<String>,
}

/// `CountChecklistTemplates`: `one`.
pub async fn count_checklist_templates<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountChecklistTemplatesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total\nFROM hr_onboarding_checklist_templates\nWHERE business_id = ?\n  AND (? IS NULL OR trigger_type = ?)")
        .bind(params.business_id)
        .bind(&params.trigger_type)
        .bind(&params.trigger_type)
        .fetch_optional(exec)
        .await
}

/// Params for `CountEmployeeChecklists` (Go `CountEmployeeChecklistsParams`).
pub struct CountEmployeeChecklistsParams {
    pub business_id: i64,
    pub employee_id: i64,
}

/// `CountEmployeeChecklists`: `one`.
pub async fn count_employee_checklists<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountEmployeeChecklistsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total\nFROM hr_onboarding_employee_checklists\nWHERE business_id = ? AND employee_id = ?")
        .bind(params.business_id)
        .bind(params.employee_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateChecklistItem` (Go `CreateChecklistItemParams`).
pub struct CreateChecklistItemParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub template_id: i64,
    pub title: String,
    pub description: Option<String>,
    pub assignee_role: String,
    pub due_offset_days: i32,
    pub requires_attachment: bool,
}

/// `CreateChecklistItem`: `execresult`.
pub async fn create_checklist_item<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateChecklistItemParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_onboarding_checklist_items (owner_id, business_id, template_id, title, description, assignee_role, due_offset_days, requires_attachment)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.template_id)
        .bind(&params.title)
        .bind(&params.description)
        .bind(&params.assignee_role)
        .bind(params.due_offset_days)
        .bind(params.requires_attachment)
        .execute(exec)
        .await
}

/// Params for `CreateChecklistTemplate` (Go `CreateChecklistTemplateParams`).
pub struct CreateChecklistTemplateParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub trigger_type: String,
    pub role_id: Option<i64>,
}

/// `CreateChecklistTemplate`: `execresult`.
pub async fn create_checklist_template<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateChecklistTemplateParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_onboarding_checklist_templates (owner_id, business_id, name, trigger_type, role_id)\nVALUES (?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.trigger_type)
        .bind(params.role_id)
        .execute(exec)
        .await
}

/// Params for `CreateEmployeeChecklist` (Go `CreateEmployeeChecklistParams`).
pub struct CreateEmployeeChecklistParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub employee_id: i64,
    pub template_id: i64,
    pub status: String,
    pub triggered_at: chrono::NaiveDate,
}

/// `CreateEmployeeChecklist`: `execresult`.
pub async fn create_employee_checklist<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateEmployeeChecklistParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_onboarding_employee_checklists (owner_id, business_id, employee_id, template_id, status, triggered_at)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.template_id)
        .bind(&params.status)
        .bind(params.triggered_at)
        .execute(exec)
        .await
}

/// Params for `CreateEmployeeChecklistItem` (Go `CreateEmployeeChecklistItemParams`).
pub struct CreateEmployeeChecklistItemParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub employee_checklist_id: i64,
    pub checklist_item_id: i64,
    pub assigned_to_employee_id: i64,
    pub due_date: chrono::NaiveDate,
    pub status: String,
}

/// `CreateEmployeeChecklistItem`: `exec`.
pub async fn create_employee_checklist_item<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateEmployeeChecklistItemParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_onboarding_employee_checklist_items (owner_id, business_id, employee_checklist_id, checklist_item_id, assigned_to_employee_id, due_date, status)\nVALUES (?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.employee_checklist_id)
        .bind(params.checklist_item_id)
        .bind(params.assigned_to_employee_id)
        .bind(params.due_date)
        .bind(&params.status)
        .execute(exec)
        .await
}

/// Params for `GetChecklistItem` (Go `GetChecklistItemParams`).
pub struct GetChecklistItemParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetChecklistItem`: `one`.
pub async fn get_checklist_item<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetChecklistItemParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, template_id, title, description, assignee_role, due_offset_days, requires_attachment, created_at, updated_at\nFROM hr_onboarding_checklist_items\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetChecklistTemplate` (Go `GetChecklistTemplateParams`).
pub struct GetChecklistTemplateParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetChecklistTemplate`: `one`.
pub async fn get_checklist_template<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetChecklistTemplateParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, trigger_type, role_id, created_at, updated_at\nFROM hr_onboarding_checklist_templates\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetEmployeeChecklist` (Go `GetEmployeeChecklistParams`).
pub struct GetEmployeeChecklistParams {
    pub business_id: i64,
    pub employee_id: i64,
    pub id: i64,
}

/// `GetEmployeeChecklist`: `one`.
pub async fn get_employee_checklist<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetEmployeeChecklistParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, template_id, status, triggered_at, created_at, updated_at\nFROM hr_onboarding_employee_checklists\nWHERE business_id = ? AND employee_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetEmployeeChecklistItem` (Go `GetEmployeeChecklistItemParams`).
pub struct GetEmployeeChecklistItemParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetEmployeeChecklistItem`: `one`.
pub async fn get_employee_checklist_item<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetEmployeeChecklistItemParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_checklist_id, checklist_item_id, assigned_to_employee_id, due_date, status, completed_at, completed_by, attachment_url, created_at, updated_at\nFROM hr_onboarding_employee_checklist_items\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetLatestEmployeeChecklist` (Go `GetLatestEmployeeChecklistParams`).
pub struct GetLatestEmployeeChecklistParams {
    pub business_id: i64,
    pub employee_id: i64,
}

/// `GetLatestEmployeeChecklist`: `one`.
pub async fn get_latest_employee_checklist<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetLatestEmployeeChecklistParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, template_id, status, triggered_at, created_at, updated_at\nFROM hr_onboarding_employee_checklists\nWHERE business_id = ? AND employee_id = ?\nORDER BY triggered_at DESC, id DESC\nLIMIT 1")
        .bind(params.business_id)
        .bind(params.employee_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListChecklistTemplates` (Go `ListChecklistTemplatesParams`).
pub struct ListChecklistTemplatesParams {
    pub business_id: i64,
    pub trigger_type: Option<String>,
    pub limit: i32,
    pub offset: i32,
}

/// `ListChecklistTemplates`: `many`.
pub async fn list_checklist_templates<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListChecklistTemplatesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, trigger_type, role_id, created_at, updated_at\nFROM hr_onboarding_checklist_templates\nWHERE business_id = ?\n  AND (? IS NULL OR trigger_type = ?)\nORDER BY name ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(&params.trigger_type)
        .bind(&params.trigger_type)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListEmployeeChecklistItems` (Go `ListEmployeeChecklistItemsParams`).
pub struct ListEmployeeChecklistItemsParams {
    pub business_id: i64,
    pub employee_checklist_id: i64,
}

/// `ListEmployeeChecklistItems`: `many`.
pub async fn list_employee_checklist_items<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListEmployeeChecklistItemsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_checklist_id, checklist_item_id, assigned_to_employee_id, due_date, status, completed_at, completed_by, attachment_url, created_at, updated_at\nFROM hr_onboarding_employee_checklist_items\nWHERE business_id = ? AND employee_checklist_id = ?\nORDER BY due_date ASC\nLIMIT 500")
        .bind(params.business_id)
        .bind(params.employee_checklist_id)
        .execute(exec)
        .await
}

/// Params for `ListEmployeeChecklists` (Go `ListEmployeeChecklistsParams`).
pub struct ListEmployeeChecklistsParams {
    pub business_id: i64,
    pub employee_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListEmployeeChecklists`: `many`.
pub async fn list_employee_checklists<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListEmployeeChecklistsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, template_id, status, triggered_at, created_at, updated_at\nFROM hr_onboarding_employee_checklists\nWHERE business_id = ? AND employee_id = ?\nORDER BY triggered_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `UpdateChecklistTemplate` (Go `UpdateChecklistTemplateParams`).
pub struct UpdateChecklistTemplateParams {
    pub name: String,
    pub trigger_type: Option<String>,
    pub role_id: Option<i64>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateChecklistTemplate`: `execresult`.
pub async fn update_checklist_template<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateChecklistTemplateParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_onboarding_checklist_templates\nSET name = ?,\n    trigger_type = COALESCE(?, trigger_type),\n    role_id = COALESCE(?, role_id)\nWHERE business_id = ? AND id = ?")
        .bind(&params.name)
        .bind(&params.trigger_type)
        .bind(params.role_id)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpdateEmployeeChecklistStatus` (Go `UpdateEmployeeChecklistStatusParams`).
pub struct UpdateEmployeeChecklistStatusParams {
    pub status: String,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateEmployeeChecklistStatus`: `exec`.
pub async fn update_employee_checklist_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateEmployeeChecklistStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE hr_onboarding_employee_checklists\nSET status = ?\nWHERE business_id = ? AND id = ?")
        .bind(&params.status)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
