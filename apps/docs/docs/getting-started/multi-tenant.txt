# Multi-Tenant Architecture

rtylr is built as a multi-tenant platform where each organization operates in complete isolation. This document covers the organizational hierarchy, roles, permissions, and addon gating system.

## Businesses

A business (organization) is the top-level tenant in rtylr. Each business has its own data, team, settings, billing, and addons. Businesses are fully isolated from each other at the database level.

## Branches & Locations

A business can have one or more branches (physical locations). Each branch has its own address, operating hours, tax settings, and inventory. Branches share the same team directory and product catalog but can have branch-specific pricing, stock levels, and settings.

## Floors

Within a branch, floors represent physical areas (e.g., Main Floor, Patio, Rooftop). Floors are used in POS for table mapping. Each floor can have its own layout with tables, stations, or rooms.

## Terminals

Terminals are POS devices registered to a branch and floor. Each terminal has a unique identifier and tracks its own session, cash drawer, and hardware configuration (receipt printer, barcode scanner, payment terminal).

## Roles

rtylr has a layered role system. Organization-level roles control access to apps and settings:

- **Owner** — Full access to everything, including billing and destructive actions. Cannot be removed.
- **Admin** — Full access to all apps and settings except billing and ownership transfer.
- **Operator** — Access to day-to-day operational apps (POS, ERP) with limited settings access.
- **Staff** — Basic access to assigned apps. Cannot modify settings or view financial data.
- **Viewer** — Read-only access to assigned apps. Cannot create, update, or delete anything.

### POS-Specific Roles

In addition to organization roles, POS has its own role layer:

- **Cashier** — Can create orders, process payments, and handle basic operations.
- **Shift Lead** — Cashier abilities plus shift management, discount approval, and limited refunds.
- **POS Manager** — Full POS access including refunds, voids, price overrides, cash drawer management, and day-end reports.

## Permissions

Permissions are granular controls within each app. They define what a user can do: view, create, update, delete, export, and app-specific actions (e.g., approve refunds, run payroll). Permissions are assigned through roles but can be customized per user when needed.

## Team Invite Flow

The team invite process works as follows:

1. An Admin or Owner goes to **Hub > Team** and clicks Invite.
2. Enter the invitee's email address and select a role.
3. Choose which apps the invitee should have access to.
4. Optionally assign to a specific branch.
5. The invitee receives an email with a link to accept the invitation and create their account.
6. Once accepted, the user appears in the team list with their assigned role and app access.

## Addon Gating

Addons are gated at the organization or branch level. When an addon is disabled, its features are hidden from the interface and its API endpoints return `403 Forbidden`. This ensures that teams only see and interact with the features their organization has enabled. Addon state is checked on every request, so changes take effect immediately.
