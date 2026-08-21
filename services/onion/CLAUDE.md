# Repository decisions

## HR time-entry ownership

`hr_time_entries.owner_id` is the business ID, matching `business_id`. This is
the convention used by the HR time-entry writer and by Flow when it persists
work-day lines into the HR table. Flow's separate `flow_work_days.owner_id`
continues to identify the resolved account owner. Any cross-module read of
`hr_time_entries` must scope by `business_id` and must not rely on `owner_id`
alone.
