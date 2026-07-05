# voxie

Protocol Buffer contracts for the rtylr platform. Voxie is the source of truth for API request and response shapes shared by the TypeScript frontend and Go backend.

## Structure

```text
voxie/
  proto/
    common/          # Shared cross-app contracts
    crm/             # Customer relations
    erp/             # Inventory and supply chain
    finance/         # Invoicing and accounting
    flow/            # Project management
    hr/              # Human resources
    hub/             # Dashboard, team, billing, developer, etc.
    insights/        # Analytics
    pos/             # POS contracts
```

Go output (`gen/go`) is generated here and committed — `onion` consumes it directly from the submodule checkout, no local generation needed. `rtylr` still generates its own TypeScript: `pnpm contracts:sync` into `packages/voxie/src`.

## Usage

```bash
make lint
make format
make generate   # regenerates gen/go — commit the result
```

## Adding New Contracts

1. Create or edit `.proto` files under `proto/{app-or-common}/{domain}/v1/`.
2. Keep each endpoint in its own `.proto` file.
3. Put shared cross-app contracts under `proto/common/`.
4. Add comments for new messages and fields.
5. Run `make lint` and `make generate`, and commit the resulting `gen/go` changes.
6. Update the Voxie submodule pointer in `onion` and `rtylr`; run `pnpm contracts:sync` in `rtylr`.

## Tools

- [buf](https://buf.build/) for linting, formatting, and breaking change detection.
- `onekit` (github.com/stackxio/onekit) generates the Go HTTP server bindings in `gen/go`.
- `rtylr` owns its own TypeScript code generation.
