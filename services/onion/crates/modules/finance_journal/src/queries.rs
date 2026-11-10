//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountJournalEntriesByStatus` (Go `CountJournalEntriesByStatusParams`).
pub struct CountJournalEntriesByStatusParams {
    pub business_id: i64,
}

/// `CountJournalEntriesByStatus`: `many`.
pub async fn count_journal_entries_by_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountJournalEntriesByStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT status, COUNT(*) AS cnt\nFROM finance_journal_entries\nWHERE business_id = ?\n  AND created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)\nGROUP BY status")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `CreateJournalEntry` (Go `CreateJournalEntryParams`).
pub struct CreateJournalEntryParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub date: chrono::NaiveDate,
    pub description: Option<String>,
}

/// `CreateJournalEntry`: `execresult`.
pub async fn create_journal_entry<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateJournalEntryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO finance_journal_entries (owner_id, business_id, date, description)\nVALUES (?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.date)
        .bind(&params.description)
        .execute(exec)
        .await
}

/// Params for `CreateJournalLine` (Go `CreateJournalLineParams`).
pub struct CreateJournalLineParams {
    pub entry_id: i64,
    pub account_id: i64,
    pub debit: Option<String>,
    pub credit: Option<String>,
}

/// `CreateJournalLine`: `exec`.
pub async fn create_journal_line<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateJournalLineParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO finance_journal_lines (entry_id, account_id, debit, credit)\nVALUES (?, ?, ?, ?)")
        .bind(params.entry_id)
        .bind(params.account_id)
        .bind(&params.debit)
        .bind(&params.credit)
        .execute(exec)
        .await
}

/// Params for `GetJournalEntry` (Go `GetJournalEntryParams`).
pub struct GetJournalEntryParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetJournalEntry`: `one`.
pub async fn get_journal_entry<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetJournalEntryParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, date, description, status, posted_at, created_at\nFROM finance_journal_entries\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetJournalLines` (Go `GetJournalLinesParams`).
pub struct GetJournalLinesParams {
    pub entry_id: i64,
}

/// `GetJournalLines`: `many`.
pub async fn get_journal_lines<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetJournalLinesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT l.id, l.entry_id, l.account_id, l.debit, l.credit, a.name AS account_name\nFROM finance_journal_lines l\nJOIN finance_journal_entries e ON e.id = l.entry_id\nLEFT JOIN finance_accounts a ON a.id = l.account_id AND a.business_id = e.business_id\nWHERE l.entry_id = ?\nORDER BY l.id ASC")
        .bind(params.entry_id)
        .execute(exec)
        .await
}

/// Params for `GetJournalLinesForEntries` (Go `GetJournalLinesForEntriesParams`).
pub struct GetJournalLinesForEntriesParams {
    pub entry_ids: Vec<i64>,
}

/// `GetJournalLinesForEntries`: `many`.
pub async fn get_journal_lines_for_entries<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetJournalLinesForEntriesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT l.id, l.entry_id, l.account_id, l.debit, l.credit, a.name AS account_name\nFROM finance_journal_lines l\nJOIN finance_journal_entries e ON e.id = l.entry_id\nLEFT JOIN finance_accounts a ON a.id = l.account_id AND a.business_id = e.business_id\nWHERE l.entry_id IN ({SLICE:entry_ids})\nORDER BY l.id ASC".to_string();

    if params.entry_ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:entry_ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:entry_ids}",
            &std::iter::repeat_n("?", params.entry_ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query;
    for value in &params.entry_ids {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `ListJournalEntries` (Go `ListJournalEntriesParams`).
pub struct ListJournalEntriesParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListJournalEntries`: `many`.
pub async fn list_journal_entries<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListJournalEntriesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, date, description, status, posted_at, created_at\nFROM finance_journal_entries\nWHERE business_id = ?\nORDER BY date DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `LockJournalAccounts` (Go `LockJournalAccountsParams`).
pub struct LockJournalAccountsParams {
    pub business_id: i64,
    pub account_ids: Vec<i64>,
}

/// `LockJournalAccounts`: `many`.
pub async fn lock_journal_accounts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LockJournalAccountsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT id\nFROM finance_accounts\nWHERE business_id = ?\n  AND id IN ({SLICE:account_ids})\n  AND is_active = TRUE\nFOR UPDATE".to_string();

    if params.account_ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:account_ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:account_ids}",
            &std::iter::repeat_n("?", params.account_ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query.bind(params.business_id);
    for value in &params.account_ids {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `PostJournalEntry` (Go `PostJournalEntryParams`).
pub struct PostJournalEntryParams {
    pub id: i64,
    pub business_id: i64,
}

/// `PostJournalEntry`: `execresult`.
pub async fn post_journal_entry<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &PostJournalEntryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE finance_journal_entries\nSET status = 'posted', posted_at = CURRENT_TIMESTAMP\nWHERE id = ? AND business_id = ? AND status = 'draft'")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `SumJournalLinesByAccount` (Go `SumJournalLinesByAccountParams`).
pub struct SumJournalLinesByAccountParams {
    pub business_id: i64,
    pub account_id: i64,
}

/// `SumJournalLinesByAccount`: `one`.
pub async fn sum_journal_lines_by_account<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumJournalLinesByAccountParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    COALESCE(SUM(CAST(l.debit AS REAL)), 0) AS total_debit,\n    COALESCE(SUM(CAST(l.credit AS REAL)), 0) AS total_credit\nFROM finance_journal_lines l\nJOIN finance_journal_entries e ON e.id = l.entry_id\nWHERE e.business_id = ?\n  AND l.account_id = ?\n  AND e.status = 'posted'")
        .bind(params.business_id)
        .bind(params.account_id)
        .fetch_optional(exec)
        .await
}
