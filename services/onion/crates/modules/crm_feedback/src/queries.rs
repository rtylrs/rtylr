//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateFeedback` (Go `CreateFeedbackParams`).
pub struct CreateFeedbackParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub customer_id: i64,
    pub r#type: String,
    pub score: Option<i32>,
    pub comment: Option<String>,
    pub source: Option<String>,
}

/// `CreateFeedback`: `execresult`.
pub async fn create_feedback<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateFeedbackParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_feedback (owner_id, business_id, customer_id, type, score, comment, source)\nVALUES (?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.customer_id)
        .bind(&params.r#type)
        .bind(params.score)
        .bind(&params.comment)
        .bind(&params.source)
        .execute(exec)
        .await
}
