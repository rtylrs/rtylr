# voxie

API contracts for the rtylr platform. Voxie is the source of truth for API request and response shapes shared by the TypeScript frontend and Go backend.

**Migration in progress:** contracts are moving off Protocol Buffers onto [onekit](https://github.com/1homsi/onekit)'s own `.onk` schema language - see "Migrating a service" below. New/actively-changing services should be migrated as they're touched; `proto/` and `api/` coexist until every service has moved.

## Structure

```text
voxie/
  proto/             # Legacy Protocol Buffer contracts - being migrated off, see api/
    common/          # Shared cross-app contracts
    crm/             # Customer relations
    erp/             # Inventory and supply chain
    finance/         # Invoicing and accounting
    flow/            # Project management
    hr/              # Human resources
    hub/             # Dashboard, team, billing, developer, etc.
    insights/        # Analytics
    pos/             # POS contracts
  api/               # .onk contracts (onekit) - the migration target
    hub/business/v1/ # first migrated service
    hub/config/v1/    # second migrated service (references business/v1.Business - a cross-package reference)
```

Go output (`gen/go`) is generated here and committed — `onion` consumes it directly from the submodule checkout, no local generation needed. `rtylr` still generates its own TypeScript: `pnpm contracts:sync` into `packages/contracts/src`.

## Usage

```bash
make lint
make format
make generate   # regenerates gen/go and gen/ts from BOTH proto/ and api/ — commit the result
```

## Adding New Contracts

New services should be written directly in `.onk` under `api/` (see "Migrating a service" below for the pattern) rather than added to `proto/`.

For anything still touching an existing `.proto` service:
1. Edit `.proto` files under `proto/{app-or-common}/{domain}/v1/`.
2. Keep each endpoint in its own `.proto` file.
3. Put shared cross-app contracts under `proto/common/`.
4. Add comments for new messages and fields.
5. Run `make lint` and `make generate`, and commit the resulting `gen/go`/`gen/ts` changes.
6. Update the Voxie submodule pointer in `onion` and `rtylr`; run `pnpm contracts:sync` in `rtylr`.

## Migrating a service to `.onk`

`hub/business/v1` is the reference migration - use it as a template:

1. Create `api/{app-or-common}/{domain}/v1/` mirroring the old `proto/` path exactly. The base HTTP path is inferred from this directory (kebab-casing multi-word segments, e.g. `hr/time_entry` → `/hr/time-entry`) - no `base_path` needed unless the route is irregular.
2. Translate the `.proto` messages and service into `.onk` (see `api/hub/business/v1/models.onk` and `service.onk`). Field-level notes:
   - `int64` fields: onk's default wire encoding is already a JSON string (matches `(onekit.http.int64_encoding) = INT64_ENCODING_STRING`), so drop that option entirely.
   - `(buf.validate.field).required = true` → `@required`.
   - `(buf.validate.field).string.min_len = 1` paired with `required` is redundant with `@required` alone (onk's `@required` already rejects empty strings) - drop the `min_len`.
   - `(buf.validate.field).int64.gt = 0` / `.gte` / `.lt` / `.lte` → `@gt(0)` / `@gte(0)` / `@lt(N)` / `@lte(N)`.
   - `(buf.validate.field).string.pattern = "..."` → `@pattern("...")`.
   - `(buf.validate.field).string.email = true` → `@email`. `.string.uri = true` → `@uri`.
   - `(buf.validate.field).repeated.min_items` / `.max_items` → `@min_items(N)` / `@max_items(N)`.
   - `(onekit.http.query).name = "..."` → `@query("...")`.
   - `optional T field [(onekit.http.nullable) = true]` → `field: T? @nullable`.
   - A cross-field/message-level constraint (e.g. `buf.validate.message.oneof`) has no onk decorator equivalent yet - needs a hand-written check in the handler for now.
3. Run `onek check .` to validate, then `onek build .` (or `make generate`) to generate.
4. Diff the generated routes against the old ones (`grep 'mux.Handle' gen/go/.../server_http.pb.go` vs the new `server.gen.go`) - they must match exactly.
5. Before deleting anything, check whether other `.proto` files import this service's types (`grep -rl "{app}/{domain}/v1" proto/`) - a proto file can't import a type that only exists in `.onk` anymore. Either migrate the dependent service in the same PR (onk resolves cross-directory references automatically - just reference the type by name, no import statement needed), or leave this service's proto types.proto in place as a shim until the dependent is migrated too.
6. Delete the old `.proto` files and their `gen/go`/`gen/ts` output for that service, and add its `proto/{path}` to `BREAKING_EXCLUDE` in the `Makefile` (deleting the file is intentional, not a regression `buf breaking` should flag).
7. Commit, and prove it end-to-end: compile the generated Go, spin up the server with a test implementation, and exercise it over real HTTP - don't just confirm it generates without errors.

## Tools

- [buf](https://buf.build/) for linting, formatting, and breaking change detection of the remaining `proto/` contracts.
- `onekit`/`onek` (github.com/1homsi/onekit) generates Go/TypeScript types, validation, HTTP server, and client code from `.onk` contracts in `api/`.
- `rtylr` owns its own TypeScript code generation.
