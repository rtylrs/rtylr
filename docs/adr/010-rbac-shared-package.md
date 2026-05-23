# ADR-010: RBAC as a Shared Package (@voxire/permissions)

**Status:** Accepted  
**Date:** 2026-03-18  
**Decision:** Role-based access control lives in a shared package consumed by all 14 apps.

## Context

rtylr has complex role requirements:

- **Platform roles**: owner, admin, operator, billing, menu_manager, staff, viewer
- **POS roles**: manager, cashier, server, bartender, kitchen
- **100+ granular capabilities** (Resource types like `menu.read`, `pos.refund`, `hr.payroll.view`)

These permissions must be checked in every app. Options:

1. **Per-app role checks** - each app implements its own permission logic
2. **Backend-only** - server returns allowed actions, frontend just renders
3. **Shared package** - single `@voxire/permissions` package with typed roles, capabilities, and guards

## Decision

`@voxire/permissions` package with:

- `roles.ts` - PlatformRole, POSRole enums with role hierarchy
- `capabilities.ts` - 100+ Resource types as a union type
- `matrix.ts` - permission matrix mapping roles to capabilities, with inheritance
- `guards.ts` - `canAccessApp()`, `canPOSAction()`, `can()`, `canAll()`, `canAny()`

## Rationale

- **Single source of truth**: Permission logic is defined once. All 14 apps import from the same package. No drift between apps.
- **Type safety**: `can(user, 'menu.read')` is type-checked. Typos in capability names are compile errors.
- **Role hierarchy**: Admin inherits operator permissions. Manager inherits cashier permissions. The matrix handles inheritance chains automatically.
- **Frontend + backend**: The same permission definitions can be used in the Go backend (via proto-generated equivalents) for server-side authorization.
- **App-level guards**: `canAccessApp('flow', user.role)` determines which apps appear in the hub menu. No hardcoded role lists.

## Trade-offs

- **Client-side enforcement is not security**: Frontend permission checks are for UX only (hiding buttons). Real authorization MUST be enforced server-side.
- **Package coupling**: All apps depend on `@voxire/permissions`. Changing a capability name requires updating all consumers.
- **Matrix complexity**: 7 platform roles x 100+ capabilities = large permission matrix. Visualization tools may be needed.

## Consequences

- Every app imports: `import { can, canAccessApp } from '@voxire/permissions'`
- Hub app uses `canAccessApp()` to show/hide app tiles
- Feature screens use `can(user, 'resource.action')` to show/hide buttons
- Backend handlers independently verify permissions (defense in depth)
