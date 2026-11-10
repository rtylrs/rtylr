//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `AddImportBatchCounts` (Go `AddImportBatchCountsParams`).
pub struct AddImportBatchCountsParams {
    pub created_count: i32,
    pub updated_count: i32,
    pub skipped_count: i32,
    pub failed_count: i32,
    pub id: i64,
    pub business_id: i64,
}

/// `AddImportBatchCounts`: `exec`.
pub async fn add_import_batch_counts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AddImportBatchCountsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_import_batches\nSET created_count = created_count + ?, updated_count = updated_count + ?,\n    skipped_count = skipped_count + ?, failed_count = failed_count + ?\nWHERE id = ? AND business_id = ?")
        .bind(params.created_count)
        .bind(params.updated_count)
        .bind(params.skipped_count)
        .bind(params.failed_count)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `AssignLeads` (Go `AssignLeadsParams`).
pub struct AssignLeadsParams {
    pub assigned_to: Option<i64>,
    pub business_id: i64,
    pub ids: Vec<i64>,
}

/// `AssignLeads`: `execresult`.
pub async fn assign_leads<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AssignLeadsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String =
        "UPDATE crm_leads SET assigned_to = ?\nWHERE business_id = ? AND id IN ({SLICE:ids})"
            .to_string();

    if params.ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:ids}",
            &std::iter::repeat_n("?", params.ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query.bind(params.assigned_to).bind(params.business_id);
    for value in &params.ids {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `AttachLeadTasksToDeal` (Go `AttachLeadTasksToDealParams`).
pub struct AttachLeadTasksToDealParams {
    pub deal_id: Option<i64>,
    pub business_id: i64,
    pub lead_id: Option<i64>,
}

/// `AttachLeadTasksToDeal`: `exec`.
pub async fn attach_lead_tasks_to_deal<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AttachLeadTasksToDealParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_tasks SET deal_id = ? WHERE business_id = ? AND lead_id = ? AND deal_id IS NULL")
        .bind(params.deal_id)
        .bind(params.business_id)
        .bind(params.lead_id)
        .execute(exec)
        .await
}

/// Params for `CallsByDay` (Go `CallsByDayParams`).
pub struct CallsByDayParams {
    pub offset_minutes: i64,
    pub business_id: i64,
    pub since: chrono::NaiveDate,
    pub rep: Option<i64>,
}

/// `CallsByDay`: `many`.
pub async fn calls_by_day<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CallsByDayParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT DATE_FORMAT(DATE_ADD(occurred_at, INTERVAL CAST(? AS SIGNED) MINUTE), '%Y-%m-%d') AS local_day,\n       COUNT(*) AS calls,\n       CAST(COALESCE(SUM(outcome IN ('connected', 'callback', 'meeting_booked', 'not_interested')), 0) AS SIGNED) AS reached\nFROM crm_touchpoints\nWHERE business_id = ? AND type = 'call' AND occurred_at >= ?\n  AND (? IS NULL OR created_by = ?)\nGROUP BY local_day\nORDER BY local_day")
        .bind(params.offset_minutes)
        .bind(params.business_id)
        .bind(params.since)
        .bind(params.rep)
        .bind(params.rep)
        .execute(exec)
        .await
}

/// Params for `CallsByRep` (Go `CallsByRepParams`).
pub struct CallsByRepParams {
    pub business_id: i64,
    pub since: chrono::NaiveDate,
    pub rep: Option<i64>,
}

/// `CallsByRep`: `many`.
pub async fn calls_by_rep<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CallsByRepParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT created_by AS rep_id,\n       COUNT(*) AS calls,\n       CAST(COALESCE(SUM(outcome IN ('connected', 'callback', 'meeting_booked', 'not_interested')), 0) AS SIGNED) AS reached\nFROM crm_touchpoints\nWHERE business_id = ? AND type = 'call' AND occurred_at >= ?\n  AND (? IS NULL OR created_by = ?)\nGROUP BY created_by")
        .bind(params.business_id)
        .bind(params.since)
        .bind(params.rep)
        .bind(params.rep)
        .execute(exec)
        .await
}

/// Params for `CountBatchLeads` (Go `CountBatchLeadsParams`).
pub struct CountBatchLeadsParams {
    pub business_id: i64,
    pub import_batch_id: Option<i64>,
}

/// `CountBatchLeads`: `one`.
pub async fn count_batch_leads<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountBatchLeadsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM crm_leads WHERE business_id = ? AND import_batch_id = ?")
        .bind(params.business_id)
        .bind(params.import_batch_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountCallsSince` (Go `CountCallsSinceParams`).
pub struct CountCallsSinceParams {
    pub business_id: i64,
    pub since: chrono::NaiveDate,
    pub filter_reps: i64,
    pub reps: Vec<i64>,
}

/// `CountCallsSince`: `one`.
pub async fn count_calls_since<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountCallsSinceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    let mut sql_template: String = "SELECT COUNT(*) FROM crm_touchpoints\nWHERE business_id = ? AND type = 'call' AND occurred_at >= ?\n  AND (CAST(? AS UNSIGNED) = 0 OR created_by IN ({SLICE:reps}))".to_string();

    if params.reps.is_empty() {
        sql_template = sql_template.replace("{SLICE:reps}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:reps}",
            &std::iter::repeat_n("?", params.reps.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query
        .bind(params.business_id)
        .bind(params.since)
        .bind(params.filter_reps);
    for value in &params.reps {
        query = query.bind(value);
    }
    query.fetch_optional(exec).await
}

/// Params for `CountConnectedCallsSince` (Go `CountConnectedCallsSinceParams`).
pub struct CountConnectedCallsSinceParams {
    pub business_id: i64,
    pub since: chrono::NaiveDate,
    pub filter_reps: i64,
    pub reps: Vec<i64>,
}

/// `CountConnectedCallsSince`: `one`.
pub async fn count_connected_calls_since<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountConnectedCallsSinceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    let mut sql_template: String = "SELECT COUNT(*) FROM crm_touchpoints\nWHERE business_id = ? AND type = 'call' AND occurred_at >= ?\n  AND outcome IN ('connected', 'callback', 'meeting_booked', 'not_interested')\n  AND (CAST(? AS UNSIGNED) = 0 OR created_by IN ({SLICE:reps}))".to_string();

    if params.reps.is_empty() {
        sql_template = sql_template.replace("{SLICE:reps}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:reps}",
            &std::iter::repeat_n("?", params.reps.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query
        .bind(params.business_id)
        .bind(params.since)
        .bind(params.filter_reps);
    for value in &params.reps {
        query = query.bind(value);
    }
    query.fetch_optional(exec).await
}

/// Params for `CountContracts` (Go `CountContractsParams`).
pub struct CountContractsParams {
    pub business_id: i64,
}

/// `CountContracts`: `one`.
pub async fn count_contracts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountContractsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM crm_contracts WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountDealContracts` (Go `CountDealContractsParams`).
pub struct CountDealContractsParams {
    pub business_id: i64,
    pub deal_id: Option<i64>,
}

/// `CountDealContracts`: `one`.
pub async fn count_deal_contracts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountDealContractsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM crm_contracts WHERE business_id = ? AND deal_id = ?")
        .bind(params.business_id)
        .bind(params.deal_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountDeals` (Go `CountDealsParams`).
pub struct CountDealsParams {
    pub business_id: i64,
    pub pipeline_id: Option<i64>,
    pub stage_id: Option<i64>,
    pub statuses: Vec<String>,
    pub assigned_to: Option<i64>,
    pub search: Option<String>,
}

/// `CountDeals`: `one`.
pub async fn count_deals<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountDealsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    let mut sql_template: String = "SELECT COUNT(*) FROM crm_deals\nWHERE business_id = ?\n  AND (? IS NULL OR pipeline_id = ?)\n  AND (? IS NULL OR stage_id = ?)\n  AND status IN ({SLICE:statuses})\n  AND (? IS NULL OR COALESCE(assigned_to, 0) = ?)\n  AND (? IS NULL OR title LIKE ?\n       OR company_name LIKE ? OR contact_name LIKE ?)".to_string();

    if params.statuses.is_empty() {
        sql_template = sql_template.replace("{SLICE:statuses}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:statuses}",
            &std::iter::repeat_n("?", params.statuses.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query
        .bind(params.business_id)
        .bind(params.pipeline_id)
        .bind(params.pipeline_id)
        .bind(params.stage_id)
        .bind(params.stage_id)
        .bind(params.assigned_to)
        .bind(params.assigned_to)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search);
    for value in &params.statuses {
        query = query.bind(value);
    }
    query.fetch_optional(exec).await
}

/// Params for `CountLeadsByStatus` (Go `CountLeadsByStatusParams`).
pub struct CountLeadsByStatusParams {
    pub today: Option<chrono::DateTime<chrono::Utc>>,
    pub month_start: Option<chrono::DateTime<chrono::Utc>>,
    pub business_id: i64,
    pub filter_reps: i64,
    pub reps: Vec<Option<i64>>,
    pub unassigned: i64,
    pub filter_sources: i64,
    pub sources: Vec<String>,
    pub filter_industries: i64,
    pub industries: Vec<Option<String>>,
    pub filter_sizes: i64,
    pub sizes: Vec<Option<String>>,
    pub filter_countries: i64,
    pub countries: Vec<Option<String>>,
    pub filter_cities: i64,
    pub cities: Vec<Option<String>>,
    pub filter_billing: i64,
    pub billing: Vec<String>,
    pub import_batch_id: Option<i64>,
    pub follow_up_from: Option<chrono::DateTime<chrono::Utc>>,
    pub follow_up_to: Option<chrono::DateTime<chrono::Utc>>,
    pub no_follow_up: i64,
    pub has_phone: Option<String>,
    pub has_email: Option<String>,
    pub has_website: Option<String>,
    pub has_linkedin: Option<String>,
    pub has_company: Option<String>,
    pub has_value: Option<String>,
    pub never_contacted: i64,
    pub attempts_min: Option<i32>,
    pub attempts_max: Option<i32>,
    pub value_min: Option<String>,
    pub value_max: Option<String>,
    pub created_from: Option<chrono::DateTime<chrono::Utc>>,
    pub created_before: Option<chrono::DateTime<chrono::Utc>>,
    pub active_from: Option<chrono::DateTime<chrono::Utc>>,
    pub active_before: Option<chrono::DateTime<chrono::Utc>>,
    pub search: Option<String>,
    pub search_digits: Option<String>,
}

/// `CountLeadsByStatus`: `many`.
pub async fn count_leads_by_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountLeadsByStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT status, COUNT(*) AS total,\n       CAST(COALESCE(SUM(next_follow_up < ?), 0) AS SIGNED) AS overdue,\n       CAST(COALESCE(SUM(next_follow_up = ?), 0) AS SIGNED) AS due_today,\n       CAST(COALESCE(SUM(converted_at >= ?), 0) AS SIGNED) AS converted\nFROM crm_leads\nWHERE business_id = ?\n  AND (CAST(? AS UNSIGNED) = 0 OR assigned_to IN ({SLICE:reps})\n       OR (CAST(? AS UNSIGNED) = 1 AND assigned_to IS NULL))\n  AND (CAST(? AS UNSIGNED) = 0 OR source IN ({SLICE:sources}))\n  AND (CAST(? AS UNSIGNED) = 0 OR industry IN ({SLICE:industries}))\n  AND (CAST(? AS UNSIGNED) = 0 OR size IN ({SLICE:sizes}))\n  AND (CAST(? AS UNSIGNED) = 0 OR country IN ({SLICE:countries}))\n  AND (CAST(? AS UNSIGNED) = 0 OR city IN ({SLICE:cities}))\n  AND (CAST(? AS UNSIGNED) = 0 OR billing IN ({SLICE:billing}))\n  AND (? IS NULL OR import_batch_id = ?)\n  AND (? IS NULL OR next_follow_up >= ?)\n  AND (? IS NULL OR next_follow_up <= ?)\n  AND (CAST(? AS UNSIGNED) = 0 OR next_follow_up IS NULL)\n  AND (? IS NULL OR IF(COALESCE(phone_digits, '') = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(COALESCE(email, '') = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(COALESCE(website, '') = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(COALESCE(linkedin, '') = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(company_name = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(COALESCE(value, 0) = 0, 'no', 'yes') = ?)\n  AND (CAST(? AS UNSIGNED) = 0 OR last_activity_at IS NULL)\n  AND (? IS NULL OR attempt_count >= ?)\n  AND (? IS NULL OR attempt_count <= ?)\n  AND (? IS NULL OR value >= ?)\n  AND (? IS NULL OR value <= ?)\n  AND (? IS NULL OR created_at >= ?)\n  AND (? IS NULL OR created_at < ?)\n  AND (? IS NULL OR last_activity_at >= ?)\n  AND (? IS NULL OR last_activity_at < ?)\n  AND (? IS NULL OR company_name LIKE ?\n       OR CONCAT_WS(' ', first_name, last_name) LIKE ?\n       OR email LIKE ? OR job_title LIKE ?\n       OR phone_digits LIKE ?)\nGROUP BY status".to_string();

    if params.reps.is_empty() {
        sql_template = sql_template.replace("{SLICE:reps}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:reps}",
            &std::iter::repeat_n("?", params.reps.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.sources.is_empty() {
        sql_template = sql_template.replace("{SLICE:sources}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:sources}",
            &std::iter::repeat_n("?", params.sources.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.industries.is_empty() {
        sql_template = sql_template.replace("{SLICE:industries}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:industries}",
            &std::iter::repeat_n("?", params.industries.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.sizes.is_empty() {
        sql_template = sql_template.replace("{SLICE:sizes}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:sizes}",
            &std::iter::repeat_n("?", params.sizes.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.countries.is_empty() {
        sql_template = sql_template.replace("{SLICE:countries}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:countries}",
            &std::iter::repeat_n("?", params.countries.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.cities.is_empty() {
        sql_template = sql_template.replace("{SLICE:cities}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:cities}",
            &std::iter::repeat_n("?", params.cities.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.billing.is_empty() {
        sql_template = sql_template.replace("{SLICE:billing}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:billing}",
            &std::iter::repeat_n("?", params.billing.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query
        .bind(params.today)
        .bind(params.today)
        .bind(params.month_start)
        .bind(params.business_id)
        .bind(params.filter_reps)
        .bind(params.unassigned)
        .bind(params.filter_sources)
        .bind(params.filter_industries)
        .bind(params.filter_sizes)
        .bind(params.filter_countries)
        .bind(params.filter_cities)
        .bind(params.filter_billing)
        .bind(params.import_batch_id)
        .bind(params.import_batch_id)
        .bind(params.follow_up_from)
        .bind(params.follow_up_from)
        .bind(params.follow_up_to)
        .bind(params.follow_up_to)
        .bind(params.no_follow_up)
        .bind(&params.has_phone)
        .bind(&params.has_phone)
        .bind(&params.has_email)
        .bind(&params.has_email)
        .bind(&params.has_website)
        .bind(&params.has_website)
        .bind(&params.has_linkedin)
        .bind(&params.has_linkedin)
        .bind(&params.has_company)
        .bind(&params.has_company)
        .bind(&params.has_value)
        .bind(&params.has_value)
        .bind(params.never_contacted)
        .bind(params.attempts_min)
        .bind(params.attempts_min)
        .bind(params.attempts_max)
        .bind(params.attempts_max)
        .bind(&params.value_min)
        .bind(&params.value_min)
        .bind(&params.value_max)
        .bind(&params.value_max)
        .bind(params.created_from)
        .bind(params.created_from)
        .bind(params.created_before)
        .bind(params.created_before)
        .bind(params.active_from)
        .bind(params.active_from)
        .bind(params.active_before)
        .bind(params.active_before)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search_digits);
    for value in &params.reps {
        query = query.bind(value);
    }
    for value in &params.sources {
        query = query.bind(value);
    }
    for value in &params.industries {
        query = query.bind(value);
    }
    for value in &params.sizes {
        query = query.bind(value);
    }
    for value in &params.countries {
        query = query.bind(value);
    }
    for value in &params.cities {
        query = query.bind(value);
    }
    for value in &params.billing {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `CountOpenTasks` (Go `CountOpenTasksParams`).
pub struct CountOpenTasksParams {
    pub today: chrono::NaiveDate,
    pub business_id: i64,
    pub assigned_to: Option<i64>,
}

/// `CountOpenTasks`: `one`.
pub async fn count_open_tasks<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountOpenTasksParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT CAST(COALESCE(SUM(due_date < ?), 0) AS SIGNED) AS overdue,\n       CAST(COALESCE(SUM(due_date = ?), 0) AS SIGNED) AS due_today,\n       CAST(COALESCE(SUM(due_date > ?), 0) AS SIGNED) AS upcoming\nFROM crm_tasks\nWHERE business_id = ? AND completed_at IS NULL\n  AND (? IS NULL OR COALESCE(assigned_to, 0) = ?)")
        .bind(params.today)
        .bind(params.today)
        .bind(params.today)
        .bind(params.business_id)
        .bind(params.assigned_to)
        .bind(params.assigned_to)
        .fetch_optional(exec)
        .await
}

/// Params for `CountPipelineDeals` (Go `CountPipelineDealsParams`).
pub struct CountPipelineDealsParams {
    pub business_id: i64,
    pub pipeline_id: i64,
}

/// `CountPipelineDeals`: `one`.
pub async fn count_pipeline_deals<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountPipelineDealsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM crm_deals WHERE business_id = ? AND pipeline_id = ?")
        .bind(params.business_id)
        .bind(params.pipeline_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountPipelines` (Go `CountPipelinesParams`).
pub struct CountPipelinesParams {
    pub business_id: i64,
}

/// `CountPipelines`: `one`.
pub async fn count_pipelines<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountPipelinesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM crm_pipelines WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountSignedContracts` (Go `CountSignedContractsParams`).
pub struct CountSignedContractsParams {
    pub business_id: i64,
}

/// `CountSignedContracts`: `one`.
pub async fn count_signed_contracts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountSignedContractsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT COUNT(*) FROM crm_contracts WHERE business_id = ? AND signed_at IS NOT NULL",
    )
    .bind(params.business_id)
    .fetch_optional(exec)
    .await
}

/// Params for `CountStageDeals` (Go `CountStageDealsParams`).
pub struct CountStageDealsParams {
    pub business_id: i64,
    pub stage_id: i64,
}

/// `CountStageDeals`: `one`.
pub async fn count_stage_deals<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountStageDealsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM crm_deals WHERE business_id = ? AND stage_id = ?")
        .bind(params.business_id)
        .bind(params.stage_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountUnsignedContracts` (Go `CountUnsignedContractsParams`).
pub struct CountUnsignedContractsParams {
    pub business_id: i64,
}

/// `CountUnsignedContracts`: `one`.
pub async fn count_unsigned_contracts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountUnsignedContractsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM crm_contracts WHERE business_id = ? AND signed_at IS NULL")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateActivity` (Go `CreateActivityParams`).
pub struct CreateActivityParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub lead_id: Option<i64>,
    pub deal_id: Option<i64>,
    pub customer_id: Option<i64>,
    pub r#type: String,
    pub outcome: String,
    pub notes: Option<String>,
    pub duration_seconds: Option<i32>,
    pub created_by: i64,
    pub occurred_at: chrono::NaiveDate,
}

/// `CreateActivity`: `execresult`.
pub async fn create_activity<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateActivityParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_touchpoints (owner_id, business_id, lead_id, deal_id, customer_id, type, outcome, notes,\n                             duration_seconds, created_by, occurred_at)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.lead_id)
        .bind(params.deal_id)
        .bind(params.customer_id)
        .bind(&params.r#type)
        .bind(&params.outcome)
        .bind(&params.notes)
        .bind(params.duration_seconds)
        .bind(params.created_by)
        .bind(params.occurred_at)
        .execute(exec)
        .await
}

/// Params for `CreateCompany` (Go `CreateCompanyParams`).
pub struct CreateCompanyParams {
    pub business_id: i64,
    pub owner_id: i64,
    pub name: String,
    pub website: Option<String>,
    pub industry: Option<String>,
    pub size: Option<String>,
    pub city: Option<String>,
    pub country: Option<String>,
    pub assigned_to: Option<i64>,
}

/// `CreateCompany`: `execresult`.
pub async fn create_company<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateCompanyParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_companies (business_id, owner_id, name, website, industry, size, city, country, assigned_to)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.business_id)
        .bind(params.owner_id)
        .bind(&params.name)
        .bind(&params.website)
        .bind(&params.industry)
        .bind(&params.size)
        .bind(&params.city)
        .bind(&params.country)
        .bind(params.assigned_to)
        .execute(exec)
        .await
}

/// Params for `CreateContract` (Go `CreateContractParams`).
pub struct CreateContractParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub lead_id: Option<i64>,
    pub deal_id: Option<i64>,
    pub plan: String,
    pub mrr: String,
    pub seats: i32,
    pub value: String,
    pub billing: String,
    pub start_date: Option<chrono::DateTime<chrono::Utc>>,
    pub end_date: Option<chrono::DateTime<chrono::Utc>>,
    pub customer_id: Option<i64>,
    pub signed_at: Option<chrono::DateTime<chrono::Utc>>,
    pub signed_by_name: Option<String>,
    pub closed_by: Option<i64>,
    pub notes: Option<String>,
}

/// `CreateContract`: `execresult`.
pub async fn create_contract<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateContractParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_contracts (owner_id, business_id, lead_id, deal_id, plan, mrr, seats,\n                           value, billing, start_date, end_date, customer_id,\n                           signed_at, signed_by_name, closed_by, notes)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.lead_id)
        .bind(params.deal_id)
        .bind(&params.plan)
        .bind(&params.mrr)
        .bind(params.seats)
        .bind(&params.value)
        .bind(&params.billing)
        .bind(params.start_date)
        .bind(params.end_date)
        .bind(params.customer_id)
        .bind(params.signed_at)
        .bind(&params.signed_by_name)
        .bind(params.closed_by)
        .bind(&params.notes)
        .execute(exec)
        .await
}

/// Params for `CreateDeal` (Go `CreateDealParams`).
pub struct CreateDealParams {
    pub business_id: i64,
    pub owner_id: i64,
    pub pipeline_id: i64,
    pub stage_id: i64,
    pub title: String,
    pub lead_id: Option<i64>,
    pub customer_id: Option<i64>,
    pub company_id: Option<i64>,
    pub contact_name: String,
    pub company_name: String,
    pub value: String,
    pub billing: String,
    pub expected_close_date: Option<chrono::DateTime<chrono::Utc>>,
    pub assigned_to: Option<i64>,
    pub source: String,
    pub status: String,
    pub closed_at: Option<chrono::DateTime<chrono::Utc>>,
    pub notes: Option<String>,
}

/// `CreateDeal`: `execresult`.
pub async fn create_deal<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateDealParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_deals (business_id, owner_id, pipeline_id, stage_id, title, lead_id, customer_id, company_id,\n                       contact_name, company_name, value, billing, expected_close_date, assigned_to, source,\n                       status, closed_at, notes)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.business_id)
        .bind(params.owner_id)
        .bind(params.pipeline_id)
        .bind(params.stage_id)
        .bind(&params.title)
        .bind(params.lead_id)
        .bind(params.customer_id)
        .bind(params.company_id)
        .bind(&params.contact_name)
        .bind(&params.company_name)
        .bind(&params.value)
        .bind(&params.billing)
        .bind(params.expected_close_date)
        .bind(params.assigned_to)
        .bind(&params.source)
        .bind(&params.status)
        .bind(params.closed_at)
        .bind(&params.notes)
        .execute(exec)
        .await
}

/// Params for `CreateDefaultPipeline` (Go `CreateDefaultPipelineParams`).
pub struct CreateDefaultPipelineParams {
    pub business_id: i64,
    pub name: String,
}

/// `CreateDefaultPipeline`: `execresult`.
pub async fn create_default_pipeline<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateDefaultPipelineParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT IGNORE INTO crm_pipelines (business_id, name, position, default_marker) VALUES (?, ?, 0, 1)")
        .bind(params.business_id)
        .bind(&params.name)
        .execute(exec)
        .await
}

/// Params for `CreateImportBatch` (Go `CreateImportBatchParams`).
pub struct CreateImportBatchParams {
    pub business_id: i64,
    pub created_by: i64,
    pub file_name: String,
    pub total_rows: i32,
}

/// `CreateImportBatch`: `execresult`.
pub async fn create_import_batch<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateImportBatchParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_import_batches (business_id, created_by, file_name, total_rows) VALUES (?, ?, ?, ?)")
        .bind(params.business_id)
        .bind(params.created_by)
        .bind(&params.file_name)
        .bind(params.total_rows)
        .execute(exec)
        .await
}

/// Params for `CreateLead` (Go `CreateLeadParams`).
pub struct CreateLeadParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub first_name: Option<String>,
    pub last_name: Option<String>,
    pub job_title: Option<String>,
    pub email: Option<String>,
    pub phone: Option<String>,
    pub phone_digits: Option<String>,
    pub company_name: String,
    pub website: Option<String>,
    pub linkedin: Option<String>,
    pub industry: Option<String>,
    pub size: Option<String>,
    pub city: Option<String>,
    pub country: Option<String>,
    pub source: String,
    pub status: String,
    pub assigned_to: Option<i64>,
    pub value: Option<String>,
    pub billing: String,
    pub notes: Option<String>,
    pub import_batch_id: Option<i64>,
}

/// `CreateLead`: `execresult`.
pub async fn create_lead<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateLeadParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_leads (owner_id, business_id, first_name, last_name, job_title, email, phone, phone_digits,\n                       company_name, website, linkedin, industry, size, city, country, source, status,\n                       assigned_to, value, billing, notes, import_batch_id)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.first_name)
        .bind(&params.last_name)
        .bind(&params.job_title)
        .bind(&params.email)
        .bind(&params.phone)
        .bind(&params.phone_digits)
        .bind(&params.company_name)
        .bind(&params.website)
        .bind(&params.linkedin)
        .bind(&params.industry)
        .bind(&params.size)
        .bind(&params.city)
        .bind(&params.country)
        .bind(&params.source)
        .bind(&params.status)
        .bind(params.assigned_to)
        .bind(&params.value)
        .bind(&params.billing)
        .bind(&params.notes)
        .bind(params.import_batch_id)
        .execute(exec)
        .await
}

/// Params for `CreatePipeline` (Go `CreatePipelineParams`).
pub struct CreatePipelineParams {
    pub business_id: i64,
    pub name: String,
    pub position: i32,
}

/// `CreatePipeline`: `execresult`.
pub async fn create_pipeline<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreatePipelineParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_pipelines (business_id, name, position) VALUES (?, ?, ?)")
        .bind(params.business_id)
        .bind(&params.name)
        .bind(params.position)
        .execute(exec)
        .await
}

/// Params for `CreateStage` (Go `CreateStageParams`).
pub struct CreateStageParams {
    pub business_id: i64,
    pub pipeline_id: i64,
    pub name: String,
    pub position: i32,
    pub probability: u8,
    pub kind: String,
}

/// `CreateStage`: `execresult`.
pub async fn create_stage<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateStageParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_pipeline_stages (business_id, pipeline_id, name, position, probability, kind)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.business_id)
        .bind(params.pipeline_id)
        .bind(&params.name)
        .bind(params.position)
        .bind(params.probability)
        .bind(&params.kind)
        .execute(exec)
        .await
}

/// Params for `CreateTask` (Go `CreateTaskParams`).
pub struct CreateTaskParams {
    pub business_id: i64,
    pub owner_id: i64,
    pub lead_id: Option<i64>,
    pub deal_id: Option<i64>,
    pub title: String,
    pub r#type: String,
    pub due_date: chrono::NaiveDate,
    pub assigned_to: Option<i64>,
    pub created_by: i64,
}

/// `CreateTask`: `execresult`.
pub async fn create_task<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateTaskParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_tasks (business_id, owner_id, lead_id, deal_id, title, type, due_date, assigned_to, created_by)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.business_id)
        .bind(params.owner_id)
        .bind(params.lead_id)
        .bind(params.deal_id)
        .bind(&params.title)
        .bind(&params.r#type)
        .bind(params.due_date)
        .bind(params.assigned_to)
        .bind(params.created_by)
        .execute(exec)
        .await
}

/// Params for `DealOutcomesByRep` (Go `DealOutcomesByRepParams`).
pub struct DealOutcomesByRepParams {
    pub business_id: i64,
    pub since: Option<chrono::DateTime<chrono::Utc>>,
    pub rep: Option<i64>,
}

/// `DealOutcomesByRep`: `many`.
pub async fn deal_outcomes_by_rep<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DealOutcomesByRepParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT CAST(COALESCE(assigned_to, 0) AS SIGNED) AS rep_id,\n       CAST(COALESCE(SUM(status = 'won'), 0) AS SIGNED) AS won,\n       CAST(COALESCE(SUM(status = 'lost'), 0) AS SIGNED) AS lost,\n       CAST(COALESCE(SUM(IF(status = 'won', CASE billing WHEN 'monthly' THEN value * 12 ELSE value END, 0)), 0) AS DECIMAL(16,2)) AS won_amount\nFROM crm_deals\nWHERE business_id = ? AND status IN ('won', 'lost') AND closed_at >= ?\n  AND (? IS NULL OR COALESCE(assigned_to, 0) = ?)\nGROUP BY rep_id")
        .bind(params.business_id)
        .bind(params.since)
        .bind(params.rep)
        .bind(params.rep)
        .execute(exec)
        .await
}

/// Params for `DealOutcomesSince` (Go `DealOutcomesSinceParams`).
pub struct DealOutcomesSinceParams {
    pub business_id: i64,
    pub pipeline_id: Option<i64>,
    pub since: Option<chrono::DateTime<chrono::Utc>>,
    pub assigned_to: Option<i64>,
}

/// `DealOutcomesSince`: `one`.
pub async fn deal_outcomes_since<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DealOutcomesSinceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT CAST(COALESCE(SUM(status = 'won'), 0) AS SIGNED) AS won,\n       CAST(COALESCE(SUM(status = 'lost'), 0) AS SIGNED) AS lost,\n       CAST(COALESCE(SUM(IF(status = 'won', CASE billing WHEN 'monthly' THEN value * 12 ELSE value END, 0)), 0) AS DECIMAL(16,2)) AS won_amount,\n       CAST(COALESCE(AVG(IF(status = 'won', DATEDIFF(closed_at, created_at), NULL)), 0) AS DECIMAL(10,1)) AS average_days_to_win\nFROM crm_deals\nWHERE business_id = ?\n  AND (? IS NULL OR pipeline_id = ?)\n  AND status IN ('won', 'lost') AND closed_at >= ?\n  AND (? IS NULL OR COALESCE(assigned_to, 0) = ?)")
        .bind(params.business_id)
        .bind(params.pipeline_id)
        .bind(params.pipeline_id)
        .bind(params.since)
        .bind(params.assigned_to)
        .bind(params.assigned_to)
        .fetch_optional(exec)
        .await
}

/// Params for `DealStageTotals` (Go `DealStageTotalsParams`).
pub struct DealStageTotalsParams {
    pub business_id: i64,
    pub pipeline_id: i64,
    pub assigned_to: Option<i64>,
    pub search: Option<String>,
}

/// `DealStageTotals`: `many`.
pub async fn deal_stage_totals<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DealStageTotalsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT stage_id, COUNT(*) AS deals,\n       CAST(COALESCE(SUM(CASE billing WHEN 'monthly' THEN value * 12 ELSE value END), 0) AS DECIMAL(16,2)) AS amount\nFROM crm_deals\nWHERE business_id = ? AND pipeline_id = ?\n  AND (? IS NULL OR COALESCE(assigned_to, 0) = ?)\n  AND (? IS NULL OR title LIKE ?\n       OR company_name LIKE ? OR contact_name LIKE ?)\nGROUP BY stage_id")
        .bind(params.business_id)
        .bind(params.pipeline_id)
        .bind(params.assigned_to)
        .bind(params.assigned_to)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .execute(exec)
        .await
}

/// Params for `DealsWonByDay` (Go `DealsWonByDayParams`).
pub struct DealsWonByDayParams {
    pub offset_minutes: i64,
    pub business_id: i64,
    pub since: Option<chrono::DateTime<chrono::Utc>>,
    pub rep: Option<i64>,
}

/// `DealsWonByDay`: `many`.
pub async fn deals_won_by_day<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DealsWonByDayParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT DATE_FORMAT(DATE_ADD(closed_at, INTERVAL CAST(? AS SIGNED) MINUTE), '%Y-%m-%d') AS local_day,\n       COUNT(*) AS won\nFROM crm_deals\nWHERE business_id = ? AND status = 'won' AND closed_at >= ?\n  AND (? IS NULL OR COALESCE(assigned_to, 0) = ?)\nGROUP BY local_day\nORDER BY local_day")
        .bind(params.offset_minutes)
        .bind(params.business_id)
        .bind(params.since)
        .bind(params.rep)
        .bind(params.rep)
        .execute(exec)
        .await
}

/// Params for `DeleteActivity` (Go `DeleteActivityParams`).
pub struct DeleteActivityParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteActivity`: `execresult`.
pub async fn delete_activity<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteActivityParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM crm_touchpoints WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `DeleteDeal` (Go `DeleteDealParams`).
pub struct DeleteDealParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteDeal`: `execresult`.
pub async fn delete_deal<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteDealParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM crm_deals WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `DeleteDealActivities` (Go `DeleteDealActivitiesParams`).
pub struct DeleteDealActivitiesParams {
    pub business_id: i64,
    pub deal_id: Option<i64>,
}

/// `DeleteDealActivities`: `exec`.
pub async fn delete_deal_activities<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteDealActivitiesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM crm_touchpoints WHERE business_id = ? AND deal_id = ?")
        .bind(params.business_id)
        .bind(params.deal_id)
        .execute(exec)
        .await
}

/// Params for `DeleteDealTasks` (Go `DeleteDealTasksParams`).
pub struct DeleteDealTasksParams {
    pub business_id: i64,
    pub deal_id: Option<i64>,
}

/// `DeleteDealTasks`: `exec`.
pub async fn delete_deal_tasks<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteDealTasksParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM crm_tasks WHERE business_id = ? AND deal_id = ?")
        .bind(params.business_id)
        .bind(params.deal_id)
        .execute(exec)
        .await
}

/// Params for `DeleteLead` (Go `DeleteLeadParams`).
pub struct DeleteLeadParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteLead`: `execresult`.
pub async fn delete_lead<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteLeadParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM crm_leads WHERE id = ? AND business_id = ? AND status <> 'qualified'")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `DeleteLeadActivities` (Go `DeleteLeadActivitiesParams`).
pub struct DeleteLeadActivitiesParams {
    pub business_id: i64,
    pub lead_ids: Vec<Option<i64>>,
}

/// `DeleteLeadActivities`: `exec`.
pub async fn delete_lead_activities<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteLeadActivitiesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String =
        "DELETE FROM crm_touchpoints WHERE business_id = ? AND lead_id IN ({SLICE:lead_ids})"
            .to_string();

    if params.lead_ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:lead_ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:lead_ids}",
            &std::iter::repeat_n("?", params.lead_ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query.bind(params.business_id);
    for value in &params.lead_ids {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `DeleteLeadTasks` (Go `DeleteLeadTasksParams`).
pub struct DeleteLeadTasksParams {
    pub business_id: i64,
    pub lead_ids: Vec<Option<i64>>,
}

/// `DeleteLeadTasks`: `exec`.
pub async fn delete_lead_tasks<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteLeadTasksParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String =
        "DELETE FROM crm_tasks WHERE business_id = ? AND lead_id IN ({SLICE:lead_ids})".to_string();

    if params.lead_ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:lead_ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:lead_ids}",
            &std::iter::repeat_n("?", params.lead_ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query.bind(params.business_id);
    for value in &params.lead_ids {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `DeleteLeads` (Go `DeleteLeadsParams`).
pub struct DeleteLeadsParams {
    pub business_id: i64,
    pub ids: Vec<i64>,
}

/// `DeleteLeads`: `execresult`.
pub async fn delete_leads<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteLeadsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "DELETE FROM crm_leads\nWHERE business_id = ? AND id IN ({SLICE:ids}) AND status <> 'qualified'".to_string();

    if params.ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:ids}",
            &std::iter::repeat_n("?", params.ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query.bind(params.business_id);
    for value in &params.ids {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `DeletePipeline` (Go `DeletePipelineParams`).
pub struct DeletePipelineParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeletePipeline`: `execresult`.
pub async fn delete_pipeline<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeletePipelineParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "DELETE FROM crm_pipelines WHERE id = ? AND business_id = ? AND default_marker IS NULL",
    )
    .bind(params.id)
    .bind(params.business_id)
    .execute(exec)
    .await
}

/// Params for `DeletePipelineStages` (Go `DeletePipelineStagesParams`).
pub struct DeletePipelineStagesParams {
    pub business_id: i64,
    pub pipeline_id: i64,
}

/// `DeletePipelineStages`: `exec`.
pub async fn delete_pipeline_stages<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeletePipelineStagesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM crm_pipeline_stages WHERE business_id = ? AND pipeline_id = ?")
        .bind(params.business_id)
        .bind(params.pipeline_id)
        .execute(exec)
        .await
}

/// Params for `DeleteStage` (Go `DeleteStageParams`).
pub struct DeleteStageParams {
    pub id: i64,
    pub business_id: i64,
    pub pipeline_id: i64,
}

/// `DeleteStage`: `execresult`.
pub async fn delete_stage<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteStageParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "DELETE FROM crm_pipeline_stages WHERE id = ? AND business_id = ? AND pipeline_id = ?",
    )
    .bind(params.id)
    .bind(params.business_id)
    .bind(params.pipeline_id)
    .execute(exec)
    .await
}

/// Params for `DeleteTask` (Go `DeleteTaskParams`).
pub struct DeleteTaskParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteTask`: `execresult`.
pub async fn delete_task<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteTaskParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM crm_tasks WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `DeleteUntouchedBatchLeads` (Go `DeleteUntouchedBatchLeadsParams`).
pub struct DeleteUntouchedBatchLeadsParams {
    pub business_id: i64,
    pub import_batch_id: Option<i64>,
}

/// `DeleteUntouchedBatchLeads`: `execresult`.
pub async fn delete_untouched_batch_leads<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteUntouchedBatchLeadsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM crm_leads\nWHERE business_id = ? AND import_batch_id = ? AND status = 'new'\n  AND last_activity_at IS NULL AND next_follow_up IS NULL AND deal_id IS NULL")
        .bind(params.business_id)
        .bind(params.import_batch_id)
        .execute(exec)
        .await
}

/// Params for `DetachLeadFromDeal` (Go `DetachLeadFromDealParams`).
pub struct DetachLeadFromDealParams {
    pub business_id: i64,
    pub deal_id: Option<i64>,
}

/// `DetachLeadFromDeal`: `exec`.
pub async fn detach_lead_from_deal<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DetachLeadFromDealParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_leads SET deal_id = NULL, status = 'connected', converted_at = NULL\nWHERE business_id = ? AND deal_id = ?")
        .bind(params.business_id)
        .bind(params.deal_id)
        .execute(exec)
        .await
}

/// Params for `FillLeadBlanks` (Go `FillLeadBlanksParams`).
pub struct FillLeadBlanksParams {
    pub first_name: Option<String>,
    pub last_name: Option<String>,
    pub job_title: Option<String>,
    pub email: Option<String>,
    pub phone: Option<String>,
    pub phone_digits: Option<String>,
    pub company_name: String,
    pub website: Option<String>,
    pub linkedin: Option<String>,
    pub industry: Option<String>,
    pub size: Option<String>,
    pub city: Option<String>,
    pub country: Option<String>,
    pub value: Option<String>,
    pub notes: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `FillLeadBlanks`: `exec`.
pub async fn fill_lead_blanks<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &FillLeadBlanksParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_leads\nSET first_name = COALESCE(first_name, ?),\n    last_name = COALESCE(last_name, ?),\n    job_title = COALESCE(job_title, ?),\n    email = COALESCE(email, ?),\n    phone = COALESCE(phone, ?),\n    phone_digits = COALESCE(phone_digits, ?),\n    company_name = COALESCE(NULLIF(company_name, ''), ?),\n    website = COALESCE(website, ?),\n    linkedin = COALESCE(linkedin, ?),\n    industry = COALESCE(industry, ?),\n    size = COALESCE(size, ?),\n    city = COALESCE(city, ?),\n    country = COALESCE(country, ?),\n    value = COALESCE(value, ?),\n    notes = COALESCE(notes, ?)\nWHERE id = ? AND business_id = ?")
        .bind(&params.first_name)
        .bind(&params.last_name)
        .bind(&params.job_title)
        .bind(&params.email)
        .bind(&params.phone)
        .bind(&params.phone_digits)
        .bind(&params.company_name)
        .bind(&params.website)
        .bind(&params.linkedin)
        .bind(&params.industry)
        .bind(&params.size)
        .bind(&params.city)
        .bind(&params.country)
        .bind(&params.value)
        .bind(&params.notes)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `FindCompanyByName` (Go `FindCompanyByNameParams`).
pub struct FindCompanyByNameParams {
    pub business_id: i64,
    pub name: String,
}

/// `FindCompanyByName`: `one`.
pub async fn find_company_by_name<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &FindCompanyByNameParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, owner_id, name, website, industry, size, phone, city, country, assigned_to, notes, created_at, updated_at FROM crm_companies WHERE business_id = ? AND name = ? ORDER BY id LIMIT 1")
        .bind(params.business_id)
        .bind(&params.name)
        .fetch_optional(exec)
        .await
}

/// Params for `FindLeadsByCompany` (Go `FindLeadsByCompanyParams`).
pub struct FindLeadsByCompanyParams {
    pub business_id: i64,
    pub companies: Vec<String>,
}

/// `FindLeadsByCompany`: `many`.
pub async fn find_leads_by_company<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &FindLeadsByCompanyParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT id, company_name, first_name, last_name FROM crm_leads\nWHERE business_id = ? AND company_name IN ({SLICE:companies})".to_string();

    if params.companies.is_empty() {
        sql_template = sql_template.replace("{SLICE:companies}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:companies}",
            &std::iter::repeat_n("?", params.companies.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query.bind(params.business_id);
    for value in &params.companies {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `FindLeadsByEmail` (Go `FindLeadsByEmailParams`).
pub struct FindLeadsByEmailParams {
    pub business_id: i64,
    pub emails: Vec<Option<String>>,
}

/// `FindLeadsByEmail`: `many`.
pub async fn find_leads_by_email<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &FindLeadsByEmailParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String =
        "SELECT id, email FROM crm_leads\nWHERE business_id = ? AND email IN ({SLICE:emails})"
            .to_string();

    if params.emails.is_empty() {
        sql_template = sql_template.replace("{SLICE:emails}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:emails}",
            &std::iter::repeat_n("?", params.emails.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query.bind(params.business_id);
    for value in &params.emails {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `FindLeadsByPhone` (Go `FindLeadsByPhoneParams`).
pub struct FindLeadsByPhoneParams {
    pub business_id: i64,
    pub phones: Vec<Option<String>>,
}

/// `FindLeadsByPhone`: `many`.
pub async fn find_leads_by_phone<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &FindLeadsByPhoneParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT id, phone_digits FROM crm_leads\nWHERE business_id = ? AND phone_digits IN ({SLICE:phones})".to_string();

    if params.phones.is_empty() {
        sql_template = sql_template.replace("{SLICE:phones}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:phones}",
            &std::iter::repeat_n("?", params.phones.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query.bind(params.business_id);
    for value in &params.phones {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `GetActivity` (Go `GetActivityParams`).
pub struct GetActivityParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetActivity`: `one`.
pub async fn get_activity<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetActivityParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, lead_id, deal_id, customer_id, contact_id, type, outcome, notes, duration_seconds, next_follow_up, created_by, occurred_at, created_at FROM crm_touchpoints WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetContract` (Go `GetContractParams`).
pub struct GetContractParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetContract`: `one`.
pub async fn get_contract<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetContractParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, lead_id, deal_id, plan, mrr, seats, value, billing, start_date, end_date, customer_id, signed_at, signed_by_name, closed_by, notes, created_at, updated_at FROM crm_contracts WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetDeal` (Go `GetDealParams`).
pub struct GetDealParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetDeal`: `one`.
pub async fn get_deal<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetDealParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, owner_id, pipeline_id, stage_id, title, lead_id, customer_id, company_id, contact_name, company_name, value, billing, expected_close_date, assigned_to, source, status, lost_reason, closed_at, notes, last_activity_at, next_follow_up, stage_changed_at, created_at, updated_at FROM crm_deals WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetDealContract` (Go `GetDealContractParams`).
pub struct GetDealContractParams {
    pub business_id: i64,
    pub deal_id: Option<i64>,
}

/// `GetDealContract`: `one`.
pub async fn get_deal_contract<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetDealContractParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, lead_id, deal_id, plan, mrr, seats, value, billing, start_date, end_date, customer_id, signed_at, signed_by_name, closed_by, notes, created_at, updated_at FROM crm_contracts WHERE business_id = ? AND deal_id = ? ORDER BY id DESC LIMIT 1")
        .bind(params.business_id)
        .bind(params.deal_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetDealForUpdate` (Go `GetDealForUpdateParams`).
pub struct GetDealForUpdateParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetDealForUpdate`: `one`.
pub async fn get_deal_for_update<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetDealForUpdateParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, owner_id, pipeline_id, stage_id, title, lead_id, customer_id, company_id, contact_name, company_name, value, billing, expected_close_date, assigned_to, source, status, lost_reason, closed_at, notes, last_activity_at, next_follow_up, stage_changed_at, created_at, updated_at FROM crm_deals WHERE id = ? AND business_id = ? FOR UPDATE")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetDefaultPipeline` (Go `GetDefaultPipelineParams`).
pub struct GetDefaultPipelineParams {
    pub business_id: i64,
}

/// `GetDefaultPipeline`: `one`.
pub async fn get_default_pipeline<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetDefaultPipelineParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, name, position, default_marker, created_at, updated_at FROM crm_pipelines WHERE business_id = ? AND default_marker = 1")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetImportBatch` (Go `GetImportBatchParams`).
pub struct GetImportBatchParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetImportBatch`: `one`.
pub async fn get_import_batch<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetImportBatchParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, created_by, file_name, total_rows, created_count, updated_count, skipped_count, failed_count, status, created_at, completed_at FROM crm_import_batches WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetImportBatchForUpdate` (Go `GetImportBatchForUpdateParams`).
pub struct GetImportBatchForUpdateParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetImportBatchForUpdate`: `one`.
pub async fn get_import_batch_for_update<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetImportBatchForUpdateParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, created_by, file_name, total_rows, created_count, updated_count, skipped_count, failed_count, status, created_at, completed_at FROM crm_import_batches WHERE id = ? AND business_id = ? FOR UPDATE")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetLead` (Go `GetLeadParams`).
pub struct GetLeadParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetLead`: `one`.
pub async fn get_lead<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetLeadParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, first_name, last_name, job_title, email, phone, phone_digits, company_name, website, linkedin, industry, size, city, country, source, status, disqualify_reason, stage, assigned_to, value, billing, mrr_potential, notes, attempt_count, last_activity_at, next_follow_up, customer_id, company_id, deal_id, converted_at, import_batch_id, created_at, updated_at FROM crm_leads WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetLeadForUpdate` (Go `GetLeadForUpdateParams`).
pub struct GetLeadForUpdateParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetLeadForUpdate`: `one`.
pub async fn get_lead_for_update<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetLeadForUpdateParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, first_name, last_name, job_title, email, phone, phone_digits, company_name, website, linkedin, industry, size, city, country, source, status, disqualify_reason, stage, assigned_to, value, billing, mrr_potential, notes, attempt_count, last_activity_at, next_follow_up, customer_id, company_id, deal_id, converted_at, import_batch_id, created_at, updated_at FROM crm_leads WHERE id = ? AND business_id = ? FOR UPDATE")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetPipeline` (Go `GetPipelineParams`).
pub struct GetPipelineParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetPipeline`: `one`.
pub async fn get_pipeline<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPipelineParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, name, position, default_marker, created_at, updated_at FROM crm_pipelines WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetStage` (Go `GetStageParams`).
pub struct GetStageParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetStage`: `one`.
pub async fn get_stage<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetStageParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, pipeline_id, name, position, probability, kind, created_at, updated_at FROM crm_pipeline_stages WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetTask` (Go `GetTaskParams`).
pub struct GetTaskParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetTask`: `one`.
pub async fn get_task<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTaskParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, owner_id, lead_id, deal_id, title, type, due_date, assigned_to, completed_at, created_by, created_at, updated_at FROM crm_tasks WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `LeadsAddedByDay` (Go `LeadsAddedByDayParams`).
pub struct LeadsAddedByDayParams {
    pub offset_minutes: i64,
    pub business_id: i64,
    pub since: chrono::NaiveDate,
    pub rep: Option<i64>,
}

/// `LeadsAddedByDay`: `many`.
pub async fn leads_added_by_day<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LeadsAddedByDayParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT DATE_FORMAT(DATE_ADD(created_at, INTERVAL CAST(? AS SIGNED) MINUTE), '%Y-%m-%d') AS local_day,\n       COUNT(*) AS leads\nFROM crm_leads\nWHERE business_id = ? AND created_at >= ?\n  AND (? IS NULL OR COALESCE(assigned_to, 0) = ?)\nGROUP BY local_day\nORDER BY local_day")
        .bind(params.offset_minutes)
        .bind(params.business_id)
        .bind(params.since)
        .bind(params.rep)
        .bind(params.rep)
        .execute(exec)
        .await
}

/// Params for `LeadsQualifiedByRep` (Go `LeadsQualifiedByRepParams`).
pub struct LeadsQualifiedByRepParams {
    pub business_id: i64,
    pub since: Option<chrono::DateTime<chrono::Utc>>,
    pub rep: Option<i64>,
}

/// `LeadsQualifiedByRep`: `many`.
pub async fn leads_qualified_by_rep<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LeadsQualifiedByRepParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT CAST(COALESCE(assigned_to, 0) AS SIGNED) AS rep_id, COUNT(*) AS qualified\nFROM crm_leads\nWHERE business_id = ? AND converted_at >= ?\n  AND (? IS NULL OR COALESCE(assigned_to, 0) = ?)\nGROUP BY rep_id")
        .bind(params.business_id)
        .bind(params.since)
        .bind(params.rep)
        .bind(params.rep)
        .execute(exec)
        .await
}

/// Params for `LinkDealContractsCustomer` (Go `LinkDealContractsCustomerParams`).
pub struct LinkDealContractsCustomerParams {
    pub customer_id: Option<i64>,
    pub business_id: i64,
    pub lead_id: Option<i64>,
}

/// `LinkDealContractsCustomer`: `exec`.
pub async fn link_deal_contracts_customer<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LinkDealContractsCustomerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_contracts SET customer_id = ?\nWHERE crm_contracts.business_id = ?\n  AND crm_contracts.deal_id IN (\n      SELECT d.id FROM crm_deals d WHERE d.business_id = ? AND d.lead_id = ?)\n  AND crm_contracts.customer_id IS NULL")
        .bind(params.customer_id)
        .bind(params.business_id)
        .bind(params.business_id)
        .bind(params.lead_id)
        .execute(exec)
        .await
}

/// Params for `LinkLeadCustomer` (Go `LinkLeadCustomerParams`).
pub struct LinkLeadCustomerParams {
    pub customer_id: Option<i64>,
    pub id: i64,
    pub business_id: i64,
}

/// `LinkLeadCustomer`: `exec`.
pub async fn link_lead_customer<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LinkLeadCustomerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_leads SET customer_id = ? WHERE id = ? AND business_id = ?")
        .bind(params.customer_id)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `LinkLeadDealsCustomer` (Go `LinkLeadDealsCustomerParams`).
pub struct LinkLeadDealsCustomerParams {
    pub customer_id: Option<i64>,
    pub business_id: i64,
    pub lead_id: Option<i64>,
}

/// `LinkLeadDealsCustomer`: `exec`.
pub async fn link_lead_deals_customer<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LinkLeadDealsCustomerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_deals SET customer_id = ?\nWHERE business_id = ? AND lead_id = ? AND customer_id IS NULL")
        .bind(params.customer_id)
        .bind(params.business_id)
        .bind(params.lead_id)
        .execute(exec)
        .await
}

/// Params for `ListCompletedTasks` (Go `ListCompletedTasksParams`).
pub struct ListCompletedTasksParams {
    pub business_id: i64,
    pub assigned_to: Option<i64>,
    pub lead_id: Option<i64>,
    pub deal_id: Option<i64>,
    pub cursor_id: Option<i64>,
    pub cursor_time: Option<chrono::DateTime<chrono::Utc>>,
    pub limit: i32,
}

/// `ListCompletedTasks`: `many`.
pub async fn list_completed_tasks<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListCompletedTasksParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT t.id, t.lead_id, t.deal_id, t.title, t.type, t.due_date, t.assigned_to, t.completed_at,\n       t.created_by, t.created_at,\n       COALESCE(l.first_name, '') AS lead_first_name, COALESCE(l.last_name, '') AS lead_last_name,\n       COALESCE(l.company_name, '') AS lead_company_name, COALESCE(l.phone, '') AS lead_phone,\n       COALESCE(d.title, '') AS deal_title\nFROM crm_tasks t\nLEFT JOIN crm_leads l ON l.id = t.lead_id AND l.business_id = t.business_id\nLEFT JOIN crm_deals d ON d.id = t.deal_id AND d.business_id = t.business_id\nWHERE t.business_id = ? AND t.completed_at IS NOT NULL\n  AND (? IS NULL OR COALESCE(t.assigned_to, 0) = ?)\n  AND (? IS NULL OR t.lead_id = ?)\n  AND (? IS NULL OR t.deal_id = ?)\n  AND (? IS NULL OR t.completed_at < ?\n       OR (t.completed_at = ? AND t.id < ?))\nORDER BY t.completed_at DESC, t.id DESC\nLIMIT ?")
        .bind(params.business_id)
        .bind(params.assigned_to)
        .bind(params.assigned_to)
        .bind(params.lead_id)
        .bind(params.lead_id)
        .bind(params.deal_id)
        .bind(params.deal_id)
        .bind(params.cursor_id)
        .bind(params.cursor_time)
        .bind(params.cursor_time)
        .bind(params.cursor_id)
        .bind(params.limit)
        .execute(exec)
        .await
}

/// Params for `ListContracts` (Go `ListContractsParams`).
pub struct ListContractsParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListContracts`: `many`.
pub async fn list_contracts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListContractsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, lead_id, deal_id, plan, mrr, seats, value, billing, start_date, end_date, customer_id, signed_at, signed_by_name, closed_by, notes, created_at, updated_at FROM crm_contracts\nWHERE business_id = ?\nORDER BY created_at DESC, id DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListDealActivities` (Go `ListDealActivitiesParams`).
pub struct ListDealActivitiesParams {
    pub business_id: i64,
    pub deal_id: Option<i64>,
    pub lead_id: Option<i64>,
    pub cursor_id: Option<i64>,
    pub cursor_time: Option<chrono::DateTime<chrono::Utc>>,
    pub limit: i32,
}

/// `ListDealActivities`: `many`.
pub async fn list_deal_activities<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListDealActivitiesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, lead_id, deal_id, customer_id, contact_id, type, outcome, notes, duration_seconds, next_follow_up, created_by, occurred_at, created_at FROM crm_touchpoints\nWHERE business_id = ?\n  AND (deal_id = ? OR (? IS NOT NULL AND lead_id = ?))\n  AND (? IS NULL OR occurred_at < ?\n       OR (occurred_at = ? AND id < ?))\nORDER BY occurred_at DESC, id DESC\nLIMIT ?")
        .bind(params.business_id)
        .bind(params.deal_id)
        .bind(params.lead_id)
        .bind(params.lead_id)
        .bind(params.cursor_id)
        .bind(params.cursor_time)
        .bind(params.cursor_time)
        .bind(params.cursor_id)
        .bind(params.limit)
        .execute(exec)
        .await
}

/// Params for `ListDealsByCloseDate` (Go `ListDealsByCloseDateParams`).
pub struct ListDealsByCloseDateParams {
    pub business_id: i64,
    pub pipeline_id: Option<i64>,
    pub stage_id: Option<i64>,
    pub statuses: Vec<String>,
    pub assigned_to: Option<i64>,
    pub search: Option<String>,
    pub cursor_id: Option<i64>,
    pub cursor_date: Option<chrono::DateTime<chrono::Utc>>,
    pub limit: i32,
}

/// `ListDealsByCloseDate`: `many`.
pub async fn list_deals_by_close_date<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListDealsByCloseDateParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT id, business_id, owner_id, pipeline_id, stage_id, title, lead_id, customer_id, company_id, contact_name, company_name, value, billing, expected_close_date, assigned_to, source, status, lost_reason, closed_at, notes, last_activity_at, next_follow_up, stage_changed_at, created_at, updated_at FROM crm_deals\nWHERE business_id = ?\n  AND (? IS NULL OR pipeline_id = ?)\n  AND (? IS NULL OR stage_id = ?)\n  AND status IN ({SLICE:statuses})\n  AND (? IS NULL OR COALESCE(assigned_to, 0) = ?)\n  AND (? IS NULL OR title LIKE ?\n       OR company_name LIKE ? OR contact_name LIKE ?)\n  AND (? IS NULL\n       OR COALESCE(expected_close_date, '9999-12-31') > ?\n       OR (COALESCE(expected_close_date, '9999-12-31') = ? AND id > ?))\nORDER BY COALESCE(expected_close_date, '9999-12-31'), id\nLIMIT ?".to_string();

    if params.statuses.is_empty() {
        sql_template = sql_template.replace("{SLICE:statuses}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:statuses}",
            &std::iter::repeat_n("?", params.statuses.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query
        .bind(params.business_id)
        .bind(params.pipeline_id)
        .bind(params.pipeline_id)
        .bind(params.stage_id)
        .bind(params.stage_id)
        .bind(params.assigned_to)
        .bind(params.assigned_to)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(params.cursor_id)
        .bind(params.cursor_date)
        .bind(params.cursor_date)
        .bind(params.cursor_id)
        .bind(params.limit);
    for value in &params.statuses {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `ListDealsByValue` (Go `ListDealsByValueParams`).
pub struct ListDealsByValueParams {
    pub business_id: i64,
    pub pipeline_id: Option<i64>,
    pub stage_id: Option<i64>,
    pub statuses: Vec<String>,
    pub assigned_to: Option<i64>,
    pub search: Option<String>,
    pub cursor_id: Option<i64>,
    pub cursor_amount: Option<String>,
    pub limit: i32,
}

/// `ListDealsByValue`: `many`.
pub async fn list_deals_by_value<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListDealsByValueParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT id, business_id, owner_id, pipeline_id, stage_id, title, lead_id, customer_id, company_id, contact_name, company_name, value, billing, expected_close_date, assigned_to, source, status, lost_reason, closed_at, notes, last_activity_at, next_follow_up, stage_changed_at, created_at, updated_at FROM crm_deals\nWHERE business_id = ?\n  AND (? IS NULL OR pipeline_id = ?)\n  AND (? IS NULL OR stage_id = ?)\n  AND status IN ({SLICE:statuses})\n  AND (? IS NULL OR COALESCE(assigned_to, 0) = ?)\n  AND (? IS NULL OR title LIKE ?\n       OR company_name LIKE ? OR contact_name LIKE ?)\n  AND (? IS NULL\n       OR value * IF(billing = 'monthly', 12, 1) < ?\n       OR (value * IF(billing = 'monthly', 12, 1) = ? AND id < ?))\nORDER BY value * IF(billing = 'monthly', 12, 1) DESC, id DESC\nLIMIT ?".to_string();

    if params.statuses.is_empty() {
        sql_template = sql_template.replace("{SLICE:statuses}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:statuses}",
            &std::iter::repeat_n("?", params.statuses.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query
        .bind(params.business_id)
        .bind(params.pipeline_id)
        .bind(params.pipeline_id)
        .bind(params.stage_id)
        .bind(params.stage_id)
        .bind(params.assigned_to)
        .bind(params.assigned_to)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(params.cursor_id)
        .bind(&params.cursor_amount)
        .bind(&params.cursor_amount)
        .bind(params.cursor_id)
        .bind(params.limit);
    for value in &params.statuses {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `ListDealsNewest` (Go `ListDealsNewestParams`).
pub struct ListDealsNewestParams {
    pub business_id: i64,
    pub pipeline_id: Option<i64>,
    pub stage_id: Option<i64>,
    pub statuses: Vec<String>,
    pub assigned_to: Option<i64>,
    pub search: Option<String>,
    pub cursor_id: Option<i64>,
    pub cursor_time: Option<chrono::DateTime<chrono::Utc>>,
    pub limit: i32,
}

/// `ListDealsNewest`: `many`.
pub async fn list_deals_newest<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListDealsNewestParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT id, business_id, owner_id, pipeline_id, stage_id, title, lead_id, customer_id, company_id, contact_name, company_name, value, billing, expected_close_date, assigned_to, source, status, lost_reason, closed_at, notes, last_activity_at, next_follow_up, stage_changed_at, created_at, updated_at FROM crm_deals\nWHERE business_id = ?\n  AND (? IS NULL OR pipeline_id = ?)\n  AND (? IS NULL OR stage_id = ?)\n  AND status IN ({SLICE:statuses})\n  AND (? IS NULL OR COALESCE(assigned_to, 0) = ?)\n  AND (? IS NULL OR title LIKE ?\n       OR company_name LIKE ? OR contact_name LIKE ?)\n  AND (? IS NULL OR created_at < ?\n       OR (created_at = ? AND id < ?))\nORDER BY created_at DESC, id DESC\nLIMIT ?".to_string();

    if params.statuses.is_empty() {
        sql_template = sql_template.replace("{SLICE:statuses}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:statuses}",
            &std::iter::repeat_n("?", params.statuses.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query
        .bind(params.business_id)
        .bind(params.pipeline_id)
        .bind(params.pipeline_id)
        .bind(params.stage_id)
        .bind(params.stage_id)
        .bind(params.assigned_to)
        .bind(params.assigned_to)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(params.cursor_id)
        .bind(params.cursor_time)
        .bind(params.cursor_time)
        .bind(params.cursor_id)
        .bind(params.limit);
    for value in &params.statuses {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `ListDeletableLeadIDs` (Go `ListDeletableLeadIDsParams`).
pub struct ListDeletableLeadIDsParams {
    pub business_id: i64,
    pub ids: Vec<i64>,
}

/// `ListDeletableLeadIDs`: `many`.
pub async fn list_deletable_lead_i_ds<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListDeletableLeadIDsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT id FROM crm_leads\nWHERE business_id = ? AND id IN ({SLICE:ids}) AND status <> 'qualified'".to_string();

    if params.ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:ids}",
            &std::iter::repeat_n("?", params.ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query.bind(params.business_id);
    for value in &params.ids {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `ListImportBatches` (Go `ListImportBatchesParams`).
pub struct ListImportBatchesParams {
    pub business_id: i64,
}

/// `ListImportBatches`: `many`.
pub async fn list_import_batches<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListImportBatchesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, business_id, created_by, file_name, total_rows, created_count, updated_count, skipped_count, failed_count, status, created_at, completed_at FROM crm_import_batches WHERE business_id = ? ORDER BY created_at DESC, id DESC LIMIT 20")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListLeadActivities` (Go `ListLeadActivitiesParams`).
pub struct ListLeadActivitiesParams {
    pub business_id: i64,
    pub lead_id: Option<i64>,
    pub cursor_id: Option<i64>,
    pub cursor_time: Option<chrono::DateTime<chrono::Utc>>,
    pub limit: i32,
}

/// `ListLeadActivities`: `many`.
pub async fn list_lead_activities<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLeadActivitiesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, lead_id, deal_id, customer_id, contact_id, type, outcome, notes, duration_seconds, next_follow_up, created_by, occurred_at, created_at FROM crm_touchpoints\nWHERE business_id = ? AND lead_id = ?\n  AND (? IS NULL OR occurred_at < ?\n       OR (occurred_at = ? AND id < ?))\nORDER BY occurred_at DESC, id DESC\nLIMIT ?")
        .bind(params.business_id)
        .bind(params.lead_id)
        .bind(params.cursor_id)
        .bind(params.cursor_time)
        .bind(params.cursor_time)
        .bind(params.cursor_id)
        .bind(params.limit)
        .execute(exec)
        .await
}

/// Params for `ListLeadCities` (Go `ListLeadCitiesParams`).
pub struct ListLeadCitiesParams {
    pub business_id: i64,
}

/// `ListLeadCities`: `many`.
pub async fn list_lead_cities<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLeadCitiesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT DISTINCT city FROM crm_leads\nWHERE business_id = ? AND city <> ''\nORDER BY city\nLIMIT 500")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListLeadCountries` (Go `ListLeadCountriesParams`).
pub struct ListLeadCountriesParams {
    pub business_id: i64,
}

/// `ListLeadCountries`: `many`.
pub async fn list_lead_countries<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLeadCountriesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT DISTINCT country FROM crm_leads\nWHERE business_id = ? AND country <> ''\nORDER BY country\nLIMIT 500")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListLeadIndustries` (Go `ListLeadIndustriesParams`).
pub struct ListLeadIndustriesParams {
    pub business_id: i64,
}

/// `ListLeadIndustries`: `many`.
pub async fn list_lead_industries<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLeadIndustriesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT DISTINCT industry FROM crm_leads\nWHERE business_id = ? AND industry <> ''\nORDER BY industry\nLIMIT 500")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListLeadSizes` (Go `ListLeadSizesParams`).
pub struct ListLeadSizesParams {
    pub business_id: i64,
}

/// `ListLeadSizes`: `many`.
pub async fn list_lead_sizes<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLeadSizesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT DISTINCT size FROM crm_leads\nWHERE business_id = ? AND size <> ''\nORDER BY size\nLIMIT 500")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListLeadSources` (Go `ListLeadSourcesParams`).
pub struct ListLeadSourcesParams {
    pub business_id: i64,
}

/// `ListLeadSources`: `many`.
pub async fn list_lead_sources<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLeadSourcesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT DISTINCT source FROM crm_leads\nWHERE business_id = ? AND source <> ''\nORDER BY source\nLIMIT 500")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListLeadsByActivity` (Go `ListLeadsByActivityParams`).
pub struct ListLeadsByActivityParams {
    pub business_id: i64,
    pub statuses: Vec<String>,
    pub filter_reps: i64,
    pub reps: Vec<Option<i64>>,
    pub unassigned: i64,
    pub filter_sources: i64,
    pub sources: Vec<String>,
    pub filter_industries: i64,
    pub industries: Vec<Option<String>>,
    pub filter_sizes: i64,
    pub sizes: Vec<Option<String>>,
    pub filter_countries: i64,
    pub countries: Vec<Option<String>>,
    pub filter_cities: i64,
    pub cities: Vec<Option<String>>,
    pub filter_billing: i64,
    pub billing: Vec<String>,
    pub import_batch_id: Option<i64>,
    pub follow_up_from: Option<chrono::DateTime<chrono::Utc>>,
    pub follow_up_to: Option<chrono::DateTime<chrono::Utc>>,
    pub no_follow_up: i64,
    pub has_phone: Option<String>,
    pub has_email: Option<String>,
    pub has_website: Option<String>,
    pub has_linkedin: Option<String>,
    pub has_company: Option<String>,
    pub has_value: Option<String>,
    pub never_contacted: i64,
    pub attempts_min: Option<i32>,
    pub attempts_max: Option<i32>,
    pub value_min: Option<String>,
    pub value_max: Option<String>,
    pub created_from: Option<chrono::DateTime<chrono::Utc>>,
    pub created_before: Option<chrono::DateTime<chrono::Utc>>,
    pub active_from: Option<chrono::DateTime<chrono::Utc>>,
    pub active_before: Option<chrono::DateTime<chrono::Utc>>,
    pub search: Option<String>,
    pub search_digits: Option<String>,
    pub cursor_id: Option<i64>,
    pub cursor_time: Option<chrono::DateTime<chrono::Utc>>,
    pub limit: i32,
}

/// `ListLeadsByActivity`: `many`.
pub async fn list_leads_by_activity<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLeadsByActivityParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT id, owner_id, business_id, first_name, last_name, job_title, email, phone, phone_digits, company_name, website, linkedin, industry, size, city, country, source, status, disqualify_reason, stage, assigned_to, value, billing, mrr_potential, notes, attempt_count, last_activity_at, next_follow_up, customer_id, company_id, deal_id, converted_at, import_batch_id, created_at, updated_at FROM crm_leads\nWHERE business_id = ?\n  AND status IN ({SLICE:statuses})\n  AND (CAST(? AS UNSIGNED) = 0 OR assigned_to IN ({SLICE:reps})\n       OR (CAST(? AS UNSIGNED) = 1 AND assigned_to IS NULL))\n  AND (CAST(? AS UNSIGNED) = 0 OR source IN ({SLICE:sources}))\n  AND (CAST(? AS UNSIGNED) = 0 OR industry IN ({SLICE:industries}))\n  AND (CAST(? AS UNSIGNED) = 0 OR size IN ({SLICE:sizes}))\n  AND (CAST(? AS UNSIGNED) = 0 OR country IN ({SLICE:countries}))\n  AND (CAST(? AS UNSIGNED) = 0 OR city IN ({SLICE:cities}))\n  AND (CAST(? AS UNSIGNED) = 0 OR billing IN ({SLICE:billing}))\n  AND (? IS NULL OR import_batch_id = ?)\n  AND (? IS NULL OR next_follow_up >= ?)\n  AND (? IS NULL OR next_follow_up <= ?)\n  AND (CAST(? AS UNSIGNED) = 0 OR next_follow_up IS NULL)\n  AND (? IS NULL OR IF(COALESCE(phone_digits, '') = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(COALESCE(email, '') = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(COALESCE(website, '') = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(COALESCE(linkedin, '') = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(company_name = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(COALESCE(value, 0) = 0, 'no', 'yes') = ?)\n  AND (CAST(? AS UNSIGNED) = 0 OR last_activity_at IS NULL)\n  AND (? IS NULL OR attempt_count >= ?)\n  AND (? IS NULL OR attempt_count <= ?)\n  AND (? IS NULL OR value >= ?)\n  AND (? IS NULL OR value <= ?)\n  AND (? IS NULL OR created_at >= ?)\n  AND (? IS NULL OR created_at < ?)\n  AND (? IS NULL OR last_activity_at >= ?)\n  AND (? IS NULL OR last_activity_at < ?)\n  AND (? IS NULL OR company_name LIKE ?\n       OR CONCAT_WS(' ', first_name, last_name) LIKE ?\n       OR email LIKE ? OR job_title LIKE ?\n       OR phone_digits LIKE ?)\n  AND (? IS NULL\n       OR COALESCE(last_activity_at, '1970-01-01 00:00:01') < ?\n       OR (COALESCE(last_activity_at, '1970-01-01 00:00:01') = ? AND id < ?))\nORDER BY COALESCE(last_activity_at, '1970-01-01 00:00:01') DESC, id DESC\nLIMIT ?".to_string();

    if params.statuses.is_empty() {
        sql_template = sql_template.replace("{SLICE:statuses}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:statuses}",
            &std::iter::repeat_n("?", params.statuses.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.reps.is_empty() {
        sql_template = sql_template.replace("{SLICE:reps}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:reps}",
            &std::iter::repeat_n("?", params.reps.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.sources.is_empty() {
        sql_template = sql_template.replace("{SLICE:sources}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:sources}",
            &std::iter::repeat_n("?", params.sources.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.industries.is_empty() {
        sql_template = sql_template.replace("{SLICE:industries}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:industries}",
            &std::iter::repeat_n("?", params.industries.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.sizes.is_empty() {
        sql_template = sql_template.replace("{SLICE:sizes}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:sizes}",
            &std::iter::repeat_n("?", params.sizes.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.countries.is_empty() {
        sql_template = sql_template.replace("{SLICE:countries}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:countries}",
            &std::iter::repeat_n("?", params.countries.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.cities.is_empty() {
        sql_template = sql_template.replace("{SLICE:cities}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:cities}",
            &std::iter::repeat_n("?", params.cities.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.billing.is_empty() {
        sql_template = sql_template.replace("{SLICE:billing}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:billing}",
            &std::iter::repeat_n("?", params.billing.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query
        .bind(params.business_id)
        .bind(params.filter_reps)
        .bind(params.unassigned)
        .bind(params.filter_sources)
        .bind(params.filter_industries)
        .bind(params.filter_sizes)
        .bind(params.filter_countries)
        .bind(params.filter_cities)
        .bind(params.filter_billing)
        .bind(params.import_batch_id)
        .bind(params.import_batch_id)
        .bind(params.follow_up_from)
        .bind(params.follow_up_from)
        .bind(params.follow_up_to)
        .bind(params.follow_up_to)
        .bind(params.no_follow_up)
        .bind(&params.has_phone)
        .bind(&params.has_phone)
        .bind(&params.has_email)
        .bind(&params.has_email)
        .bind(&params.has_website)
        .bind(&params.has_website)
        .bind(&params.has_linkedin)
        .bind(&params.has_linkedin)
        .bind(&params.has_company)
        .bind(&params.has_company)
        .bind(&params.has_value)
        .bind(&params.has_value)
        .bind(params.never_contacted)
        .bind(params.attempts_min)
        .bind(params.attempts_min)
        .bind(params.attempts_max)
        .bind(params.attempts_max)
        .bind(&params.value_min)
        .bind(&params.value_min)
        .bind(&params.value_max)
        .bind(&params.value_max)
        .bind(params.created_from)
        .bind(params.created_from)
        .bind(params.created_before)
        .bind(params.created_before)
        .bind(params.active_from)
        .bind(params.active_from)
        .bind(params.active_before)
        .bind(params.active_before)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search_digits)
        .bind(params.cursor_id)
        .bind(params.cursor_time)
        .bind(params.cursor_time)
        .bind(params.cursor_id)
        .bind(params.limit);
    for value in &params.statuses {
        query = query.bind(value);
    }
    for value in &params.reps {
        query = query.bind(value);
    }
    for value in &params.sources {
        query = query.bind(value);
    }
    for value in &params.industries {
        query = query.bind(value);
    }
    for value in &params.sizes {
        query = query.bind(value);
    }
    for value in &params.countries {
        query = query.bind(value);
    }
    for value in &params.cities {
        query = query.bind(value);
    }
    for value in &params.billing {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `ListLeadsByCompany` (Go `ListLeadsByCompanyParams`).
pub struct ListLeadsByCompanyParams {
    pub business_id: i64,
    pub statuses: Vec<String>,
    pub filter_reps: i64,
    pub reps: Vec<Option<i64>>,
    pub unassigned: i64,
    pub filter_sources: i64,
    pub sources: Vec<String>,
    pub filter_industries: i64,
    pub industries: Vec<Option<String>>,
    pub filter_sizes: i64,
    pub sizes: Vec<Option<String>>,
    pub filter_countries: i64,
    pub countries: Vec<Option<String>>,
    pub filter_cities: i64,
    pub cities: Vec<Option<String>>,
    pub filter_billing: i64,
    pub billing: Vec<String>,
    pub import_batch_id: Option<i64>,
    pub follow_up_from: Option<chrono::DateTime<chrono::Utc>>,
    pub follow_up_to: Option<chrono::DateTime<chrono::Utc>>,
    pub no_follow_up: i64,
    pub has_phone: Option<String>,
    pub has_email: Option<String>,
    pub has_website: Option<String>,
    pub has_linkedin: Option<String>,
    pub has_company: Option<String>,
    pub has_value: Option<String>,
    pub never_contacted: i64,
    pub attempts_min: Option<i32>,
    pub attempts_max: Option<i32>,
    pub value_min: Option<String>,
    pub value_max: Option<String>,
    pub created_from: Option<chrono::DateTime<chrono::Utc>>,
    pub created_before: Option<chrono::DateTime<chrono::Utc>>,
    pub active_from: Option<chrono::DateTime<chrono::Utc>>,
    pub active_before: Option<chrono::DateTime<chrono::Utc>>,
    pub search: Option<String>,
    pub search_digits: Option<String>,
    pub cursor_id: Option<i64>,
    pub cursor_text: Option<String>,
    pub limit: i32,
}

/// `ListLeadsByCompany`: `many`.
pub async fn list_leads_by_company<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLeadsByCompanyParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT id, owner_id, business_id, first_name, last_name, job_title, email, phone, phone_digits, company_name, website, linkedin, industry, size, city, country, source, status, disqualify_reason, stage, assigned_to, value, billing, mrr_potential, notes, attempt_count, last_activity_at, next_follow_up, customer_id, company_id, deal_id, converted_at, import_batch_id, created_at, updated_at FROM crm_leads\nWHERE business_id = ?\n  AND status IN ({SLICE:statuses})\n  AND (CAST(? AS UNSIGNED) = 0 OR assigned_to IN ({SLICE:reps})\n       OR (CAST(? AS UNSIGNED) = 1 AND assigned_to IS NULL))\n  AND (CAST(? AS UNSIGNED) = 0 OR source IN ({SLICE:sources}))\n  AND (CAST(? AS UNSIGNED) = 0 OR industry IN ({SLICE:industries}))\n  AND (CAST(? AS UNSIGNED) = 0 OR size IN ({SLICE:sizes}))\n  AND (CAST(? AS UNSIGNED) = 0 OR country IN ({SLICE:countries}))\n  AND (CAST(? AS UNSIGNED) = 0 OR city IN ({SLICE:cities}))\n  AND (CAST(? AS UNSIGNED) = 0 OR billing IN ({SLICE:billing}))\n  AND (? IS NULL OR import_batch_id = ?)\n  AND (? IS NULL OR next_follow_up >= ?)\n  AND (? IS NULL OR next_follow_up <= ?)\n  AND (CAST(? AS UNSIGNED) = 0 OR next_follow_up IS NULL)\n  AND (? IS NULL OR IF(COALESCE(phone_digits, '') = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(COALESCE(email, '') = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(COALESCE(website, '') = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(COALESCE(linkedin, '') = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(company_name = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(COALESCE(value, 0) = 0, 'no', 'yes') = ?)\n  AND (CAST(? AS UNSIGNED) = 0 OR last_activity_at IS NULL)\n  AND (? IS NULL OR attempt_count >= ?)\n  AND (? IS NULL OR attempt_count <= ?)\n  AND (? IS NULL OR value >= ?)\n  AND (? IS NULL OR value <= ?)\n  AND (? IS NULL OR created_at >= ?)\n  AND (? IS NULL OR created_at < ?)\n  AND (? IS NULL OR last_activity_at >= ?)\n  AND (? IS NULL OR last_activity_at < ?)\n  AND (? IS NULL OR company_name LIKE ?\n       OR CONCAT_WS(' ', first_name, last_name) LIKE ?\n       OR email LIKE ? OR job_title LIKE ?\n       OR phone_digits LIKE ?)\n  AND (? IS NULL OR company_name > ?\n       OR (company_name = ? AND id > ?))\nORDER BY company_name, id\nLIMIT ?".to_string();

    if params.statuses.is_empty() {
        sql_template = sql_template.replace("{SLICE:statuses}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:statuses}",
            &std::iter::repeat_n("?", params.statuses.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.reps.is_empty() {
        sql_template = sql_template.replace("{SLICE:reps}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:reps}",
            &std::iter::repeat_n("?", params.reps.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.sources.is_empty() {
        sql_template = sql_template.replace("{SLICE:sources}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:sources}",
            &std::iter::repeat_n("?", params.sources.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.industries.is_empty() {
        sql_template = sql_template.replace("{SLICE:industries}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:industries}",
            &std::iter::repeat_n("?", params.industries.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.sizes.is_empty() {
        sql_template = sql_template.replace("{SLICE:sizes}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:sizes}",
            &std::iter::repeat_n("?", params.sizes.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.countries.is_empty() {
        sql_template = sql_template.replace("{SLICE:countries}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:countries}",
            &std::iter::repeat_n("?", params.countries.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.cities.is_empty() {
        sql_template = sql_template.replace("{SLICE:cities}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:cities}",
            &std::iter::repeat_n("?", params.cities.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.billing.is_empty() {
        sql_template = sql_template.replace("{SLICE:billing}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:billing}",
            &std::iter::repeat_n("?", params.billing.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query
        .bind(params.business_id)
        .bind(params.filter_reps)
        .bind(params.unassigned)
        .bind(params.filter_sources)
        .bind(params.filter_industries)
        .bind(params.filter_sizes)
        .bind(params.filter_countries)
        .bind(params.filter_cities)
        .bind(params.filter_billing)
        .bind(params.import_batch_id)
        .bind(params.import_batch_id)
        .bind(params.follow_up_from)
        .bind(params.follow_up_from)
        .bind(params.follow_up_to)
        .bind(params.follow_up_to)
        .bind(params.no_follow_up)
        .bind(&params.has_phone)
        .bind(&params.has_phone)
        .bind(&params.has_email)
        .bind(&params.has_email)
        .bind(&params.has_website)
        .bind(&params.has_website)
        .bind(&params.has_linkedin)
        .bind(&params.has_linkedin)
        .bind(&params.has_company)
        .bind(&params.has_company)
        .bind(&params.has_value)
        .bind(&params.has_value)
        .bind(params.never_contacted)
        .bind(params.attempts_min)
        .bind(params.attempts_min)
        .bind(params.attempts_max)
        .bind(params.attempts_max)
        .bind(&params.value_min)
        .bind(&params.value_min)
        .bind(&params.value_max)
        .bind(&params.value_max)
        .bind(params.created_from)
        .bind(params.created_from)
        .bind(params.created_before)
        .bind(params.created_before)
        .bind(params.active_from)
        .bind(params.active_from)
        .bind(params.active_before)
        .bind(params.active_before)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search_digits)
        .bind(params.cursor_id)
        .bind(&params.cursor_text)
        .bind(&params.cursor_text)
        .bind(params.cursor_id)
        .bind(params.limit);
    for value in &params.statuses {
        query = query.bind(value);
    }
    for value in &params.reps {
        query = query.bind(value);
    }
    for value in &params.sources {
        query = query.bind(value);
    }
    for value in &params.industries {
        query = query.bind(value);
    }
    for value in &params.sizes {
        query = query.bind(value);
    }
    for value in &params.countries {
        query = query.bind(value);
    }
    for value in &params.cities {
        query = query.bind(value);
    }
    for value in &params.billing {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `ListLeadsByFollowUp` (Go `ListLeadsByFollowUpParams`).
pub struct ListLeadsByFollowUpParams {
    pub business_id: i64,
    pub statuses: Vec<String>,
    pub filter_reps: i64,
    pub reps: Vec<Option<i64>>,
    pub unassigned: i64,
    pub filter_sources: i64,
    pub sources: Vec<String>,
    pub filter_industries: i64,
    pub industries: Vec<Option<String>>,
    pub filter_sizes: i64,
    pub sizes: Vec<Option<String>>,
    pub filter_countries: i64,
    pub countries: Vec<Option<String>>,
    pub filter_cities: i64,
    pub cities: Vec<Option<String>>,
    pub filter_billing: i64,
    pub billing: Vec<String>,
    pub import_batch_id: Option<i64>,
    pub follow_up_from: Option<chrono::DateTime<chrono::Utc>>,
    pub follow_up_to: Option<chrono::DateTime<chrono::Utc>>,
    pub no_follow_up: i64,
    pub has_phone: Option<String>,
    pub has_email: Option<String>,
    pub has_website: Option<String>,
    pub has_linkedin: Option<String>,
    pub has_company: Option<String>,
    pub has_value: Option<String>,
    pub never_contacted: i64,
    pub attempts_min: Option<i32>,
    pub attempts_max: Option<i32>,
    pub value_min: Option<String>,
    pub value_max: Option<String>,
    pub created_from: Option<chrono::DateTime<chrono::Utc>>,
    pub created_before: Option<chrono::DateTime<chrono::Utc>>,
    pub active_from: Option<chrono::DateTime<chrono::Utc>>,
    pub active_before: Option<chrono::DateTime<chrono::Utc>>,
    pub search: Option<String>,
    pub search_digits: Option<String>,
    pub cursor_id: Option<i64>,
    pub cursor_time: Option<chrono::DateTime<chrono::Utc>>,
    pub limit: i32,
}

/// `ListLeadsByFollowUp`: `many`.
pub async fn list_leads_by_follow_up<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLeadsByFollowUpParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT id, owner_id, business_id, first_name, last_name, job_title, email, phone, phone_digits, company_name, website, linkedin, industry, size, city, country, source, status, disqualify_reason, stage, assigned_to, value, billing, mrr_potential, notes, attempt_count, last_activity_at, next_follow_up, customer_id, company_id, deal_id, converted_at, import_batch_id, created_at, updated_at FROM crm_leads\nWHERE business_id = ?\n  AND status IN ({SLICE:statuses})\n  AND (CAST(? AS UNSIGNED) = 0 OR assigned_to IN ({SLICE:reps})\n       OR (CAST(? AS UNSIGNED) = 1 AND assigned_to IS NULL))\n  AND (CAST(? AS UNSIGNED) = 0 OR source IN ({SLICE:sources}))\n  AND (CAST(? AS UNSIGNED) = 0 OR industry IN ({SLICE:industries}))\n  AND (CAST(? AS UNSIGNED) = 0 OR size IN ({SLICE:sizes}))\n  AND (CAST(? AS UNSIGNED) = 0 OR country IN ({SLICE:countries}))\n  AND (CAST(? AS UNSIGNED) = 0 OR city IN ({SLICE:cities}))\n  AND (CAST(? AS UNSIGNED) = 0 OR billing IN ({SLICE:billing}))\n  AND (? IS NULL OR import_batch_id = ?)\n  AND (? IS NULL OR next_follow_up >= ?)\n  AND (? IS NULL OR next_follow_up <= ?)\n  AND (CAST(? AS UNSIGNED) = 0 OR next_follow_up IS NULL)\n  AND (? IS NULL OR IF(COALESCE(phone_digits, '') = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(COALESCE(email, '') = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(COALESCE(website, '') = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(COALESCE(linkedin, '') = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(company_name = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(COALESCE(value, 0) = 0, 'no', 'yes') = ?)\n  AND (CAST(? AS UNSIGNED) = 0 OR last_activity_at IS NULL)\n  AND (? IS NULL OR attempt_count >= ?)\n  AND (? IS NULL OR attempt_count <= ?)\n  AND (? IS NULL OR value >= ?)\n  AND (? IS NULL OR value <= ?)\n  AND (? IS NULL OR created_at >= ?)\n  AND (? IS NULL OR created_at < ?)\n  AND (? IS NULL OR last_activity_at >= ?)\n  AND (? IS NULL OR last_activity_at < ?)\n  AND (? IS NULL OR company_name LIKE ?\n       OR CONCAT_WS(' ', first_name, last_name) LIKE ?\n       OR email LIKE ? OR job_title LIKE ?\n       OR phone_digits LIKE ?)\n  AND (? IS NULL\n       OR COALESCE(next_follow_up, '9999-12-31') > ?\n       OR (COALESCE(next_follow_up, '9999-12-31') = ? AND id > ?))\nORDER BY COALESCE(next_follow_up, '9999-12-31'), id\nLIMIT ?".to_string();

    if params.statuses.is_empty() {
        sql_template = sql_template.replace("{SLICE:statuses}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:statuses}",
            &std::iter::repeat_n("?", params.statuses.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.reps.is_empty() {
        sql_template = sql_template.replace("{SLICE:reps}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:reps}",
            &std::iter::repeat_n("?", params.reps.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.sources.is_empty() {
        sql_template = sql_template.replace("{SLICE:sources}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:sources}",
            &std::iter::repeat_n("?", params.sources.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.industries.is_empty() {
        sql_template = sql_template.replace("{SLICE:industries}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:industries}",
            &std::iter::repeat_n("?", params.industries.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.sizes.is_empty() {
        sql_template = sql_template.replace("{SLICE:sizes}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:sizes}",
            &std::iter::repeat_n("?", params.sizes.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.countries.is_empty() {
        sql_template = sql_template.replace("{SLICE:countries}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:countries}",
            &std::iter::repeat_n("?", params.countries.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.cities.is_empty() {
        sql_template = sql_template.replace("{SLICE:cities}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:cities}",
            &std::iter::repeat_n("?", params.cities.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.billing.is_empty() {
        sql_template = sql_template.replace("{SLICE:billing}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:billing}",
            &std::iter::repeat_n("?", params.billing.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query
        .bind(params.business_id)
        .bind(params.filter_reps)
        .bind(params.unassigned)
        .bind(params.filter_sources)
        .bind(params.filter_industries)
        .bind(params.filter_sizes)
        .bind(params.filter_countries)
        .bind(params.filter_cities)
        .bind(params.filter_billing)
        .bind(params.import_batch_id)
        .bind(params.import_batch_id)
        .bind(params.follow_up_from)
        .bind(params.follow_up_from)
        .bind(params.follow_up_to)
        .bind(params.follow_up_to)
        .bind(params.no_follow_up)
        .bind(&params.has_phone)
        .bind(&params.has_phone)
        .bind(&params.has_email)
        .bind(&params.has_email)
        .bind(&params.has_website)
        .bind(&params.has_website)
        .bind(&params.has_linkedin)
        .bind(&params.has_linkedin)
        .bind(&params.has_company)
        .bind(&params.has_company)
        .bind(&params.has_value)
        .bind(&params.has_value)
        .bind(params.never_contacted)
        .bind(params.attempts_min)
        .bind(params.attempts_min)
        .bind(params.attempts_max)
        .bind(params.attempts_max)
        .bind(&params.value_min)
        .bind(&params.value_min)
        .bind(&params.value_max)
        .bind(&params.value_max)
        .bind(params.created_from)
        .bind(params.created_from)
        .bind(params.created_before)
        .bind(params.created_before)
        .bind(params.active_from)
        .bind(params.active_from)
        .bind(params.active_before)
        .bind(params.active_before)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search_digits)
        .bind(params.cursor_id)
        .bind(params.cursor_time)
        .bind(params.cursor_time)
        .bind(params.cursor_id)
        .bind(params.limit);
    for value in &params.statuses {
        query = query.bind(value);
    }
    for value in &params.reps {
        query = query.bind(value);
    }
    for value in &params.sources {
        query = query.bind(value);
    }
    for value in &params.industries {
        query = query.bind(value);
    }
    for value in &params.sizes {
        query = query.bind(value);
    }
    for value in &params.countries {
        query = query.bind(value);
    }
    for value in &params.cities {
        query = query.bind(value);
    }
    for value in &params.billing {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `ListLeadsNewest` (Go `ListLeadsNewestParams`).
pub struct ListLeadsNewestParams {
    pub business_id: i64,
    pub statuses: Vec<String>,
    pub filter_reps: i64,
    pub reps: Vec<Option<i64>>,
    pub unassigned: i64,
    pub filter_sources: i64,
    pub sources: Vec<String>,
    pub filter_industries: i64,
    pub industries: Vec<Option<String>>,
    pub filter_sizes: i64,
    pub sizes: Vec<Option<String>>,
    pub filter_countries: i64,
    pub countries: Vec<Option<String>>,
    pub filter_cities: i64,
    pub cities: Vec<Option<String>>,
    pub filter_billing: i64,
    pub billing: Vec<String>,
    pub import_batch_id: Option<i64>,
    pub follow_up_from: Option<chrono::DateTime<chrono::Utc>>,
    pub follow_up_to: Option<chrono::DateTime<chrono::Utc>>,
    pub no_follow_up: i64,
    pub has_phone: Option<String>,
    pub has_email: Option<String>,
    pub has_website: Option<String>,
    pub has_linkedin: Option<String>,
    pub has_company: Option<String>,
    pub has_value: Option<String>,
    pub never_contacted: i64,
    pub attempts_min: Option<i32>,
    pub attempts_max: Option<i32>,
    pub value_min: Option<String>,
    pub value_max: Option<String>,
    pub created_from: Option<chrono::DateTime<chrono::Utc>>,
    pub created_before: Option<chrono::DateTime<chrono::Utc>>,
    pub active_from: Option<chrono::DateTime<chrono::Utc>>,
    pub active_before: Option<chrono::DateTime<chrono::Utc>>,
    pub search: Option<String>,
    pub search_digits: Option<String>,
    pub cursor_id: Option<i64>,
    pub cursor_time: Option<chrono::DateTime<chrono::Utc>>,
    pub limit: i32,
}

/// `ListLeadsNewest`: `many`.
pub async fn list_leads_newest<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLeadsNewestParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT id, owner_id, business_id, first_name, last_name, job_title, email, phone, phone_digits, company_name, website, linkedin, industry, size, city, country, source, status, disqualify_reason, stage, assigned_to, value, billing, mrr_potential, notes, attempt_count, last_activity_at, next_follow_up, customer_id, company_id, deal_id, converted_at, import_batch_id, created_at, updated_at FROM crm_leads\nWHERE business_id = ?\n  AND status IN ({SLICE:statuses})\n  AND (CAST(? AS UNSIGNED) = 0 OR assigned_to IN ({SLICE:reps})\n       OR (CAST(? AS UNSIGNED) = 1 AND assigned_to IS NULL))\n  AND (CAST(? AS UNSIGNED) = 0 OR source IN ({SLICE:sources}))\n  AND (CAST(? AS UNSIGNED) = 0 OR industry IN ({SLICE:industries}))\n  AND (CAST(? AS UNSIGNED) = 0 OR size IN ({SLICE:sizes}))\n  AND (CAST(? AS UNSIGNED) = 0 OR country IN ({SLICE:countries}))\n  AND (CAST(? AS UNSIGNED) = 0 OR city IN ({SLICE:cities}))\n  AND (CAST(? AS UNSIGNED) = 0 OR billing IN ({SLICE:billing}))\n  AND (? IS NULL OR import_batch_id = ?)\n  AND (? IS NULL OR next_follow_up >= ?)\n  AND (? IS NULL OR next_follow_up <= ?)\n  AND (CAST(? AS UNSIGNED) = 0 OR next_follow_up IS NULL)\n  AND (? IS NULL OR IF(COALESCE(phone_digits, '') = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(COALESCE(email, '') = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(COALESCE(website, '') = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(COALESCE(linkedin, '') = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(company_name = '', 'no', 'yes') = ?)\n  AND (? IS NULL OR IF(COALESCE(value, 0) = 0, 'no', 'yes') = ?)\n  AND (CAST(? AS UNSIGNED) = 0 OR last_activity_at IS NULL)\n  AND (? IS NULL OR attempt_count >= ?)\n  AND (? IS NULL OR attempt_count <= ?)\n  AND (? IS NULL OR value >= ?)\n  AND (? IS NULL OR value <= ?)\n  AND (? IS NULL OR created_at >= ?)\n  AND (? IS NULL OR created_at < ?)\n  AND (? IS NULL OR last_activity_at >= ?)\n  AND (? IS NULL OR last_activity_at < ?)\n  AND (? IS NULL OR company_name LIKE ?\n       OR CONCAT_WS(' ', first_name, last_name) LIKE ?\n       OR email LIKE ? OR job_title LIKE ?\n       OR phone_digits LIKE ?)\n  AND (? IS NULL OR created_at < ?\n       OR (created_at = ? AND id < ?))\nORDER BY created_at DESC, id DESC\nLIMIT ?".to_string();

    if params.statuses.is_empty() {
        sql_template = sql_template.replace("{SLICE:statuses}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:statuses}",
            &std::iter::repeat_n("?", params.statuses.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.reps.is_empty() {
        sql_template = sql_template.replace("{SLICE:reps}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:reps}",
            &std::iter::repeat_n("?", params.reps.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.sources.is_empty() {
        sql_template = sql_template.replace("{SLICE:sources}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:sources}",
            &std::iter::repeat_n("?", params.sources.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.industries.is_empty() {
        sql_template = sql_template.replace("{SLICE:industries}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:industries}",
            &std::iter::repeat_n("?", params.industries.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.sizes.is_empty() {
        sql_template = sql_template.replace("{SLICE:sizes}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:sizes}",
            &std::iter::repeat_n("?", params.sizes.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.countries.is_empty() {
        sql_template = sql_template.replace("{SLICE:countries}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:countries}",
            &std::iter::repeat_n("?", params.countries.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.cities.is_empty() {
        sql_template = sql_template.replace("{SLICE:cities}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:cities}",
            &std::iter::repeat_n("?", params.cities.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    if params.billing.is_empty() {
        sql_template = sql_template.replace("{SLICE:billing}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:billing}",
            &std::iter::repeat_n("?", params.billing.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query
        .bind(params.business_id)
        .bind(params.filter_reps)
        .bind(params.unassigned)
        .bind(params.filter_sources)
        .bind(params.filter_industries)
        .bind(params.filter_sizes)
        .bind(params.filter_countries)
        .bind(params.filter_cities)
        .bind(params.filter_billing)
        .bind(params.import_batch_id)
        .bind(params.import_batch_id)
        .bind(params.follow_up_from)
        .bind(params.follow_up_from)
        .bind(params.follow_up_to)
        .bind(params.follow_up_to)
        .bind(params.no_follow_up)
        .bind(&params.has_phone)
        .bind(&params.has_phone)
        .bind(&params.has_email)
        .bind(&params.has_email)
        .bind(&params.has_website)
        .bind(&params.has_website)
        .bind(&params.has_linkedin)
        .bind(&params.has_linkedin)
        .bind(&params.has_company)
        .bind(&params.has_company)
        .bind(&params.has_value)
        .bind(&params.has_value)
        .bind(params.never_contacted)
        .bind(params.attempts_min)
        .bind(params.attempts_min)
        .bind(params.attempts_max)
        .bind(params.attempts_max)
        .bind(&params.value_min)
        .bind(&params.value_min)
        .bind(&params.value_max)
        .bind(&params.value_max)
        .bind(params.created_from)
        .bind(params.created_from)
        .bind(params.created_before)
        .bind(params.created_before)
        .bind(params.active_from)
        .bind(params.active_from)
        .bind(params.active_before)
        .bind(params.active_before)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search_digits)
        .bind(params.cursor_id)
        .bind(params.cursor_time)
        .bind(params.cursor_time)
        .bind(params.cursor_id)
        .bind(params.limit);
    for value in &params.statuses {
        query = query.bind(value);
    }
    for value in &params.reps {
        query = query.bind(value);
    }
    for value in &params.sources {
        query = query.bind(value);
    }
    for value in &params.industries {
        query = query.bind(value);
    }
    for value in &params.sizes {
        query = query.bind(value);
    }
    for value in &params.countries {
        query = query.bind(value);
    }
    for value in &params.cities {
        query = query.bind(value);
    }
    for value in &params.billing {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `ListOpenTasks` (Go `ListOpenTasksParams`).
pub struct ListOpenTasksParams {
    pub business_id: i64,
    pub assigned_to: Option<i64>,
    pub due_from: Option<chrono::DateTime<chrono::Utc>>,
    pub due_to: Option<chrono::DateTime<chrono::Utc>>,
    pub lead_id: Option<i64>,
    pub deal_id: Option<i64>,
    pub cursor_id: Option<i64>,
    pub cursor_date: Option<chrono::DateTime<chrono::Utc>>,
    pub limit: i32,
}

/// `ListOpenTasks`: `many`.
pub async fn list_open_tasks<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListOpenTasksParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT t.id, t.lead_id, t.deal_id, t.title, t.type, t.due_date, t.assigned_to, t.completed_at,\n       t.created_by, t.created_at,\n       COALESCE(l.first_name, '') AS lead_first_name, COALESCE(l.last_name, '') AS lead_last_name,\n       COALESCE(l.company_name, '') AS lead_company_name, COALESCE(l.phone, '') AS lead_phone,\n       COALESCE(d.title, '') AS deal_title\nFROM crm_tasks t\nLEFT JOIN crm_leads l ON l.id = t.lead_id AND l.business_id = t.business_id\nLEFT JOIN crm_deals d ON d.id = t.deal_id AND d.business_id = t.business_id\nWHERE t.business_id = ? AND t.completed_at IS NULL\n  AND (? IS NULL OR COALESCE(t.assigned_to, 0) = ?)\n  AND (? IS NULL OR t.due_date >= ?)\n  AND (? IS NULL OR t.due_date <= ?)\n  AND (? IS NULL OR t.lead_id = ?)\n  AND (? IS NULL OR t.deal_id = ?)\n  AND (? IS NULL OR t.due_date > ?\n       OR (t.due_date = ? AND t.id > ?))\nORDER BY t.due_date, t.id\nLIMIT ?")
        .bind(params.business_id)
        .bind(params.assigned_to)
        .bind(params.assigned_to)
        .bind(params.due_from)
        .bind(params.due_from)
        .bind(params.due_to)
        .bind(params.due_to)
        .bind(params.lead_id)
        .bind(params.lead_id)
        .bind(params.deal_id)
        .bind(params.deal_id)
        .bind(params.cursor_id)
        .bind(params.cursor_date)
        .bind(params.cursor_date)
        .bind(params.cursor_id)
        .bind(params.limit)
        .execute(exec)
        .await
}

/// Params for `ListPipelineStages` (Go `ListPipelineStagesParams`).
pub struct ListPipelineStagesParams {
    pub business_id: i64,
}

/// `ListPipelineStages`: `many`.
pub async fn list_pipeline_stages<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListPipelineStagesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, business_id, pipeline_id, name, position, probability, kind, created_at, updated_at FROM crm_pipeline_stages WHERE business_id = ? ORDER BY pipeline_id, position, id")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListPipelines` (Go `ListPipelinesParams`).
pub struct ListPipelinesParams {
    pub business_id: i64,
}

/// `ListPipelines`: `many`.
pub async fn list_pipelines<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListPipelinesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, business_id, name, position, default_marker, created_at, updated_at FROM crm_pipelines WHERE business_id = ? ORDER BY position, id")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListSignedContracts` (Go `ListSignedContractsParams`).
pub struct ListSignedContractsParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListSignedContracts`: `many`.
pub async fn list_signed_contracts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListSignedContractsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, lead_id, deal_id, plan, mrr, seats, value, billing, start_date, end_date, customer_id, signed_at, signed_by_name, closed_by, notes, created_at, updated_at FROM crm_contracts\nWHERE business_id = ? AND signed_at IS NOT NULL\nORDER BY created_at DESC, id DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListStageDeals` (Go `ListStageDealsParams`).
pub struct ListStageDealsParams {
    pub business_id: i64,
    pub pipeline_id: i64,
    pub stage_id: i64,
    pub assigned_to: Option<i64>,
    pub search: Option<String>,
    pub cursor_id: Option<i64>,
    pub cursor_time: Option<chrono::DateTime<chrono::Utc>>,
    pub limit: i32,
}

/// `ListStageDeals`: `many`.
pub async fn list_stage_deals<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListStageDealsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, business_id, owner_id, pipeline_id, stage_id, title, lead_id, customer_id, company_id, contact_name, company_name, value, billing, expected_close_date, assigned_to, source, status, lost_reason, closed_at, notes, last_activity_at, next_follow_up, stage_changed_at, created_at, updated_at FROM crm_deals\nWHERE business_id = ? AND pipeline_id = ?\n  AND stage_id = ?\n  AND (? IS NULL OR COALESCE(assigned_to, 0) = ?)\n  AND (? IS NULL OR title LIKE ?\n       OR company_name LIKE ? OR contact_name LIKE ?)\n  AND (? IS NULL OR created_at < ?\n       OR (created_at = ? AND id < ?))\nORDER BY created_at DESC, id DESC\nLIMIT ?")
        .bind(params.business_id)
        .bind(params.pipeline_id)
        .bind(params.stage_id)
        .bind(params.assigned_to)
        .bind(params.assigned_to)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(&params.search)
        .bind(params.cursor_id)
        .bind(params.cursor_time)
        .bind(params.cursor_time)
        .bind(params.cursor_id)
        .bind(params.limit)
        .execute(exec)
        .await
}

/// Params for `ListUnsignedContracts` (Go `ListUnsignedContractsParams`).
pub struct ListUnsignedContractsParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListUnsignedContracts`: `many`.
pub async fn list_unsigned_contracts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListUnsignedContractsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, lead_id, deal_id, plan, mrr, seats, value, billing, start_date, end_date, customer_id, signed_at, signed_by_name, closed_by, notes, created_at, updated_at FROM crm_contracts\nWHERE business_id = ? AND signed_at IS NULL\nORDER BY created_at DESC, id DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `MarkLeadConverted` (Go `MarkLeadConvertedParams`).
pub struct MarkLeadConvertedParams {
    pub company_id: Option<i64>,
    pub deal_id: Option<i64>,
    pub converted_at: Option<chrono::DateTime<chrono::Utc>>,
    pub id: i64,
    pub business_id: i64,
}

/// `MarkLeadConverted`: `exec`.
pub async fn mark_lead_converted<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MarkLeadConvertedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_leads\nSET status = 'qualified', disqualify_reason = NULL, company_id = ?, deal_id = ?, converted_at = ?\nWHERE id = ? AND business_id = ?")
        .bind(params.company_id)
        .bind(params.deal_id)
        .bind(params.converted_at)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `MoveDeal` (Go `MoveDealParams`).
pub struct MoveDealParams {
    pub pipeline_id: i64,
    pub stage_id: i64,
    pub status: String,
    pub closed_at: Option<chrono::DateTime<chrono::Utc>>,
    pub lost_reason: Option<String>,
    pub stage_changed_at: chrono::NaiveDate,
    pub id: i64,
    pub business_id: i64,
}

/// `MoveDeal`: `exec`.
pub async fn move_deal<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MoveDealParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_deals\nSET pipeline_id = ?, stage_id = ?, status = ?, closed_at = ?, lost_reason = ?, stage_changed_at = ?\nWHERE id = ? AND business_id = ?")
        .bind(params.pipeline_id)
        .bind(params.stage_id)
        .bind(&params.status)
        .bind(params.closed_at)
        .bind(&params.lost_reason)
        .bind(params.stage_changed_at)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `RecordDealWork` (Go `RecordDealWorkParams`).
pub struct RecordDealWorkParams {
    pub last_activity_at: Option<chrono::DateTime<chrono::Utc>>,
    pub id: i64,
    pub business_id: i64,
}

/// `RecordDealWork`: `exec`.
pub async fn record_deal_work<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RecordDealWorkParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_deals SET last_activity_at = ? WHERE id = ? AND business_id = ?")
        .bind(params.last_activity_at)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `RecordLeadWork` (Go `RecordLeadWorkParams`).
pub struct RecordLeadWorkParams {
    pub status: String,
    pub last_activity_at: Option<chrono::DateTime<chrono::Utc>>,
    pub attempt_count: i32,
    pub id: i64,
    pub business_id: i64,
}

/// `RecordLeadWork`: `exec`.
pub async fn record_lead_work<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RecordLeadWorkParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_leads\nSET status = ?, last_activity_at = ?, attempt_count = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.status)
        .bind(params.last_activity_at)
        .bind(params.attempt_count)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `RenamePipeline` (Go `RenamePipelineParams`).
pub struct RenamePipelineParams {
    pub name: String,
    pub id: i64,
    pub business_id: i64,
}

/// `RenamePipeline`: `exec`.
pub async fn rename_pipeline<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RenamePipelineParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_pipelines SET name = ? WHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `SetDealNextFollowUp` (Go `SetDealNextFollowUpParams`).
pub struct SetDealNextFollowUpParams {
    pub id: i64,
    pub business_id: i64,
}

/// `SetDealNextFollowUp`: `exec`.
pub async fn set_deal_next_follow_up<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SetDealNextFollowUpParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_deals\nSET next_follow_up = (\n    SELECT MIN(t.due_date) FROM crm_tasks t\n    WHERE t.business_id = crm_deals.business_id AND t.deal_id = crm_deals.id AND t.completed_at IS NULL\n)\nWHERE crm_deals.id = ? AND crm_deals.business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `SetImportBatchStatus` (Go `SetImportBatchStatusParams`).
pub struct SetImportBatchStatusParams {
    pub status: String,
    pub completed_at: Option<chrono::DateTime<chrono::Utc>>,
    pub id: i64,
    pub business_id: i64,
}

/// `SetImportBatchStatus`: `exec`.
pub async fn set_import_batch_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SetImportBatchStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_import_batches SET status = ?, completed_at = ? WHERE id = ? AND business_id = ?")
        .bind(&params.status)
        .bind(params.completed_at)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `SetLeadNextFollowUp` (Go `SetLeadNextFollowUpParams`).
pub struct SetLeadNextFollowUpParams {
    pub id: i64,
    pub business_id: i64,
}

/// `SetLeadNextFollowUp`: `exec`.
pub async fn set_lead_next_follow_up<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SetLeadNextFollowUpParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_leads\nSET next_follow_up = (\n    SELECT MIN(t.due_date) FROM crm_tasks t\n    WHERE t.business_id = crm_leads.business_id AND t.lead_id = crm_leads.id AND t.completed_at IS NULL\n)\nWHERE crm_leads.id = ? AND crm_leads.business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `SetLeadsStatus` (Go `SetLeadsStatusParams`).
pub struct SetLeadsStatusParams {
    pub status: String,
    pub disqualify_reason: Option<String>,
    pub business_id: i64,
    pub ids: Vec<i64>,
}

/// `SetLeadsStatus`: `execresult`.
pub async fn set_leads_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SetLeadsStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "UPDATE crm_leads SET status = ?, disqualify_reason = ?\nWHERE business_id = ? AND id IN ({SLICE:ids}) AND status <> 'qualified'".to_string();

    if params.ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:ids}",
            &std::iter::repeat_n("?", params.ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query
        .bind(&params.status)
        .bind(&params.disqualify_reason)
        .bind(params.business_id);
    for value in &params.ids {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `SetTaskCompletedAt` (Go `SetTaskCompletedAtParams`).
pub struct SetTaskCompletedAtParams {
    pub completed_at: Option<chrono::DateTime<chrono::Utc>>,
    pub id: i64,
    pub business_id: i64,
}

/// `SetTaskCompletedAt`: `exec`.
pub async fn set_task_completed_at<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SetTaskCompletedAtParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_tasks SET completed_at = ? WHERE id = ? AND business_id = ?")
        .bind(params.completed_at)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateContract` (Go `UpdateContractParams`).
pub struct UpdateContractParams {
    pub plan: String,
    pub mrr: String,
    pub seats: i32,
    pub value: String,
    pub billing: String,
    pub start_date: Option<chrono::DateTime<chrono::Utc>>,
    pub end_date: Option<chrono::DateTime<chrono::Utc>>,
    pub customer_id: Option<i64>,
    pub signed_at: Option<chrono::DateTime<chrono::Utc>>,
    pub signed_by_name: Option<String>,
    pub closed_by: Option<i64>,
    pub notes: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateContract`: `exec`.
pub async fn update_contract<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateContractParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_contracts\nSET plan = ?, mrr = ?, seats = ?, value = ?, billing = ?, start_date = ?, end_date = ?,\n    customer_id = ?, signed_at = ?, signed_by_name = ?, closed_by = ?, notes = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.plan)
        .bind(&params.mrr)
        .bind(params.seats)
        .bind(&params.value)
        .bind(&params.billing)
        .bind(params.start_date)
        .bind(params.end_date)
        .bind(params.customer_id)
        .bind(params.signed_at)
        .bind(&params.signed_by_name)
        .bind(params.closed_by)
        .bind(&params.notes)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateDeal` (Go `UpdateDealParams`).
pub struct UpdateDealParams {
    pub title: String,
    pub contact_name: String,
    pub company_name: String,
    pub value: String,
    pub billing: String,
    pub expected_close_date: Option<chrono::DateTime<chrono::Utc>>,
    pub assigned_to: Option<i64>,
    pub source: String,
    pub notes: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateDeal`: `exec`.
pub async fn update_deal<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateDealParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_deals\nSET title = ?, contact_name = ?, company_name = ?, value = ?, billing = ?, expected_close_date = ?,\n    assigned_to = ?, source = ?, notes = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.title)
        .bind(&params.contact_name)
        .bind(&params.company_name)
        .bind(&params.value)
        .bind(&params.billing)
        .bind(params.expected_close_date)
        .bind(params.assigned_to)
        .bind(&params.source)
        .bind(&params.notes)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateLead` (Go `UpdateLeadParams`).
pub struct UpdateLeadParams {
    pub first_name: Option<String>,
    pub last_name: Option<String>,
    pub job_title: Option<String>,
    pub email: Option<String>,
    pub phone: Option<String>,
    pub phone_digits: Option<String>,
    pub company_name: String,
    pub website: Option<String>,
    pub linkedin: Option<String>,
    pub industry: Option<String>,
    pub size: Option<String>,
    pub city: Option<String>,
    pub country: Option<String>,
    pub source: String,
    pub status: String,
    pub disqualify_reason: Option<String>,
    pub assigned_to: Option<i64>,
    pub value: Option<String>,
    pub billing: String,
    pub notes: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateLead`: `exec`.
pub async fn update_lead<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateLeadParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_leads\nSET first_name = ?, last_name = ?, job_title = ?, email = ?, phone = ?, phone_digits = ?,\n    company_name = ?, website = ?, linkedin = ?, industry = ?, size = ?, city = ?, country = ?,\n    source = ?, status = ?, disqualify_reason = ?, assigned_to = ?, value = ?, billing = ?, notes = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.first_name)
        .bind(&params.last_name)
        .bind(&params.job_title)
        .bind(&params.email)
        .bind(&params.phone)
        .bind(&params.phone_digits)
        .bind(&params.company_name)
        .bind(&params.website)
        .bind(&params.linkedin)
        .bind(&params.industry)
        .bind(&params.size)
        .bind(&params.city)
        .bind(&params.country)
        .bind(&params.source)
        .bind(&params.status)
        .bind(&params.disqualify_reason)
        .bind(params.assigned_to)
        .bind(&params.value)
        .bind(&params.billing)
        .bind(&params.notes)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateStage` (Go `UpdateStageParams`).
pub struct UpdateStageParams {
    pub name: String,
    pub position: i32,
    pub probability: u8,
    pub id: i64,
    pub business_id: i64,
    pub pipeline_id: i64,
}

/// `UpdateStage`: `exec`.
pub async fn update_stage<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateStageParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_pipeline_stages SET name = ?, position = ?, probability = ?\nWHERE id = ? AND business_id = ? AND pipeline_id = ?")
        .bind(&params.name)
        .bind(params.position)
        .bind(params.probability)
        .bind(params.id)
        .bind(params.business_id)
        .bind(params.pipeline_id)
        .execute(exec)
        .await
}

/// Params for `UpdateTask` (Go `UpdateTaskParams`).
pub struct UpdateTaskParams {
    pub title: String,
    pub r#type: String,
    pub due_date: chrono::NaiveDate,
    pub assigned_to: Option<i64>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateTask`: `exec`.
pub async fn update_task<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateTaskParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_tasks SET title = ?, type = ?, due_date = ?, assigned_to = ? WHERE id = ? AND business_id = ?")
        .bind(&params.title)
        .bind(&params.r#type)
        .bind(params.due_date)
        .bind(params.assigned_to)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
