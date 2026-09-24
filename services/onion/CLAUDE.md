# Repository decisions

Run `make generate` before baseline build, vet, lint, and test commands.

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

## Runtime configuration and operations

`.env.example` is the source-of-truth operator index for environment inputs;
run `python3 scripts/check_env_example.py` after adding a production lookup.
Cloudflare R2-style stores use the S3-compatible `S3_BUCKET_NAME`,
`S3_REGION`, `S3_ENDPOINT`, and `S3_FORCE_PATH_STYLE` variables. The old
`R2_*` names are not read by the application.

The API rate limit is 1,000 requests per second per client key, configured in
`internal/platform/server/http/web_server.go`. Logs are sent to Better Stack
when configured; there is no Axiom integration. `scripts/reset-db.sh` delegates
to the canonical `scripts/migrate.sh`; `scripts/goose.sh` is only a
compatibility wrapper for `up`, and `scripts/check_schema_drift.py` scans all
nested module SQL files.
