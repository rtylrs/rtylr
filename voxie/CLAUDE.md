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
- **Atomic commits** — one logical unit per commit. A new proto field and its comments belong together.
- Commit message format: short imperative, no conventional prefix. `add staff user payload proto` not `feat: add staff user payload proto`.
- **`gen/go` and `gen/ts` are both committed.** CI (`.github/workflows/contracts.yml`) regenerates them from `proto/` on every push to `main` and fails PRs on drift — never hand-edit either. `rtylr` and `pos` consume `gen/ts` directly from the submodule; there is no separate frontend generation step anymore.
- After committing to `voxie`, update the submodule ref in both `rtylr` and `onion` repos.

### Definition of Done

A proto change is only complete when ALL of the following are true:

- `buf lint` passes with zero errors
- `buf breaking` passes against the last tagged version (no breaking changes unless intentional and documented)
- All new fields have a comment describing their purpose
- `make generate` was run and `gen/go` + `gen/ts` are committed alongside the proto change (CI will also catch drift, but don't rely on it)
- Submodule refs updated in `onion`, `rtylr`, and `pos`
- `@voxire/contracts/package.json` exports updated in `packages` (shared by `rtylr` and `pos`) if a new `.proto` file was added
- If a new domain or structural pattern was introduced, it is documented in `rtylr/docs/`

### PR Pre-Merge Checklist

Run through this before merging any proto change:

- [ ] `buf lint` passes with zero output
- [ ] `buf breaking --against '.git#branch=main'` passes (or breaking change is intentional and a v2 package was created)
- [ ] All new fields have inline comments
- [ ] `make generate` run and `gen/go` + `gen/ts` committed (CI enforces this on push to `main`, but PRs must already be clean)
- [ ] Submodule ref updated in `onion` — no local regeneration needed there anymore, `gen/go` ships pre-built
- [ ] Submodule ref updated in `packages` (the `voxie` submodule at `packages/contracts/voxie`) — no local generation needed there either, `rtylr` and `pos` both pick it up via the shared `packages` submodule
- [ ] New export paths added to `packages/contracts/package.json` (if new `.proto` file added)
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
| Build tool | buf | v2 | Linting, formatting, breaking-change detection |
| Validation | bufbuild/protovalidate | via `buf.yaml` deps | Field validation rules on request messages; enforced at runtime via `buf.build/go/protovalidate` |
| HTTP contracts | 1homsi/onekit | pinned tag (see `Makefile` `ONEKIT_VERSION`) | Real Go + proto dependency, not vendored. `onekit/http` annotations define HTTP routes on services; `protoc-gen-onekit-go-http` generates the server bindings into `gen/go` |
| Go codegen | protoc-gen-go + protoc-gen-onekit-go-http | pinned | Owned by `voxie` — `make generate` runs it and the output is committed to `gen/go` |
| TypeScript codegen | ts-proto + protoc-gen-onekit-ts-client | pinned | Owned by `voxie` — `make generate` runs both and the output is committed to `gen/ts`. ts-proto generates message types, onekit generates the typed fetch client per service. |
| Go module | `github.com/voxire/voxie` | — | Imported in `onion` via `replace`; `gen/go` ships pre-built, no local generation needed in `onion` |
| TS package | `@voxire/contracts` (in the `packages` repo) | — | `packages/contracts` carries `voxie` as its own submodule and re-exports `gen/ts` directly; `rtylr` and `pos` both consume it through the shared `packages` submodule, no local generation needed |

### Adding Dependencies

Runtime dependencies are allowed here now, but only for what the generated output itself needs at runtime: Go side currently `github.com/1homsi/onekit`, `buf.build/go/protovalidate`, and transitive deps; JS side is `ts-proto` and `typescript` as devDependencies only (generated `gen/ts` has zero runtime npm dependencies — onekit's ts-client is self-contained, using only `fetch`). Don't add anything voxie's own proto/codegen tooling doesn't require — application logic still belongs in `onion` and the frontend apps.

---

## 4. ARCHITECTURE & PATTERNS

### Repository Layout

```
voxie/
├── proto/                      Source of truth — edit these files
│   ├── common/                  Shared cross-app contracts
│   │   ├── auth/v1/             login, register, verify_email, reset_password, google_auth, invite_accept, ...
│   │   ├── oauth/v1/            OAuth token issuing
│   │   └── pagination/v1/
│   ├── crm/{capability}/v1/
│   ├── erp/{capability}/v1/
│   ├── finance/{capability}/v1/
│   ├── flow/{capability}/v1/
│   ├── hr/{capability}/v1/
│   ├── hub/{capability}/{subdomain}/v1/
│   ├── insights/{capability}/v1/
│   └── pos/{capability}/{subdomain}/v1/
├── gen/go/                     Generated Go output — committed, regenerated by `make generate` / CI
├── gen/ts/                     Generated TypeScript output — committed, regenerated by `make generate` / CI
├── buf.yaml                    buf v2 config — module path proto/, deps: bufbuild/protovalidate
├── buf.gen.yaml                codegen plugin config (Go + TS plugins)
├── buf.lock                    locked buf dependencies
├── go.mod                      requires github.com/1homsi/onekit + protovalidate runtime deps
├── package.json                 devDependencies: ts-proto, typescript (codegen tooling only, no runtime deps)
├── scripts/generate_validation_metadata.py   parses proto/ directly, outputs gen/ts/validation
├── .github/workflows/contracts.yml   lint / breaking / generate+typecheck+commit CI
└── Makefile                    clean / lint / format / breaking / generate / typecheck targets
```

### Proto Conventions

- **Package naming:** mirror the proto path below `proto/`, using dots. For example, `proto/hub/team/members/v1/list_members.proto` uses `package hub.team.members.v1;`.
- **Go option:** mirror the generated path. For example, `option go_package = "github.com/voxire/voxie/gen/go/hub/team/members/v1";`.
- **Field naming:** `snake_case` in proto. Generated Go uses camelCase struct fields with `json:"snake_case"` tags.
- **Message naming:** PascalCase. e.g., `StaffUserPayload`, `ListTeamRequest`, `InviteUserResponse`.
- **Version suffix:** always `v1` in the package path.
- **Payload vs Request/Response:** Use `Payload` for reusable data types, `Request`/`Response` for operation-specific envelopes.
- **All new fields must have a comment.**
- **Endpoint files:** each endpoint request/response pair gets its own proto file named after the operation, such as `login.proto`, `list_members.proto`, or `process_payment.proto`.
- **Reusable types:** payload-only messages live in a local `types.proto` beside the capability that owns them.
- **Shared contracts:** cross-app contracts live under `proto/common/{domain}/v1/`, not under an app root.

### Generation Ownership

Both Go and TypeScript are owned and committed by `voxie` itself — neither consumer regenerates anything locally anymore:

- **Go (`gen/go`):** `make generate` (or CI on push to `main`) runs `protoc-gen-go` + `protoc-gen-onekit-go-http`. `onion` consumes the submodule at whatever commit is checked out.
- **TypeScript (`gen/ts`):** the same `make generate` run also runs `protoc-gen-ts_proto` (message types) + `protoc-gen-onekit-ts-client` (a typed fetch client per service, self-contained — no runtime npm dependencies) + `scripts/generate_validation_metadata.py` (frontend-side `buf.validate` rule enforcement, parsed directly from `proto/`, output to `gen/ts/validation`). `packages/contracts` (in the `packages` repo, shared by `rtylr` and `pos`) carries `voxie` as its own submodule and re-exports `gen/ts` directly.

### How Consumers Import These Types

Import paths include the capability path:

**Go backend (`onion`):**

```go
import membersv1 "github.com/voxire/voxie/gen/go/hub/team/members/v1"

return handlers.ListUsersResponse{
    Users: []*membersv1.StaffUserPayload{ ... },
}, nil
```

**TypeScript frontends (`rtylr`, `pos`) via `@voxire/contracts`:**

```ts
import type { StaffUserPayload } from '@voxire/contracts/hub/team/members/v1/types';
import type { LoginRequest } from '@voxire/contracts/common/auth/v1/login';
import { FloorServiceClient } from '@voxire/contracts/pos/floor/v1/service_client';
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
3. Coordinate removal across `onion` (stop writing) and `rtylr` (stop reading) before dropping the `reserved` block.

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

**Do not add export paths to `packages/contracts/package.json` unless the corresponding `.proto` file exists in `voxie/proto/`.** Stub exports cause TypeScript build failures downstream.

Order of operations when adding a new `.proto` file:
1. Create the `.proto` file in `proto/common/{domain}/v1/` for shared contracts or `proto/{app}/{capability}/{subdomain}/v1/` for app-owned contracts.
2. Run `make lint`.
3. Run `make generate` and commit the `.proto` source together with the regenerated `gen/go` and `gen/ts`.
4. In `onion`: bump the `voxie` submodule pointer — no generation step needed there.
5. In `packages` (shared by `rtylr` and `pos`): bump the `voxie` submodule pointer at `packages/contracts/voxie`; add the export path to `packages/contracts/package.json` if a new `.proto` file was added. No local generation step needed there either.

---

## 5. CONTRACT DOMAINS

| Domain | Package examples | Covers |
|---|---|---|
| `common` | `common.auth.v1`, `common.oauth.v1`, `common.pagination.v1` | Shared auth (login, register, verify email, reset password, Google auth, invite accept), OAuth token issuing, pagination |
| `hub` | `hub.team.members.v1`, `hub.business.v1` | Team, billing, config, permissions, API keys, webhooks, notifications, security |
| `erp` | `erp.product.v1`, `erp.inventory.v1` | Products, inventory, suppliers, purchase orders |
| `hr` | `hr.staff.v1`, `hr.employees.payroll.runs.v1` | Staff, departments, shifts, payroll, time off |
| `crm` | `crm.customer.v1`, `crm.loyalty.v1` | Customers, loyalty, campaigns, segments |
| `finance` | `finance.invoices.v1`, `finance.expenses.v1` | Invoices, expenses, quotes, bills, bank accounts, journal, reports |
| `flow` | `flow.workspace.v1`, `flow.task.v1` | Workspaces, projects, tasks, labels |
| `pos` | `pos.orders.v1`, `pos.orders.payments.v1` | Orders, tables, sessions, terminals, floors, KDS, combos, modifiers, gift cards, loyalty, receipts |
| `insights` | `insights.dashboard.v1`, `insights.product.v1` | Analytics deep-dive |

---

## 6. CODE QUALITY RULES — NON-NEGOTIABLE

1. **Never hand-edit `gen/go`.** It's generated and committed — regenerate it with `make generate`, never patch the output directly.
2. **All new message fields must have a comment** describing their purpose.
3. **Field names are snake_case** in proto — no exceptions.
4. **Never remove or renumber existing fields** without a deprecation plan. Use `reserved` instead.
5. **Never rename an existing field** without creating a v2 package.
6. **Run `make generate` and commit `gen/go` + `gen/ts`** after proto changes.
7. **Run `buf lint` before committing.** `buf breaking` must pass against main.
8. **New proto files follow** `proto/common/{domain}/v1/{endpoint}.proto` or `proto/{app}/{capability}/{subdomain}/v1/{endpoint}.proto`. One endpoint per file.
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
| `@voxire/contracts/package.json` exports include `book/v1/*`, `store/v1/*`, and legacy `auth/v1/auth` paths with no corresponding proto files | `packages/contracts/package.json` (in rtylr) | Medium |
| Go module is declared as `v0.0.0` — no semantic versioning enforced | `go.mod` | Low |

---

## 9. SESSION STARTUP CHECKLIST

1. `git pull origin main` on the `voxie` repo
2. Check if `onion` and `packages` (the `voxie` submodule at `packages/contracts/voxie`) have pending submodule ref updates
3. Re-read this CLAUDE.md
4. Confirm `buf` is installed: `buf --version`

---

## 10. ENVIRONMENT SETUP

### Prerequisites

| Tool | Version | Install |
|---|---|---|
| buf | v1.65.0 | `go install github.com/bufbuild/buf/cmd/buf@v1.65.0` |
| protoc-gen-go | latest | `make install-plugins` |
| protoc-gen-onekit-go-http | pinned (`ONEKIT_VERSION` in Makefile) | `make install-plugins` |
| protoc-gen-onekit-ts-client | pinned (`ONEKIT_VERSION` in Makefile) | `make install-plugins` |
| protoc-gen-ts_proto | pinned (`package.json`) | `make install-js` (pnpm) |

### Generate Consumer Code

```bash
# In voxie/ — regenerates gen/go and gen/ts, and commits both
make generate
```

There is nothing to run in `onion`, `packages`, `rtylr`, or `pos` — they all consume the committed output directly from the submodule.

### Add a new proto file

1. Create `proto/common/{domain}/v1/{endpoint}.proto` for shared contracts or `proto/{app}/{capability}/{subdomain}/v1/{endpoint}.proto` for app-owned contracts.
2. Run `make lint`.
3. Run `make generate` and commit the `.proto` file together with the regenerated `gen/go` and `gen/ts`.
4. In `onion`: `cd voxie && git pull && cd ..` → `git add voxie` → import from the generated capability path under `gen/go/`. No local generation step needed.
5. In `packages`: `cd contracts/voxie && git pull && cd ../..` → `git add contracts/voxie` → add export to `packages/contracts/package.json` if needed → import from `@voxire/contracts/{capability path}`. No local generation step needed.

### Submodule update workflow

```bash
# In onion/
cd voxie && git pull origin main && cd ..
git add voxie
git commit -m "update voxie contracts"

# In packages/ (consumed by both rtylr and pos)
cd contracts/voxie && git pull origin main && cd ../..
git add contracts/voxie
git commit -m "update voxie contracts"
```
