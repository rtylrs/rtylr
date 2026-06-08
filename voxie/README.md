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

Generated code is intentionally not committed here. Consumers generate their own language output from the checked-out submodule:

- `rtylr`: `pnpm contracts:sync` generates TypeScript into `packages/contracts/src`.
- `onion`: `make contracts` generates Go into the local `voxie/gen/go` checkout before Go builds/tests.

## Usage

```bash
make lint
make format
```

## Adding New Contracts

1. Create or edit `.proto` files under `proto/{app-or-common}/{domain}/v1/`.
2. Keep each endpoint in its own `.proto` file.
3. Put shared cross-app contracts under `proto/common/`.
4. Add comments for new messages and fields.
5. Run `make lint`.
6. Update the Voxie submodule in consumer repos and run their generation commands.

## Tools

- [buf](https://buf.build/) for linting, formatting, and breaking change detection.
- Consumer repos own language-specific code generation.
