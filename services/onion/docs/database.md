# Database

The rtylr backend uses MySQL. All schema files live in `onion/internal/platform/database/schemas/` organized by module.

- **Production credentials (host, database, user, password):** stored as deployment environment variables only. Never commit them to this repository or any documentation. Operators retrieve them from the deployment platform's secret store.
- **Engine:** MySQL (InnoDB)
- **Primary keys:** `BIGINT NOT NULL AUTO_INCREMENT`
- **Multi-tenancy:** Every table that is tenant-scoped has an `owner_id BIGINT NOT NULL` column referencing `users.id`

---

## Running Migrations

Use `scripts/migrate.sh`. It applies all schema files in the correct order using `IF NOT EXISTS` so it is safe to run repeatedly.

```bash
cd onion/

export DB_HOST=<host>
export DB_PORT=3306
export DB_USER=<user>
export DB_PASSWORD=<password>
export DB_NAME=<database>
# Pull real values from the deployment platform's secret store — never hardcode.

./scripts/migrate.sh
```

Migration order is defined once in
`internal/platform/database/schemas/migration_order.sh` and consumed by both
schema migration entry points. The Onion runner applies module-owned
`internal/modules/*/sql/migration_*.sql` files after the shared schema
directories, so those migrations are tracked and retried consistently.

---

## Shared Tables

### `users`
Core user accounts. Every person who logs in to any rtylr app has a row here.

| Column | Type | Notes |
|--------|------|-------|
| `id` | BIGINT PK | |
| `email` | VARCHAR(320) UNIQUE | Login identifier |
| `password_hash` | VARCHAR(512) | bcrypt |
| `full_name` | VARCHAR(255) NULL | |
| `signup_ip_address` | VARCHAR(45) NULL | |
| `signup_device_info` | TEXT NULL | |
| `created_at` | TIMESTAMP | |
| `updated_at` | TIMESTAMP | |

### `user_app_permissions`
Controls which apps each staff user can access and what role they have within a business account.

| Column | Type | Notes |
|--------|------|-------|
| `id` | BIGINT PK | |
| `owner_id` | BIGINT | The business owner's user ID |
| `user_id` | BIGINT | The staff member's user ID |
| `apps` | JSON | Array of AppId strings: `["pos","erp","hr"]` |
| `role` | ENUM | `admin`, `operator`, `staff`, `viewer` |
| `is_active` | BOOLEAN | |
| `created_at` / `updated_at` | TIMESTAMP | |

Unique constraint: `(owner_id, user_id)` — one permission row per user per business.

### `config`
Business configuration (branding, contact info, feature flags).

| Column | Type | Notes |
|--------|------|-------|
| `id` | BIGINT PK | |
| `owner_id` | INT | |
| `site_name` | TEXT NULL | |
| `logo_url` | TEXT NULL | |
| `favicon_url` | TEXT NULL | |
| `meta_tags` | JSON NULL | |
| `contact_email` | TEXT | |
| `contact_number` | TEXT | |
| `instagram` | TEXT NULL | |
| `maintenance_mode` | BOOLEAN | |
| `maintenance_notice` | TEXT NULL | |
| `addons_enabled` | BOOLEAN | |
| `enable_reviews` | BOOLEAN | |
| `enable_wishlist` | BOOLEAN | |
| `enable_newsletter_signup` | BOOLEAN | |
| `about_content` | TEXT | |
| `footer_sub_text` | TEXT NULL | |

### `themes`
Per-business theme/color configuration. Used by Hub's design editor.

Stores full color palette: `primary_color`, `secondary_color`, `accent_color`, `background_color`, `surface_color`, `error_color`, `success_color`, `warning_color`, `info_color`, `text_primary_color`, `text_secondary_color`, font settings, and breakpoints.

---

## ERP Tables

Schema: `schemas/erp/schema.sql`

### `erp_categories`
| Column | Type |
|--------|------|
| `id` | BIGINT PK |
| `owner_id` | BIGINT |
| `name` | VARCHAR(255) |
| `description` | TEXT |

### `erp_suppliers`
| Column | Type | Notes |
|--------|------|-------|
| `id` | BIGINT PK | |
| `owner_id` | BIGINT | |
| `name` | VARCHAR(255) | |
| `email` | VARCHAR(255) | |
| `phone` | VARCHAR(50) | |
| `address` | TEXT | |
| `city`, `country` | VARCHAR(100) | |
| `tax_id` | VARCHAR(100) | |
| `payment_terms` | VARCHAR(100) | |
| `is_active` | BOOLEAN | |

### `erp_products`
| Column | Type | Notes |
|--------|------|-------|
| `id` | BIGINT PK | |
| `owner_id` | BIGINT | |
| `category_id` | BIGINT NULL | FK → erp_categories |
| `supplier_id` | BIGINT NULL | FK → erp_suppliers |
| `name` | VARCHAR(255) | |
| `sku` | VARCHAR(100) | |
| `barcode` | VARCHAR(100) | |
| `unit` | VARCHAR(50) | Default: `unit` |
| `cost_price` | DECIMAL(10,2) | |
| `sale_price` | DECIMAL(10,2) | |
| `stock_quantity` | DECIMAL(10,3) | |
| `low_stock_threshold` | DECIMAL(10,3) | Alert threshold |
| `reorder_quantity` | DECIMAL(10,3) | |
| `is_active` | BOOLEAN | |
| `track_inventory` | BOOLEAN | |

### `erp_inventory_transactions`
Audit log for every stock movement.

| Column | Type | Notes |
|--------|------|-------|
| `type` | ENUM | `purchase`, `sale`, `adjustment`, `return`, `waste` |
| `quantity_change` | DECIMAL(10,3) | Signed delta |
| `quantity_before` | DECIMAL(10,3) | |
| `quantity_after` | DECIMAL(10,3) | |
| `reference_id` | BIGINT | FK to source entity |
| `reference_type` | VARCHAR(50) | e.g. `purchase_order` |

### `erp_purchase_orders`
| Column | Type | Notes |
|--------|------|-------|
| `status` | ENUM | `draft`, `sent`, `partial`, `received`, `cancelled` |
| `po_number` | VARCHAR(100) | |
| `subtotal`, `tax_amount`, `total` | DECIMAL(10,2) | |
| `expected_delivery` | DATE | |
| `received_at` | TIMESTAMP NULL | |

### `erp_purchase_order_items`
Line items for each purchase order. Columns: `purchase_order_id`, `product_id`, `quantity_ordered`, `quantity_received`, `unit_cost`, `total_cost`.

---

## HR Tables

Schema: `schemas/hr/schema.sql`

### `hr_departments`
| Column | Type |
|--------|------|
| `owner_id` | BIGINT |
| `name` | VARCHAR(255) |
| `manager_staff_id` | BIGINT NULL |

### `hr_staff`
| Column | Type | Notes |
|--------|------|-------|
| `user_id` | BIGINT NULL | Links to `users` table if staff has a login |
| `department_id` | BIGINT NULL | |
| `first_name`, `last_name` | VARCHAR | |
| `employment_type` | ENUM | `full_time`, `part_time`, `contractor`, `intern` |
| `status` | ENUM | `active`, `inactive`, `on_leave`, `terminated` |
| `salary` | DECIMAL(10,2) NULL | Annual salary |
| `hourly_rate` | DECIMAL(10,2) NULL | |
| `overtime_rate` | DECIMAL(10,2) NULL | |
| `start_date` | DATE | |
| `apps` | JSON | Which rtylr apps this staff member can access |

### `hr_shifts`
| Column | Type | Notes |
|--------|------|-------|
| `staff_id` | BIGINT | |
| `date` | DATE | |
| `start_time` / `end_time` | TIME | |
| `status` | ENUM | `scheduled`, `in_progress`, `completed`, `absent`, `cancelled` |
| `clock_in_at` / `clock_out_at` | TIMESTAMP NULL | |
| `hours_worked` | DECIMAL(5,2) | |
| `overtime_hours` | DECIMAL(5,2) | |

### `hr_time_off_requests`
| Column | Type | Notes |
|--------|------|-------|
| `type` | ENUM | `vacation`, `sick`, `personal`, `unpaid`, `other` |
| `status` | ENUM | `pending`, `approved`, `rejected`, `cancelled` |
| `days_count` | DECIMAL(4,1) | |
| `approved_by` | BIGINT NULL | Staff ID of approver |

### `hr_payroll_records`
| Column | Type | Notes |
|--------|------|-------|
| `period_start` / `period_end` | DATE | |
| `period_label` | VARCHAR(50) | e.g. `"April 2026"` |
| `regular_hours` / `overtime_hours` | DECIMAL(6,2) | |
| `base_pay`, `overtime_pay`, `bonuses`, `deductions`, `tax_withheld`, `net_pay` | DECIMAL(10,2) | |
| `status` | ENUM | `draft`, `processed`, `paid` |

---

## CRM Tables

Schema: `schemas/crm/schema.sql`

### `crm_customers`
| Column | Type | Notes |
|--------|------|-------|
| `loyalty_points` | INT | Current balance |
| `total_spent` | DECIMAL(10,2) | Lifetime spend |
| `visit_count` | INT | |
| `last_visit_at` | TIMESTAMP NULL | |
| `loyalty_tier` | ENUM | `bronze`, `silver`, `gold`, `platinum` |
| `tags` | JSON | Array of string tags |
| `gender` | ENUM | `male`, `female`, `other`, `prefer_not_to_say` |

### `crm_loyalty_transactions`
| Column | Type | Notes |
|--------|------|-------|
| `type` | ENUM | `earn`, `redeem`, `adjust`, `expire` |
| `points` | INT | Delta |
| `balance_before` / `balance_after` | INT | |
| `reference_id` / `reference_type` | — | Source entity |

### `crm_loyalty_config`
One row per business. Key fields:

| Column | Default | Notes |
|--------|---------|-------|
| `points_per_currency_unit` | 1.0 | Points earned per 1 unit of currency |
| `currency_per_point` | 0.01 | Value when redeeming |
| `bronze_min` / `silver_min` / `gold_min` / `platinum_min` | 0 / 500 / 1500 / 5000 | Tier thresholds |
| `bronze_discount_pct` … `platinum_discount_pct` | 0 / 5 / 10 / 15 | Tier discount % |
| `points_expiry_days` | 365 | |

### `crm_segments`
Audience segments with JSON rule arrays.

| Column | Type |
|--------|------|
| `rules` | JSON |
| `customer_count` | INT |
| `last_calculated_at` | TIMESTAMP NULL |

### `crm_campaigns`
| Column | Type | Notes |
|--------|------|-------|
| `channel` | ENUM | `email`, `sms`, `push`, `whatsapp` |
| `status` | ENUM | `draft`, `scheduled`, `active`, `paused`, `completed`, `cancelled` |
| `segment_id` | BIGINT NULL | Target audience |
| `audience_count` / `sent_count` / `delivered_count` / `opened_count` / `clicked_count` | INT | Metrics |

### `crm_campaign_sends`
Per-customer send tracking: `status` ENUM (`queued`, `sent`, `delivered`, `opened`, `clicked`, `bounced`, `failed`).

---

## Finance Tables

Schema: `schemas/finance/schema.sql`

### `finance_invoices`
| Column | Type | Notes |
|--------|------|-------|
| `invoice_number` | VARCHAR(100) | |
| `status` | ENUM | `draft`, `sent`, `viewed`, `partial`, `paid`, `overdue`, `cancelled`, `void` |
| `discount_type` | ENUM | `fixed`, `percentage` |
| `subtotal`, `discount_amount`, `tax_amount`, `total`, `amount_paid`, `amount_due` | DECIMAL(10,2) | |
| `currency` | VARCHAR(10) | Default `USD` |

### `finance_invoice_items`
Line items. Columns: `description`, `quantity`, `unit`, `unit_price`, `tax_rate`, `discount_pct`, `total`, `sort_order`.

### `finance_payments`
| Column | Type | Notes |
|--------|------|-------|
| `method` | ENUM | `cash`, `bank_transfer`, `card`, `cheque`, `online`, `other` |
| `payment_date` | DATE | |
| `reference` | VARCHAR(255) | |

### `finance_expenses`
| Column | Type | Notes |
|--------|------|-------|
| `category_id` | BIGINT NULL | FK → finance_expense_categories |
| `payment_method` | ENUM | `cash`, `card`, `bank_transfer`, `other` |
| `receipt_url` | TEXT | |
| `is_tax_deductible` | BOOLEAN | |

### `finance_tax_config`
| Column | Type | Notes |
|--------|------|-------|
| `type` | ENUM | `vat`, `sales_tax`, `gst`, `other` |
| `rate` | DECIMAL(5,2) | |
| `is_default` | BOOLEAN | |

### `finance_recurring_invoices`
| Column | Type | Notes |
|--------|------|-------|
| `frequency` | ENUM | `weekly`, `monthly`, `quarterly`, `yearly` |
| `next_invoice_date` | DATE | |
| `end_date` | DATE NULL | |

---

## Flow Tables

Schema: `schemas/flow/schema.sql`

### `flow_workspaces`
| Column | Notes |
|--------|-------|
| `icon` | Emoji, default `🚀` |
| `color` | Hex, default `#4583C4` |
| `is_default` | BOOLEAN |
| `created_by` | BIGINT (user ID) |

### `flow_workspace_members`
Unique `(workspace_id, user_id)`. Role: `admin`, `member`, `viewer`.

### `flow_projects`
| Column | Notes |
|--------|-------|
| `workspace_id` | BIGINT |
| `status` | `active`, `archived`, `completed` |
| `start_date` / `due_date` | DATE NULL |
| `created_by` | BIGINT |

### `flow_tasks`
| Column | Notes |
|--------|-------|
| `parent_task_id` | BIGINT NULL — enables subtasks |
| `status` | `todo`, `in_progress`, `in_review`, `done`, `cancelled` |
| `priority` | `none`, `low`, `medium`, `high`, `urgent` |
| `assignee_id` | BIGINT NULL |
| `reporter_id` | BIGINT |
| `estimated_hours` / `actual_hours` | DECIMAL(6,2) |
| `labels` | JSON — array of label IDs |
| `sort_order` | INT |

### `flow_task_comments`
Simple text comments on tasks. `is_edited` BOOLEAN tracks edits.

### `flow_task_activity`
Audit log. `action` VARCHAR(100), `old_value` / `new_value` TEXT.

---

## POS Tables

Schema: `schemas/pos/schema.sql`

### `pos_locations`
Physical locations / branches. Columns: `name`, `address`, `phone`, `is_active`.

### `pos_tables`
| Column | Notes |
|--------|-------|
| `location_id` | BIGINT NULL |
| `capacity` | INT, default 4 |
| `status` | `available`, `occupied`, `reserved`, `cleaning` |
| `current_session_order_id` | BIGINT NULL |
| `qr_code_url` | TEXT |

### `pos_sessions`
Cash register sessions (opened and closed by staff).

| Column | Notes |
|--------|-------|
| `staff_id` | BIGINT |
| `status` | `open`, `closed` |
| `opening_cash` / `closing_cash` / `expected_cash` / `cash_difference` | DECIMAL(10,2) |
| `total_sales` / `total_orders` | Aggregates |
| `total_cash_sales` / `total_card_sales` / `total_online_sales` | Payment breakdowns |
| `total_refunds` | DECIMAL(10,2) |

### `pos_orders`
| Column | Notes |
|--------|-------|
| `type` | `dine_in`, `takeaway`, `delivery`, `online` |
| `status` | `pending`, `confirmed`, `preparing`, `ready`, `served`, `paid`, `cancelled`, `refunded` |
| `payment_method` | `cash`, `card`, `online`, `split` |
| `payment_status` | `unpaid`, `partial`, `paid`, `refunded` |
| `customer_id` | BIGINT NULL — links to crm_customers |
| `discount_amount` / `tax_amount` / `total` | DECIMAL(10,2) |

### `pos_order_items`
| Column | Notes |
|--------|-------|
| `menu_item_id` | BIGINT NULL |
| `quantity` | DECIMAL(8,3) |
| `status` | `pending`, `preparing`, `ready`, `served`, `cancelled` |
| `discount_pct` / `tax_rate` | DECIMAL |

### `pos_refunds`
Simple refund record. Columns: `order_id`, `amount`, `reason`, `refunded_by`, `refunded_at`.

---

## Key Relationships

```
users
  └─ user_app_permissions (owner_id, user_id)
  └─ config (owner_id)
  └─ themes (owner_id)

erp_products → erp_categories (category_id)
erp_products → erp_suppliers (supplier_id)
erp_purchase_orders → erp_suppliers (supplier_id)
erp_purchase_order_items → erp_purchase_orders / erp_products

hr_staff → hr_departments (department_id)
hr_shifts → hr_staff (staff_id)
hr_payroll_records → hr_staff (staff_id)
hr_time_off_requests → hr_staff (staff_id)

crm_loyalty_transactions → crm_customers (customer_id)
crm_campaign_sends → crm_campaigns / crm_customers
crm_campaigns → crm_segments (segment_id)

finance_invoice_items → finance_invoices (invoice_id)
finance_payments → finance_invoices (invoice_id)
finance_expenses → finance_expense_categories (category_id)
finance_recurring_invoices → finance_invoices (template_invoice_id)

flow_workspace_members → flow_workspaces / users
flow_projects → flow_workspaces (workspace_id)
flow_project_members → flow_projects / users
flow_tasks → flow_projects / flow_workspaces
flow_tasks → flow_tasks (parent_task_id)  ← subtasks
flow_task_comments → flow_tasks
flow_task_activity → flow_tasks

pos_tables → pos_locations (location_id)
pos_sessions → pos_locations / hr_staff
pos_orders → pos_sessions / pos_tables / crm_customers
pos_order_items → pos_orders
pos_refunds → pos_orders
```
