//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountExpensesByCategory` (Go `CountExpensesByCategoryParams`).
pub struct CountExpensesByCategoryParams {
    pub business_id: i64,
}

/// `CountExpensesByCategory`: `many`.
pub async fn count_expenses_by_category<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountExpensesByCategoryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT COALESCE(CAST(category_id AS CHAR), 'uncategorised') AS category_key, COUNT(*) AS cnt\nFROM finance_expenses\nWHERE business_id = ?\n  AND created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)\nGROUP BY category_id")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `CountExpensesInCategory` (Go `CountExpensesInCategoryParams`).
pub struct CountExpensesInCategoryParams {
    pub business_id: i64,
    pub category_id: Option<i64>,
}

/// `CountExpensesInCategory`: `one`.
pub async fn count_expenses_in_category<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountExpensesInCategoryParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM finance_expenses\nWHERE business_id = ? AND category_id = ?")
        .bind(params.business_id)
        .bind(params.category_id)
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
    sqlx::query("INSERT INTO finance_expenses (owner_id, business_id, category_id, description, amount, currency, expense_date, vendor, payment_method, receipt_url, is_tax_deductible, tax_amount, notes, status)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending')")
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
}

/// `CreateExpenseCategory`: `execresult`.
pub async fn create_expense_category<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateExpenseCategoryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO finance_expense_categories (owner_id, business_id, name, color, icon)\nVALUES (?, ?, ?, ?, 'tag')")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.color)
        .execute(exec)
        .await
}

/// Params for `DeleteExpense` (Go `DeleteExpenseParams`).
pub struct DeleteExpenseParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteExpense`: `execresult`.
pub async fn delete_expense<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteExpenseParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "DELETE FROM finance_expenses WHERE id = ? AND business_id = ? AND status <> 'approved'",
    )
    .bind(params.id)
    .bind(params.business_id)
    .execute(exec)
    .await
}

/// Params for `DeleteExpenseCategory` (Go `DeleteExpenseCategoryParams`).
pub struct DeleteExpenseCategoryParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteExpenseCategory`: `execresult`.
pub async fn delete_expense_category<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteExpenseCategoryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM finance_expense_categories\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `FindExpenseCategoryByName` (Go `FindExpenseCategoryByNameParams`).
pub struct FindExpenseCategoryByNameParams {
    pub business_id: i64,
    pub name: String,
}

/// `FindExpenseCategoryByName`: `one`.
pub async fn find_expense_category_by_name<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &FindExpenseCategoryByNameParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, color, created_at\nFROM finance_expense_categories\nWHERE business_id = ? AND name = ?\nORDER BY id\nLIMIT 1")
        .bind(params.business_id)
        .bind(&params.name)
        .fetch_optional(exec)
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
    sqlx::query("SELECT id, owner_id, business_id, category_id, description, amount, currency, expense_date, vendor, payment_method, receipt_url, is_tax_deductible, tax_amount, notes, status, created_at, updated_at\nFROM finance_expenses\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetExpenseCategoryByID` (Go `GetExpenseCategoryByIDParams`).
pub struct GetExpenseCategoryByIDParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetExpenseCategoryByID`: `one`.
pub async fn get_expense_category_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetExpenseCategoryByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, color, created_at\nFROM finance_expense_categories\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
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
    sqlx::query("SELECT id, owner_id, business_id, name, color, created_at\nFROM finance_expense_categories\nWHERE business_id = ?\nORDER BY name, id")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListExpenses` (Go `ListExpensesParams`).
pub struct ListExpensesParams {
    pub business_id: i64,
}

/// `ListExpenses`: `many`.
pub async fn list_expenses<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListExpensesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, category_id, description, amount, currency, expense_date, vendor, payment_method, receipt_url, is_tax_deductible, tax_amount, notes, status, created_at, updated_at\nFROM finance_expenses\nWHERE business_id = ?\nORDER BY expense_date DESC, id DESC\nLIMIT 500")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListExpensesPaginated` (Go `ListExpensesPaginatedParams`).
pub struct ListExpensesPaginatedParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListExpensesPaginated`: `many`.
pub async fn list_expenses_paginated<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListExpensesPaginatedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, category_id, description, amount, currency, expense_date, vendor, payment_method, receipt_url, is_tax_deductible, tax_amount, notes, status, created_at, updated_at\nFROM finance_expenses\nWHERE business_id = ?\nORDER BY expense_date DESC, id DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `SumExpensesByBusiness` (Go `SumExpensesByBusinessParams`).
pub struct SumExpensesByBusinessParams {
    pub business_id: i64,
}

/// `SumExpensesByBusiness`: `one`.
pub async fn sum_expenses_by_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumExpensesByBusinessParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(CAST(amount AS DECIMAL(14,2))), 0) AS total\nFROM finance_expenses\nWHERE business_id = ?")
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

/// `UpdateExpense`: `execresult`.
pub async fn update_expense<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateExpenseParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE finance_expenses SET category_id = ?, description = ?, amount = ?, currency = ?, expense_date = ?, vendor = ?, payment_method = ?, receipt_url = ?, is_tax_deductible = ?, tax_amount = ?, notes = ?\nWHERE id = ? AND business_id = ? AND status <> 'approved'")
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

/// Params for `UpdateExpenseCategory` (Go `UpdateExpenseCategoryParams`).
pub struct UpdateExpenseCategoryParams {
    pub name: String,
    pub color: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateExpenseCategory`: `exec`.
pub async fn update_expense_category<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateExpenseCategoryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE finance_expense_categories\nSET name = ?, color = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(&params.color)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateExpenseStatus` (Go `UpdateExpenseStatusParams`).
pub struct UpdateExpenseStatusParams {
    pub status: Option<String>,
    pub id: i64,
    pub business_id: i64,
    pub current_status: Option<String>,
}

/// `UpdateExpenseStatus`: `execresult`.
pub async fn update_expense_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateExpenseStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE finance_expenses\nSET status = ?\nWHERE id = ?\n  AND business_id = ?\n  AND status = ?")
        .bind(&params.status)
        .bind(params.id)
        .bind(params.business_id)
        .bind(&params.current_status)
        .execute(exec)
        .await
}
