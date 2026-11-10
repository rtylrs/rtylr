//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountTickets` (Go `CountTicketsParams`).
pub struct CountTicketsParams {
    pub business_id: i64,
}

/// `CountTickets`: `one`.
pub async fn count_tickets<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountTicketsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM support_tickets WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountTicketsByStatus` (Go `CountTicketsByStatusParams`).
pub struct CountTicketsByStatusParams {
    pub business_id: i64,
    pub status: String,
}

/// `CountTicketsByStatus`: `one`.
pub async fn count_tickets_by_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountTicketsByStatusParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM support_tickets WHERE business_id = ? AND status = ?")
        .bind(params.business_id)
        .bind(&params.status)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateMessage` (Go `CreateMessageParams`).
pub struct CreateMessageParams {
    pub ticket_id: i64,
    pub sender_type: String,
    pub sender_name: String,
    pub body: String,
    pub is_internal: Option<bool>,
}

/// `CreateMessage`: `execresult`.
pub async fn create_message<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateMessageParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO support_messages (ticket_id, sender_type, sender_name, body, is_internal)\nVALUES (?, ?, ?, ?, ?)")
        .bind(params.ticket_id)
        .bind(&params.sender_type)
        .bind(&params.sender_name)
        .bind(&params.body)
        .bind(params.is_internal)
        .execute(exec)
        .await
}

/// Params for `CreateTicket` (Go `CreateTicketParams`).
pub struct CreateTicketParams {
    pub business_id: i64,
    pub customer_name: String,
    pub customer_email: String,
    pub subject: String,
    pub status: String,
    pub priority: String,
    pub assigned_to: Option<i64>,
    pub channel: Option<String>,
}

/// `CreateTicket`: `execresult`.
pub async fn create_ticket<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateTicketParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO support_tickets (business_id, customer_name, customer_email, subject, status, priority, assigned_to, channel)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.business_id)
        .bind(&params.customer_name)
        .bind(&params.customer_email)
        .bind(&params.subject)
        .bind(&params.status)
        .bind(&params.priority)
        .bind(params.assigned_to)
        .bind(&params.channel)
        .execute(exec)
        .await
}

/// Params for `FindSupportCustomerByEmail` (Go `FindSupportCustomerByEmailParams`).
pub struct FindSupportCustomerByEmailParams {
    pub business_id: i64,
    pub email: Option<String>,
}

/// `FindSupportCustomerByEmail`: `one`.
pub async fn find_support_customer_by_email<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &FindSupportCustomerByEmailParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM support_customers\nWHERE business_id = ? AND email = ?\nORDER BY id ASC\nLIMIT 1")
        .bind(params.business_id)
        .bind(&params.email)
        .fetch_optional(exec)
        .await
}

/// Params for `GetTicketBusiness` (Go `GetTicketBusinessParams`).
pub struct GetTicketBusinessParams {
    pub id: i64,
}

/// `GetTicketBusiness`: `one`.
pub async fn get_ticket_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTicketBusinessParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT business_id FROM support_tickets WHERE id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetTicketByID` (Go `GetTicketByIDParams`).
pub struct GetTicketByIDParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetTicketByID`: `one`.
pub async fn get_ticket_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTicketByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, customer_name, customer_email, subject, status, priority, assigned_to, channel, created_at, updated_at\nFROM support_tickets WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `IncrementSupportCustomer` (Go `IncrementSupportCustomerParams`).
pub struct IncrementSupportCustomerParams {
    pub column1: String,
    pub name: String,
    pub column3: String,
    pub email: Option<String>,
    pub business_id: i64,
    pub id: i64,
}

/// `IncrementSupportCustomer`: `exec`.
pub async fn increment_support_customer<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &IncrementSupportCustomerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE support_customers\nSET name = CASE WHEN ? <> '' THEN ? ELSE name END,\n    email = CASE WHEN ? <> '' THEN ? ELSE email END,\n    total_tickets = total_tickets + 1,\n    last_ticket_at = NOW()\nWHERE business_id = ? AND id = ?")
        .bind(&params.column1)
        .bind(&params.name)
        .bind(&params.column3)
        .bind(&params.email)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `InsertSupportCustomer` (Go `InsertSupportCustomerParams`).
pub struct InsertSupportCustomerParams {
    pub business_id: i64,
    pub name: String,
    pub email: Option<String>,
}

/// `InsertSupportCustomer`: `exec`.
pub async fn insert_support_customer<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertSupportCustomerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO support_customers (business_id, name, email, total_tickets, last_ticket_at)\nVALUES (?, ?, ?, 1, NOW())")
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.email)
        .execute(exec)
        .await
}

/// Params for `ListMessages` (Go `ListMessagesParams`).
pub struct ListMessagesParams {
    pub ticket_id: i64,
    pub business_id: i64,
}

/// `ListMessages`: `many`.
pub async fn list_messages<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListMessagesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT m.id, m.ticket_id, m.sender_type, m.sender_name, m.body, m.is_internal, m.created_at\nFROM support_messages m\nJOIN support_tickets t ON t.id = m.ticket_id\nWHERE m.ticket_id = ? AND t.business_id = ?\nORDER BY m.created_at ASC")
        .bind(params.ticket_id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListTickets` (Go `ListTicketsParams`).
pub struct ListTicketsParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListTickets`: `many`.
pub async fn list_tickets<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListTicketsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, business_id, customer_name, customer_email, subject, status, priority, assigned_to, channel, created_at, updated_at\nFROM support_tickets WHERE business_id = ?\nORDER BY created_at DESC LIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListTicketsByStatus` (Go `ListTicketsByStatusParams`).
pub struct ListTicketsByStatusParams {
    pub business_id: i64,
    pub status: String,
    pub limit: i32,
    pub offset: i32,
}

/// `ListTicketsByStatus`: `many`.
pub async fn list_tickets_by_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListTicketsByStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, business_id, customer_name, customer_email, subject, status, priority, assigned_to, channel, created_at, updated_at\nFROM support_tickets WHERE business_id = ? AND status = ?\nORDER BY created_at DESC LIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(&params.status)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `UpdateTicket` (Go `UpdateTicketParams`).
pub struct UpdateTicketParams {
    pub column1: String,
    pub subject: String,
    pub status: String,
    pub priority: String,
    pub assigned_to: Option<i64>,
    pub channel: Option<String>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateTicket`: `exec`.
pub async fn update_ticket<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateTicketParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE support_tickets SET\n    subject = CASE WHEN ? <> '' THEN ? ELSE subject END,\n    status = COALESCE(?, status),\n    priority = COALESCE(?, priority),\n    assigned_to = COALESCE(?, assigned_to),\n    channel = COALESCE(?, channel)\nWHERE business_id = ? AND id = ?")
        .bind(&params.column1)
        .bind(&params.subject)
        .bind(&params.status)
        .bind(&params.priority)
        .bind(params.assigned_to)
        .bind(&params.channel)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
