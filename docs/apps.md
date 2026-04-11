# Apps

Each rtylr app is a standalone React + Vite SPA served on its own subdomain. They share authentication via a cookie on `.rtylr.com` and share code through packages in `rtylr/packages/`.

---

## Hub — `hub.rtylr.com` (port 5175)

**Purpose:** Universal command center for business owners and admins. Hub is the primary management app — it controls menus, analytics, QR codes, visual design, settings, and user management for the whole account.

**Allowed roles:** `owner`, `admin`, `operator`

### Routes

| Path                 | Screen                         |
| -------------------- | ------------------------------ |
| `/`                  | Dashboard (analytics overview) |
| `/menus`             | Menu list                      |
| `/menu`              | Menu editor                    |
| `/menu/:id`          | Specific menu editor           |
| `/create-item`       | Create menu item               |
| `/create-banner`     | Create menu banner             |
| `/orders`            | Order history                  |
| `/qr-codes`          | QR code list                   |
| `/qr-codes/create`   | QR code generator              |
| `/design`            | Theme / branding designer      |
| `/analytics`         | Full analytics                 |
| `/settings/account`  | Account profile                |
| `/settings/general`  | Business general settings      |
| `/settings/industry` | Business type / industry       |
| `/settings/domain`   | Custom domain                  |
| `/settings/email`    | Email settings                 |
| `/settings/billing`  | Billing & subscription         |
| `/settings/addons`   | Add-ons                        |

### Notes

- Hub reads `config` and `theme` data from the API on startup (`GET /get-configs`)
- Supports i18n via `i18next`; language is persisted in a `language` cookie
- Has a fullscreen mode (persisted preference)

---

## Auth — `auth.rtylr.com` (port 5176)

**Purpose:** Cross-app authentication. All apps redirect here when no `token` cookie is found. After login the token is set as a cookie on `.rtylr.com` and the user is redirected back via the `?redirect=<appId>` query parameter.

**Allowed roles:** All (public-facing)

### Routes

| Path                  | Screen             |
| --------------------- | ------------------ |
| `/`                   | Login              |
| `/register`           | Registration       |
| `/password-reset`     | Password reset     |
| `/email-verification` | Email verification |

### Notes

- Token is set as a cookie with `domain: .rtylr.com` so all subdomains share it
- The `?redirect=<appId>` param tells Auth where to send the user after login (e.g., `?redirect=pos` → redirect to `pos.rtylr.com`)

---

## Menu — `menu.rtylr.com` (port 5177)

**Purpose:** Customer-facing digital menu. Publicly accessible (no auth required). Customers scan a QR code and land here to browse the menu, add items, and place orders.

**Allowed roles:** None (public)

### Routes

The menu app uses public routes that display the storefront menu. Route structure is dynamic based on business slug/id. Screen directories: `Public/`, `Dev/`.

### Notes

- Uses Redux for state (legacy pattern, unlike other apps which use Zustand)
- Imports `@voxire/menu` for shared menu display components

---

## POS — `pos.rtylr.com` (port 5178)

**Purpose:** Point-of-sale terminal for floor and counter sales. Staff use this to take orders, manage tables, process payments, and close shifts.

**Allowed roles:** `owner`, `admin`, `operator`, `staff`

### Routes

| Path      | Screen                              |
| --------- | ----------------------------------- |
| `/`       | Terminal (order entry)              |
| `/orders` | Active/recent orders                |
| `/tables` | Table map and status                |
| `/close`  | Close session / cash reconciliation |

### Notes

- Auth guard checks for `auth_token` cookie and redirects to `http://localhost:5176/auth/login?redirect=pos` in dev
- Integrates with POS backend routes for sessions, tables, orders, payments, and refunds
- Kitchen display integration via order item status updates

---

## ERP — `erp.rtylr.com` (port 5179)

**Purpose:** Inventory management, purchasing, and supplier relations. Tracks stock levels, cost of goods, and purchase orders.

**Allowed roles:** `owner`, `admin`, `operator`

### Routes

| Path          | Screen                          |
| ------------- | ------------------------------- |
| `/`           | Dashboard                       |
| `/inventory`  | Inventory list and stock levels |
| `/purchasing` | Purchase orders                 |
| `/suppliers`  | Supplier management             |

### Notes

- Low-stock alerts are driven by the `low_stock_threshold` field on `erp_products`
- Inventory adjustments are recorded in `erp_inventory_transactions` with a typed reason (`purchase`, `sale`, `adjustment`, `return`, `waste`)

---

## HR — `hr.rtylr.com` (port 5180)

**Purpose:** Human resources management — staff records, shift scheduling, payroll processing, and time-off requests.

**Allowed roles:** `owner`, `admin`, `operator`, `staff` (staff can view their own data)

### Routes

| Path           | Screen                  |
| -------------- | ----------------------- |
| `/`            | Dashboard (HR overview) |
| `/staff`       | Staff directory         |
| `/shifts`      | Shift schedule          |
| `/payroll`     | Payroll records         |
| `/departments` | Department management   |

### Notes

- Staff can clock in/out via `POST /hr/shifts/clock-in` and `POST /hr/shifts/clock-out`
- Time-off requests have a workflow: `pending → approved/rejected`
- Payroll records track regular hours, overtime, bonuses, deductions, and net pay

---

## CRM — `crm.rtylr.com` (port 5181)

**Purpose:** Customer relationship management — customer profiles, loyalty program, marketing campaigns, and audience segments.

**Allowed roles:** `owner`, `admin`, `operator`

### Routes

| Path         | Screen                           |
| ------------ | -------------------------------- |
| `/`          | Dashboard (CRM overview)         |
| `/customers` | Customer list                    |
| `/loyalty`   | Loyalty program config and tiers |
| `/campaigns` | Email/SMS/push campaigns         |
| `/segments`  | Audience segments with rules     |

### Notes

- Loyalty tiers: `bronze` → `silver` → `gold` → `platinum` (thresholds configurable via `crm_loyalty_config`)
- Segments use JSON rule arrays stored in `crm_segments.rules`
- Campaign channels: `email`, `sms`, `push`, `whatsapp`

---

## Finance — `finance.rtylr.com` (port 5182)

**Purpose:** Financial management — invoicing, expense tracking, profit & loss reporting, and tax configuration.

**Allowed roles:** `owner`, `admin`

### Routes

| Path        | Screen                         |
| ----------- | ------------------------------ |
| `/`         | Dashboard (financial overview) |
| `/invoices` | Invoice list and creation      |
| `/expenses` | Expense tracking               |
| `/pl`       | Profit & loss report           |
| `/tax`      | Tax config and reporting       |

### Notes

- Invoices support multiple statuses: `draft → sent → viewed → partial → paid → overdue → cancelled → void`
- Recurring invoices can be scheduled weekly, monthly, quarterly, or yearly
- Tax config supports `vat`, `sales_tax`, `gst`, and `other` types

---

## Flow — `flow.rtylr.com` (port 5183)

**Purpose:** Project and task management for teams — similar to Linear or Asana. Workspaces contain projects, projects contain tasks, tasks support subtasks and comments.

**Allowed roles:** `owner`, `admin`, `operator`, `staff`

### Routes

| Path                                | Screen                      |
| ----------------------------------- | --------------------------- |
| `/`                                 | Redirect to `/w/default`    |
| `/w/:workspaceId`                   | Workspace dashboard         |
| `/w/:workspaceId/p/:projectId`      | Project board (kanban)      |
| `/w/:workspaceId/p/:projectId/list` | Project list view           |
| `/inbox`                            | Task inbox / assigned to me |
| `/settings`                         | Flow settings               |

### Notes

- Task statuses: `todo → in_progress → in_review → done → cancelled`
- Task priorities: `none`, `low`, `medium`, `high`, `urgent`
- Labels are workspace-scoped with configurable colors
- Tasks support `parent_task_id` for subtask nesting

---

## Landing — `rtylr.com` (port 4200)

**Purpose:** Marketing site. Built with Next.js 15 (SSR). No authentication required.

**Allowed roles:** None (public)

### Notes

- Next.js 15 with App Router
- Served separately from the Vite apps
- Dev port 4200 is assigned in `APP_CONFIG`
