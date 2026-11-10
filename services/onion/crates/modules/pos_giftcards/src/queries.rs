//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateGiftCard` (Go `CreateGiftCardParams`).
pub struct CreateGiftCardParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub customer_id: Option<i64>,
    pub code: String,
    pub initial_amount: String,
    pub current_balance: String,
    pub customer_name: Option<String>,
    pub customer_email: Option<String>,
    pub expires_at: Option<chrono::DateTime<chrono::Utc>>,
}

/// `CreateGiftCard`: `execresult`.
pub async fn create_gift_card<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateGiftCardParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_gift_cards (owner_id, business_id, customer_id, code, initial_amount, current_balance, customer_name, customer_email, expires_at)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.customer_id)
        .bind(&params.code)
        .bind(&params.initial_amount)
        .bind(&params.current_balance)
        .bind(&params.customer_name)
        .bind(&params.customer_email)
        .bind(params.expires_at)
        .execute(exec)
        .await
}

/// Params for `GetGiftCardByCode` (Go `GetGiftCardByCodeParams`).
pub struct GetGiftCardByCodeParams {
    pub business_id: i64,
    pub code: String,
}

/// `GetGiftCardByCode`: `one`.
pub async fn get_gift_card_by_code<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetGiftCardByCodeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, customer_id, code, initial_amount, current_balance,\n       customer_name, customer_email, is_active, expires_at, created_at, updated_at\nFROM pos_gift_cards WHERE business_id = ? AND code = ?")
        .bind(params.business_id)
        .bind(&params.code)
        .fetch_optional(exec)
        .await
}

/// Params for `GetGiftCardByCodeForUpdate` (Go `GetGiftCardByCodeForUpdateParams`).
pub struct GetGiftCardByCodeForUpdateParams {
    pub business_id: i64,
    pub code: String,
}

/// `GetGiftCardByCodeForUpdate`: `one`.
pub async fn get_gift_card_by_code_for_update<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetGiftCardByCodeForUpdateParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, customer_id, code, initial_amount, current_balance,\n       customer_name, customer_email, is_active, expires_at, created_at, updated_at\nFROM pos_gift_cards WHERE business_id = ? AND code = ? FOR UPDATE")
        .bind(params.business_id)
        .bind(&params.code)
        .fetch_optional(exec)
        .await
}

/// Params for `LinkCustomerToGiftCard` (Go `LinkCustomerToGiftCardParams`).
pub struct LinkCustomerToGiftCardParams {
    pub customer_id: Option<i64>,
    pub business_id: i64,
    pub id: i64,
}

/// `LinkCustomerToGiftCard`: `exec`.
pub async fn link_customer_to_gift_card<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LinkCustomerToGiftCardParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_gift_cards\nSET customer_id = ?\nWHERE business_id = ? AND id = ?")
        .bind(params.customer_id)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `ListGiftCardsByOwner` (Go `ListGiftCardsByOwnerParams`).
pub struct ListGiftCardsByOwnerParams {
    pub business_id: i64,
}

/// `ListGiftCardsByOwner`: `many`.
pub async fn list_gift_cards_by_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListGiftCardsByOwnerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, customer_id, code, initial_amount, current_balance,\n       customer_name, customer_email, is_active, expires_at, created_at, updated_at\nFROM pos_gift_cards WHERE business_id = ? ORDER BY created_at DESC\nLIMIT 200")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateGiftCardBalance` (Go `UpdateGiftCardBalanceParams`).
pub struct UpdateGiftCardBalanceParams {
    pub current_balance: String,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateGiftCardBalance`: `exec`.
pub async fn update_gift_card_balance<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateGiftCardBalanceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_gift_cards SET current_balance = ? WHERE id = ? AND business_id = ?")
        .bind(&params.current_balance)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
