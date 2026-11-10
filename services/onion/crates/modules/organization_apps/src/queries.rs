//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `ListEnabledAddonIDs` (Go `ListEnabledAddonIDsParams`).
pub struct ListEnabledAddonIDsParams {
    pub business_id: i64,
}

/// `ListEnabledAddonIDs`: `many`.
pub async fn list_enabled_addon_i_ds<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListEnabledAddonIDsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT addon_id\nFROM business_addons\nWHERE business_id = ? AND is_enabled = 1\nLIMIT 100")
        .bind(params.business_id)
        .execute(exec)
        .await
}
