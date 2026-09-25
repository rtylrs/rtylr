# RTYLR Role, License, and Permission Model

## Implementation plan

Status: design and implementation plan

This document is the handoff plan for a future implementation session. It is
based on the current `rtylr` frontend, the `onion` backend, the shared
permission package, the existing HR and POS role systems, and the current
cross-app architecture documents.

No implementation is included in this document. The next session should use
this plan as the source of truth, update the affected contracts and code in
dependency order, and keep the migration reversible until the new model has
been verified in production-like tenants.

## 1. Objective

Replace the current single global role plus flat application list with a hub
role and independent application role assignments.

The system must support all of the following:

- Dash as the common corporate hub for every application
- A global Dash role for common workspace and corporate permissions
- Different roles in different applications for the same person
- Multiple roles in an application where that application supports it
- Corporate hierarchy inside every application
- Custom roles created immediately, not in a later phase
- Granular view, create, edit, delete, approve, publish, send, export, and
  manage permissions
- Location, department, team, project, and other scope restrictions
- Independent role upgrades, downgrades, license changes, and removals
- POS cashiers with optional void and refund capabilities
- Flow Team Leaders, Managers, and Super Managers
- Server-side authorization that cannot be bypassed by hiding or showing UI
- Complete audit history for role, license, permission, and scope changes
- A reversible migration from the current role model

## 2. Confirmed product decisions

These decisions came from the product discussion and should not be silently
changed during implementation.

### 2.1 Dash is the hub

Dash is not a normal licensed application. Every business user has a Dash
identity and can enter Dash according to their Dash role.

Applications such as POS, ERP, HR, CRM, Finance, Flow, Insights, Menu
management, and Recruit receive independent licenses and app roles.

The public customer-facing Menu pages and public Recruit pages are not
employee applications and do not need internal employee role checks.

### 2.2 Global Dash roles

The standard global Dash roles are:

- Owner
- Admin
- Manager
- Member
- Viewer

The Dash role controls the common hub. It also participates in the corporate
hierarchy and can provide app-role recommendations or defaults, but it does
not force the same app role in every application.

Examples:

- Dash Manager, Flow Manager, Finance Viewer
- Dash Manager, Flow Super Manager, HR Employee
- Dash Member, POS Senior Cashier, CRM Sales User
- Dash Viewer, Insights Viewer

### 2.3 Application roles

Every selected application must have one or more app roles selected during
user creation. The app role is independent per application.

The same user can be a Manager in Flow and a Viewer in Finance. Changing the
Finance role must not change the Flow role.

Removing a license revokes the application access immediately. The previous
role assignment remains as inactive history so that access can be restored or
audited without losing the previous state.

### 2.4 Multiple roles

Multiple application roles should be supported where the application can
combine them safely. Permissions from assigned roles are combined.

Each application can declare its assignment rule:

- `many`: multiple roles combine, with a union of permissions
- `one`: one effective primary role is required
- `primary_plus_addons`: one primary role plus additional capability roles

Initial application rules:

- POS: `primary_plus_addons`
- Flow: `many`, with scope restrictions handled internally
- HR: `one` effective HR employee role
- ERP: `many`
- Finance: `many`
- CRM: `many`
- Insights: `many`
- Menu management: `many`
- Recruit: `many`

The UI should show a primary display role even when multiple roles are
assigned.

### 2.5 Role hierarchy

Every application has a corporate hierarchy. The default levels are:

1. Admin
2. Super Manager
3. Manager
4. Team Leader or Supervisor
5. Operator, Contributor, or Specialist
6. Viewer

Not every application must show every level. The application can use the
most understandable job title for each level.

The hierarchy level is metadata used for inheritance, delegation, and display.
It must not be the only authorization check. Actual permissions and scope are
always evaluated.

### 2.6 Custom roles

Custom roles are available in the first release of this model.

A custom role contains:

- Application or Dash hub ownership
- Display name and description
- Corporate hierarchy level
- Optional parent role
- Exact permissions
- Scope rules
- Whether it can manage users in its own application
- Whether it can create or edit other custom roles
- Creation, update, archive, and assignment history

Custom roles must be assignable to users and groups. A custom role that is
assigned to users cannot be hard-deleted. It must be archived or migrated.

Dash is the only control plane for cross-app role and license management. HR
permission-role definitions remain managed in HR as a deliberate exception,
and Flow project membership remains managed in Flow as project scope.

This does not prevent applications from managing access to resources inside
their own workflows. For example, Flow may allow an authorized manager to add
an employee to another project, assign the employee to a task, or add the
employee to an operational team. That changes Flow resource membership only.
It must not change the user's Dash role, Flow app role, application license,
global permissions, corporate hierarchy, or cross-app assignment.

Each app-owned resource-access action requires a dedicated permission, scope
check, audit event, and server-side authorization. Resource membership must not
be modeled as a hidden role or used as an alternate role-management API.

### 2.7 Role changes and upgrades

Role assignment is mutable by design.

The system must support:

- Promotion and downgrade
- Adding an application
- Removing an application
- Changing one app role without touching other apps
- Adding or removing a secondary role
- Moving a scope from one location, department, team, or project to another
- Editing a custom role and immediately recalculating its assignees
- Deactivating a user and revoking all effective app access
- Restoring a previously removed assignment
- Auditing the before and after state

Explicit app assignments must not be overwritten by a Dash role change.
Assignments created only as defaults must remain distinguishable from explicit
assignments.

## 3. How Dash roles relate to app roles

Dash roles are useful even when they do not automatically assign the same role
in every app.

Dash roles provide four important functions:

1. They control access to the shared Dash hub.
2. They define common corporate hierarchy and delegation authority.
3. They control who can invite users, assign licenses, create roles, and manage
   groups.
4. They can provide suggested app-role defaults without overriding explicit
   app decisions.

### 3.1 Default and explicit assignments

An application assignment must record its source:

- `explicit`: selected directly during invitation or editing
- `dash_default`: suggested or inherited from a Dash role
- `group`: provided by a group membership
- `system`: Owner or Admin access override

Recommended resolution:

1. Owner and Admin receive the system-level application access override when
   the business subscribes to the application.
2. For other users, an app must be explicitly licensed or granted through a
   group.
3. A selected app must have at least one app role.
4. Explicit user assignments override Dash defaults.
5. Group roles combine with direct user roles unless the business chooses a
   replacement policy for that group.
6. Multiple compatible roles combine their permissions.
7. Scope restrictions are applied after role permissions are resolved.
8. Removing the license removes effective access even if the inactive role
   assignment is retained for history.

Dash Manager therefore does not automatically become Manager everywhere. A
business may configure recommendations such as:

```text
Dash Manager defaults:
  Flow: Manager
  Finance: Viewer
  HR: Employee
  CRM: no default
```

These defaults are suggestions or inherited assignments only. The invitation
form must still show the role for each selected application and permit an
explicit choice.

### 3.2 Owner and Admin access

Owner and Admin can access all subscribed applications without a normal
per-user app assignment. This is an authorization override, not a replacement
for business subscription entitlements.

The implementation must keep these concepts separate:

- Is the business subscribed to the application?
- Does this user have a license assignment?
- Does this user have an application role?
- Is this user Owner or Admin with an access override?

If billing later counts seats separately, that billing calculation must use the
license model and must not be inferred only from authorization.

### 3.3 Dash Manager

Dash Manager is the fifth global role. Its exact permissions should be
customizable, but the standard role should be below Admin.

Dash Manager can be given:

- Broad visibility into the hub
- Team and organization visibility
- App assignment visibility
- Audit visibility
- Operational group management
- Limited team actions

Dash Manager must not automatically receive:

- Owner transfer
- Global role administration
- Unrestricted license administration
- Security administration
- API key management
- Billing ownership actions
- Business deletion

These permissions should be represented as individual Dash permissions so that
the business can create a stronger or weaker custom Dash Manager role.

## 4. Standard hierarchy and role templates

The role names below are starting system templates. All of them can be copied
into custom roles and adjusted.

### 4.1 POS

- POS Admin
- Location Manager
- POS Manager
- Shift Supervisor
- Cashier
- Senior Cashier
- Server
- Bartender
- Kitchen
- Reports Viewer

`Senior Cashier` is a standard capability role or add-on that includes the
Cashier role plus optional void and refund permissions.

### 4.2 ERP

- ERP Admin
- Super Manager
- Inventory Manager
- Warehouse Manager
- Purchasing Manager
- Catalog Manager
- Inventory Operator
- Purchasing Operator
- Catalog Editor
- ERP Viewer

### 4.3 Finance

- Finance Admin
- Super Manager
- Finance Manager
- Accountant
- Accounts Payable Manager
- Billing Specialist
- Invoicing Operator
- Financial Analyst
- Finance Viewer

The old global Billing role should become a Finance role. Dash billing pages
should use a separate Dash permission or a custom Dash Billing Manager role.

### 4.4 CRM

- CRM Admin
- Super Manager
- CRM Manager
- Sales Manager
- Sales Representative
- Customer Success User
- Marketing Manager
- Marketing Operator
- Loyalty Manager
- Customer Data Steward
- CRM Viewer

### 4.5 HR

- HR Admin
- Super Manager
- HR Manager
- HR Staff
- Payroll Manager
- Employee
- Executive Viewer
- HR Viewer

HR should have one effective role per employee. Multiple permission needs should
be combined into one custom HR role rather than assigning several HR employee
roles.

### 4.6 Flow

- Flow Admin
- Super Manager
- Program Manager
- Project Manager
- Team Leader
- Contributor
- Task Operator
- Flow Viewer

Use `Super Manager` as the standard display title for the person who manages
Managers. Permit a business to display a custom title such as Program Manager,
Portfolio Manager, or Head of Projects while retaining the same internal role
level and permission structure.

### 4.7 Insights

- Insights Admin
- Super Manager
- Analytics Manager
- Analyst
- Executive Viewer
- Insights Viewer

### 4.8 Menu management

- Menu Admin
- Super Manager
- Menu Manager
- Menu Editor
- Branding Editor
- Menu Viewer

### 4.9 Recruit

- Recruit Admin
- Super Manager
- Recruiting Manager
- Recruiter
- Recruiting Coordinator
- Sourcer
- Interviewer
- Recruit Viewer

## 5. Current role inventory and migration mapping

The current shared platform roles are defined in
`packages/types/src/user.ts`, `packages/permissions/src/roles.ts`, and
`packages/permissions/src/matrix.ts`.

### 5.1 Current shared roles

| Current role    | Current behavior                                                                                               | Migration treatment                                                                                                                       |
| --------------- | -------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `owner`         | Full account access, all inherited permissions, owner-only payroll execution, reserved ownership               | Keep as Dash Owner. Preserve owner invariants.                                                                                            |
| `admin`         | Broad Dash and cross-app administration, team management, settings, approvals, terminals, sensitive operations | Keep as Dash Admin. Give system app access to all subscribed apps.                                                                        |
| `operator`      | Cross-app operational access across POS, ERP, CRM, HR, Finance, Flow, Insights, Menu, and some Dash functions  | Do not map blindly. Split into app-specific Manager, Super Manager, or Operator roles based on current permissions and review exceptions. |
| `billing`       | Dash billing plus Finance invoices, expenses, and reports                                                      | Convert to Dash Member plus Finance Billing Specialist, with a separate Dash billing permission where required.                           |
| `menu_manager`  | Menu and design access, currently inherits Staff permissions                                                   | Convert to Dash Member plus Menu Manager. Remove unrelated inherited POS, Flow, and HR access.                                            |
| `store_manager` | Intended as a POS location manager, with location flags; currently inherits Staff permissions                  | Convert to Dash Member plus POS Location Manager and preserve assigned locations.                                                         |
| `staff`         | Basic POS, Flow, and HR self-service access                                                                    | Convert to Dash Member plus explicit POS, Flow, and HR roles based on the user’s current assignments.                                     |
| `viewer`        | Dash dashboard and Insights dashboard access                                                                   | Convert to Dash Viewer. Preserve explicitly granted app viewers.                                                                          |

The current matrix has an inheritance issue where Admin receives Billing
permissions through Operator, despite the invitation copy saying Admin excludes
billing. The migration must use explicit new permissions and remove accidental
inheritance.

### 5.2 Current POS roles

The current POS roles are:

- `manager`: void, refund, discounts, cash pickup, employee management,
  terminal management, and reports
- `cashier`: drawer access, cash drops, and discounts
- `server`: basic drawer and service operations
- `bartender`: service role without default management capabilities
- `kitchen`: kitchen role without default management capabilities

Keep these as POS templates, but split the individual actions. Do not make
void and refund exclusive to POS Manager.

### 5.3 Current Flow roles

Flow currently has business roles `owner`, `admin`, `member`, and `viewer`, plus
project roles `admin`, `member`, and `viewer`.

The current code also grants a business-role bypass to global Owner, Admin,
and Operator. Replace this with the new Flow role and scope resolver.

The project role should not be another role selector during Dash user
creation. It should become a project scope policy inside Flow.

### 5.4 Current HR roles

HR contains two overlapping vocabularies:

- Setup defaults: Employee, Manager, HR Admin, Finance/Payroll Manager,
  Executive
- Helper roles: Administrator, Manager, HR Staff, Employee, Viewer

HR also has custom roles stored in `hr_roles` and assignments in
`hr_employee_roles`.

Consolidate these into the app-role model. Keep the existing HR permission
catalog as source material, but expose one coherent HR role system to Dash and
HR.

### 5.5 Current Recruit roles

Recruit contains labels for Recruiter, Recruiting Coordinator, Sourcer,
Recruiting Manager, and Interviewer. These are currently team labels rather
than a complete authorization model. Convert them into Recruit role templates
and add permissions for candidate data, postings, submissions, page content,
and branding.

### 5.6 Current custom roles and group permissions

Dash currently has a custom role screen, but custom roles are not fully
assignable or enforced. HR has a more complete custom role flow. Groups have
separate resource permission levels.

The new model must make one role system authoritative and either migrate or
retire the old group permission vocabulary after dual-read verification.

Backend helper values such as `console_admin` that are outside normal tenant
invites must be documented as system-console roles, not migrated into tenant
Dash roles.

## 6. Permission model

### 6.1 Permission dimensions

Every application feature must distinguish at least:

- `view`: show the page, navigation item, and permitted data
- `create`: create a record
- `edit`: update a record
- `delete`: delete a record
- `archive`: archive or restore a record
- `approve`: approve a workflow or financial action
- `publish`: publish public content
- `send`: send an invoice, campaign, or communication
- `export`: export data
- `assign`: assign work, people, locations, or resources
- `manage`: manage settings or the feature as a whole
- `configure`: change application configuration
- `execute`: run payroll or another irreversible process
- `override`: bypass an operational restriction
- `sensitive.view`: view protected fields

Hiding a route is not permission enforcement. Each action must be protected in
the service and handler layer as well as hidden in the frontend where
appropriate.

### 6.2 Permission identifiers

Keep a stable canonical identifier for every permission. The current code uses
colon-separated resources such as `finance:invoices:create`. The new catalog
may continue that convention, but it must add the missing actions and avoid
using an app-level permission as a substitute for every child permission.

Examples:

```text
finance:invoices:view
finance:invoices:create
finance:invoices:send
finance:accounts:view
finance:accounts:manage
finance:bank_feeds:sensitive_view
```

An app-level assignment grants entry to the app. It does not automatically
grant every action inside the app.

### 6.3 Permission inheritance

System roles may inherit from parent roles. Custom roles may optionally inherit
from a parent role, but the permission editor must show the effective result.

Changing a parent role affects its children and assigned users. Archiving a
parent role requires a migration choice for child roles.

No role may grant permissions outside its application unless it is a Dash hub
role with an explicitly defined hub permission.

## 7. Complete feature and hide/unhide catalog

This catalog is the permission inventory to implement. For each item, the
frontend should hide navigation and controls the role cannot use, while the
backend must deny direct requests.

The standard role behavior is:

- App Admin: all feature visibility and actions in the app
- Super Manager: broad visibility, cross-team or cross-scope management, and
  operational approvals, without global security or license administration
- Manager: full visibility within scope, operational management, assignments,
  and selected approvals
- Team Leader or Supervisor: team-level visibility and actions
- Operator or Specialist: functional actions only
- Viewer: view-only access to permitted data
- Custom role: exact permissions selected by the business

### 7.1 Dash hub

The Dash routes are in `apps/dash/src/router.tsx`.

#### Workspace and personal areas

- Dashboard overview
- Apps catalog
- Application launchers
- Notifications
- Chat and assistant
- Help center
- Theme preferences
- Onboarding
- Account profile
- Business switcher

Permissions:

- View own profile
- View hub dashboard
- View app catalog
- View notifications
- Use assistant
- View help
- Change personal preferences
- Switch businesses

#### Team and access management

- View team directory
- View member details
- Invite users
- Bulk invite users
- Edit member profile
- Change Dash role
- Assign app licenses
- Assign app roles
- Remove app access
- Deactivate users
- Reactivate users
- Remove users
- View access requests
- Approve access requests
- Reject access requests
- Manage groups
- Create groups
- Edit groups
- Delete or archive groups
- Add group members
- Remove group members
- Assign group roles
- Assign group licenses

#### Roles and permissions

- View standard Dash roles
- View application roles
- View permission matrix
- Create custom Dash roles
- Create custom application roles
- Clone roles
- Edit custom roles
- Archive custom roles
- Assign roles to users
- Assign roles to groups
- Configure parent role
- Configure hierarchy level
- Configure app defaults
- Configure scope requirements
- View affected users before saving
- View role assignment history

#### Business and account settings

- Account settings
- Businesses
- General settings
- Industry
- Custom domain
- Email settings
- AI usage
- Subscription
- Add-ons
- Billing
- Payment settings

#### Developer, integrations, and security

- Integrations
- API keys
- Webhooks
- Security settings
- Audit log
- Activity history

#### Menu management inside Dash

- Menus
- Menu items
- Categories and availability
- Banners
- QR codes
- Design
- Branding
- Menu orders
- Menu analytics

Dash-specific sensitive permissions include billing details, API keys, security
settings, audit data, employee personal data, and role administration.

### 7.2 POS

The POS frontend is not present in this checkout. POS backend roles and
contracts are present, so the POS frontend repository must be included in the
implementation handoff.

#### Sales operations

- View orders
- Create orders
- Edit orders
- Hold orders
- Resume held orders
- Void open orders
- Refund completed payments
- Apply discounts
- Override discounts
- Manage tables
- Manage receipts
- Reprint receipts
- Open sessions
- Close sessions

#### Cash operations

- Open drawer
- Cash drop
- Cash pickup
- View cash activity
- Approve cash activity
- View shift totals
- Reconcile drawer

#### Restaurant operations

- Kitchen display
- Bar workflow
- Table service
- Order routing
- Production status
- Service notes

#### POS administration

- Manage POS employees
- Manage employee PINs
- Manage terminals
- Configure terminal behavior
- View POS reports
- Configure location settings
- Assign POS locations
- Manage location-level POS access

#### Special cashier capabilities

Void and refund must be independent permissions:

- `pos:orders:void`
- `pos:orders:refund`
- `pos:orders:refund:approve`
- `pos:orders:void:override`

Create a standard Senior Cashier template or capability role that inherits
Cashier and can receive void and refund permissions without receiving every
POS Manager permission.

Add policy constraints:

- Maximum refund amount
- Maximum discount amount
- Refund approval threshold
- Void reason required
- Refund reason required
- Allowed transaction age
- Allowed locations
- Allowed terminals
- Shift-only restriction

Every override, void, refund, and approval must be audited.

### 7.3 ERP

The ERP routes are in `apps/erp/src/router.tsx`.

#### Catalog

- Dashboard
- Products
- Product details
- Categories
- Brands
- Variants
- Price lists
- Barcodes

Actions:

- View products
- Create products
- Edit products
- Delete products
- Manage categories
- Manage brands
- Manage variants
- Manage prices
- Generate barcodes
- Export catalog

#### Inventory and warehouses

- Inventory
- Warehouses
- Stock transfers
- Stock takes
- Goods receiving
- Expiry tracking
- Low stock alerts
- Reorder suggestions
- Inventory valuation

Actions:

- View stock
- Adjust stock
- Transfer stock
- Approve stock adjustments
- Perform stock takes
- Receive stock
- Manage warehouses
- Configure reorder levels
- View valuation
- Export inventory

#### Purchasing

- Purchase orders
- Purchase order details
- Create purchase orders
- Edit purchase orders
- Approve purchase orders
- Suppliers
- Create suppliers
- Edit suppliers
- Receive purchases

#### ERP configuration

- ERP settings
- Product configuration
- Inventory configuration
- Warehouse configuration
- Purchasing configuration

ERP roles should split Catalog, Inventory, Warehouse, Purchasing, and
management responsibilities instead of assigning all of them to Operator.

### 7.4 Finance

The Finance routes are in `apps/finance/src/router.tsx`.

#### Invoicing and quotes

- Finance dashboard
- Invoice list
- Invoice details
- Create invoice
- Edit invoice
- Delete or cancel invoice
- Send invoice
- Recurring invoices
- Quotes
- Create quote
- Edit quote
- Send quote

#### Expenses and bills

- Expenses
- Expense details
- Create expense
- Edit expense
- Delete expense
- Approve expense
- Expense categories
- Bills
- Create bill
- Edit bill
- Approve bill

#### Banking and accounting

- Bank feeds
- Bank account connections
- Reconciliation
- Profit and loss
- Balance sheet
- Cash flow
- Journal entries
- Chart of accounts
- Account configuration

#### Tax and reports

- Tax settings
- Tax reports
- Financial reports
- Export financial reports
- View sensitive financial fields

Finance roles must independently control:

- Transaction visibility
- Invoice creation
- Invoice sending
- Expense approval
- Bill approval
- Bank feed visibility
- Reconciliation
- Journal entries
- Chart of accounts
- Tax configuration
- Financial report export
- Sensitive account and banking information

### 7.5 CRM

The CRM routes are in `apps/crm/src/router.tsx`.

#### Customer data

- Customer dashboard
- Customer list
- Customer profiles
- Create customers
- Edit customers
- Delete customers
- Import customers
- Merge customers
- Export customers
- View consent and communication preferences
- Manage consent

#### Loyalty

- Loyalty dashboard
- Loyalty members
- Rewards
- Tiers
- Configure loyalty rules
- Issue or adjust rewards

#### Marketing

- Campaign list
- Email campaigns
- SMS campaigns
- Campaign details
- Create campaigns
- Edit campaigns
- Pause campaigns
- Send campaigns
- Templates
- Segments
- Segment builder
- Automations
- Automation builder

#### Sales add-on

- Sales pipeline
- Leads
- Lead details
- Contracts
- Assign leads
- Move pipeline stages
- Export sales data

#### Customer engagement and settings

- Feedback
- Customer communication history
- CRM settings
- Add-on access

CRM roles should separate Sales, Marketing, Customer Success, Loyalty, Data
Steward, and CRM administration.

### 7.6 HR

The HR routes are in `apps/hr/src/router.tsx`.

#### Employee lifecycle

- Employees
- Staff directory
- Employee details
- Reporting lines
- Departments
- Locations
- Organization chart
- Onboarding
- Onboarding templates
- Offboarding
- Documents
- Equipment
- Asset register
- Benefits
- Training

#### Workforce management

- Shifts
- Schedule
- Shift roster
- Attendance
- Time logging
- Timesheets
- Timesheet approvals
- Calendar
- Time off
- Leave requests
- Leave approvals
- Overtime
- Operations

#### Performance

- Reviews
- Review cycles
- One-on-ones
- Performance
- Goals
- My review

#### Payroll and compensation

- Payroll
- Payroll runs
- Payroll details
- Payslips
- Payroll configuration
- Compensation information
- Bonuses
- Deductions

#### Reports and HR settings

- HR reports
- Employee reports
- Department reports
- Leave policy
- HR configuration
- Roles and permissions

The HR permission catalog must include:

- View own profile
- View own schedule
- Request time off
- View team
- Approve time off
- Manage schedules
- Manage employees
- Manage roles
- Manage departments
- Manage locations
- Manage HR configuration
- View reports
- Manage payroll
- Manage benefits
- Manage approval chains

Sensitive HR permissions must separately protect salary, compensation,
payroll, documents, benefits, disciplinary records, and personal data.

### 7.7 Flow

The Flow routes are in `apps/flow/src/router.tsx`.

#### Projects and tasks

- Project list
- Project dashboard
- Project board
- Project list view
- Project overview
- Project timeline
- Project calendar
- Project settings
- Task list
- Task details
- Create tasks
- Edit tasks
- Delete tasks
- Assign tasks
- My issues
- Inbox
- Triage
- Archive

#### Planning

- Today
- My day
- Timeline
- Calendar
- Milestones
- Cycles
- Views
- Roadmap
- Flow insights

#### People and time

- Team
- Directory
- Presence
- Leave
- Team time
- Time reports
- Profile

#### Flow administration

- Manage Flow team
- Manage Team Leaders
- Manage Managers
- Manage Super Managers
- Manage project assignments
- Manage views
- Manage cycles
- Manage roadmaps
- Manage Flow settings
- Manage Flow roles

Flow-specific hierarchy:

- Team Leader: one team or delivery group
- Manager: one or more projects or teams
- Super Manager: multiple Managers, cross-project delivery, and cross-team
  reporting
- Flow Admin: Flow configuration, roles, integrations, and complete app
  administration

Project membership must be a scope restriction, not a second role selector in
Dash. A user can be a Flow Super Manager over selected projects while having
no rights over unrelated projects.

### 7.8 Insights

The Insights routes are in `apps/insights/src/router.tsx`.

- Main dashboard
- Revenue analytics
- Customer analytics
- Product analytics
- Operations analytics
- Financial analytics
- Reports
- Saved reports
- Report sharing
- Report export
- Insights settings

Separate permissions are needed for:

- Basic dashboard view
- Revenue view
- Financial view
- Customer data view
- Operations view
- Saved report creation
- Report sharing
- Report export
- Analytics configuration

Financial analytics and exports should be independently controllable.

### 7.9 Menu management

The public Menu app is customer-facing. Its customer features are not internal
employee permissions:

- Browse menu
- View products
- Build cart
- Checkout
- Order status
- Customer profile

The internal management surface is currently in Dash:

- Menus
- Menu items
- Categories
- Availability
- Banners
- QR codes
- Design
- Branding
- Menu orders
- Menu analytics

The old global `menu_manager` role becomes Menu Manager.

### 7.10 Recruit

Recruit has public and authenticated surfaces.

Public surfaces:

- Careers landing page
- Department pages
- Open positions
- Career details
- Application form
- Application documents
- Custom application questions

Authenticated administrative surfaces:

- Careers management
- Create job posting
- Edit job posting
- Publish job posting
- Unpublish job posting
- Archive job posting
- View candidate submissions
- Review candidate details
- Change candidate status
- Export candidate data
- Manage page sections
- Manage page branding
- Manage application form fields

Candidate personal data, resumes, compensation, and interview feedback need
separate sensitive permissions.

### 7.11 Auth and public infrastructure

Auth routes are public infrastructure:

- Login
- Registration
- Password reset
- Email verification
- Invitation acceptance
- OAuth consent
- Authentication callbacks

These should not be treated as a normal licensed app. Invitation acceptance
must still validate the invitation, business, assigned roles, and expiration.

## 8. Data model

The current `user_app_permissions` table stores one role and one JSON app list
per user. That structure cannot represent independent roles, multiple roles,
role history, or scopes cleanly.

### 8.1 Recommended entities

Use normalized tables or equivalent domain entities:

#### Business membership

```text
business_members
  id
  business_id
  user_id
  dash_role_id
  status
  created_at
  updated_at
```

There must be one active membership per user and business.

#### Role definitions

```text
roles
  id
  business_id nullable
  app_id nullable
  key
  name
  description
  level
  parent_role_id nullable
  assignment_mode
  is_system
  is_archived
  created_by
  created_at
  updated_at
```

`app_id = null` identifies a Dash hub role. System roles are seeded and
business custom roles use the business id.

#### Role permissions

```text
role_permissions
  role_id
  permission_key
  effect
```

Use an explicit effect model only if deny rules are required. If there is no
deny requirement, store grants and keep the evaluator union-based.

#### Application license assignments

```text
app_license_assignments
  id
  business_id
  user_id
  app_id
  status
  source
  granted_by
  granted_at
  revoked_by nullable
  revoked_at nullable
```

This separates subscription and user access semantics from role definitions.

#### Role assignments

```text
role_assignments
  id
  business_id
  user_id nullable
  group_id nullable
  app_id
  role_id
  source
  is_primary
  status
  granted_by
  granted_at
  revoked_by nullable
  revoked_at nullable
```

An assignment must target either a user or a group. A unique constraint should
prevent duplicate active assignments of the same role to the same target.

#### Scopes

```text
role_assignment_scopes
  assignment_id
  scope_type
  scope_id
```

Examples:

- `location:12`
- `department:4`
- `team:7`
- `project:91`
- `business:*`

Do not put all scope logic into an unqueryable JSON blob. Scope checks need to
be auditable and enforced in backend services.

#### Defaults

```text
dash_role_app_defaults
  dash_role_id
  app_id
  default_role_id nullable
  default_enabled
```

Defaults should be distinguishable from explicit assignments and must never
silently overwrite explicit assignments.

#### Audit history

```text
authorization_audit_events
  id
  business_id
  actor_user_id
  target_user_id nullable
  target_group_id nullable
  action
  app_id nullable
  role_id nullable
  before_json
  after_json
  reason nullable
  created_at
```

Store audit events for role creation, role changes, license changes, scope
changes, access approvals, access revocations, and sensitive overrides.

### 8.2 Existing data compatibility

Keep the current `role` and `apps` fields during migration. Mark them as
legacy and dual-write them only while old clients still exist.

Do not drop or rewrite legacy columns until:

- All apps read the new model
- The backend resolver reads the new model
- Shadow comparison shows no unexplained differences
- Existing users have been reviewed
- Rollback has been tested

## 9. Contract and API changes

The current invitation contracts accept one `role` and an `apps` string list.
The new contracts must carry structured app assignments.

### 9.1 Shared user profile

Extend the current user response with:

- `dash_role`
- `dash_role_id` if role ids are exposed
- `app_assignments`
- `app_id`
- `role_ids`
- `primary_role_id`
- `assignment_source`
- `scope`
- `effective_app_access`
- `permission_version`

Do not require every frontend to receive the complete permission matrix if the
server can provide a safe resolved capability list. The backend remains the
authority.

### 9.2 Invite user

Replace the single role and flat apps with a structure equivalent to:

```text
InviteUserRequest
  email
  full_name
  dash_role
  app_assignments[]
    app_id
    role_ids[]
    primary_role_id
    scopes[]
  group_ids[]
  password optional
```

The server must validate:

- The inviter can assign the Dash role
- The inviter can assign each app role
- The business owns or subscribes to each app
- Every assignment has a valid role for that app
- Scopes are valid for that app
- Owner cannot be granted through a normal invitation

### 9.3 Update permissions

Support patch operations that can independently change:

- Dash role
- One app license
- One app role
- One primary role
- One scope
- Group membership

The update response should return the complete effective assignment state so
the frontend can refresh without reconstructing permissions locally.

### 9.4 Access requests

Access requests must include:

- App id
- Requested role ids
- Requested scope
- Optional reason
- Status
- Approver
- Approval timestamp

Approval must not silently preserve an unrelated global role as the app role.

### 9.5 Custom role APIs

Provide endpoints or RPCs for:

- List roles by hub or app
- Get role detail
- Create role
- Update role
- Clone role
- Archive role
- List role assignments
- Assign role to user
- Assign role to group
- Remove role assignment
- Preview permission impact
- List audit history

All list endpoints must be paginated according to repository rules.

### 9.6 Contract compatibility

During migration, accept the old request shape and translate it into the new
model. Emit both old and new response fields until every consumer is migrated.

Update the source contracts in the Voxie contract repository and regenerate the
generated clients. Do not hand-edit generated output as the final change.

## 10. Backend authorization implementation

### 10.1 Central authorization service

Create one backend authorization service that resolves:

- Business membership
- Dash role
- Subscription entitlement
- Direct app licenses
- Direct app roles
- Group licenses and roles
- Role inheritance
- Custom role permissions
- Scopes
- Sensitive-data restrictions
- Action-specific constraints

The service should expose operations equivalent to:

```text
CanHub(user, permission)
CanApp(user, app, permission, resource)
HasRole(user, app, role)
HasScope(user, app, resource, scope)
ResolveEffectiveAccess(user)
```

### 10.2 JWT and session behavior

Do not rely on a single stale role in the JWT for application authorization.
The token should at minimum identify the user, business, and session. Role
resolution should come from the authorization service or a versioned cache.

If claims are retained for performance, include an authorization version and
invalidate or refresh it after role, license, or scope changes.

### 10.3 Enforcement location

Authorization must be enforced in backend services and handlers, following the
existing onion layers:

```text
handler -> application service -> repository
```

The service layer should authorize the action before mutating data. The
repository should remain responsible for data access and SQL.

### 10.4 App-specific enforcement

#### POS

- Resolve primary POS role
- Combine approved capability add-ons
- Check location scope
- Check terminal scope
- Check amount and age thresholds
- Require reasons for configured overrides
- Write audit events

#### Flow

- Resolve Flow app role
- Check team, manager, project, or portfolio scope
- Treat project membership as scope
- Remove the old global Operator bypass
- Allow Super Manager to manage Managers only within scope

#### HR

- Resolve one HR employee role
- Check employee, department, and location scope
- Protect compensation and payroll fields separately
- Revoke access when an employee is terminated
- Ensure HR role configuration does not create a second global role

#### Finance

- Separate viewing from creating, sending, approving, reconciling, and
  configuring
- Protect bank feeds, accounts, journals, tax, and sensitive financial data

#### CRM

- Separate customer data, marketing, loyalty, sales, and automation functions
- Protect imports, merges, deletion, consent, and exports

#### ERP

- Separate catalog, inventory, warehouses, purchasing, receiving, and stock
  approval
- Protect destructive product and inventory actions

#### Insights

- Separate dashboard view, financial analytics, saved reports, sharing, and
  export

#### Recruit

- Separate job posting management, candidate submissions, candidate sensitive
  data, page content, and branding

### 10.5 Access revocation

Implement immediate revocation for:

- License removal
- User deactivation
- Employee termination
- Business membership removal
- Group removal
- Role archive or replacement
- Scope removal

The HR termination flow must call the shared hub access revocation service.
This closes the current security gap where termination does not remove app
access.

## 11. Shared frontend implementation

### 11.1 Shared types

Update `@voxire/types` with:

- `DashRole`
- `AppId`
- `RoleDefinition`
- `AppRoleAssignment`
- `LicenseAssignment`
- `RoleScope`
- `PermissionKey`
- `AssignmentSource`
- `EffectiveAccess`

Remove reliance on an unstructured single `User.role` for app actions.

### 11.2 Auth and guards

Update `@voxire/auth` so the store contains normalized hub and app access.

AppGuard must:

- Check app subscription and effective access
- Allow Owner and Admin app override
- Check explicit assignments for other users
- Fail closed when access is absent

RoleGuard should accept an app context and permission key rather than only a
global role list.

### 11.3 Permission evaluator

Update `@voxire/permissions` to evaluate:

- Hub permissions
- App permissions
- Multiple roles
- Inherited roles
- Scope
- POS capability constraints

Remove the current behavior where assigning an app causes any `app:<id>` check
to pass without checking child permissions.

### 11.4 Navigation and route guards

Every application should have:

- Navigation filtering by `feature.view`
- Page-level route guards
- Control-level action guards
- Backend error handling for forbidden requests

The route should not merely redirect all unknown permissions to the dashboard.
It should preserve a consistent forbidden state where appropriate.

## 12. Dash user experience implementation

### 12.1 Invite user flow

Keep the existing personal details and authentication steps. Replace the
single role and flat license sections with:

1. Personal details
2. Authentication
3. Dash role
4. Application licenses
5. App role and scope configuration per selected app
6. Group assignment
7. Effective access preview
8. Submit

Each app row should show:

- App name
- License toggle
- Role picker
- Multiple-role option where supported
- Primary role
- Scope picker
- Sensitive permission warning
- Access summary

Example:

```text
Flow       [licensed]  Team Leader, Contributor   Teams: Product, Support
Finance    [licensed]  Viewer                     Business-wide
HR         [licensed]  Employee                   Location: Beirut
POS        [licensed]  Senior Cashier              Locations: 1, 3
```

### 12.2 Admin and Owner display

Owner and Admin should see all subscribed apps as inherited access. The UI
should distinguish:

- Inherited Admin access
- Explicit user license
- Group license
- App role

Do not make inherited access look like a missing assignment.

### 12.3 Member detail and editing

Replace the current single role editor with:

- Dash role editor
- App assignment list
- Role chips per app
- Scope editor
- Add role
- Remove role
- Change primary role
- Revoke license
- Restore assignment
- Effective permission preview
- Assignment history

Changing POS role must not silently rewrite locations. Changing a location
must not silently change the POS role.

### 12.4 Custom role builder

Dash should provide a role builder with:

- Hub or app selection
- Role name
- Job-title display name
- Description
- Parent role
- Hierarchy level
- Permission groups
- Individual action checkboxes
- Sensitive-data permissions
- Scope requirements
- Assignment management permissions
- Impact preview
- Save and archive

The builder must prevent a user from granting permissions they are not allowed
to delegate.

The builder is authoritative for Dash roles and all cross-app application
roles. HR permission-role definitions remain in HR, and Flow project
membership remains in Flow. Those exceptions must be linked and visible from
Dash without creating a second competing role definition.

### 12.5 Groups

Groups should assign:

- Dash role only when explicitly allowed
- App licenses
- App roles
- Scopes

The effective permission preview must show whether access comes from a direct
assignment or a group.

### 12.6 Bulk invitations

Bulk invite should support one Dash role plus structured per-app assignments.
If different users need different roles, provide a review table before
submission rather than silently applying one broad role.

## 13. App-by-app frontend migration

### 13.1 Dash

- Replace role dropdown with the five Dash roles and custom Dash roles
- Replace flat `selectedApps` with structured app assignments
- Add role picker per app
- Add scopes
- Add custom role builder
- Add effective permission preview
- Update Team, MemberDetail, Groups, Permissions, Roles, and access requests
- Remove role-derived POS location flags
- Use explicit POS role and location assignments

### 13.2 POS

- Coordinate with the separate POS frontend repository
- Read the new app assignments
- Add POS role and capability selectors
- Add Senior Cashier template
- Add void and refund controls independent of Manager
- Add reason and threshold controls
- Hide and guard terminal, employee, report, and cash controls by permission

### 13.3 ERP

- Replace platform-role checks in `erpPermissions.ts`
- Add feature and action permission hooks
- Filter all navigation groups
- Guard product, inventory, warehouse, purchasing, and settings screens
- Add destructive-action confirmations where permissions allow deletion or
  stock adjustment

### 13.4 Finance

- Replace platform-role checks in `financePermissions.ts`
- Add separate invoice, expense, bill, banking, journal, tax, and report
  permissions
- Hide and guard send, approve, reconcile, configure, and export actions
- Add sensitive financial data checks

### 13.5 CRM

- Replace platform-role checks in `crmPermissions.ts`
- Separate customer, loyalty, marketing, sales, data stewardship, and
  automation permissions
- Protect import, merge, delete, send, configure, and export actions

### 13.6 HR

- Replace global-role checks with the unified HR app role
- Consolidate `hrAccessUtils` and the HR setup role vocabulary
- Preserve HR permission-role definitions and the existing HR role editor
- Keep HR permission-role management in HR as the deliberate app-owned
  exception
- Let Dash manage the HR license and display the effective HR role
- Enforce one effective HR employee role
- Protect payroll, compensation, documents, benefits, and personal data
- Connect termination to hub access revocation

### 13.7 Flow

- Replace business-role bypass behavior with Flow app roles
- Add Team Leader, Manager, and Super Manager templates
- Keep project membership managed in Flow as internal scope
- Keep Flow app-role management in Dash
- Add cross-project and cross-team permission checks
- Hide roadmap, cycles, team, reporting, and settings based on permissions

### 13.8 Insights

- Replace global role export checks
- Add dashboard, analytics-area, report, share, and export permissions
- Protect financial analytics separately

### 13.9 Menu

- Keep public Menu customer flows outside employee RBAC
- Protect internal menu management in Dash with Menu roles
- Add Menu Manager, Menu Editor, Branding Editor, and Viewer permissions

### 13.10 Recruit

- Protect authenticated admin routes using Recruit roles
- Add recruiter, coordinator, sourcer, interviewer, and manager permissions
- Separate candidate data from page branding and job posting permissions

## 14. Migration strategy

### Phase 0: preparation

- Create this ADR or link it from the relevant architecture documentation
- Inventory all current role checks and app access checks
- Inventory all backend handler checks
- Freeze new role names while the migration is implemented
- Add a feature flag for the new authorization resolver
- Seed system Dash and app role templates

### Phase 1: schema and contracts

- Add new role, membership, license, assignment, scope, and audit tables
- Add indexes for business, user, app, status, and active assignments
- Add contract types and generated clients
- Keep legacy columns and old request fields
- Add data migration tests

### Phase 2: dual-write and backfill

- Backfill Dash roles from current global roles
- Backfill app licenses from `apps[]`
- Backfill obvious app roles from current role permissions
- Create review records for ambiguous Operator users
- Convert Billing, Menu Manager, Store Manager, Staff, and Viewer into explicit
  app assignments
- Preserve POS location assignments
- Preserve Flow project membership
- Preserve HR custom role assignments
- Dual-write changes to old and new models

### Phase 3: backend resolver

- Implement the new resolver behind the feature flag
- Compare old and new decisions in shadow mode
- Log differences without changing the response
- Resolve all expected differences caused by removing accidental inheritance
- Add direct API authorization tests

### Phase 4: Dash migration

- Update the profile response
- Update the auth store
- Update invite and edit flows
- Update role and custom role pages
- Update groups and access requests
- Add the effective access preview
- Test all five Dash roles

### Phase 5: application migration

Migrate in this order:

1. Shared permission package
2. POS backend and separate POS frontend
3. Flow
4. HR
5. Finance
6. ERP
7. CRM
8. Insights
9. Menu management
10. Recruit

Each app migration must update navigation, routes, controls, API calls, and
tests together.

### Phase 6: enforcement switch

- Enable the new backend resolver for an internal tenant
- Enable it for a small tenant cohort
- Monitor authorization differences and forbidden responses
- Enable the new Dash UI only after backend authorization is active
- Expand rollout gradually
- Keep legacy fields and the old resolver available for rollback

### Phase 7: cleanup

Only after rollout verification:

- Stop writing legacy role and app fields
- Remove old platform-role app checks
- Remove flat app access assumptions
- Remove accidental inheritance paths
- Remove obsolete group permission vocabulary
- Drop legacy columns only in a separately reviewed migration
- Update all architecture and cross-app documentation

## 15. Role-change behavior

### 15.1 Changing Dash role

When Dash role changes:

- Recalculate Dash permissions immediately
- Apply new defaults only to assignments marked as default or inherited
- Keep explicit app roles unchanged
- Recheck whether the user can manage roles or licenses
- Write an audit event
- Refresh cached access

### 15.2 Changing app role

When an app role changes:

- Recalculate only that app’s permissions
- Keep other app assignments unchanged
- Keep license status unchanged
- Recheck scope
- Refresh the user session or permission version
- Write an audit event

### 15.3 Editing a custom role

When a custom role changes:

- Show affected users and groups before saving
- Apply the new permission set to all active assignments
- Recalculate cached access
- Keep the previous role definition in audit history
- Do not silently remove a license

### 15.4 Removing a license

When a license is removed:

- Revoke app access immediately
- Revoke all active app actions
- Retain role and scope history as inactive
- Remove app navigation
- Invalidate app sessions if necessary
- Write an audit event

### 15.5 Multiple roles

For applications that support multiple roles:

- Combine permissions using union semantics
- Combine scopes carefully
- Show the permission source for each effective permission
- Do not treat a lower role as a deny against a higher role unless explicit
  deny rules are introduced

For HR, use one effective role and create a custom combined role when an
employee needs multiple functional permission sets.

## 16. Security requirements

- Server-side authorization is mandatory
- Frontend hiding is not a security boundary
- Fail closed when role, license, or scope data is missing
- Do not trust an app id from the browser without checking entitlement
- Do not trust a role id without checking that it belongs to the business and
  application
- Do not allow a manager to grant permissions they cannot delegate
- Protect sensitive fields independently from page visibility
- Audit every role, license, scope, and high-risk action change
- Invalidate or refresh permission caches after changes
- Revoke access when a user is deactivated or terminated
- Preserve tenant boundaries for custom roles and scopes
- Require confirmation for destructive permission grants
- Do not expose hidden data through exports, search, reports, or assistant tools

## 17. Testing plan

### 17.1 Permission resolver unit tests

Test:

- Owner access
- Admin access
- Dash Manager hub access
- Member with no app license
- Viewer with no app license
- Member with Flow Manager and Finance Viewer
- Explicit role overriding a Dash default
- Group role combined with direct role
- Multiple compatible roles unioning permissions
- HR one-role restriction
- Location scope
- Department scope
- Project scope
- Role inheritance
- Archived role
- Removed license
- Deactivated user
- Terminated employee
- Missing role or missing app assignment
- Cross-business role id rejection

### 17.2 POS tests

- Cashier cannot void by default
- Senior Cashier can void
- Senior Cashier can refund when assigned
- Void and refund remain independent
- Refund threshold requires approval
- Location restriction works
- Terminal restriction works
- Reasons are required when configured
- Audit event contains the required information

### 17.3 Flow tests

- Contributor can edit assigned tasks
- Team Leader can manage assigned team
- Manager can manage assigned projects
- Super Manager can manage assigned Managers
- Super Manager cannot manage unrelated projects
- Super Manager cannot manage Dash roles
- Flow Admin can configure Flow
- Project scope is enforced without exposing a second role selector

### 17.4 App frontend tests

For every app:

- Unauthorized navigation item is hidden
- Direct route access is forbidden
- Unauthorized action button is hidden or disabled
- Direct API denial is handled
- Viewer cannot mutate data
- Manager can perform only manager permissions
- Custom role changes refresh access
- License removal removes the app from the hub

### 17.5 Migration tests

- Every existing global role can be backfilled
- Ambiguous Operator assignments are reported
- Existing POS locations survive
- Existing HR roles survive
- Existing Flow project membership survives
- Existing group memberships survive
- Dual-read old and new results are compared
- Rollback restores old resolver behavior

## 18. Verification and quality gates

Before declaring the implementation complete:

- Run the repository’s real build command
- Run frontend unit and integration tests
- Run backend tests
- Run contract generation and contract audits
- Run migration tests against a copy of representative tenant data
- Verify every list endpoint added for role assignments is paginated
- Run static searches for old single-role app checks
- Run static searches for flat `apps[]` authorization decisions
- Run static searches for frontend-only permission gates
- Verify no generated contract output was hand-edited
- Verify no secrets or environment files changed
- Verify English and Arabic user-facing strings for every new UI label
- Verify role and permission labels use i18n
- Verify the mobile Dash invitation and role editor at 390px
- Verify audit events for every high-risk change

Use the package manager declared by the repository. This repository is a pnpm
workspace, so do not use npm to modify dependencies or lockfiles.

## 19. Implementation commit order

Keep changes atomic and in dependency order:

1. Add the architecture decision and role vocabulary documentation
2. Add shared role, assignment, scope, and permission types
3. Add contract definitions and generated client updates
4. Add database migrations and repository models
5. Add backend role definitions and seed data
6. Add backend authorization resolver
7. Add backend custom role and assignment APIs
8. Add audit and revocation services
9. Update shared frontend auth and permission packages
10. Update Dash user creation and editing
11. Update Dash role builder and group assignments
12. Update POS backend and POS frontend
13. Update Flow
14. Update HR
15. Update Finance
16. Update ERP
17. Update CRM
18. Update Insights
19. Update Menu management
20. Update Recruit
21. Add migration backfill and shadow comparison
22. Add end-to-end authorization tests
23. Remove legacy paths after rollout approval

Use short imperative commit subjects without conventional-commit prefixes. Do
not combine unrelated fixes into these commits.

## 20. Acceptance criteria

The work is complete only when all of these are true:

- Dash is the common hub for every business user
- The five global Dash roles exist and are enforced
- Dash role and app role are separate but related
- A user can be Flow Super Manager and Finance Viewer at the same time
- Dash role changes do not overwrite explicit app roles
- Every licensed app has an app-role selector
- Every app has a documented hierarchy
- POS Cashier can optionally receive void and refund permissions
- POS thresholds, reasons, locations, and audits work
- Flow has Team Leader, Manager, and Super Manager behavior
- Flow project access is enforced as Flow-managed scope rather than a second
  Dash role
- HR uses one unified permission-role model, with role definitions remaining in
  HR
- Dash is the only control plane for cross-app roles, licenses, and scopes
- Custom roles can be created, edited, assigned, archived, and audited
- Custom roles can be assigned to users and groups
- Multiple roles combine where the app supports them
- HR prevents multiple conflicting employee roles
- Role and license upgrades take effect immediately
- License removal revokes app access immediately
- Employee termination revokes shared app access
- All listed pages and actions have view and action permissions
- Hidden UI elements cannot be reached through direct API calls
- Backend enforcement is authoritative
- Existing users are migrated without losing access or gaining unexplained
  access
- Rollback has been tested
- All role and permission changes are audited
- Documentation and generated contracts match the implementation

## 21. Final decisions to preserve in the next session

Use these choices unless the product owner explicitly changes them:

- The hub role is called a Dash role
- Dash has Owner, Admin, Manager, Member, and Viewer
- Super Manager is a supported app role title, especially in Flow
- Dash Manager does not automatically become Manager in every app
- App roles are selected independently per licensed app
- Explicit app assignments survive Dash role changes
- POS Cashier can receive optional void and refund permissions
- Flow Super Manager manages Managers within its assigned scope
- Flow project membership is managed in Flow as internal scope, not a second
  Dash role selector
- HR permission-role definitions remain managed in HR
- Custom roles are included in the initial implementation
- App access changes are immediate and auditable
- Role hierarchy, permissions, and scopes are separate but combined during
  authorization

## 22. Review additions and boundary conditions

This section records items that must be handled explicitly during
implementation. They are easy to miss because they cross multiple apps or sit
outside the visible role selector.

### 22.1 Product surface coverage

The root README describes thirteen product surfaces, but the current `rtylr`
checkout contains ten Vite apps. The implementation session must confirm each
consumer before switching the resolver globally.

Tenant-facing surfaces in scope:

- Dash hub
- Auth infrastructure
- POS terminal
- ERP
- HR
- CRM
- Finance
- Flow
- Recruit
- Insights
- Customer-facing Menu
- Internal Menu management in Dash

Additional surfaces requiring an explicit boundary decision:

- POS frontend in the separate POS repository
- Mobile client in the `pablo` repository
- Docs application, which is public or internal documentation rather than a
  tenant business application
- Landing application, which is marketing rather than tenant authorization
- Booking and Support backend domains, which currently have no corresponding
  app surface in this checkout
- Internal Console, which uses console administration roles and must not be
  treated as a tenant Dash role

The same backend authorization concepts must be consumable by the POS and
mobile clients. The implementation is not complete if only the ten local Vite
apps use the new model.

### 22.2 Corporate reporting hierarchy versus permission hierarchy

The company reporting chart and the role hierarchy are different concepts.

Reporting hierarchy answers:

- Who reports to whom?
- Which team or department does a person belong to?
- Which Managers are supervised by a Super Manager?
- Which locations, projects, or departments fall under a leader?

Permission hierarchy answers:

- What actions can a role perform?
- Which role can delegate another role?
- Which role can approve an action?

Do not grant access only because a user is listed as somebody’s manager in the
org chart. A manager must have both the required app permission and the scope
that contains the relevant team, department, project, or location.

HR reporting lines and org-chart data can provide a default scope suggestion,
but the role assignment remains the authorization source. Changing a reporting
line must not silently grant access unless the business explicitly enables
that policy.

Flow Super Manager should therefore have:

- A role level above Manager
- Explicit permission to manage Managers
- An explicit team, department, program, or project scope
- A reporting relationship only when the business wants reporting lines to
  influence the scope

### 22.3 Feature matrix artifact before coding

Before implementing UI gates, create a machine-readable permission inventory
and a human-readable matrix.

Each row must contain:

- App
- Feature area
- Page or route
- Navigation key
- Permission key
- View action
- Create action
- Edit action
- Delete or archive action
- Approve action
- Publish or send action
- Export action
- Sensitive-data action
- Scope type
- Standard roles that receive it
- Whether custom roles may receive it
- Backend handler or service enforcing it
- Frontend component hiding it

Generate the matrix from the canonical permission definitions where possible.
Keep a reviewed CSV or Markdown output for product and QA review. This is the
artifact that proves every feature can be hidden and unhidden per role.

No route or navigation item should be declared complete until it appears in
this inventory.

### 22.4 Row-level and field-level authorization

Page-level visibility is not enough. The backend must also filter returned
records and fields.

Examples:

- A Team Leader sees only members in assigned teams.
- A Location Manager sees only POS locations in scope.
- A Finance Viewer may see reports but not bank account identifiers.
- An HR Manager may see assigned employees but not payroll fields unless the
  role has compensation access.
- A Recruiter may view candidate pipeline data but not compensation or private
  interview feedback.
- A Flow Manager may see assigned projects but not unrelated projects.

Every list, detail, search, export, report, assistant query, background job,
and bulk operation must apply the same row and field rules.

Repositories should receive authorized scope filters from the application
service rather than fetching all rows and filtering only in the browser.

### 22.5 Approval and separation-of-duties policies

High-risk actions need separate permission and policy checks.

Do not treat `manage` as automatically including `approve` or `execute`.

Examples:

- POS void versus POS void approval
- POS refund versus high-value refund approval
- Finance expense creation versus expense approval
- Finance invoice creation versus invoice sending
- Finance reconciliation preparation versus completion
- HR payroll preparation versus payroll execution
- Recruit candidate review versus candidate final decision

The business must be able to configure:

- Approval threshold
- Required number of approvers
- Whether self-approval is allowed
- Whether the creator can approve their own action
- Required reason
- Required attachment or evidence
- Escalation path
- Expiration of pending approval

These policies need audit events and must be enforced in backend services.

### 22.6 License, subscription, and seat behavior

Authorization and billing must use separate models.

Define and test:

- Business subscription entitlement
- Application enabled or disabled state
- User license assignment
- Group license assignment
- Owner and Admin access override
- Paid seat count
- Seat limit reached behavior
- Trial expiration behavior
- Add-on dependency behavior
- License transfer between users
- License removal from a deactivated user
- What happens when an application subscription is canceled

Owner and Admin may access subscribed applications without normal user-level
license assignments, but the business must still be entitled to the app. If
the business has seat billing, the seat counter must be explicit and audited.

### 22.7 Invitation and pending-user behavior

An invitation is not an active authorization grant until accepted.

The invitation flow must define:

- Whether assignments are stored on the invitation or only on acceptance
- Whether a changed role is re-evaluated when the invite is accepted
- Whether a revoked invite removes all pending assignments
- Whether an expired invite can be resent with new assignments
- Whether a user can accept an invite after their role was downgraded
- Whether the invitation displays all app roles and scopes
- Whether the invitation can be edited without sending a new email

Recommended behavior:

- Store the proposed assignment snapshot on the invitation
- Revalidate the snapshot against current business subscriptions and role
  definitions at acceptance time
- Apply the current active assignment only after successful acceptance
- Record both the invitation snapshot and final assignment in the audit log
- Grant no app access while the invitation is pending

### 22.8 Group conflict and assignment precedence

The plan requires direct and group assignments to combine, but the exact
conflict policy must be implemented consistently.

Define:

- Whether a group can grant an app license without a direct user assignment
- Whether removing a direct assignment leaves group access active
- Whether scopes combine as a union
- Whether a group can provide the primary display role
- Whether a direct assignment can suppress a group role
- Whether a group role can be archived while members still depend on it
- Whether a user can see the source of every effective permission

Recommended rule:

- Direct and group permissions combine by union
- Direct assignments may mark the primary role
- Scope is the union of explicitly granted scopes
- A license is active if any valid source grants it
- Explicit revocation is a separate deny or suppression record only if the
  business requires it

Do not accidentally implement a mixture of replacement and union semantics.

### 22.9 Role delegation and privilege ceilings

Create a role-administration matrix independent from app data permissions.

At minimum:

| Actor             | Can manage Dash roles |  Can manage app roles |  Can manage app users |        Can manage licenses |
| ----------------- | --------------------: | --------------------: | --------------------: | -------------------------: |
| Owner             |                   Yes |                   Yes |                   Yes |                        Yes |
| Dash Admin        |                   Yes |                   Yes |                   Yes |                        Yes |
| Dash Manager      |          Configurable |          Configurable |          Configurable |               Configurable |
| App Admin         |                    No |               Own app |               Own app | Only if explicitly granted |
| App Super Manager |                    No | Own app, within scope | Own app, within scope |              No by default |
| App Manager       |                    No | Own app, within scope | Own app, within scope |              No by default |
| Other roles       |                    No |                    No |                    No |                         No |

An actor cannot create, assign, or delegate a permission above their own
delegation ceiling. App authority must not leak into Dash authority.

### 22.10 Role lifecycle and versioning

Add explicit role lifecycle behavior:

- Draft role
- Active role
- Archived role
- Replacement role
- System role protection
- Custom role version
- Assignment migration
- Effective date
- Optional future-dated promotion or downgrade

Editing a role should create a versioned audit record. A role assignment should
be able to identify the role version that produced its effective permissions.

When a role is archived, the UI must require one of:

- Migrate assignments to a replacement role
- Keep assignments on the archived role until manually changed
- Revoke the assignments with explicit confirmation

System Owner, Admin, and other security-critical templates must not be
modifiable in ways that remove their required invariants.

### 22.11 Permission cache and distributed invalidation

Role changes must reach every consumer, not only the current Dash browser.

Implement one of:

- Authorization version in the user profile and token
- Permission-change event on the backend event bus
- Cache invalidation by business and user
- Short-lived authorization cache with immediate revocation path

Test invalidation for:

- Open web sessions
- POS sessions
- Mobile sessions
- WebSocket subscriptions
- Background jobs
- Scheduled reports
- Assistant or Caret requests
- API keys and webhooks

An access removal must not wait for a long-lived cache to expire.

### 22.12 POS offline and terminal behavior

Because POS is a Tauri terminal surface, the separate POS repository must be
audited for offline behavior before implementation.

Define:

- How role assignments are cached locally
- How long offline permissions remain valid
- Whether refunds and voids are allowed offline
- How a revoked cashier is blocked after reconnection
- How local actions are queued and audited
- How terminal and location scope are synchronized
- Whether high-risk actions require online authorization

The safest default is to require a current server authorization for refunds,
cash pickup, employee management, terminal management, and other high-risk
actions when the terminal is offline.

### 22.13 HR lifecycle integration

Link the hub business member, HR employee, and application assignments without
creating a second independent identity.

Define events for:

- Employee created
- Employee hired
- Employee transferred
- Employee promoted
- Employee moved to a department
- Employee moved to a location
- Employee terminated
- Employee rehired
- Business membership deactivated

Termination should revoke login and app access immediately while retaining the
audit trail. Rehire should require an explicit restoration decision rather
than automatically restoring all old access.

### 22.14 Audit retention and operational reporting

The audit system must specify:

- Retention period
- Immutable fields
- Actor and target identity
- Business context
- Before and after state
- Reason and approval reference
- Request id and correlation id
- Source client, such as Dash, POS, mobile, or API
- Export permission
- Audit search filters
- Whether audit records can be deleted

Provide Dash reports for:

- Current effective access by user
- Current effective access by app
- Users with high-risk permissions
- Users with access from groups
- Recently changed roles
- Recently revoked access
- Archived roles with active dependents
- Users with no app license
- Users with Admin or Super Manager access

### 22.15 Explicit final decisions before implementation

The next session should confirm these remaining policy choices before changing
the schema:

1. Use `Super Manager` as the standard Flow display title, with custom display
   titles allowed.
2. Allow multiple roles per app only where that app declares a safe assignment
   mode.
3. Keep one effective HR role per employee.
4. Treat direct app role assignments as persistent when Dash role changes.
5. Keep default assignments distinguishable from explicit assignments.
6. Revoke access immediately on license removal, termination, or deactivation.
7. Preserve inactive assignments and audit history after revocation.
8. Use union semantics for direct and group permissions unless an explicit
   suppression policy is added.
9. Keep project membership, location access, and department access as scope,
   not additional visible global roles.
10. Require a generated feature-by-role matrix before app migration begins.

### 22.16 Latest authority decisions

The final authority boundary is:

- Dash is the only place for cross-app role and license management.
- Dash manages Dash roles, application roles, POS roles, POS capability
  add-ons, app licenses, cross-app scopes, group assignments, and delegation
  rules.
- Apps display effective access and link back to Dash for Dash-owned changes.
- Apps continue to manage normal business data and operational workflows.
- Apps may manage resource-level membership and access inside their own
  workflows when the actor has the required app permission and resource scope.
  For example, Flow may add a user to another project or assign a task.
- Resource-level membership changes do not modify Dash roles, app roles,
  licenses, global permissions, hierarchy, or cross-app assignments.
- App Managers may manage app team membership and operational data only within
  their permitted scope.
- HR owns HR permission-role definitions and its existing permission editor.
- Flow owns project membership and project-level access as operational scope.
- POS terminals consume roles from Dash and do not create or edit roles locally.
- HR job titles remain ordinary employee data and remain editable in HR.
- Public Menu and Recruit roles remain outside internal employee RBAC.

Do not implement this with a blanket HTTP 403 rule for every mutation outside
Dash. Route each mutation according to domain ownership. Dash-controlled
changes should return a typed managed-in-Dash response when a client reaches a
non-authoritative surface. HR and Flow retain their stated app-owned mutation
surfaces, including Flow's resource-membership workflows.

### 22.17 Direct credential provisioning

Dash must offer both invitation and direct-credential provisioning.

Invitation mode:

- Send an invitation to the work email.
- Let the user create their password.
- Verify the email.
- Enroll MFA before sensitive app access.

Direct-credential mode:

- Create a username from Dash.
- Create a temporary password or a secure temporary-password challenge.
- Require password change at first login.
- Verify the user’s own work email.
- Support email OTP as a second-factor option.
- Preserve compatibility with authenticator-app or WebAuthn MFA where
  supported.

Never store or return a permanent password in Dash, an app, an audit event, or
an analytics payload. Audit provisioning mode, email verification, MFA
enrollment, password reset, and recovery changes without recording secrets.
