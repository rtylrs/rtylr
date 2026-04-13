# voxie

Protocol Buffer contracts for the rtylr platform. Single source of truth for API request/response types, shared between frontend (TypeScript) and backend (Go).

## Structure

```
voxie/
  proto/
    common/          # Shared types (pagination, errors, timestamps)
    auth/            # Authentication contracts
    hub/             # Hub (dashboard, team, billing, developer, etc.)
    pos/             # POS terminal contracts
    erp/             # Inventory & supply chain
    hr/              # Human resources
    crm/             # Customer relations
    finance/         # Invoicing & accounting
    flow/            # Project management
    insights/        # Analytics
  gen/
    go/              # Generated Go types
    ts/              # Generated TypeScript types
```

## Usage

### Generate code

```bash
make generate
```

### As a git submodule

**Frontend (rtylr):**
```bash
git submodule add git@github.com:voxire/voxie.git voxie
```

**Backend (onion):**
```bash
git submodule add git@github.com:voxire/voxie.git voxie
```

Then import generated types:

**TypeScript:**
```typescript
import { LoginRequest, LoginResponse } from '../voxie/gen/ts/auth/auth';
import { ListTeamResponse, TeamMember } from '../voxie/gen/ts/hub/user';
```

**Go:**
```go
import hub "github.com/voxire/voxie/gen/go/hub"
```

## Adding new contracts

1. Create or edit `.proto` files in `proto/{module}/`
2. Run `make generate`
3. Commit both `.proto` and `gen/` files
4. Pull submodule updates in frontend + backend repos

## Tools

- [buf](https://buf.build/) for linting, generation, and breaking change detection
- [ts-proto](https://github.com/stephenh/ts-proto) for TypeScript output
- [protoc-gen-go](https://pkg.go.dev/google.golang.org/protobuf/cmd/protoc-gen-go) for Go output
