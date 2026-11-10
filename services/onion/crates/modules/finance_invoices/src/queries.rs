//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountInvoicesByCustomer` (Go `CountInvoicesByCustomerParams`).
pub struct CountInvoicesByCustomerParams {
    pub business_id: i64,
    pub customer_id: Option<i64>,
}

/// `CountInvoicesByCustomer`: `one`.
pub async fn count_invoices_by_customer<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountInvoicesByCustomerParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM finance_invoices WHERE business_id = ? AND customer_id = ?")
        .bind(params.business_id)
        .bind(params.customer_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountInvoicesByStatus` (Go `CountInvoicesByStatusParams`).
pub struct CountInvoicesByStatusParams {
    pub business_id: i64,
}

/// `CountInvoicesByStatus`: `many`.
pub async fn count_invoices_by_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountInvoicesByStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT status, COUNT(*) AS cnt\nFROM finance_invoices\nWHERE business_id = ?\n  AND created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)\nGROUP BY status")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `CountOverdueInvoices` (Go `CountOverdueInvoicesParams`).
pub struct CountOverdueInvoicesParams {
    pub business_id: i64,
    pub due_date: chrono::NaiveDate,
}

/// `CountOverdueInvoices`: `one`.
pub async fn count_overdue_invoices<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountOverdueInvoicesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*)\nFROM finance_invoices\nWHERE business_id = ?\n  AND status NOT IN ('paid', 'void', 'draft')\n  AND due_date < ?")
        .bind(params.business_id)
        .bind(params.due_date)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateExpense` (Go `CreateExpenseParams`).
pub struct CreateExpenseParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub category_id: Option<i64>,
    pub description: String,
    pub amount: String,
    pub currency: Option<String>,
    pub expense_date: chrono::NaiveDate,
    pub vendor: Option<String>,
    pub payment_method: String,
    pub receipt_url: Option<String>,
    pub is_tax_deductible: Option<bool>,
    pub tax_amount: Option<String>,
    pub notes: Option<String>,
}

/// `CreateExpense`: `execresult`.
pub async fn create_expense<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateExpenseParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO finance_expenses\n  (owner_id, business_id, category_id, description, amount, currency, expense_date, vendor,\n   payment_method, receipt_url, is_tax_deductible, tax_amount, notes)\nVALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.category_id)
        .bind(&params.description)
        .bind(&params.amount)
        .bind(&params.currency)
        .bind(params.expense_date)
        .bind(&params.vendor)
        .bind(&params.payment_method)
        .bind(&params.receipt_url)
        .bind(params.is_tax_deductible)
        .bind(&params.tax_amount)
        .bind(&params.notes)
        .execute(exec)
        .await
}

/// Params for `CreateExpenseCategory` (Go `CreateExpenseCategoryParams`).
pub struct CreateExpenseCategoryParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub color: Option<String>,
    pub icon: Option<String>,
}

/// `CreateExpenseCategory`: `execresult`.
pub async fn create_expense_category<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateExpenseCategoryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO finance_expense_categories (owner_id, business_id, name, color, icon)\nVALUES (?,?,?,?,?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.color)
        .bind(&params.icon)
        .execute(exec)
        .await
}

/// Params for `CreateInvoice` (Go `CreateInvoiceParams`).
pub struct CreateInvoiceParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub invoice_number: String,
    pub client_name: String,
    pub client_email: Option<String>,
    pub client_address: Option<String>,
    pub client_tax_id: Option<String>,
    pub customer_id: Option<i64>,
    pub status: String,
    pub issue_date: chrono::NaiveDate,
    pub due_date: chrono::NaiveDate,
    pub currency: Option<String>,
    pub subtotal: Option<String>,
    pub discount_type: String,
    pub discount_value: Option<String>,
    pub discount_amount: Option<String>,
    pub tax_rate: Option<String>,
    pub tax_amount: Option<String>,
    pub total: Option<String>,
    pub amount_paid: Option<String>,
    pub amount_due: Option<String>,
    pub notes: Option<String>,
    pub terms: Option<String>,
}

/// `CreateInvoice`: `execresult`.
pub async fn create_invoice<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateInvoiceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO finance_invoices\n  (owner_id, business_id, invoice_number, client_name, client_email, client_address, client_tax_id, customer_id,\n   status, issue_date, due_date, currency, subtotal, discount_type, discount_value,\n   discount_amount, tax_rate, tax_amount, total, amount_paid, amount_due, notes, terms)\nVALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.invoice_number)
        .bind(&params.client_name)
        .bind(&params.client_email)
        .bind(&params.client_address)
        .bind(&params.client_tax_id)
        .bind(params.customer_id)
        .bind(&params.status)
        .bind(params.issue_date)
        .bind(params.due_date)
        .bind(&params.currency)
        .bind(&params.subtotal)
        .bind(&params.discount_type)
        .bind(&params.discount_value)
        .bind(&params.discount_amount)
        .bind(&params.tax_rate)
        .bind(&params.tax_amount)
        .bind(&params.total)
        .bind(&params.amount_paid)
        .bind(&params.amount_due)
        .bind(&params.notes)
        .bind(&params.terms)
        .execute(exec)
        .await
}

/// Params for `CreatePayment` (Go `CreatePaymentParams`).
pub struct CreatePaymentParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub invoice_id: i64,
    pub amount: String,
    pub currency: Option<String>,
    pub payment_date: chrono::NaiveDate,
    pub method: String,
    pub reference: Option<String>,
    pub notes: Option<String>,
}

/// `CreatePayment`: `execresult`.
pub async fn create_payment<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreatePaymentParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO finance_payments\n  (owner_id, business_id, invoice_id, amount, currency, payment_date, method, reference, notes)\nVALUES (?,?,?,?,?,?,?,?,?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.invoice_id)
        .bind(&params.amount)
        .bind(&params.currency)
        .bind(params.payment_date)
        .bind(&params.method)
        .bind(&params.reference)
        .bind(&params.notes)
        .execute(exec)
        .await
}

/// Params for `CreateTaxConfig` (Go `CreateTaxConfigParams`).
pub struct CreateTaxConfigParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub rate: String,
    pub r#type: String,
    pub is_default: Option<bool>,
    pub is_active: Option<bool>,
}

/// `CreateTaxConfig`: `execresult`.
pub async fn create_tax_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateTaxConfigParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO finance_tax_config (owner_id, business_id, name, rate, type, is_default, is_active)\nVALUES (?,?,?,?,?,?,?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.rate)
        .bind(&params.r#type)
        .bind(params.is_default)
        .bind(params.is_active)
        .execute(exec)
        .await
}

/// Params for `DeleteExpense` (Go `DeleteExpenseParams`).
pub struct DeleteExpenseParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteExpense`: `exec`.
pub async fn delete_expense<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteExpenseParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM finance_expenses WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `DeleteInvoice` (Go `DeleteInvoiceParams`).
pub struct DeleteInvoiceParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteInvoice`: `exec`.
pub async fn delete_invoice<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteInvoiceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "DELETE FROM finance_invoices WHERE id = ? AND business_id = ? AND status = 'draft'",
    )
    .bind(params.id)
    .bind(params.business_id)
    .execute(exec)
    .await
}

/// Params for `DeleteInvoiceItems` (Go `DeleteInvoiceItemsParams`).
pub struct DeleteInvoiceItemsParams {
    pub invoice_id: i64,
}

/// `DeleteInvoiceItems`: `exec`.
pub async fn delete_invoice_items<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteInvoiceItemsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM finance_invoice_items WHERE invoice_id = ?")
        .bind(params.invoice_id)
        .execute(exec)
        .await
}

/// Params for `GetExpenseByID` (Go `GetExpenseByIDParams`).
pub struct GetExpenseByIDParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetExpenseByID`: `one`.
pub async fn get_expense_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetExpenseByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, category_id, description, amount, currency, expense_date, vendor,\n       payment_method, receipt_url, is_tax_deductible, tax_amount, notes, status, created_at, updated_at\nFROM finance_expenses\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetInvoiceByID` (Go `GetInvoiceByIDParams`).
pub struct GetInvoiceByIDParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetInvoiceByID`: `one`.
pub async fn get_invoice_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetInvoiceByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, invoice_number, client_name, client_email, client_address, client_tax_id, customer_id,\n       status, issue_date, due_date, paid_date, reporting_date, currency, subtotal, discount_type,\n       discount_value, discount_amount, tax_rate, tax_amount, total, amount_paid, amount_due,\n       notes, terms, created_at, updated_at\nFROM finance_invoices\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetInvoiceByIDForUpdate` (Go `GetInvoiceByIDForUpdateParams`).
pub struct GetInvoiceByIDForUpdateParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetInvoiceByIDForUpdate`: `one`.
pub async fn get_invoice_by_id_for_update<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetInvoiceByIDForUpdateParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, invoice_number, client_name, client_email, client_address, client_tax_id, customer_id,\n       status, issue_date, due_date, paid_date, reporting_date, currency, subtotal, discount_type,\n       discount_value, discount_amount, tax_rate, tax_amount, total, amount_paid, amount_due,\n       notes, terms, created_at, updated_at\nFROM finance_invoices\nWHERE id = ? AND business_id = ?\nFOR UPDATE")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetInvoiceItems` (Go `GetInvoiceItemsParams`).
pub struct GetInvoiceItemsParams {
    pub invoice_id: i64,
}

/// `GetInvoiceItems`: `many`.
pub async fn get_invoice_items<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetInvoiceItemsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, invoice_id, description, quantity, unit, unit_price, tax_rate, discount_pct, total, sort_order\nFROM finance_invoice_items\nWHERE invoice_id = ?\nORDER BY sort_order ASC, id ASC\nLIMIT 500")
        .bind(params.invoice_id)
        .execute(exec)
        .await
}

/// Params for `GetInvoiceStatusForUpdate` (Go `GetInvoiceStatusForUpdateParams`).
pub struct GetInvoiceStatusForUpdateParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetInvoiceStatusForUpdate`: `one`.
pub async fn get_invoice_status_for_update<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetInvoiceStatusForUpdateParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT status\nFROM finance_invoices\nWHERE id = ? AND business_id = ?\nFOR UPDATE",
    )
    .bind(params.id)
    .bind(params.business_id)
    .fetch_optional(exec)
    .await
}

/// Params for `GetPaymentByReference` (Go `GetPaymentByReferenceParams`).
pub struct GetPaymentByReferenceParams {
    pub business_id: i64,
    pub invoice_id: i64,
    pub reference: Option<String>,
}

/// `GetPaymentByReference`: `one`.
pub async fn get_payment_by_reference<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPaymentByReferenceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, invoice_id, amount, currency, payment_date, method, reference, notes, created_at\nFROM finance_payments\nWHERE business_id = ? AND invoice_id = ? AND reference = ?")
        .bind(params.business_id)
        .bind(params.invoice_id)
        .bind(&params.reference)
        .fetch_optional(exec)
        .await
}

/// Params for `GetPaymentsByInvoice` (Go `GetPaymentsByInvoiceParams`).
pub struct GetPaymentsByInvoiceParams {
    pub invoice_id: i64,
}

/// `GetPaymentsByInvoice`: `many`.
pub async fn get_payments_by_invoice<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPaymentsByInvoiceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, invoice_id, amount, currency, payment_date, method, reference, notes, created_at\nFROM finance_payments\nWHERE invoice_id = ?\nORDER BY payment_date ASC, id ASC\nLIMIT 500")
        .bind(params.invoice_id)
        .execute(exec)
        .await
}

/// Params for `ListExpenseCategories` (Go `ListExpenseCategoriesParams`).
pub struct ListExpenseCategoriesParams {
    pub business_id: i64,
}

/// `ListExpenseCategories`: `many`.
pub async fn list_expense_categories<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListExpenseCategoriesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, color, icon, created_at\nFROM finance_expense_categories\nWHERE business_id = ?\nORDER BY name ASC\nLIMIT 200")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListInvoicesByCustomer` (Go `ListInvoicesByCustomerParams`).
pub struct ListInvoicesByCustomerParams {
    pub business_id: i64,
    pub customer_id: Option<i64>,
    pub limit: i32,
    pub offset: i32,
}

/// `ListInvoicesByCustomer`: `many`.
pub async fn list_invoices_by_customer<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListInvoicesByCustomerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, invoice_number, client_name, client_email, client_address, client_tax_id, customer_id,\n       status, issue_date, due_date, paid_date, reporting_date, currency, subtotal, discount_type,\n       discount_value, discount_amount, tax_rate, tax_amount, total, amount_paid, amount_due,\n       notes, terms, created_at, updated_at\nFROM finance_invoices\nWHERE business_id = ? AND customer_id = ?\nORDER BY created_at DESC, id DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.customer_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListOverdueInvoices` (Go `ListOverdueInvoicesParams`).
pub struct ListOverdueInvoicesParams {
    pub business_id: i64,
    pub due_date: chrono::NaiveDate,
    pub limit: i32,
    pub offset: i32,
}

/// `ListOverdueInvoices`: `many`.
pub async fn list_overdue_invoices<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListOverdueInvoicesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, invoice_number, client_name, client_email, client_address, client_tax_id, customer_id,\n       status, issue_date, due_date, paid_date, reporting_date, currency, subtotal, discount_type,\n       discount_value, discount_amount, tax_rate, tax_amount, total, amount_paid, amount_due,\n       notes, terms, created_at, updated_at\nFROM finance_invoices\nWHERE business_id = ?\n  AND status NOT IN ('paid', 'void', 'draft')\n  AND due_date < ?\nORDER BY due_date ASC, id ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.due_date)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListRecentExpenses` (Go `ListRecentExpensesParams`).
pub struct ListRecentExpensesParams {
    pub business_id: i64,
    pub limit: i32,
}

/// `ListRecentExpenses`: `many`.
pub async fn list_recent_expenses<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListRecentExpensesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, description, amount, expense_date, status\nFROM finance_expenses\nWHERE business_id = ? AND status <> 'rejected'\nORDER BY expense_date DESC, id DESC\nLIMIT ?")
        .bind(params.business_id)
        .bind(params.limit)
        .execute(exec)
        .await
}

/// Params for `ListRecentPaidInvoices` (Go `ListRecentPaidInvoicesParams`).
pub struct ListRecentPaidInvoicesParams {
    pub business_id: i64,
    pub limit: i32,
}

/// `ListRecentPaidInvoices`: `many`.
pub async fn list_recent_paid_invoices<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListRecentPaidInvoicesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, invoice_number, client_name, total, paid_date\nFROM finance_invoices\nWHERE business_id = ? AND status = 'paid' AND paid_date IS NOT NULL\nORDER BY paid_date DESC, id DESC\nLIMIT ?")
        .bind(params.business_id)
        .bind(params.limit)
        .execute(exec)
        .await
}

/// Params for `ListTaxConfigs` (Go `ListTaxConfigsParams`).
pub struct ListTaxConfigsParams {
    pub business_id: i64,
}

/// `ListTaxConfigs`: `many`.
pub async fn list_tax_configs<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListTaxConfigsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, rate, type, is_default, is_active, created_at\nFROM finance_tax_config\nWHERE business_id = ?\nORDER BY is_default DESC, name ASC\nLIMIT 200")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `MonthlyExpenseTotals` (Go `MonthlyExpenseTotalsParams`).
pub struct MonthlyExpenseTotalsParams {
    pub business_id: i64,
    pub from_expense_date: chrono::NaiveDate,
    pub to_expense_date: chrono::NaiveDate,
}

/// `MonthlyExpenseTotals`: `many`.
pub async fn monthly_expense_totals<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MonthlyExpenseTotalsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT DATE_FORMAT(expense_date, '%Y-%m') AS month, COALESCE(SUM(amount),0) AS total\nFROM finance_expenses\nWHERE business_id = ? AND status = 'approved' AND expense_date BETWEEN ? AND ?\nGROUP BY DATE_FORMAT(expense_date, '%Y-%m')")
        .bind(params.business_id)
        .bind(params.from_expense_date)
        .bind(params.to_expense_date)
        .execute(exec)
        .await
}

/// Params for `MonthlyPaidInvoiceTotals` (Go `MonthlyPaidInvoiceTotalsParams`).
pub struct MonthlyPaidInvoiceTotalsParams {
    pub business_id: i64,
    pub from_paid_date: Option<chrono::DateTime<chrono::Utc>>,
    pub to_paid_date: Option<chrono::DateTime<chrono::Utc>>,
}

/// `MonthlyPaidInvoiceTotals`: `many`.
pub async fn monthly_paid_invoice_totals<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MonthlyPaidInvoiceTotalsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT DATE_FORMAT(paid_date, '%Y-%m') AS month, COALESCE(SUM(total),0) AS total\nFROM finance_invoices\nWHERE business_id = ? AND status = 'paid' AND paid_date BETWEEN ? AND ?\nGROUP BY DATE_FORMAT(paid_date, '%Y-%m')")
        .bind(params.business_id)
        .bind(params.from_paid_date)
        .bind(params.to_paid_date)
        .execute(exec)
        .await
}

/// Params for `SumExpenses` (Go `SumExpensesParams`).
pub struct SumExpensesParams {
    pub business_id: i64,
    pub from_expense_date: chrono::NaiveDate,
    pub to_expense_date: chrono::NaiveDate,
}

/// `SumExpenses`: `one`.
pub async fn sum_expenses<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumExpensesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(amount),0) FROM finance_expenses WHERE business_id = ? AND status = 'approved' AND expense_date BETWEEN ? AND ?")
        .bind(params.business_id)
        .bind(params.from_expense_date)
        .bind(params.to_expense_date)
        .fetch_optional(exec)
        .await
}

/// Params for `SumExpensesByCategory` (Go `SumExpensesByCategoryParams`).
pub struct SumExpensesByCategoryParams {
    pub business_id: i64,
    pub from_expense_date: chrono::NaiveDate,
    pub to_expense_date: chrono::NaiveDate,
}

/// `SumExpensesByCategory`: `many`.
pub async fn sum_expenses_by_category<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumExpensesByCategoryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT e.category_id, COALESCE(c.name, '') AS name, COALESCE(c.color, '') AS color, COALESCE(SUM(e.amount),0) AS total\nFROM finance_expenses e\nLEFT JOIN finance_expense_categories c ON c.id = e.category_id AND c.business_id = e.business_id\nWHERE e.business_id = ? AND e.status = 'approved' AND e.expense_date BETWEEN ? AND ?\nGROUP BY e.category_id, c.name, c.color\nORDER BY total DESC")
        .bind(params.business_id)
        .bind(params.from_expense_date)
        .bind(params.to_expense_date)
        .execute(exec)
        .await
}

/// Params for `SumOverdueAmount` (Go `SumOverdueAmountParams`).
pub struct SumOverdueAmountParams {
    pub business_id: i64,
}

/// `SumOverdueAmount`: `one`.
pub async fn sum_overdue_amount<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumOverdueAmountParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(amount_due),0)\nFROM finance_invoices\nWHERE business_id = ?\n  AND status IN ('sent', 'viewed', 'partial', 'overdue')\n  AND due_date < CURRENT_DATE\n  AND amount_due > 0")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `SumPaidInvoices` (Go `SumPaidInvoicesParams`).
pub struct SumPaidInvoicesParams {
    pub business_id: i64,
    pub from_paid_date: Option<chrono::DateTime<chrono::Utc>>,
    pub to_paid_date: Option<chrono::DateTime<chrono::Utc>>,
}

/// `SumPaidInvoices`: `one`.
pub async fn sum_paid_invoices<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumPaidInvoicesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(total),0) FROM finance_invoices WHERE business_id = ? AND status = 'paid' AND paid_date BETWEEN ? AND ?")
        .bind(params.business_id)
        .bind(params.from_paid_date)
        .bind(params.to_paid_date)
        .fetch_optional(exec)
        .await
}

/// Params for `SumPendingExpenses` (Go `SumPendingExpensesParams`).
pub struct SumPendingExpensesParams {
    pub business_id: i64,
    pub from_expense_date: chrono::NaiveDate,
    pub to_expense_date: chrono::NaiveDate,
}

/// `SumPendingExpenses`: `one`.
pub async fn sum_pending_expenses<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumPendingExpensesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(amount),0) AS total, COUNT(*) AS expense_count\nFROM finance_expenses\nWHERE business_id = ? AND status = 'pending' AND expense_date BETWEEN ? AND ?")
        .bind(params.business_id)
        .bind(params.from_expense_date)
        .bind(params.to_expense_date)
        .fetch_optional(exec)
        .await
}

/// Params for `SumTaxCollected` (Go `SumTaxCollectedParams`).
pub struct SumTaxCollectedParams {
    pub business_id: i64,
    pub from_paid_date: Option<chrono::DateTime<chrono::Utc>>,
    pub to_paid_date: Option<chrono::DateTime<chrono::Utc>>,
}

/// `SumTaxCollected`: `one`.
pub async fn sum_tax_collected<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumTaxCollectedParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(tax_amount),0) FROM finance_invoices WHERE business_id = ? AND status = 'paid' AND paid_date BETWEEN ? AND ?")
        .bind(params.business_id)
        .bind(params.from_paid_date)
        .bind(params.to_paid_date)
        .fetch_optional(exec)
        .await
}

/// Params for `SumTaxOwed` (Go `SumTaxOwedParams`).
pub struct SumTaxOwedParams {
    pub business_id: i64,
    pub from_expense_date: chrono::NaiveDate,
    pub to_expense_date: chrono::NaiveDate,
}

/// `SumTaxOwed`: `one`.
pub async fn sum_tax_owed<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumTaxOwedParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(tax_amount),0) FROM finance_expenses WHERE business_id = ? AND status = 'approved' AND is_tax_deductible = TRUE AND expense_date BETWEEN ? AND ?")
        .bind(params.business_id)
        .bind(params.from_expense_date)
        .bind(params.to_expense_date)
        .fetch_optional(exec)
        .await
}

/// Params for `SumTotalReceivable` (Go `SumTotalReceivableParams`).
pub struct SumTotalReceivableParams {
    pub business_id: i64,
}

/// `SumTotalReceivable`: `one`.
pub async fn sum_total_receivable<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumTotalReceivableParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(amount_due),0) FROM finance_invoices WHERE business_id = ? AND status IN ('sent','viewed','partial','overdue')")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `UpdateExpense` (Go `UpdateExpenseParams`).
pub struct UpdateExpenseParams {
    pub category_id: Option<i64>,
    pub description: String,
    pub amount: String,
    pub currency: Option<String>,
    pub expense_date: chrono::NaiveDate,
    pub vendor: Option<String>,
    pub payment_method: String,
    pub receipt_url: Option<String>,
    pub is_tax_deductible: Option<bool>,
    pub tax_amount: Option<String>,
    pub notes: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateExpense`: `exec`.
pub async fn update_expense<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateExpenseParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE finance_expenses SET\n  category_id = ?, description = ?, amount = ?, currency = ?, expense_date = ?,\n  vendor = ?, payment_method = ?, receipt_url = ?, is_tax_deductible = ?,\n  tax_amount = ?, notes = ?\nWHERE id = ? AND business_id = ?")
        .bind(params.category_id)
        .bind(&params.description)
        .bind(&params.amount)
        .bind(&params.currency)
        .bind(params.expense_date)
        .bind(&params.vendor)
        .bind(&params.payment_method)
        .bind(&params.receipt_url)
        .bind(params.is_tax_deductible)
        .bind(&params.tax_amount)
        .bind(&params.notes)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateInvoice` (Go `UpdateInvoiceParams`).
pub struct UpdateInvoiceParams {
    pub client_name: String,
    pub client_email: Option<String>,
    pub client_address: Option<String>,
    pub client_tax_id: Option<String>,
    pub customer_id: Option<i64>,
    pub issue_date: chrono::NaiveDate,
    pub due_date: chrono::NaiveDate,
    pub currency: Option<String>,
    pub subtotal: Option<String>,
    pub discount_type: String,
    pub discount_value: Option<String>,
    pub discount_amount: Option<String>,
    pub tax_rate: Option<String>,
    pub tax_amount: Option<String>,
    pub total: Option<String>,
    pub amount_due: Option<String>,
    pub notes: Option<String>,
    pub terms: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateInvoice`: `exec`.
pub async fn update_invoice<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateInvoiceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE finance_invoices SET\n  client_name = ?, client_email = ?, client_address = ?, client_tax_id = ?, customer_id = ?,\n  issue_date = ?, due_date = ?, currency = ?, subtotal = ?, discount_type = ?,\n  discount_value = ?, discount_amount = ?, tax_rate = ?, tax_amount = ?,\n  total = ?, amount_due = ?, notes = ?, terms = ?\nWHERE id = ? AND business_id = ? AND status = 'draft'")
        .bind(&params.client_name)
        .bind(&params.client_email)
        .bind(&params.client_address)
        .bind(&params.client_tax_id)
        .bind(params.customer_id)
        .bind(params.issue_date)
        .bind(params.due_date)
        .bind(&params.currency)
        .bind(&params.subtotal)
        .bind(&params.discount_type)
        .bind(&params.discount_value)
        .bind(&params.discount_amount)
        .bind(&params.tax_rate)
        .bind(&params.tax_amount)
        .bind(&params.total)
        .bind(&params.amount_due)
        .bind(&params.notes)
        .bind(&params.terms)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateInvoicePaymentTotals` (Go `UpdateInvoicePaymentTotalsParams`).
pub struct UpdateInvoicePaymentTotalsParams {
    pub amount_paid: Option<String>,
    pub amount_due: Option<String>,
    pub status: String,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateInvoicePaymentTotals`: `exec`.
pub async fn update_invoice_payment_totals<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateInvoicePaymentTotalsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE finance_invoices SET amount_paid = ?, amount_due = ?, status = ? WHERE id = ? AND business_id = ?")
        .bind(&params.amount_paid)
        .bind(&params.amount_due)
        .bind(&params.status)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateInvoicePaymentTotalsWithPaidDate` (Go `UpdateInvoicePaymentTotalsWithPaidDateParams`).
pub struct UpdateInvoicePaymentTotalsWithPaidDateParams {
    pub amount_paid: Option<String>,
    pub amount_due: Option<String>,
    pub status: String,
    pub paid_date: Option<chrono::DateTime<chrono::Utc>>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateInvoicePaymentTotalsWithPaidDate`: `exec`.
pub async fn update_invoice_payment_totals_with_paid_date<
    'e,
    E: sqlx::Executor<'e, Database = MySql>,
>(
    exec: E,
    params: &UpdateInvoicePaymentTotalsWithPaidDateParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE finance_invoices SET amount_paid = ?, amount_due = ?, status = ?, paid_date = ? WHERE id = ? AND business_id = ?")
        .bind(&params.amount_paid)
        .bind(&params.amount_due)
        .bind(&params.status)
        .bind(params.paid_date)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateInvoiceStatus` (Go `UpdateInvoiceStatusParams`).
pub struct UpdateInvoiceStatusParams {
    pub status: String,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateInvoiceStatus`: `execresult`.
pub async fn update_invoice_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateInvoiceStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE finance_invoices\nSET status = ?\nWHERE id = ? AND business_id = ?\n  AND status NOT IN ('paid', 'partial', 'void', 'cancelled')")
        .bind(&params.status)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
