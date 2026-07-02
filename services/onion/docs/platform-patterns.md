# Onion Platform Patterns

These primitives adapt the strongest backend patterns we found in Koussa without
forcing a rewrite of the current modules.

## Runmodes

New runmodes should describe themselves with `internal/platform/runmode.Spec`:
name, database tags, web-server requirement, optional gRPC config, setup,
startup, and route registration. Existing commands can adopt it gradually.

## Jobs

Each module should own background jobs under:

```text
internal/modules/<module>/internal/jobs
```

The module exposes:

```go
func (m *Module) RegisterJobs(ctx context.Context, registry *jobs.Registry)
```

API boot calls `jobs.AutoRegister(ctx, modules)` for discovery/logging. Worker
boot registers the job-owning modules and starts the scheduler.

Jobs can declare a cron schedule directly in their module package:

```go
jobs.Job{
    Name: "billing.sync",
    Schedule: jobs.ScheduleFromEnv("BILLING_SYNC_SCHEDULE", "0 * * * *"),
    Handler: service.Sync,
}
```

Worker knobs:

- `RUN_JOB=<job.name>` runs one job and exits.
- `JOB_TIMEOUT=30m` sets the default per-job timeout.
- `SCHEDULER_ENABLED=false` disables cron execution.
- `SCHEDULER_DISABLED=true` hard-disables cron execution.

Scheduled execution uses a local in-process duplicate guard plus a MySQL named
lock, so two worker replicas do not run the same job at the same time.

## Cache Invalidation

Use `internal/platform/cache/invalidation.RegisterGroups` for service-owned
groups. Default pattern-backed groups are registered at cache init for user,
business, product, config, subscription, and related scopes.

Workers subscribe to both pattern invalidations and group invalidations so a
cache wipe on one replica reaches local in-memory cache on the others.

## Database Tags

New SQL-backed code can use:

```go
dbclient.TaggedPool(dbclient.TagMain)
dbclient.TaggedPool(dbclient.TagReplica)
dbclient.TaggedPool(dbclient.TagLogging)
```

The wrapper satisfies sqlc's DBTX shape and stamps query contexts with the
logical database tag.

## Probes And Profiling

`cmd/api`, `cmd/upload`, `cmd/license`, and `cmd/worker` start a probe server
when `PROBE_PORT` is set or when running in Kubernetes. pprof starts only when
`PPROF_ENABLED=true` or `PPROF_PORT` is set.

## Secure Keys

`internal/platform/securekey` owns random key generation, hashing,
constant-time matching, expiry, revocation, and rotated-key grace validation.

## Typed Workers

`internal/platform/workerqueue` provides a typed JSON worker engine with bounded
workers, ack/nack, panic recovery, and structured message logging.
`internal/platform/sqsqueue` adapts AWS SQS into that interface.

## Module Scaffolding

Use:

```bash
make generate-module NAME=<module_name>
```

The generated module follows the current Onion package shape:
`application`, `domain`, `infrastructure`, `routes`, and `sql`.

## oneKit HTTP Boundary

New or migrated HTTP endpoints should move toward:

```text
frontend <-> oneKit generated client/server <-> Onion application service
```

The generated oneKit server owns request binding, path/query/body decoding,
header checks, and protovalidate validation. Onion handlers should stop
duplicating that validation once an endpoint is on a generated proto service.

Use `internal/platform/onekit` as the Onion bridge:

- `onekit.NewMux()` creates the private `http.ServeMux` passed to generated
  `WithMux(...)` options.
- `onekit.ErrorHandler` keeps generated route failures in the existing Onion
  JSON error envelope.
- `onekit.Mount(...)` registers generated `"METHOD /path"` patterns on chi so
  normal Onion middleware still runs.
- `onekit.Serve(...)` adapts the current `Serve(ctx, req) (res, *ServerError)`
  handler shape while modules migrate internally.

Contract generation remains backwards-compatible:

```bash
make contracts
ONION_REQUIRE_ONEKIT=1 make contracts
```

`make contracts` always generates Go proto types. If
`protoc-gen-onekit-go-http` is installed, it also generates oneKit HTTP server
files. `ONION_REQUIRE_ONEKIT=1` makes missing oneKit generation a hard failure
for modules that have moved to generated HTTP services.
