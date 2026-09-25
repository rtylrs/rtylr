# Tenant Index Matrix

Issue #170 adds business-leading indexes only where a live Onion query filters
by `business_id` and the existing owner/workspace-leading keys cannot serve its
order, range, grouping, or join shape.

| Domain | Query shape | Index |
| --- | --- | --- |
| Booking appointments | `business_id` plus start-time list/range | `idx_booking_appts_business_start (business_id, start_time, id)` |
| Booking appointments | `business_id`, status, start-time range | `idx_booking_appts_business_status_start (business_id, status, start_time)` |
| Store orders | `business_id` ordered by creation time | `idx_store_orders_business_created (business_id, created_at, id)` |
| Flow tasks | `business_id`, completed status, completion window | `idx_flow_tasks_business_status_completed (business_id, status, completed_at)` |
| Flow tasks | `business_id` grouped by assignee | `idx_flow_tasks_business_assignee (business_id, assignee_id)` |
| Flow roadmap | business-scoped projects joined to root tasks | `idx_flow_projects_business_created (business_id, created_at, id)` and `idx_flow_tasks_project_parent (project_id, parent_task_id)` |
| Finance bills | business-scoped due-date queue and pending totals | `idx_finance_bills_business_due (business_id, due_date, id)` and `idx_finance_bills_business_status (business_id, status)` |
| Finance bank accounts | business-scoped balance aggregation | `idx_finance_bank_accounts_business (business_id)` |
| Finance journal entries | business-scoped history and status totals | `idx_finance_journal_entries_business_date (business_id, date, id)` and `idx_finance_journal_entries_business_status (business_id, status)` |

All indexes are additive and mirrored by idempotent schema migrations. The
schema PR documents rollback as an explicit `DROP INDEX` only after checking
production query plans; existing owner/workspace-leading indexes remain for
their separate access paths.
