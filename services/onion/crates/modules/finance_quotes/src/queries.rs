//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountQuotesByStatus` (Go `CountQuotesByStatusParams`).
pub struct CountQuotesByStatusParams {
    pub business_id: i64,
}

/// `CountQuotesByStatus`: `many`.
pub async fn count_quotes_by_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountQuotesByStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT status, COUNT(*) AS cnt\nFROM finance_quotes\nWHERE business_id = ?\n  AND created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)\nGROUP BY status")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `CreateQuote` (Go `CreateQuoteParams`).
pub struct CreateQuoteParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub customer_name: String,
    pub customer_email: Option<String>,
    pub status: String,
    pub items: String,
    pub subtotal: Option<String>,
    pub tax: Option<String>,
    pub discount: Option<String>,
    pub total: Option<String>,
    pub valid_until: Option<chrono::DateTime<chrono::Utc>>,
    pub notes: Option<String>,
}

/// `CreateQuote`: `execresult`.
pub async fn create_quote<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateQuoteParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO finance_quotes\n(owner_id, business_id, customer_name, customer_email, status, items, subtotal, tax, discount, total, valid_until, notes)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.customer_name)
        .bind(&params.customer_email)
        .bind(&params.status)
        .bind(&params.items)
        .bind(&params.subtotal)
        .bind(&params.tax)
        .bind(&params.discount)
        .bind(&params.total)
        .bind(params.valid_until)
        .bind(&params.notes)
        .execute(exec)
        .await
}

/// Params for `DeleteQuote` (Go `DeleteQuoteParams`).
pub struct DeleteQuoteParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteQuote`: `execresult`.
pub async fn delete_quote<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteQuoteParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM finance_quotes\nWHERE id = ? AND business_id = ? AND status IN ('draft', 'sent')")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `GetQuote` (Go `GetQuoteParams`).
pub struct GetQuoteParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetQuote`: `one`.
pub async fn get_quote<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetQuoteParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, customer_name, customer_email, status,\n       items, subtotal, tax, discount, total, valid_until, notes, created_at, updated_at\nFROM finance_quotes\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsertInvoiceFromQuote` (Go `InsertInvoiceFromQuoteParams`).
pub struct InsertInvoiceFromQuoteParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub invoice_number: String,
    pub client_name: String,
    pub client_email: Option<String>,
    pub status: String,
    pub issue_date: chrono::NaiveDate,
    pub due_date: chrono::NaiveDate,
    pub subtotal: Option<String>,
    pub tax_amount: Option<String>,
    pub discount_amount: Option<String>,
    pub total: Option<String>,
    pub amount_due: Option<String>,
    pub notes: Option<String>,
}

/// `InsertInvoiceFromQuote`: `execresult`.
pub async fn insert_invoice_from_quote<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertInvoiceFromQuoteParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO finance_invoices\n(owner_id, business_id, invoice_number, client_name, client_email, status,\n issue_date, due_date, subtotal, tax_amount, discount_amount, total,\n amount_due, notes)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.invoice_number)
        .bind(&params.client_name)
        .bind(&params.client_email)
        .bind(&params.status)
        .bind(params.issue_date)
        .bind(params.due_date)
        .bind(&params.subtotal)
        .bind(&params.tax_amount)
        .bind(&params.discount_amount)
        .bind(&params.total)
        .bind(&params.amount_due)
        .bind(&params.notes)
        .execute(exec)
        .await
}

/// Params for `InsertInvoiceItemFromQuote` (Go `InsertInvoiceItemFromQuoteParams`).
pub struct InsertInvoiceItemFromQuoteParams {
    pub invoice_id: i64,
    pub description: String,
    pub quantity: String,
    pub unit: Option<String>,
    pub unit_price: String,
    pub tax_rate: Option<String>,
    pub discount_pct: Option<String>,
    pub total: String,
    pub sort_order: Option<i32>,
}

/// `InsertInvoiceItemFromQuote`: `exec`.
pub async fn insert_invoice_item_from_quote<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertInvoiceItemFromQuoteParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO finance_invoice_items\n(invoice_id, description, quantity, unit, unit_price, tax_rate, discount_pct, total, sort_order)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.invoice_id)
        .bind(&params.description)
        .bind(&params.quantity)
        .bind(&params.unit)
        .bind(&params.unit_price)
        .bind(&params.tax_rate)
        .bind(&params.discount_pct)
        .bind(&params.total)
        .bind(params.sort_order)
        .execute(exec)
        .await
}

/// Params for `ListQuotes` (Go `ListQuotesParams`).
pub struct ListQuotesParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListQuotes`: `many`.
pub async fn list_quotes<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListQuotesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, customer_name, customer_email, status,\n       items, subtotal, tax, discount, total, valid_until, notes, created_at, updated_at\nFROM finance_quotes\nWHERE business_id = ?\nORDER BY created_at DESC, id DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `MarkQuoteConverted` (Go `MarkQuoteConvertedParams`).
pub struct MarkQuoteConvertedParams {
    pub id: i64,
    pub business_id: i64,
}

/// `MarkQuoteConverted`: `execresult`.
pub async fn mark_quote_converted<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MarkQuoteConvertedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE finance_quotes\nSET status = 'converted'\nWHERE id = ? AND business_id = ? AND status <> 'converted'")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateQuote` (Go `UpdateQuoteParams`).
pub struct UpdateQuoteParams {
    pub customer_name: String,
    pub customer_email: Option<String>,
    pub status: String,
    pub items: String,
    pub subtotal: Option<String>,
    pub tax: Option<String>,
    pub discount: Option<String>,
    pub total: Option<String>,
    pub valid_until: Option<chrono::DateTime<chrono::Utc>>,
    pub notes: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateQuote`: `execresult`.
pub async fn update_quote<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateQuoteParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE finance_quotes\nSET customer_name = ?, customer_email = ?, status = ?, items = ?,\n    subtotal = ?, tax = ?, discount = ?, total = ?, valid_until = ?, notes = ?\nWHERE id = ? AND business_id = ? AND status IN ('draft', 'sent')")
        .bind(&params.customer_name)
        .bind(&params.customer_email)
        .bind(&params.status)
        .bind(&params.items)
        .bind(&params.subtotal)
        .bind(&params.tax)
        .bind(&params.discount)
        .bind(&params.total)
        .bind(params.valid_until)
        .bind(&params.notes)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateQuoteStatus` (Go `UpdateQuoteStatusParams`).
pub struct UpdateQuoteStatusParams {
    pub status: String,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateQuoteStatus`: `execresult`.
pub async fn update_quote_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateQuoteStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE finance_quotes\nSET status = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.status)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
