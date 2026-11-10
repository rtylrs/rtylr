//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountAppointmentForOwner` (Go `CountAppointmentForOwnerParams`).
pub struct CountAppointmentForOwnerParams {
    pub id: i64,
    pub business_id: i64,
}

/// `CountAppointmentForOwner`: `one`.
pub async fn count_appointment_for_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountAppointmentForOwnerParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM booking_appointments WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateReminder` (Go `CreateReminderParams`).
pub struct CreateReminderParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub appointment_id: Option<i64>,
    pub r#type: String,
    pub scheduled_at: Option<chrono::DateTime<chrono::Utc>>,
}

/// `CreateReminder`: `execresult`.
pub async fn create_reminder<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateReminderParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO booking_reminders (owner_id, business_id, appointment_id, type, scheduled_at) VALUES (?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.appointment_id)
        .bind(&params.r#type)
        .bind(params.scheduled_at)
        .execute(exec)
        .await
}

/// Params for `ListReminders` (Go `ListRemindersParams`).
pub struct ListRemindersParams {
    pub business_id: i64,
}

/// `ListReminders`: `many`.
pub async fn list_reminders<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListRemindersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, appointment_id, type, scheduled_at, sent_at, created_at\nFROM booking_reminders WHERE business_id = ? ORDER BY scheduled_at DESC LIMIT 100")
        .bind(params.business_id)
        .execute(exec)
        .await
}
