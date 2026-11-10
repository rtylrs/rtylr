//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateReview` (Go `CreateReviewParams`).
pub struct CreateReviewParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub staff_id: i64,
    pub reviewer_id: Option<i64>,
    pub period: Option<String>,
    pub rating: Option<String>,
    pub goals_rating: Option<String>,
    pub values_rating: Option<String>,
    pub skills_rating: Option<String>,
    pub notes: Option<String>,
    pub status: String,
}

/// `CreateReview`: `execresult`.
pub async fn create_review<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateReviewParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_reviews\n(owner_id, business_id, staff_id, reviewer_id, period, rating, goals_rating, values_rating, skills_rating, notes, status)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.staff_id)
        .bind(params.reviewer_id)
        .bind(&params.period)
        .bind(&params.rating)
        .bind(&params.goals_rating)
        .bind(&params.values_rating)
        .bind(&params.skills_rating)
        .bind(&params.notes)
        .bind(&params.status)
        .execute(exec)
        .await
}

/// Params for `ListReviews` (Go `ListReviewsParams`).
pub struct ListReviewsParams {
    pub business_id: i64,
}

/// `ListReviews`: `many`.
pub async fn list_reviews<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListReviewsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, staff_id, reviewer_id, period,\n       rating, goals_rating, values_rating, skills_rating,\n       notes, status, created_at, updated_at\nFROM hr_reviews\nWHERE business_id = ?\nORDER BY created_at DESC, id DESC\nLIMIT 500")
        .bind(params.business_id)
        .execute(exec)
        .await
}
