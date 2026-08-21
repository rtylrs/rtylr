# Repository decisions

## HR time-entry ownership

`hr_time_entries.owner_id` is the business ID, matching `business_id`. This is
the convention used by the HR time-entry writer and by Flow when it persists
work-day lines into the HR table. Flow's separate `flow_work_days.owner_id`
continues to identify the resolved account owner. Any cross-module read of
`hr_time_entries` must scope by `business_id` and must not rely on `owner_id`
alone.

## Migration safety

Migration scripts must stop on a non-zero database exit, must not record a
failed file as applied, and must report applied, skipped, and failed counts.
Schema files should still be idempotent so a stopped migration can be safely
retried after the underlying error is corrected.
