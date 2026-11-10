//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateAccount` (Go `CreateAccountParams`).
pub struct CreateAccountParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub code: Option<String>,
    pub r#type: String,
    pub currency: Option<String>,
    pub balance: Option<String>,
    pub is_active: Option<bool>,
}

/// `CreateAccount`: `execresult`.
pub async fn create_account<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateAccountParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO finance_accounts (owner_id, business_id, name, code, type, currency, balance, is_active)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.code)
        .bind(&params.r#type)
        .bind(&params.currency)
        .bind(&params.balance)
        .bind(params.is_active)
        .execute(exec)
        .await
}

/// Params for `DeleteAccount` (Go `DeleteAccountParams`).
pub struct DeleteAccountParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteAccount`: `execresult`.
pub async fn delete_account<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteAccountParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM finance_accounts\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `FindLedgerAccount` (Go `FindLedgerAccountParams`).
pub struct FindLedgerAccountParams {
    pub business_id: i64,
    pub code: Option<String>,
    pub account_type: String,
    pub currency: Option<String>,
}

/// `FindLedgerAccount`: `one`.
pub async fn find_ledger_account<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &FindLedgerAccountParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id\nFROM finance_accounts\nWHERE business_id = ?\n  AND code = ?\n  AND type = ?\n  AND COALESCE(NULLIF(currency, ''), 'USD') = ?\nORDER BY id ASC\nLIMIT 1")
        .bind(params.business_id)
        .bind(&params.code)
        .bind(&params.account_type)
        .bind(&params.currency)
        .fetch_optional(exec)
        .await
}

/// Params for `GetPayrollAccountMapping` (Go `GetPayrollAccountMappingParams`).
pub struct GetPayrollAccountMappingParams {
    pub business_id: i64,
}

/// `GetPayrollAccountMapping`: `one`.
pub async fn get_payroll_account_mapping<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPayrollAccountMappingParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT business_id, expense_account_id, payable_account_id, deductions_payable_account_id\nFROM finance_payroll_account_mappings\nWHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListAccounts` (Go `ListAccountsParams`).
pub struct ListAccountsParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListAccounts`: `many`.
pub async fn list_accounts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListAccountsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT a.id, a.owner_id, a.business_id, a.name, a.code, a.type, a.currency, a.balance, a.is_active, a.created_at,\n       CAST(COALESCE(p.debit, 0) - COALESCE(p.credit, 0) AS DECIMAL(15,2)) AS posted_net_debit\nFROM finance_accounts a\nLEFT JOIN (\n    SELECT l.account_id,\n           SUM(CAST(l.debit AS DECIMAL(15,2))) AS debit,\n           SUM(CAST(l.credit AS DECIMAL(15,2))) AS credit\n    FROM finance_journal_lines l\n    JOIN finance_journal_entries e ON e.id = l.entry_id\n    WHERE e.business_id = ?\n      AND e.status = 'posted'\n    GROUP BY l.account_id\n) p ON p.account_id = a.id\nWHERE a.business_id = ?\nORDER BY a.type ASC, a.code ASC, a.name ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListAccountsForTree` (Go `ListAccountsForTreeParams`).
pub struct ListAccountsForTreeParams {
    pub business_id: i64,
}

/// `ListAccountsForTree`: `many`.
pub async fn list_accounts_for_tree<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListAccountsForTreeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, code, parent_id, type, currency, balance, is_active, created_at\nFROM finance_accounts\nWHERE business_id = ?\nORDER BY type ASC, code ASC, name ASC\nLIMIT 1000")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateAccount` (Go `UpdateAccountParams`).
pub struct UpdateAccountParams {
    pub name: String,
    pub code: Option<String>,
    pub r#type: String,
    pub currency: Option<String>,
    pub balance: Option<String>,
    pub is_active: Option<bool>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateAccount`: `execresult`.
pub async fn update_account<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateAccountParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE finance_accounts\nSET name = ?, code = ?, type = ?, currency = ?, balance = ?, is_active = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(&params.code)
        .bind(&params.r#type)
        .bind(&params.currency)
        .bind(&params.balance)
        .bind(params.is_active)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
