# Onion Platform Patterns

These primitives adapt the strongest backend patterns we found in Koussa without
forcing a rewrite of the current modules.

## Command lifecycle

Commands own process startup and shutdown. Keep command-specific wiring in
`cmd/<command>` and implement the narrow server interfaces from
`internal/platform/server/http` where needed. Shared server behavior belongs
in the platform package; module construction and dependencies belong in the
command composition root.

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

## Probes And Profiling

`cmd/api`, `cmd/upload`, and `cmd/worker` start a probe server
when `PROBE_PORT` is set or when running in Kubernetes. pprof starts only when
`PPROF_ENABLED=true` or `PPROF_PORT` is set. The destructive probe `POST
/stop` requires the `PROBE_STOP_TOKEN` value in the `X-Probe-Token` header;
without that token it is disabled. pprof binds to loopback so operators should
use port-forwarding rather than exposing it on a pod or host interface.

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

HTTP endpoints are contract-first:

```text
frontend <-> oneKit generated client/server <-> Onion application service
```

Each backend module should expose one or more Voxie `.onk` service definitions.
The schema root is `voxie/api`, but that filesystem directory is not an HTTP
prefix: `api/pos/floor/v1/service.onk` generates routes below
`/pos/floor/v1`. Each RPC decorator supplies the remaining operation path,
giving a full route of `/{app}/{module}/{version}/{route}`.

Onion modules implement the generated `FooServiceServer` interfaces directly.
Registration uses the shared oneKit mux:

1. Expose `RegisterAPI(*http.ServeMux)` on the module.
2. Call generated `RegisterFooServiceServer(...)` with `WithMux(...)`.
3. Let `AutoRegisterAPIs` register every module on the shared mux and
   `MountOneKit` attach it once at the router root. Do not add an `/api` mount
   or duplicate generated paths in chi.
4. Keep authorization, tenancy resolution, and domain orchestration in the
   module implementation.

The generated oneKit server owns request binding, path/query/body decoding,
header checks, and contract validation. Module implementations should not
redefine HTTP request structs for endpoints that have generated service
contracts.

### Errors from oneKit handlers

Return errors through `internal/platform/onekiterrors` so module
implementations consistently use the shared backend error constructors:

```go
return nil, onekiterrors.New(http.StatusBadRequest, "Name is required")
```

Do not reintroduce the removed protobuf-era `WithErrorHandler` or per-route
adapter setup. The generated server owns its JSON error envelope.

Legacy chi handlers use the same `internal/platform/onekiterrors` constructors.
The former `internal/platform/errors` server-error contract was removed during
the oneKit migration and must not be reintroduced for new legacy handlers.

Contract generation is owned by the `voxie` repo, not `onion`. `gen/go` ships
pre-committed in the `voxie` submodule — bump the submodule pointer to pick up
new or changed contracts, there is nothing to run locally in `onion`.
