//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountActiveCustomers` (Go `CountActiveCustomersParams`).
pub struct CountActiveCustomersParams {
    pub business_id: i64,
}

/// `CountActiveCustomers`: `one`.
pub async fn count_active_customers<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountActiveCustomersParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM crm_customers WHERE business_id = ? AND is_active = TRUE")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountActiveLoyaltyMembers` (Go `CountActiveLoyaltyMembersParams`).
pub struct CountActiveLoyaltyMembersParams {
    pub business_id: i64,
}

/// `CountActiveLoyaltyMembers`: `one`.
pub async fn count_active_loyalty_members<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountActiveLoyaltyMembersParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM crm_customers WHERE business_id = ? AND loyalty_points > 0 AND is_active = TRUE")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountActiveMembersOnTier` (Go `CountActiveMembersOnTierParams`).
pub struct CountActiveMembersOnTierParams {
    pub business_id: i64,
    pub loyalty_tier: String,
}

/// `CountActiveMembersOnTier`: `one`.
pub async fn count_active_members_on_tier<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountActiveMembersOnTierParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM crm_customers\nWHERE business_id = ? AND is_active = TRUE AND loyalty_tier = ?")
        .bind(params.business_id)
        .bind(&params.loyalty_tier)
        .fetch_optional(exec)
        .await
}

/// Params for `CountAllLoyaltyTransactions` (Go `CountAllLoyaltyTransactionsParams`).
pub struct CountAllLoyaltyTransactionsParams {
    pub business_id: i64,
}

/// `CountAllLoyaltyTransactions`: `one`.
pub async fn count_all_loyalty_transactions<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountAllLoyaltyTransactionsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM crm_loyalty_transactions WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountCustomers` (Go `CountCustomersParams`).
pub struct CountCustomersParams {
    pub business_id: i64,
}

/// `CountCustomers`: `one`.
pub async fn count_customers<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountCustomersParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM crm_customers WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountCustomersCreatedBetween` (Go `CountCustomersCreatedBetweenParams`).
pub struct CountCustomersCreatedBetweenParams {
    pub business_id: i64,
    pub date_from: Option<chrono::DateTime<chrono::Utc>>,
    pub date_to: Option<chrono::DateTime<chrono::Utc>>,
}

/// `CountCustomersCreatedBetween`: `one`.
pub async fn count_customers_created_between<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountCustomersCreatedBetweenParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*)\nFROM crm_customers\nWHERE business_id = ?\n  AND created_at >= ?\n  AND created_at <= ?")
        .bind(params.business_id)
        .bind(params.date_from)
        .bind(params.date_to)
        .fetch_optional(exec)
        .await
}

/// Params for `CountLoyaltyTiers` (Go `CountLoyaltyTiersParams`).
pub struct CountLoyaltyTiersParams {
    pub business_id: i64,
}

/// `CountLoyaltyTiers`: `one`.
pub async fn count_loyalty_tiers<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountLoyaltyTiersParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM crm_loyalty_tiers WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountLoyaltyTransactions` (Go `CountLoyaltyTransactionsParams`).
pub struct CountLoyaltyTransactionsParams {
    pub business_id: i64,
    pub customer_id: i64,
}

/// `CountLoyaltyTransactions`: `one`.
pub async fn count_loyalty_transactions<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountLoyaltyTransactionsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT COUNT(*) FROM crm_loyalty_transactions WHERE business_id = ? AND customer_id = ?",
    )
    .bind(params.business_id)
    .bind(params.customer_id)
    .fetch_optional(exec)
    .await
}

/// Params for `CountNewCustomersThisMonth` (Go `CountNewCustomersThisMonthParams`).
pub struct CountNewCustomersThisMonthParams {
    pub business_id: i64,
    pub created_at_start: Option<chrono::DateTime<chrono::Utc>>,
    pub created_at_end: Option<chrono::DateTime<chrono::Utc>>,
}

/// `CountNewCustomersThisMonth`: `one`.
pub async fn count_new_customers_this_month<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountNewCustomersThisMonthParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM crm_customers\nWHERE business_id = ?\n  AND created_at >= ?\n  AND created_at < ?")
        .bind(params.business_id)
        .bind(params.created_at_start)
        .bind(params.created_at_end)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateCustomer` (Go `CreateCustomerParams`).
pub struct CreateCustomerParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub first_name: String,
    pub last_name: Option<String>,
    pub email: Option<String>,
    pub phone: Option<String>,
    pub date_of_birth: Option<chrono::DateTime<chrono::Utc>>,
    pub gender: String,
    pub address: Option<String>,
    pub city: Option<String>,
    pub country: Option<String>,
    pub tags: String,
    pub notes: Option<String>,
    pub consent: String,
}

/// `CreateCustomer`: `execresult`.
pub async fn create_customer<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateCustomerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_customers\n  (owner_id, business_id, first_name, last_name, email, phone, date_of_birth, gender,\n\t   address, city, country, tags, notes, consent, loyalty_tier, is_active)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'bronze', TRUE)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.first_name)
        .bind(&params.last_name)
        .bind(&params.email)
        .bind(&params.phone)
        .bind(params.date_of_birth)
        .bind(&params.gender)
        .bind(&params.address)
        .bind(&params.city)
        .bind(&params.country)
        .bind(&params.tags)
        .bind(&params.notes)
        .bind(&params.consent)
        .execute(exec)
        .await
}

/// Params for `CreateLoyaltyTier` (Go `CreateLoyaltyTierParams`).
pub struct CreateLoyaltyTierParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub min_points: Option<i32>,
    pub discount_pct: Option<String>,
    pub perks: String,
    pub sort_order: Option<i32>,
}

/// `CreateLoyaltyTier`: `execresult`.
pub async fn create_loyalty_tier<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateLoyaltyTierParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_loyalty_tiers (owner_id, business_id, name, min_points, discount_pct, perks, sort_order)\nVALUES (?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(params.min_points)
        .bind(&params.discount_pct)
        .bind(&params.perks)
        .bind(params.sort_order)
        .execute(exec)
        .await
}

/// Params for `CreateLoyaltyTransaction` (Go `CreateLoyaltyTransactionParams`).
pub struct CreateLoyaltyTransactionParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub customer_id: i64,
    pub r#type: String,
    pub points: i32,
    pub balance_before: i32,
    pub balance_after: i32,
    pub description: Option<String>,
    pub reference_id: Option<i64>,
    pub reference_type: Option<String>,
}

/// `CreateLoyaltyTransaction`: `exec`.
pub async fn create_loyalty_transaction<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateLoyaltyTransactionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_loyalty_transactions\n  (owner_id, business_id, customer_id, type, points, balance_before, balance_after, description, reference_id, reference_type)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.customer_id)
        .bind(&params.r#type)
        .bind(params.points)
        .bind(params.balance_before)
        .bind(params.balance_after)
        .bind(&params.description)
        .bind(params.reference_id)
        .bind(&params.reference_type)
        .execute(exec)
        .await
}

/// Params for `DeleteCustomer` (Go `DeleteCustomerParams`).
pub struct DeleteCustomerParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteCustomer`: `execresult`.
pub async fn delete_customer<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteCustomerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM crm_customers WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `DeleteCustomerCampaignSends` (Go `DeleteCustomerCampaignSendsParams`).
pub struct DeleteCustomerCampaignSendsParams {
    pub business_id: i64,
    pub customer_id: i64,
}

/// `DeleteCustomerCampaignSends`: `exec`.
pub async fn delete_customer_campaign_sends<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteCustomerCampaignSendsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM crm_campaign_sends WHERE business_id = ? AND customer_id = ?")
        .bind(params.business_id)
        .bind(params.customer_id)
        .execute(exec)
        .await
}

/// Params for `DeleteCustomerFeedback` (Go `DeleteCustomerFeedbackParams`).
pub struct DeleteCustomerFeedbackParams {
    pub business_id: i64,
    pub customer_id: i64,
}

/// `DeleteCustomerFeedback`: `exec`.
pub async fn delete_customer_feedback<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteCustomerFeedbackParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM crm_feedback WHERE business_id = ? AND customer_id = ?")
        .bind(params.business_id)
        .bind(params.customer_id)
        .execute(exec)
        .await
}

/// Params for `DeleteCustomerLoyaltyTransactions` (Go `DeleteCustomerLoyaltyTransactionsParams`).
pub struct DeleteCustomerLoyaltyTransactionsParams {
    pub business_id: i64,
    pub customer_id: i64,
}

/// `DeleteCustomerLoyaltyTransactions`: `exec`.
pub async fn delete_customer_loyalty_transactions<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteCustomerLoyaltyTransactionsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM crm_loyalty_transactions WHERE business_id = ? AND customer_id = ?")
        .bind(params.business_id)
        .bind(params.customer_id)
        .execute(exec)
        .await
}

/// Params for `DeleteCustomerTimeline` (Go `DeleteCustomerTimelineParams`).
pub struct DeleteCustomerTimelineParams {
    pub business_id: i64,
    pub customer_id: i64,
}

/// `DeleteCustomerTimeline`: `exec`.
pub async fn delete_customer_timeline<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteCustomerTimelineParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM crm_customer_timeline WHERE business_id = ? AND customer_id = ?")
        .bind(params.business_id)
        .bind(params.customer_id)
        .execute(exec)
        .await
}

/// Params for `DeleteLoyaltyTier` (Go `DeleteLoyaltyTierParams`).
pub struct DeleteLoyaltyTierParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteLoyaltyTier`: `execresult`.
pub async fn delete_loyalty_tier<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteLoyaltyTierParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM crm_loyalty_tiers WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `DetachCustomerFromActivities` (Go `DetachCustomerFromActivitiesParams`).
pub struct DetachCustomerFromActivitiesParams {
    pub business_id: i64,
    pub customer_id: Option<i64>,
}

/// `DetachCustomerFromActivities`: `exec`.
pub async fn detach_customer_from_activities<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DetachCustomerFromActivitiesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE crm_touchpoints SET customer_id = NULL WHERE business_id = ? AND customer_id = ?",
    )
    .bind(params.business_id)
    .bind(params.customer_id)
    .execute(exec)
    .await
}

/// Params for `DetachCustomerFromContacts` (Go `DetachCustomerFromContactsParams`).
pub struct DetachCustomerFromContactsParams {
    pub business_id: i64,
    pub customer_id: Option<i64>,
}

/// `DetachCustomerFromContacts`: `exec`.
pub async fn detach_customer_from_contacts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DetachCustomerFromContactsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE crm_contacts SET customer_id = NULL WHERE business_id = ? AND customer_id = ?",
    )
    .bind(params.business_id)
    .bind(params.customer_id)
    .execute(exec)
    .await
}

/// Params for `DetachCustomerFromContracts` (Go `DetachCustomerFromContractsParams`).
pub struct DetachCustomerFromContractsParams {
    pub business_id: i64,
    pub customer_id: Option<i64>,
}

/// `DetachCustomerFromContracts`: `exec`.
pub async fn detach_customer_from_contracts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DetachCustomerFromContractsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE crm_contracts SET customer_id = NULL WHERE business_id = ? AND customer_id = ?",
    )
    .bind(params.business_id)
    .bind(params.customer_id)
    .execute(exec)
    .await
}

/// Params for `DetachCustomerFromDeals` (Go `DetachCustomerFromDealsParams`).
pub struct DetachCustomerFromDealsParams {
    pub business_id: i64,
    pub customer_id: Option<i64>,
}

/// `DetachCustomerFromDeals`: `exec`.
pub async fn detach_customer_from_deals<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DetachCustomerFromDealsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_deals SET customer_id = NULL WHERE business_id = ? AND customer_id = ?")
        .bind(params.business_id)
        .bind(params.customer_id)
        .execute(exec)
        .await
}

/// Params for `DetachCustomerFromLeads` (Go `DetachCustomerFromLeadsParams`).
pub struct DetachCustomerFromLeadsParams {
    pub business_id: i64,
    pub customer_id: Option<i64>,
}

/// `DetachCustomerFromLeads`: `exec`.
pub async fn detach_customer_from_leads<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DetachCustomerFromLeadsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_leads SET customer_id = NULL WHERE business_id = ? AND customer_id = ?")
        .bind(params.business_id)
        .bind(params.customer_id)
        .execute(exec)
        .await
}

/// Params for `FindCustomerForCapture` (Go `FindCustomerForCaptureParams`).
pub struct FindCustomerForCaptureParams {
    pub business_id: i64,
    pub email: String,
    pub phone_input: Option<String>,
}

/// `FindCustomerForCapture`: `one`.
pub async fn find_customer_for_capture<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &FindCustomerForCaptureParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM crm_customers\nWHERE business_id = ?\n  AND ((? <> '' AND LOWER(email) = LOWER(?))\n    OR (? <> '' AND phone = ?))\nORDER BY CASE WHEN ? <> '' AND LOWER(email) = LOWER(?) THEN 0 ELSE 1 END\nLIMIT 1")
        .bind(params.business_id)
        .bind(&params.email)
        .bind(&params.email)
        .bind(&params.phone_input)
        .bind(&params.phone_input)
        .bind(&params.email)
        .bind(&params.email)
        .fetch_optional(exec)
        .await
}

/// Params for `GetCampaignSendStats` (Go `GetCampaignSendStatsParams`).
pub struct GetCampaignSendStatsParams {
    pub campaign_id: i64,
    pub business_id: i64,
}

/// `GetCampaignSendStats`: `one`.
pub async fn get_campaign_send_stats<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetCampaignSendStatsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    COALESCE(SUM(CASE WHEN status = 'bounced' THEN 1 ELSE 0 END), 0) AS bounced_count,\n    COALESCE(SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END), 0) AS failed_count\nFROM crm_campaign_sends WHERE campaign_id = ? AND business_id = ?")
        .bind(params.campaign_id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetCustomerByID` (Go `GetCustomerByIDParams`).
pub struct GetCustomerByIDParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetCustomerByID`: `one`.
pub async fn get_customer_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetCustomerByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, first_name, last_name, email, phone, date_of_birth, gender,\n       address, city, country, loyalty_points, total_spent, visit_count, last_visit_at,\n       loyalty_tier, tags, notes, COALESCE(consent, JSON_OBJECT()) AS consent, is_active, created_at, updated_at\nFROM crm_customers WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetCustomerTimeline` (Go `GetCustomerTimelineParams`).
pub struct GetCustomerTimelineParams {
    pub business_id: i64,
    pub customer_id: i64,
}

/// `GetCustomerTimeline`: `many`.
pub async fn get_customer_timeline<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetCustomerTimelineParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, customer_id, type, description, COALESCE(metadata, JSON_OBJECT()) AS metadata, created_at\nFROM crm_customer_timeline\nWHERE business_id = ? AND customer_id = ?\nORDER BY created_at DESC\nLIMIT 100")
        .bind(params.business_id)
        .bind(params.customer_id)
        .execute(exec)
        .await
}

/// Params for `GetLoyaltyConfig` (Go `GetLoyaltyConfigParams`).
pub struct GetLoyaltyConfigParams {
    pub business_id: i64,
}

/// `GetLoyaltyConfig`: `one`.
pub async fn get_loyalty_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetLoyaltyConfigParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, is_enabled, points_per_currency_unit, currency_per_point,\n       bronze_min, silver_min, gold_min, platinum_min,\n       bronze_discount_pct, silver_discount_pct, gold_discount_pct, platinum_discount_pct,\n       points_expiry_days, updated_at\nFROM crm_loyalty_config WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetLoyaltyTierByID` (Go `GetLoyaltyTierByIDParams`).
pub struct GetLoyaltyTierByIDParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetLoyaltyTierByID`: `one`.
pub async fn get_loyalty_tier_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetLoyaltyTierByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, min_points, discount_pct, perks, sort_order, created_at, updated_at\nFROM crm_loyalty_tiers WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsertCustomerFromCapture` (Go `InsertCustomerFromCaptureParams`).
pub struct InsertCustomerFromCaptureParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub first_name: String,
    pub last_name: Option<String>,
    pub email: Option<String>,
    pub phone: Option<String>,
    pub total_spent: Option<String>,
    pub visit_count: Option<i32>,
    pub last_visit_at: Option<chrono::DateTime<chrono::Utc>>,
}

/// `InsertCustomerFromCapture`: `execresult`.
pub async fn insert_customer_from_capture<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertCustomerFromCaptureParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_customers\n  (owner_id, business_id, first_name, last_name, email, phone, total_spent, visit_count, last_visit_at, loyalty_tier, tags, is_active)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'bronze', JSON_ARRAY(), TRUE)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.first_name)
        .bind(&params.last_name)
        .bind(&params.email)
        .bind(&params.phone)
        .bind(&params.total_spent)
        .bind(params.visit_count)
        .bind(params.last_visit_at)
        .execute(exec)
        .await
}

/// Params for `InsertTimelineEntry` (Go `InsertTimelineEntryParams`).
pub struct InsertTimelineEntryParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub customer_id: i64,
    pub r#type: String,
    pub description: Option<String>,
    pub metadata: String,
}

/// `InsertTimelineEntry`: `exec`.
pub async fn insert_timeline_entry<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertTimelineEntryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_customer_timeline (owner_id, business_id, customer_id, type, description, metadata)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.customer_id)
        .bind(&params.r#type)
        .bind(&params.description)
        .bind(&params.metadata)
        .execute(exec)
        .await
}

/// Params for `InsightsCRMLastChange` (Go `InsightsCRMLastChangeParams`).
pub struct InsightsCRMLastChangeParams {
    pub business_id: i64,
}

/// `InsightsCRMLastChange`: `one`.
pub async fn insights_crm_last_change<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsCRMLastChangeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT CAST(COALESCE(MAX(updated_at), '1000-01-01') AS DATETIME) AS last_change\nFROM crm_customers\nWHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsCountCustomers` (Go `InsightsCountCustomersParams`).
pub struct InsightsCountCustomersParams {
    pub business_id: i64,
}

/// `InsightsCountCustomers`: `one`.
pub async fn insights_count_customers<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsCountCustomersParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS count FROM crm_customers WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsCountCustomersAsOf` (Go `InsightsCountCustomersAsOfParams`).
pub struct InsightsCountCustomersAsOfParams {
    pub business_id: i64,
    pub as_of: chrono::NaiveDate,
}

/// `InsightsCountCustomersAsOf`: `one`.
pub async fn insights_count_customers_as_of<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsCountCustomersAsOfParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS count\nFROM crm_customers\nWHERE business_id = ?\n  AND created_at < DATE_ADD(?, INTERVAL 1 DAY)")
        .bind(params.business_id)
        .bind(params.as_of)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsCountNewCustomers` (Go `InsightsCountNewCustomersParams`).
pub struct InsightsCountNewCustomersParams {
    pub business_id: i64,
    pub date_from: Option<chrono::DateTime<chrono::Utc>>,
    pub date_to: chrono::NaiveDate,
}

/// `InsightsCountNewCustomers`: `one`.
pub async fn insights_count_new_customers<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsCountNewCustomersParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS count\nFROM crm_customers\nWHERE business_id = ?\n  AND created_at >= ?\n  AND created_at < DATE_ADD(?, INTERVAL 1 DAY)")
        .bind(params.business_id)
        .bind(params.date_from)
        .bind(params.date_to)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsCustomerContacts` (Go `InsightsCustomerContactsParams`).
pub struct InsightsCustomerContactsParams {
    pub business_id: i64,
    pub ids: Vec<i64>,
}

/// `InsightsCustomerContacts`: `many`.
pub async fn insights_customer_contacts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsCustomerContactsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT id,\n       COALESCE(first_name, '') AS first_name,\n       COALESCE(last_name, '') AS last_name,\n       COALESCE(email, '') AS email\nFROM crm_customers\nWHERE business_id = ?\n  AND id IN ({SLICE:ids})".to_string();

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

/// Params for `ListAllLoyaltyTransactions` (Go `ListAllLoyaltyTransactionsParams`).
pub struct ListAllLoyaltyTransactionsParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListAllLoyaltyTransactions`: `many`.
pub async fn list_all_loyalty_transactions<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListAllLoyaltyTransactionsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, customer_id, type, points, balance_before, balance_after,\n       description, reference_id, reference_type, created_at\nFROM crm_loyalty_transactions\nWHERE business_id = ?\nORDER BY created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListCustomerTimelinePage` (Go `ListCustomerTimelinePageParams`).
pub struct ListCustomerTimelinePageParams {
    pub business_id: i64,
    pub customer_id: i64,
    pub cursor_created_at: Option<chrono::DateTime<chrono::Utc>>,
    pub cursor_id: i64,
    pub limit: i32,
}

/// `ListCustomerTimelinePage`: `many`.
pub async fn list_customer_timeline_page<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListCustomerTimelinePageParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, customer_id, type, description, COALESCE(metadata, JSON_OBJECT()) AS metadata, created_at\nFROM crm_customer_timeline\nWHERE business_id = ?\n  AND customer_id = ?\n  AND (\n    ? IS NULL\n    OR created_at < ?\n    OR (created_at = ? AND id < ?)\n  )\nORDER BY created_at DESC, id DESC\nLIMIT ?")
        .bind(params.business_id)
        .bind(params.customer_id)
        .bind(params.cursor_created_at)
        .bind(params.cursor_created_at)
        .bind(params.cursor_created_at)
        .bind(params.cursor_id)
        .bind(params.limit)
        .execute(exec)
        .await
}

/// Params for `ListCustomers` (Go `ListCustomersParams`).
pub struct ListCustomersParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListCustomers`: `many`.
pub async fn list_customers<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListCustomersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, first_name, last_name, email, phone, date_of_birth, gender,\n       address, city, country, loyalty_points, total_spent, visit_count, last_visit_at,\n       loyalty_tier, tags, notes, COALESCE(consent, JSON_OBJECT()) AS consent, is_active, created_at, updated_at\nFROM crm_customers WHERE business_id = ?\nORDER BY created_at DESC LIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListLoyaltyTiers` (Go `ListLoyaltyTiersParams`).
pub struct ListLoyaltyTiersParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListLoyaltyTiers`: `many`.
pub async fn list_loyalty_tiers<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLoyaltyTiersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, min_points, discount_pct, perks, sort_order, created_at, updated_at\nFROM crm_loyalty_tiers WHERE business_id = ?\nORDER BY sort_order ASC, id ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListLoyaltyTransactions` (Go `ListLoyaltyTransactionsParams`).
pub struct ListLoyaltyTransactionsParams {
    pub business_id: i64,
    pub customer_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListLoyaltyTransactions`: `many`.
pub async fn list_loyalty_transactions<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLoyaltyTransactionsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, customer_id, type, points, balance_before, balance_after,\n       description, reference_id, reference_type, created_at\nFROM crm_loyalty_transactions\nWHERE business_id = ? AND customer_id = ?\nORDER BY created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.customer_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListTopCustomers` (Go `ListTopCustomersParams`).
pub struct ListTopCustomersParams {
    pub business_id: i64,
}

/// `ListTopCustomers`: `many`.
pub async fn list_top_customers<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListTopCustomersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, first_name, last_name, email, total_spent, loyalty_points, loyalty_tier\nFROM crm_customers WHERE business_id = ? AND is_active = TRUE\nORDER BY total_spent DESC LIMIT 10")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `MigrateLoyaltyTransactions` (Go `MigrateLoyaltyTransactionsParams`).
pub struct MigrateLoyaltyTransactionsParams {
    pub customer_id: i64,
    pub customer_id_2: i64,
    pub business_id: i64,
}

/// `MigrateLoyaltyTransactions`: `exec`.
pub async fn migrate_loyalty_transactions<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MigrateLoyaltyTransactionsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_loyalty_transactions SET customer_id = ? WHERE customer_id = ? AND business_id = ?")
        .bind(params.customer_id)
        .bind(params.customer_id_2)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `MigrateTimeline` (Go `MigrateTimelineParams`).
pub struct MigrateTimelineParams {
    pub customer_id: i64,
    pub customer_id_2: i64,
    pub business_id: i64,
}

/// `MigrateTimeline`: `exec`.
pub async fn migrate_timeline<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MigrateTimelineParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_customer_timeline SET customer_id = ? WHERE customer_id = ? AND business_id = ?")
        .bind(params.customer_id)
        .bind(params.customer_id_2)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `MoveCustomerActivities` (Go `MoveCustomerActivitiesParams`).
pub struct MoveCustomerActivitiesParams {
    pub to_customer_id: Option<i64>,
    pub from_customer_id: Option<i64>,
    pub business_id: i64,
}

/// `MoveCustomerActivities`: `exec`.
pub async fn move_customer_activities<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MoveCustomerActivitiesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE crm_touchpoints SET customer_id = ?\nWHERE customer_id = ? AND business_id = ?",
    )
    .bind(params.to_customer_id)
    .bind(params.from_customer_id)
    .bind(params.business_id)
    .execute(exec)
    .await
}

/// Params for `MoveCustomerCampaignSends` (Go `MoveCustomerCampaignSendsParams`).
pub struct MoveCustomerCampaignSendsParams {
    pub to_customer_id: i64,
    pub from_customer_id: i64,
    pub business_id: i64,
}

/// `MoveCustomerCampaignSends`: `exec`.
pub async fn move_customer_campaign_sends<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MoveCustomerCampaignSendsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE crm_campaign_sends SET customer_id = ?\nWHERE customer_id = ? AND business_id = ?",
    )
    .bind(params.to_customer_id)
    .bind(params.from_customer_id)
    .bind(params.business_id)
    .execute(exec)
    .await
}

/// Params for `MoveCustomerContacts` (Go `MoveCustomerContactsParams`).
pub struct MoveCustomerContactsParams {
    pub to_customer_id: Option<i64>,
    pub from_customer_id: Option<i64>,
    pub business_id: i64,
}

/// `MoveCustomerContacts`: `exec`.
pub async fn move_customer_contacts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MoveCustomerContactsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE crm_contacts SET customer_id = ?\nWHERE customer_id = ? AND business_id = ?",
    )
    .bind(params.to_customer_id)
    .bind(params.from_customer_id)
    .bind(params.business_id)
    .execute(exec)
    .await
}

/// Params for `MoveCustomerContracts` (Go `MoveCustomerContractsParams`).
pub struct MoveCustomerContractsParams {
    pub to_customer_id: Option<i64>,
    pub from_customer_id: Option<i64>,
    pub business_id: i64,
}

/// `MoveCustomerContracts`: `exec`.
pub async fn move_customer_contracts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MoveCustomerContractsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE crm_contracts SET customer_id = ?\nWHERE customer_id = ? AND business_id = ?",
    )
    .bind(params.to_customer_id)
    .bind(params.from_customer_id)
    .bind(params.business_id)
    .execute(exec)
    .await
}

/// Params for `MoveCustomerDeals` (Go `MoveCustomerDealsParams`).
pub struct MoveCustomerDealsParams {
    pub to_customer_id: Option<i64>,
    pub from_customer_id: Option<i64>,
    pub business_id: i64,
}

/// `MoveCustomerDeals`: `exec`.
pub async fn move_customer_deals<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MoveCustomerDealsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_deals SET customer_id = ?\nWHERE customer_id = ? AND business_id = ?")
        .bind(params.to_customer_id)
        .bind(params.from_customer_id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `MoveCustomerFeedback` (Go `MoveCustomerFeedbackParams`).
pub struct MoveCustomerFeedbackParams {
    pub to_customer_id: i64,
    pub from_customer_id: i64,
    pub business_id: i64,
}

/// `MoveCustomerFeedback`: `exec`.
pub async fn move_customer_feedback<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MoveCustomerFeedbackParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE crm_feedback SET customer_id = ?\nWHERE customer_id = ? AND business_id = ?",
    )
    .bind(params.to_customer_id)
    .bind(params.from_customer_id)
    .bind(params.business_id)
    .execute(exec)
    .await
}

/// Params for `MoveCustomerLeads` (Go `MoveCustomerLeadsParams`).
pub struct MoveCustomerLeadsParams {
    pub to_customer_id: Option<i64>,
    pub from_customer_id: Option<i64>,
    pub business_id: i64,
}

/// `MoveCustomerLeads`: `exec`.
pub async fn move_customer_leads<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MoveCustomerLeadsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_leads SET customer_id = ?\nWHERE customer_id = ? AND business_id = ?")
        .bind(params.to_customer_id)
        .bind(params.from_customer_id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `RecordCustomerCaptureEvent` (Go `RecordCustomerCaptureEventParams`).
pub struct RecordCustomerCaptureEventParams {
    pub event_id: u64,
    pub business_id: i64,
}

/// `RecordCustomerCaptureEvent`: `execresult`.
pub async fn record_customer_capture_event<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RecordCustomerCaptureEventParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "INSERT IGNORE INTO crm_customer_capture_events (event_id, business_id)\nVALUES (?, ?)",
    )
    .bind(params.event_id)
    .bind(params.business_id)
    .execute(exec)
    .await
}

/// Params for `RecordVisit` (Go `RecordVisitParams`).
pub struct RecordVisitParams {
    pub total_spent: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `RecordVisit`: `exec`.
pub async fn record_visit<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RecordVisitParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_customers\nSET visit_count = visit_count + 1,\n    last_visit_at = NOW(),\n    total_spent = total_spent + ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.total_spent)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateCustomer` (Go `UpdateCustomerParams`).
pub struct UpdateCustomerParams {
    pub first_name: String,
    pub last_name: Option<String>,
    pub email: Option<String>,
    pub phone: Option<String>,
    pub date_of_birth: Option<chrono::DateTime<chrono::Utc>>,
    pub gender: String,
    pub address: Option<String>,
    pub city: Option<String>,
    pub country: Option<String>,
    pub tags: String,
    pub notes: Option<String>,
    pub consent: String,
    pub is_active: Option<bool>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateCustomer`: `execresult`.
pub async fn update_customer<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateCustomerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_customers\nSET first_name = ?, last_name = ?, email = ?, phone = ?, date_of_birth = ?,\n\t    gender = ?, address = ?, city = ?, country = ?, tags = ?, notes = ?, consent = ?, is_active = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.first_name)
        .bind(&params.last_name)
        .bind(&params.email)
        .bind(&params.phone)
        .bind(params.date_of_birth)
        .bind(&params.gender)
        .bind(&params.address)
        .bind(&params.city)
        .bind(&params.country)
        .bind(&params.tags)
        .bind(&params.notes)
        .bind(&params.consent)
        .bind(params.is_active)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateCustomerFromCapture` (Go `UpdateCustomerFromCaptureParams`).
pub struct UpdateCustomerFromCaptureParams {
    pub first_name: String,
    pub last_name: Option<String>,
    pub email: Option<String>,
    pub phone: Option<String>,
    pub total_spent_delta: Option<String>,
    pub visit_count_delta: Option<i32>,
    pub last_visit_at: Option<chrono::DateTime<chrono::Utc>>,
    pub business_id: i64,
    pub customer_id: i64,
}

/// `UpdateCustomerFromCapture`: `exec`.
pub async fn update_customer_from_capture<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateCustomerFromCaptureParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_customers\nSET first_name = CASE WHEN ? <> '' THEN ? ELSE first_name END,\n    last_name = CASE WHEN ? <> '' THEN ? ELSE last_name END,\n    email = CASE WHEN ? <> '' THEN ? ELSE email END,\n    phone = CASE WHEN ? <> '' THEN ? ELSE phone END,\n    total_spent = total_spent + ?,\n    visit_count = visit_count + ?,\n    last_visit_at = COALESCE(?, last_visit_at),\n    is_active = TRUE\nWHERE business_id = ? AND id = ?")
        .bind(&params.first_name)
        .bind(&params.first_name)
        .bind(&params.last_name)
        .bind(&params.last_name)
        .bind(&params.email)
        .bind(&params.email)
        .bind(&params.phone)
        .bind(&params.phone)
        .bind(&params.total_spent_delta)
        .bind(params.visit_count_delta)
        .bind(params.last_visit_at)
        .bind(params.business_id)
        .bind(params.customer_id)
        .execute(exec)
        .await
}

/// Params for `UpdateCustomerLoyalty` (Go `UpdateCustomerLoyaltyParams`).
pub struct UpdateCustomerLoyaltyParams {
    pub loyalty_points: Option<i32>,
    pub loyalty_tier: String,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateCustomerLoyalty`: `exec`.
pub async fn update_customer_loyalty<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateCustomerLoyaltyParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_customers SET loyalty_points = ?, loyalty_tier = ? WHERE id = ? AND business_id = ?")
        .bind(params.loyalty_points)
        .bind(&params.loyalty_tier)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateCustomerMergeStats` (Go `UpdateCustomerMergeStatsParams`).
pub struct UpdateCustomerMergeStatsParams {
    pub loyalty_points: Option<i32>,
    pub total_spent: Option<String>,
    pub visit_count: Option<i32>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateCustomerMergeStats`: `exec`.
pub async fn update_customer_merge_stats<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateCustomerMergeStatsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_customers\nSET loyalty_points = ?, total_spent = ?, visit_count = ?\nWHERE id = ? AND business_id = ?")
        .bind(params.loyalty_points)
        .bind(&params.total_spent)
        .bind(params.visit_count)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateLoyaltyTier` (Go `UpdateLoyaltyTierParams`).
pub struct UpdateLoyaltyTierParams {
    pub name: String,
    pub min_points: Option<i32>,
    pub discount_pct: Option<String>,
    pub perks: String,
    pub sort_order: Option<i32>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateLoyaltyTier`: `execresult`.
pub async fn update_loyalty_tier<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateLoyaltyTierParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_loyalty_tiers SET name = ?, min_points = ?, discount_pct = ?, perks = ?, sort_order = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(params.min_points)
        .bind(&params.discount_pct)
        .bind(&params.perks)
        .bind(params.sort_order)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpsertLoyaltyConfig` (Go `UpsertLoyaltyConfigParams`).
pub struct UpsertLoyaltyConfigParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub is_enabled: Option<bool>,
    pub points_per_currency_unit: Option<String>,
    pub currency_per_point: Option<String>,
    pub bronze_min: Option<i32>,
    pub silver_min: Option<i32>,
    pub gold_min: Option<i32>,
    pub platinum_min: Option<i32>,
    pub bronze_discount_pct: Option<String>,
    pub silver_discount_pct: Option<String>,
    pub gold_discount_pct: Option<String>,
    pub platinum_discount_pct: Option<String>,
    pub points_expiry_days: Option<i32>,
}

/// `UpsertLoyaltyConfig`: `exec`.
pub async fn upsert_loyalty_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertLoyaltyConfigParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_loyalty_config\n  (owner_id, business_id, is_enabled, points_per_currency_unit, currency_per_point,\n   bronze_min, silver_min, gold_min, platinum_min,\n   bronze_discount_pct, silver_discount_pct, gold_discount_pct, platinum_discount_pct,\n   points_expiry_days)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)\nON DUPLICATE KEY UPDATE\n  is_enabled = VALUES(is_enabled),\n  points_per_currency_unit = VALUES(points_per_currency_unit),\n  currency_per_point = VALUES(currency_per_point),\n  bronze_min = VALUES(bronze_min),\n  silver_min = VALUES(silver_min),\n  gold_min = VALUES(gold_min),\n  platinum_min = VALUES(platinum_min),\n  bronze_discount_pct = VALUES(bronze_discount_pct),\n  silver_discount_pct = VALUES(silver_discount_pct),\n  gold_discount_pct = VALUES(gold_discount_pct),\n  platinum_discount_pct = VALUES(platinum_discount_pct),\n  points_expiry_days = VALUES(points_expiry_days)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.is_enabled)
        .bind(&params.points_per_currency_unit)
        .bind(&params.currency_per_point)
        .bind(params.bronze_min)
        .bind(params.silver_min)
        .bind(params.gold_min)
        .bind(params.platinum_min)
        .bind(&params.bronze_discount_pct)
        .bind(&params.silver_discount_pct)
        .bind(&params.gold_discount_pct)
        .bind(&params.platinum_discount_pct)
        .bind(params.points_expiry_days)
        .execute(exec)
        .await
}
