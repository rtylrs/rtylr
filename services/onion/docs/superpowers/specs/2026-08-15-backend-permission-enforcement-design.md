# Backend permission enforcement — first pass

## Problem

`hub_roles` (the business-wide custom-role matrix edited via Dash's
Permissions/Roles screens, `internal/modules/roles`) is stored correctly but
never read by any request-handling code outside the `roles` module itself —
confirmed by `grep -rln "modules/roles" internal --include="*.go" | grep -v
"^internal/modules/roles/"` returning nothing. Worse: there is no column
anywhere linking a user to a `hub_roles` row, so the custom-roles feature is
not just unenforced, it's unassignable. Fixing that (a migration, a Team UI
assignment flow, a resolution helper) is a separate, larger project and is
out of scope here.

What _is_ real and assignable today is `AuthUser.Role` (`device.User.Role`),
which the JWT is minted with from `businesses.owner_id` (→ `"owner"`) or
`user_app_permissions.role` otherwise
(`internal/modules/user/internal/infrastructure/repository.go:191`,
`GetUserRole`). The frontend's `packages/permissions/src/capabilities.ts` +
`matrix.ts` already define a fine-grained resource:action matrix for exactly
this 8-role enum (`viewer/staff/store_manager/menu_manager/billing/operator
/admin/owner`) — but it exists only in the frontend bundle; nothing on the
backend enforces it.

An initial audit named several "unenforced" resources. Tracing the actual
call paths (not just grepping for the resource name) found most of them
already covered by a different, real mechanism:

- `hr_leave.ApproveLeave`/`RejectLeave` already call
  `authorizedLeaveDecision`, which checks `hr_core.HasPermission(ctx,
ownerID, employeeID, "approve_time_off")`
  (`internal/modules/hr_leave/internal/application/service.go:290-305`).
- `hr_payroll`'s four handlers already check `manage_payroll` via the same
  hr_core mechanism.
- `customers.DeleteCustomer`/`ExportCustomers`/`MergeCustomers` already call
  `requireCRMManager` (owner/admin/console_admin only) —
  `internal/modules/customers/internal/routes/handlers/authorization.go`.

Two gaps are real and confirmed by reading the handler bodies directly:

1. **`hr_time.ApproveTimesheet` / `RejectTimesheet`** — zero permission
   check. The `Service` has no `PermissionProvider` dependency at all (unlike
   `hr_leave`, which already has one wired).
2. **`invoices` module** (`CreateInvoice`, `SendInvoice`, and others) — zero
   role checks anywhere in the module; no `device.User.Role`,
   `IsBusinessAdmin`, or local helper exists.

## Scope

In scope:

- Wire `hr_time.ApproveTimesheet`/`RejectTimesheet` to `hr_core`'s existing
  `HasPermission(ctx, businessID, employeeID, "approve_time_off")`, following
  the exact adapter pattern `hr_leave` and `hr_payroll` already use.
- Add a small `internal/platform/permissions` package that ports the
  8-role-hierarchy walk from `matrix.ts`/`guards.ts` for the _specific_
  resources checked below — not the full ~90-key matrix, since unused
  entries would drift from the TS source with nothing to catch it.
- Gate `invoices.CreateInvoice` on `finance:invoices:create` and
  `invoices.SendInvoice` on `finance:invoices:send` — both keys already exist
  in `capabilities.ts`, so no new business rule is invented.

Explicitly out of scope (documented in the PR description, not silently
dropped):

- `hub_roles` custom-role enforcement — needs a user→role assignment
  mechanism that doesn't exist yet; a separate, larger project.
- `invoices` update/delete/status-change handlers — `capabilities.ts` has no
  distinct action key for these (only `finance:invoices:create` and
  `finance:invoices:send` are split out); inventing one is a product
  decision, not something to decide silently while wiring enforcement.
- `crm:customers:manage` granting `operator` (per the frontend matrix) vs.
  the backend's current owner/admin-only `requireCRMManager` — a
  stricter-than-intended mismatch, not a security hole. Flagged, not fixed.

## Design

### 1. `hr_time` permission wiring

Mirror `hr_leave`'s existing shape exactly:

- `internal/modules/hr_time/module.go`: add a `HasPermission(ctx,
businessID, employeeID, permissionKey)` method to the existing
  `hrCoreApprovalAdapter` (it already wraps `*hrcore.Module`, which already
  exposes `HasPermission`; the adapter just doesn't forward it yet).
- `internal/modules/hr_time/internal/domain`: add a `PermissionProvider`
  interface (`HasPermission`, `GetEmployeeIDByUserID`) matching
  `hr_leave`'s.
- `internal/modules/hr_time/internal/application/service.go`: add a
  `permissions domain.PermissionProvider` field, thread it through
  `NewService`, and add a `permissionApproveTimeOff = "approve_time_off"`
  constant (same string hr_leave already uses — timesheet and leave
  approval share the same hr_core-seeded concept).
- In `ApproveTimesheet` and `RejectTimesheet`: resolve `approverID` →
  `employeeID` via `s.permissions.GetEmployeeIDByUserID`, then check
  `s.permissions.HasPermission(ctx, businessID, employeeID,
permissionApproveTimeOff)`, returning a new
  `domain.ErrApproverNotAllowed` sentinel when false. Deny-first: unresolvable
  employee ID or a permission-check error both deny, they never fall through
  to allow.
- `module.go`'s `NewModule` passes the same adapter instance for both the
  approval-chain and permission dependencies, exactly as `hr_leave` does.

### 2. `internal/platform/permissions` (new package)

```go
package permissions

type Resource string

const (
    ResourceFinanceInvoicesCreate Resource = "finance:invoices:create"
    ResourceFinanceInvoicesSend   Resource = "finance:invoices:send"
)

// roleHierarchy mirrors packages/permissions/src/roles.ts ROLE_HIERARCHY.
var roleHierarchy = []string{
    "viewer", "staff", "store_manager", "menu_manager", "billing",
    "operator", "admin", "owner",
}

// directPermissions is a PARTIAL mirror of packages/permissions/src/matrix.ts
// DIRECT_PERMISSIONS — only the resources this package currently enforces.
// Extend it (and add a Resource constant above) before checking a new key.
var directPermissions = map[string][]Resource{
    "operator": {ResourceFinanceInvoicesCreate, ResourceFinanceInvoicesSend},
}

// HasCapability reports whether role (a PlatformRole string, case-insensitive)
// grants resource, walking the hierarchy the same way matrix.ts's
// getPermissions does, plus the owner bypass guards.ts's can() applies.
func HasCapability(role string, resource Resource) bool {
    role = strings.ToLower(strings.TrimSpace(role))
    if role == "owner" {
        return true
    }
    idx := indexOf(roleHierarchy, role)
    if idx < 0 {
        return false
    }
    for i := 0; i <= idx; i++ {
        for _, r := range directPermissions[roleHierarchy[i]] {
            if r == resource {
                return true
            }
        }
    }
    return false
}
```

No dependency on `internal/platform/auth` — takes a plain role string so it
stays a pure, independently unit-testable function, callable from any module
without import-cycle risk.

### 3. `invoices` handlers

In `create_invoice_handler.go` and `send_invoice_handler.go`'s `Serve`,
immediately after the existing `device.User == nil` check:

```go
if !permissions.HasCapability(device.User.Role, permissions.ResourceFinanceInvoicesCreate) {
    return CreateInvoiceResponse{}, onekiterrors.PermissionDenied()
}
```

(matching resource constant for send). No change to `Validate`, tenant
resolution, or the service layer — this is a handler-layer gate, consistent
with how `requireCRMManager` and `IsBusinessAdmin` are applied elsewhere in
the codebase.

## Testing

- `internal/platform/permissions`: table-driven unit tests covering the
  hierarchy walk (viewer/staff/etc. denied, operator/admin/owner allowed) and
  the owner bypass.
- `hr_time`: extend the existing service test suite the same way
  `hr_leave`'s tests cover `authorizedLeaveDecision` — an employee without
  `approve_time_off` gets denied, one with it succeeds, `RejectTimesheet`
  covered the same way.
- `invoices`: handler-level test (or integration test, matching whatever
  `invoices` already uses) confirming a non-operator role gets
  `PermissionDenied` on create/send, operator/admin/owner succeed.

## PR framing

The PR description must state plainly what shipped vs. what remains:

- Shipped: `hr_time` approve/reject now enforce `approve_time_off`; invoice
  create/send now enforce role via the new capability package.
- Already covered (no change needed, documented so reviewers don't assume
  otherwise): `hr_leave`, `hr_payroll`, CRM customer delete/export/merge.
- Explicitly not covered by this PR: `hub_roles` custom-role enforcement
  (needs an assignment mechanism first), invoice update/delete/status
  changes (no matching frontend capability key exists to enforce against
  yet), and the CRM operator-vs-admin-only mismatch.
