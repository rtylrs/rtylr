# CLAUDE.md — voxie Contracts

## 1. PROJECT IDENTITY

- **Product:** RTYLR — commerce operating system
- **Repo role:** Protocol Buffer contracts. Single source of truth for all API request/response types shared between the Go backend (`onion`) and the TypeScript frontend (`rtylr`). Neither repo writes its own type definitions for cross-boundary types — they import from here.
- **Built by:** Voxire — Abed El-Fattah Amouneh & Mohammad Homsi
- **Business email:** info@voxire.com
- **WhatsApp:** +961 3 940 708

---

## 2. GIT WORKFLOW — RULES

- Feature branches are the **default**. Prefix: `feat/`, `fix/`, `chore/`, `refactor/`
- **Direct push to `main`** is acceptable ONLY for: trivial fixes to `.proto` comments or formatting. Structural or type-breaking changes always go through a branch.
- **No `.md` files** in commits except `CLAUDE.md` itself.
- **Atomic commits** — one logical unit per commit. A new proto field + its generated code = one commit.
- Commit message format: short imperative, no conventional prefix. `add staff user payload proto` not `feat: add staff user payload proto`.
- **ALWAYS commit both the `.proto` source AND the generated `gen/` files together in one commit.** The generated code is committed — consumers should not need `buf` or `protoc` installed.
- After committing to `voxie`, update the submodule ref in both `rtylr` and `onion` repos.

### Definition of Done

A proto change is only complete when ALL of the following are true:

- `buf lint` passes with zero errors
- `buf breaking` passes against the last tagged version (no breaking changes unless intentional and documented)
- All new fields have a comment describing their purpose
- Generated `gen/go/` and `gen/ts/` files are committed alongside the `.proto` file
- Submodule refs updated in both `onion` and `rtylr`
- `@voxire/contracts/package.json` exports updated in `rtylr` if a new `.proto` file was added
- If a new domain or structural pattern was introduced, it is documented in `rtylr/docs/`

### PR Pre-Merge Checklist

Run through this before merging any proto change:

- [ ] `buf lint` passes with zero output
- [ ] `buf breaking --against '.git#branch=main'` passes (or breaking change is intentional and a v2 package was created)
- [ ] All new fields have inline comments
- [ ] `gen/go/` and `gen/ts/` files were regenerated with `make generate` and committed
- [ ] Submodule ref updated in `onion` (`cd voxie && git pull && cd .. && git add voxie`)
- [ ] Submodule ref updated in `rtylr` (`cd voxie && git pull && cd .. && git add voxie && pnpm contracts:sync`)
- [ ] New export paths added to `packages/contracts/package.json` in `rtylr` (if new `.proto` file added)
- [ ] `rtylr/docs/` updated if the change affects the API contract

### When You Introduce a New Pattern

If you create a new proto convention (a new naming pattern, a new use of `oneof`, a new enum structure, etc.), document it in this CLAUDE.md before the commit is made.

### Updating Documentation

**When you add a new domain, new messages, or make breaking changes that affect the API contract, update `rtylr/docs/`** to reflect the new types available and any migration notes.

---

## 3. TECH STACK

| Component | Technology | Version | Notes |
|---|---|---|---|
| Schema language | Protocol Buffers 3 | proto3 | All `.proto` files use `syntax = "proto3"` |
| Build tool | buf | v2 | Linting, generation, breaking-change detection |
| Go codegen | protoc-gen-go | latest | Generates `.pb.go` files |
| TypeScript codegen | ts-proto | latest | Generates TypeScript interfaces (types only, no runtime) |
| Go module | `github.com/voxire/voxie` | — | Imported in `onion` via `replace` directive |
| TS package | `@voxire/contracts` | — | Re-exports from `voxie/gen/ts/` via `packages/contracts/` in `rtylr` |

### Adding Dependencies

This repo has no runtime dependencies and should stay that way. `buf`, `protoc-gen-go`, and `protoc-gen-ts_proto` are codegen tools only. Do not add Go runtime dependencies to `go.mod`. If you think you need a runtime dependency here, it belongs in `onion` instead.

---

## 4. ARCHITECTURE & PATTERNS

### Repository Layout

```
voxie/
├── proto/                      Source of truth — edit these files
│   ├── auth/v1/
│   │   └── auth.proto
│   ├── common/v1/
│   │   ├── error.proto
│   │   └── pagination.proto
│   ├── crm/v1/
│   ├── erp/v1/
│   ├── finance/v1/
│   ├── flow/v1/
│   ├── hr/v1/
│   ├── hub/v1/
│   ├── insights/v1/
│   └── pos/v1/
├── gen/                        Generated — do not edit directly
│   ├── go/                     Go structs (`.pb.go`)
│   └── ts/                     TypeScript interfaces (`.ts`)
├── buf.gen.yaml
├── buf.yaml
├── go.mod
└── Makefile                    `make generate`
```

### Proto Conventions

- **Package naming:** `{domain}.v1` — e.g., `package hub.v1;`
- **Go option:** `option go_package = "github.com/voxire/voxie/gen/go/{domain}/v1;{domain}v1";`
- **Field naming:** `snake_case` in proto. Generated Go uses camelCase struct fields with `json:"snake_case"` tags.
- **Message naming:** PascalCase. e.g., `StaffUserPayload`, `ListTeamRequest`, `InviteUserResponse`.
- **Version suffix:** always `v1` in the package path.
- **Payload vs Request/Response:** Use `Payload` for reusable data types, `Request`/`Response` for operation-specific envelopes.
- **All new fields must have a comment.**

### Generation Config (`buf.gen.yaml`)

```yaml
version: v2
plugins:
  - local: protoc-gen-go
    out: gen/go
    opt:
      - paths=source_relative

  - local: protoc-gen-ts_proto
    out: gen/ts
    opt:
      - esModuleInterop=true
      - outputServices=false
      - env=browser
      - useExactTypes=false
      - exportCommonSymbols=false
      - onlyTypes=true   # TypeScript types only — zero runtime code
```

### How Consumers Import These Types

**Go backend (`onion`):**

```go
import hubv1 "github.com/voxire/voxie/gen/go/hub/v1"

return handlers.ListUsersResponse{
    Users: []*hubv1.StaffUserPayload{ ... },
}, nil
```

**TypeScript frontend (`rtylr`):**

```ts
import type { StaffUserPayload } from '@voxire/contracts/hub/v1/team';
import type { ProductPayload }   from '@voxire/contracts/erp/v1/products';
```

### Deprecating and Removing Fields

**Never delete a field.** This breaks the JSON wire format for consumers still reading it. Follow this process:

1. **Comment the field as deprecated:** `// Deprecated: use new_field_name instead. Remove after YYYY-MM-DD.`
2. **Reserve the field number and name:**
   ```protobuf
   message UserPayload {
     reserved 5;
     reserved "old_field";
   }
   ```
3. Keep the field in `gen/` until all consumers have been updated.
4. Coordinate removal across `onion` (stop writing) and `rtylr` (stop reading) before dropping the `reserved` block.

### Enum Evolution Rules

Enums are the most dangerous part of a proto schema to change. Follow these rules:

- **Adding a new enum value is backwards compatible.** Old consumers will receive the value as the unknown default (0). Document new values clearly.
- **Never rename an existing enum value.** The name is part of the JSON serialisation — renaming it is a breaking change equivalent to renaming a field.
- **Never remove an enum value.** Reserve the number and name instead:
  ```protobuf
  enum OrderStatus {
    ORDER_STATUS_UNSPECIFIED = 0;
    ORDER_STATUS_PENDING     = 1;
    ORDER_STATUS_COMPLETED   = 2;
    reserved 3;              // was: ORDER_STATUS_CANCELLED (removed 2025-01-10)
    reserved "ORDER_STATUS_CANCELLED";
  }
  ```
- **Never reuse a reserved enum number.** Future additions must use the next available number.

### proto3 Default Values and Optionality

`proto3` uses zero values for all missing fields: `0` for integers, `""` for strings, `false` for bools. This creates ambiguity — **`active: false` could mean "inactive" or "the field was not set".**

When the distinction matters, use `optional`:

```protobuf
message ProductPayload {
  int64  id      = 1; // never optional — always present
  string name    = 2; // never optional — always present
  optional bool active = 3; // nullable — absence is meaningful
}
```

Generated Go code will use a pointer (`*bool`) for `optional` fields. TypeScript will use `boolean | undefined`. Document in a comment why a field is `optional`.

Do **not** use `google.protobuf.BoolValue` wrappers — the `optional` keyword achieves the same with less boilerplate.

### When to Create a v2 Package

Create `proto/{domain}/v2/` only for genuinely breaking changes that cannot be handled via deprecation:

| Change | Breaking? | Action |
|---|---|---|
| Add a new field | No | Add to v1 |
| Mark a field deprecated | No | Add comment + `reserved` |
| Rename a field | **Yes** | Create v2 |
| Rename a message | **Yes** | Create v2 |
| Change a field's type | **Yes** | Create v2 |
| Change semantic meaning of a field | **Yes** | Create v2 |
| Add a new enum value | No | Add to v1 |
| Rename/remove an enum value | **Yes** | Create v2 |

### Exports Must Have a Corresponding Proto File

**Do not add export paths to `packages/contracts/package.json` in `rtylr` unless the corresponding `.proto` file exists in `voxie/proto/`.** Stub exports cause TypeScript build failures downstream.

Order of operations when adding a new `.proto` file:
1. Create the `.proto` file in `proto/{domain}/v1/`.
2. Run `make generate`.
3. Commit both together.
4. Add the export path to `packages/contracts/package.json` in `rtylr`.
5. Run `pnpm contracts:sync` in `rtylr`.

---

## 5. CONTRACT DOMAINS

| Domain | Proto package | Covers |
|---|---|---|
| `auth` | `auth.v1` | Login, register, token, password reset |
| `common` | `common.v1` | Error envelope, pagination |
| `hub` | `hub.v1` | Team, billing, config, design, analytics, permissions, API keys, webhooks, notifications, security |
| `erp` | `erp.v1` | Products, inventory, suppliers, purchase orders |
| `hr` | `hr.v1` | Staff, departments, shifts, payroll, time off |
| `crm` | `crm.v1` | Customers, loyalty, campaigns, segments |
| `finance` | `finance.v1` | Invoices, expenses, quotes, bills, bank accounts, journal, reports |
| `flow` | `flow.v1` | Workspaces, projects, tasks, labels |
| `pos` | `pos.v1` | Orders, tables, sessions, terminals, floors, KDS, combos, modifiers, gift cards, loyalty, receipts |
| `insights` | `insights.v1` | Analytics deep-dive |
| `support` | `support.v1` | Helpdesk tickets, inbox, knowledge base |

---

## 6. CODE QUALITY RULES — NON-NEGOTIABLE

1. **Never edit files in `gen/`** directly. They will be overwritten on the next `make generate`.
2. **All new message fields must have a comment** describing their purpose.
3. **Field names are snake_case** in proto — no exceptions.
4. **Never remove or renumber existing fields** without a deprecation plan. Use `reserved` instead.
5. **Never rename an existing field** without creating a v2 package.
6. **Commit generated code alongside proto changes** — always in the same commit.
7. **Run `buf lint` before committing.** `buf breaking` must pass against main.
8. **New proto files follow** `proto/{domain}/v1/{concern}.proto`. One concern per file.
9. **Never add `packages/contracts/package.json` export paths without a corresponding `.proto` source file.**

---

## 7. SECURITY RULES

- **No secrets in proto definitions.** Proto files define types, not values. Never embed API keys, tokens, or credentials as default field values or comments.
- **Sensitive fields must be documented.** If a proto message carries sensitive data (passwords, tokens, PII), add a comment on the field flagging this: `// Sensitive: do not log this field.`
- **Field renaming is a security concern too.** A renamed field could cause consumers to silently stop reading it, leading to auth or permission checks being skipped. Treat renames with the same care as breaking changes.
- **Deprecate, don't delete, auth-related fields.** If you are removing a field used in authentication or authorisation, coordinate the deprecation window carefully with both `onion` and `rtylr` to avoid a security gap.

---

## 8. KNOWN ISSUES

| Issue | Location | Priority |
|---|---|---|
| `@voxire/contracts/package.json` exports include `book/v1/*` and `store/v1/*` paths with no corresponding proto files | `packages/contracts/package.json` (in rtylr) | Medium |
| `buf.yaml` module config not documented — linting rules are opaque | `buf.yaml` | Low |
| No CI workflow to enforce `buf breaking` checks on PRs | — | High |
| Go module is declared as `v0.0.0` — no semantic versioning enforced | `go.mod` | Low |

---

## 9. SESSION STARTUP CHECKLIST

1. `git pull origin main` on the `voxie` repo
2. Check if `onion` and `rtylr` have pending submodule ref updates
3. Re-read this CLAUDE.md
4. Confirm `buf` is installed: `buf --version`

---

## 10. ENVIRONMENT SETUP

### Prerequisites

| Tool | Version | Install |
|---|---|---|
| buf | v2 | https://buf.build/docs/installation |
| protoc-gen-go | latest | `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest` |
| protoc-gen-ts_proto | latest | `npm install -g ts-proto` |

### Generate all code

```bash
make generate
# or directly:
buf generate
```

### Add a new proto file

1. Create `proto/{domain}/v1/{concern}.proto`.
2. Run `make generate`.
3. Commit both the `.proto` file and all changed `gen/` files.
4. In `rtylr`: `cd voxie && git pull && cd ..` → add export to `packages/contracts/package.json` → `pnpm contracts:sync`.
5. In `onion`: `cd voxie && git pull && cd ..` → import from `gen/go/{domain}/v1`.

### Submodule update workflow

```bash
# In onion/
cd voxie && git pull origin main && cd ..
git add voxie
git commit "update voxie contracts"

# In rtylr/
cd voxie && git pull origin main && cd ..
git add voxie
git commit "update voxie contracts"
pnpm contracts:sync
```
