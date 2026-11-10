//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `GetDaySummaryByDate` (Go `GetDaySummaryByDateParams`).
pub struct GetDaySummaryByDateParams {
    pub business_id: i64,
    pub paid_at: Option<chrono::DateTime<chrono::Utc>>,
    pub paid_at_end: Option<chrono::DateTime<chrono::Utc>>,
}

/// `GetDaySummaryByDate`: `one`.
pub async fn get_day_summary_by_date<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetDaySummaryByDateParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    COALESCE(SUM(total), 0) AS total_sales,\n    COALESCE(SUM(tax_amount), 0) AS total_tax,\n    COALESCE(SUM(tip_amount), 0) AS total_tips,\n    COUNT(*) AS total_orders\nFROM pos_orders\nWHERE business_id = ? AND payment_status = 'paid'\n  AND paid_at >= ?\n  AND paid_at < ?")
        .bind(params.business_id)
        .bind(params.paid_at)
        .bind(params.paid_at_end)
        .fetch_optional(exec)
        .await
}

/// Params for `GetPaymentBreakdownByDate` (Go `GetPaymentBreakdownByDateParams`).
pub struct GetPaymentBreakdownByDateParams {
    pub business_id: i64,
    pub paid_at: Option<chrono::DateTime<chrono::Utc>>,
    pub paid_at_end: Option<chrono::DateTime<chrono::Utc>>,
}

/// `GetPaymentBreakdownByDate`: `many`.
pub async fn get_payment_breakdown_by_date<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPaymentBreakdownByDateParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT payment_method, COALESCE(SUM(total), 0) AS method_total\nFROM pos_orders\nWHERE business_id = ? AND payment_status = 'paid'\n  AND paid_at >= ?\n  AND paid_at < ?\nGROUP BY payment_method\nLIMIT 200")
        .bind(params.business_id)
        .bind(params.paid_at)
        .bind(params.paid_at_end)
        .execute(exec)
        .await
}

/// Params for `GetRefundTotalByDate` (Go `GetRefundTotalByDateParams`).
pub struct GetRefundTotalByDateParams {
    pub business_id: i64,
    pub refunded_at: Option<chrono::DateTime<chrono::Utc>>,
    pub refunded_at_end: Option<chrono::DateTime<chrono::Utc>>,
}

/// `GetRefundTotalByDate`: `one`.
pub async fn get_refund_total_by_date<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetRefundTotalByDateParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(r.amount), 0) AS total_refunds FROM pos_refunds r\nJOIN pos_orders o ON o.id = r.order_id AND o.business_id = r.business_id\nWHERE r.business_id = ?\n  AND refunded_at >= ?\n  AND refunded_at < ?")
        .bind(params.business_id)
        .bind(params.refunded_at)
        .bind(params.refunded_at_end)
        .fetch_optional(exec)
        .await
}
