//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountBenefitEnrollmentsByEmployee` (Go `CountBenefitEnrollmentsByEmployeeParams`).
pub struct CountBenefitEnrollmentsByEmployeeParams {
    pub business_id: i64,
    pub employee_id: i64,
}

/// `CountBenefitEnrollmentsByEmployee`: `one`.
pub async fn count_benefit_enrollments_by_employee<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountBenefitEnrollmentsByEmployeeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT COUNT(*)\nFROM hr_benefit_enrollments\nWHERE business_id = ? AND employee_id = ?",
    )
    .bind(params.business_id)
    .bind(params.employee_id)
    .fetch_optional(exec)
    .await
}

/// Params for `CreateBenefitEnrollment` (Go `CreateBenefitEnrollmentParams`).
pub struct CreateBenefitEnrollmentParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub employee_id: i64,
    pub plan_id: i64,
}

/// `CreateBenefitEnrollment`: `execresult`.
pub async fn create_benefit_enrollment<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateBenefitEnrollmentParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_benefit_enrollments (owner_id, business_id, employee_id, plan_id, status)\nVALUES (?, ?, ?, ?, 'active')")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.plan_id)
        .execute(exec)
        .await
}

/// Params for `CreateBenefitPlan` (Go `CreateBenefitPlanParams`).
pub struct CreateBenefitPlanParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub r#type: String,
    pub provider: Option<String>,
    pub cost_employer: Option<String>,
    pub cost_employee: Option<String>,
    pub is_active: Option<bool>,
}

/// `CreateBenefitPlan`: `execresult`.
pub async fn create_benefit_plan<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateBenefitPlanParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_benefit_plans (owner_id, business_id, name, type, provider, cost_employer, cost_employee, is_active)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.r#type)
        .bind(&params.provider)
        .bind(&params.cost_employer)
        .bind(&params.cost_employee)
        .bind(params.is_active)
        .execute(exec)
        .await
}

/// Params for `GetBenefitEnrollmentByID` (Go `GetBenefitEnrollmentByIDParams`).
pub struct GetBenefitEnrollmentByIDParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetBenefitEnrollmentByID`: `one`.
pub async fn get_benefit_enrollment_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetBenefitEnrollmentByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, plan_id, status, enrolled_at, cancelled_at\nFROM hr_benefit_enrollments\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetBenefitPlanIDForBusiness` (Go `GetBenefitPlanIDForBusinessParams`).
pub struct GetBenefitPlanIDForBusinessParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetBenefitPlanIDForBusiness`: `one`.
pub async fn get_benefit_plan_id_for_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetBenefitPlanIDForBusinessParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id\nFROM hr_benefit_plans\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListBenefitEnrollmentsByEmployee` (Go `ListBenefitEnrollmentsByEmployeeParams`).
pub struct ListBenefitEnrollmentsByEmployeeParams {
    pub business_id: i64,
    pub employee_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListBenefitEnrollmentsByEmployee`: `many`.
pub async fn list_benefit_enrollments_by_employee<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListBenefitEnrollmentsByEmployeeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, employee_id, plan_id, status, enrolled_at, cancelled_at\nFROM hr_benefit_enrollments\nWHERE business_id = ? AND employee_id = ?\nORDER BY enrolled_at DESC, id DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.employee_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListBenefitPlans` (Go `ListBenefitPlansParams`).
pub struct ListBenefitPlansParams {
    pub business_id: i64,
}

/// `ListBenefitPlans`: `many`.
pub async fn list_benefit_plans<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListBenefitPlansParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, type, provider, cost_employer, cost_employee, is_active, created_at\nFROM hr_benefit_plans\nWHERE business_id = ?\nORDER BY name ASC\nLIMIT 100")
        .bind(params.business_id)
        .execute(exec)
        .await
}
