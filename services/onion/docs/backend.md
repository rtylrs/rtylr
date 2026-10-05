# Onion backend

Onion is the Go backend in `services/onion`, with module path `github.com/voxire/onion`. Read `go.mod` for the required Go version. `cmd/api` composes HTTP capabilities; `cmd/worker` composes background jobs and event consumers. API contracts and generated transports are maintained in the repository's `voxie` project.

## Boundaries

`internal/modules/<context>` owns one business context. Its child packages group cohesive capabilities. Domain values and rules are independent of database and HTTP libraries. Application code depends on consumer-defined ports. Infrastructure implements those ports and maps persistence values. API adapters map transport requests and responses explicitly. Use only the layers a capability needs.

Mutable business state has one owner. Keep its schemas, migrations, queries, sqlc configuration, repositories and transaction boundaries together. Cross-context SQL joins, writes, persistence imports and shared SQL transactions violate that boundary. Owner APIs and durable facts support integration. Consumer projections and cache invalidation belong to the consumer.

See [persistence ownership](sqlc-query-ownership.md) for the context map, consistency guarantees, pricing reads, POS inventory facts and migration identities. The architecture check and [exception register](architecture-exceptions.json) provide mechanical enforcement and specific accepted deviations.

## Contracts and registration

Change HTTP contracts in `voxie/api` and regenerate them with `make -C voxie generate` from the repository root. Do not edit or commit ignored generated contracts. Keep generated SQL models private to repositories and map them to application/domain values.

The composition roots initialize modules, connect consumer ports and register transports. They do not implement business adapters or persistence operations. Consumer adapters live with their modules. Import aliases are used for genuine package-name collisions; directories and package declarations should otherwise agree.

Resolve tenant scope from authenticated context and enforce permissions on the server. A visible client action does not grant authorization. Return failures honestly, including unresolved business scope and unavailable dependencies.

## Validation

From `services/onion`:

```sh
make architecture-check
go test ./...
./scripts/test-db.sh
```

The existing database test script uses an isolated test database, replays owned migration sources and loads synthetic fixtures. Inspect its environment settings before running it. Do not point tests at production. Generated sqlc output must be regenerated after query or schema changes. Database migrations and service deployment are separate operations.
