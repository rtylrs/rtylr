//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `ListDashboardWidgets` (Go `ListDashboardWidgetsParams`).
pub struct ListDashboardWidgetsParams {
    pub business_id: i64,
}

/// `ListDashboardWidgets`: `many`.
pub async fn list_dashboard_widgets<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListDashboardWidgetsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    id, owner_id, business_id, role_id, widget_type,\n    position_x, position_y, width, height,\n    config, created_at, updated_at\nFROM hr_dashboard_widgets\nWHERE business_id = ?\nORDER BY position_y ASC, position_x ASC\nLIMIT 500")
        .bind(params.business_id)
        .execute(exec)
        .await
}
