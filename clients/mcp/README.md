# rtylr MCP Server

An [MCP (Model Context Protocol)](https://modelcontextprotocol.io) server that gives AI assistants like Claude access to your rtylr workspace — tasks, projects, customers, invoices, inventory, and more.

## What it does

Once connected, your AI assistant can:

- **Flow** — list workspaces and projects, create/update/delete tasks, view your assigned tasks
- **CRM** — look up customers, view campaigns, check loyalty tiers
- **Finance** — list invoices and expenses, generate P&L reports, create invoices
- **ERP** — browse inventory, look up products and suppliers, view purchase orders
- **POS** — list orders and menu items, manage sessions, check gift cards

## Prerequisites

- Node.js 18+
- A rtylr account at [rtylr.com](https://rtylr.com)

## Installation

```bash
npm install -g @rtylr/mcp
```

Or run directly without installing:

```bash
npx @rtylr/mcp
```

## Usage with Claude Desktop

Add this to your `claude_desktop_config.json` (found at `~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "rtylr": {
      "command": "npx",
      "args": ["-y", "@rtylr/mcp"]
    }
  }
}
```

Restart Claude Desktop. The first time you ask Claude to interact with rtylr, it will open your browser to authorize access.

## Usage with other MCP clients

```bash
rtylr-mcp
```

The server communicates over stdio using the standard MCP protocol. Point any MCP-compatible client at it.

## Authentication

Authorization uses [PKCE OAuth 2.0](https://oauth.net/2/pkce/) — no client secret required.

On first use:

1. Your browser opens the API's `/oauth/authorize` page (`RTYLR_API_URL`, default `https://api.voxire.com`), which redirects to the RTYLR sign-in and consent screen (`sso.rtylr.com` in production)
2. You approve the requested permissions
3. The token is saved to `~/.rtylr-mcp/config.json` (readable only by you)
4. Tokens refresh automatically

To sign out and re-authorize, delete `~/.rtylr-mcp/config.json`, or call the `auth_logout` tool, which also revokes the refresh token.

### Scopes

The token is limited to the permissions you approve on the consent screen. The server enforces them on every request:

| Scope                            | Allows                        |
| -------------------------------- | ----------------------------- |
| `flow:read` / `flow:write`       | Flow routes (`/flow/*`)       |
| `crm:read` / `crm:write`         | CRM routes (`/crm/*`)         |
| `finance:read` / `finance:write` | Finance routes (`/finance/*`) |
| `erp:read` / `erp:write`         | ERP routes (`/erp/*`)         |
| `pos:read` / `pos:write`         | POS routes (`/pos/*`)         |

`read` covers `GET` requests; `write` covers every method and implies `read`. Account, security and settings routes are never reachable with an OAuth token. A request outside the granted scopes returns `403` with `WWW-Authenticate: Bearer error="insufficient_scope"`. When a new release asks for a scope your saved grant lacks, the next tool call reopens the browser to approve it.

## Available tools

The server registers 98 tools (`src/tools/index.ts`).

### Flow

| Tool                      | Description                                                                            |
| ------------------------- | -------------------------------------------------------------------------------------- |
| `flow_list_comments`      | List all comments on a task.                                                           |
| `flow_create_comment`     | Add a comment to a task.                                                               |
| `flow_my_tasks`           | List tasks assigned to the authenticated user, grouped or filtered by status.          |
| `flow_list_projects`      | List projects in a workspace.                                                          |
| `flow_get_project`        | Get a single project by ID.                                                            |
| `flow_create_project`     | Create a new project in a workspace.                                                   |
| `flow_update_project`     | Update a project name, description, or color.                                          |
| `flow_delete_project`     | Delete a project and all its tasks permanently.                                        |
| `flow_list_tasks`         | List tasks in a project or workspace. Supports filtering by status and priority.       |
| `flow_get_task`           | Get a single task by ID, including its description, labels, subtasks, and comments.    |
| `flow_create_task`        | Create a new task.                                                                     |
| `flow_update_task`        | Update an existing task — title, status, priority, assignee, due date, or description. |
| `flow_delete_task`        | Permanently delete a task.                                                             |
| `flow_update_task_status` | Quickly move a task to a new status (shortcut for flow_update_task).                   |
| `flow_list_workspaces`    | List all Flow workspaces the authenticated user has access to.                         |
| `flow_create_workspace`   | Create a new Flow workspace.                                                           |

### CRM

| Tool                       | Description                                                                              |
| -------------------------- | ---------------------------------------------------------------------------------------- |
| `crm_list_automations`     | List CRM automations (triggered workflows like welcome emails, re-engagement sequences). |
| `crm_toggle_automation`    | Enable or disable an automation.                                                         |
| `crm_list_campaigns`       | List marketing campaigns. Filter by status.                                              |
| `crm_get_campaign`         | Get a single campaign by ID, including stats.                                            |
| `crm_create_campaign`      | Create a new marketing campaign.                                                         |
| `crm_update_campaign`      | Update a campaign (only possible while in draft status).                                 |
| `crm_list_customers`       | List CRM customers. Supports search by name or email.                                    |
| `crm_get_customer`         | Get a CRM customer by ID, including their loyalty status and purchase history.           |
| `crm_create_customer`      | Add a new customer to the CRM.                                                           |
| `crm_update_customer`      | Update a customer record.                                                                |
| `crm_delete_customer`      | Delete a customer from the CRM.                                                          |
| `crm_list_feedback`        | List customer feedback and reviews.                                                      |
| `crm_list_loyalty_tiers`   | List loyalty program tiers (Bronze, Silver, Gold, etc.).                                 |
| `crm_create_loyalty_tier`  | Create a new loyalty tier.                                                               |
| `crm_list_loyalty_members` | List customers enrolled in the loyalty program.                                          |
| `crm_assign_loyalty_tier`  | Assign or change a customer's loyalty tier.                                              |
| `crm_list_loyalty_rewards` | List redeemable loyalty rewards.                                                         |
| `crm_list_segments`        | List customer segments (audience groups used for targeting campaigns).                   |
| `crm_create_segment`       | Create a new customer segment with filter rules.                                         |

### Finance

| Tool                           | Description                                              |
| ------------------------------ | -------------------------------------------------------- |
| `finance_list_bills`           | List accounts-payable bills.                             |
| `finance_get_bill`             | Get a single bill by ID.                                 |
| `finance_create_bill`          | Create a new bill (accounts payable).                    |
| `finance_update_bill`          | Update a bill.                                           |
| `finance_delete_bill`          | Delete a bill.                                           |
| `finance_list_expenses`        | List finance expenses.                                   |
| `finance_create_expense`       | Record a new expense.                                    |
| `finance_delete_expense`       | Delete an expense.                                       |
| `finance_list_invoices`        | List finance invoices. Filter by status.                 |
| `finance_get_invoice`          | Get a single invoice by ID.                              |
| `finance_create_invoice`       | Create a new invoice.                                    |
| `finance_update_invoice`       | Update an invoice.                                       |
| `finance_delete_invoice`       | Delete an invoice (only possible while in draft status). |
| `finance_list_quotes`          | List finance quotes / estimates.                         |
| `finance_create_quote`         | Create a new quote.                                      |
| `finance_update_quote`         | Update a quote.                                          |
| `finance_delete_quote`         | Delete a quote.                                          |
| `finance_convert_quote`        | Convert an accepted quote into an invoice.               |
| `finance_profit_loss`          | Get profit & loss report for a date range.               |
| `finance_balance_sheet`        | Get balance sheet as of a given date.                    |
| `finance_cash_flow`            | Get cash flow statement for a date range.                |
| `finance_tax_config`           | Get tax configuration (rates and rules).                 |
| `finance_chart_of_accounts`    | Get the chart of accounts (hierarchical tree).           |
| `finance_list_journal_entries` | List journal entries.                                    |
| `finance_post_journal_entry`   | Post a draft journal entry (moves it to posted status).  |

### ERP

| Tool                         | Description                                          |
| ---------------------------- | ---------------------------------------------------- |
| `erp_inventory_report`       | Get current inventory levels and stock report.       |
| `erp_list_price_lists`       | List ERP price lists.                                |
| `erp_create_price_list`      | Create a new price list.                             |
| `erp_list_stock_transfers`   | List stock transfers between locations.              |
| `erp_list_stock_takes`       | List stock take / physical inventory count sessions. |
| `erp_list_products`          | List ERP products / inventory items.                 |
| `erp_get_product`            | Get a single ERP product by ID.                      |
| `erp_get_product_by_barcode` | Look up an ERP product by barcode.                   |
| `erp_create_product`         | Create a new ERP product.                            |
| `erp_update_product`         | Update an ERP product.                               |
| `erp_delete_product`         | Delete an ERP product.                               |
| `erp_list_purchase_orders`   | List ERP purchase orders.                            |
| `erp_get_purchase_order`     | Get a single purchase order by ID.                   |
| `erp_create_purchase_order`  | Create a new purchase order.                         |
| `erp_approve_purchase_order` | Approve a purchase order.                            |
| `erp_list_suppliers`         | List ERP suppliers.                                  |
| `erp_create_supplier`        | Create a new supplier.                               |
| `erp_update_supplier`        | Update a supplier.                                   |
| `erp_delete_supplier`        | Delete a supplier.                                   |

### POS

| Tool                                | Description                                                                                 |
| ----------------------------------- | ------------------------------------------------------------------------------------------- |
| `pos_get_gift_card`                 | Get a gift card by code, including balance and status.                                      |
| `pos_redeem_gift_card`              | Redeem a gift card against an order.                                                        |
| `pos_kds_list_orders`               | List orders on the Kitchen Display System (KDS), optionally filtered by station and status. |
| `pos_list_menu_items`               | List POS menu items. Supports filtering by category and search.                             |
| `pos_get_menu_item`                 | Get a POS menu item by ID.                                                                  |
| `pos_create_menu_item`              | Create a new POS menu item.                                                                 |
| `pos_update_menu_item`              | Update a POS menu item.                                                                     |
| `pos_toggle_menu_item_availability` | Toggle the availability of a POS menu item.                                                 |
| `pos_list_orders`                   | List POS orders. Supports filtering by status.                                              |
| `pos_get_order`                     | Get a POS order by ID.                                                                      |
| `pos_create_order`                  | Create a new POS order.                                                                     |
| `pos_update_order_status`           | Update the status of a POS order.                                                           |
| `pos_void_order`                    | Void a POS order.                                                                           |
| `pos_list_sessions`                 | List POS sessions. Supports filtering by location and status.                               |
| `pos_get_session`                   | Get a POS session by ID.                                                                    |
| `pos_open_session`                  | Open a new POS session.                                                                     |
| `pos_close_session`                 | Close a POS session.                                                                        |

### Auth

| Tool          | Description                                                                                           |
| ------------- | ----------------------------------------------------------------------------------------------------- |
| `auth_status` | Check authentication status — returns current user info if logged in, or indicates not authenticated. |
| `auth_logout` | Log out — revokes the current access token and clears local credentials.                              |

## Configuration

| Environment variable | Default                  | Description                                                           |
| -------------------- | ------------------------ | --------------------------------------------------------------------- |
| `RTYLR_API_URL`      | `https://api.voxire.com` | Override the API base URL for local development or a test environment |

## Development

The server lives in the `rtylr` monorepo at `clients/mcp`. From the repository root:

```bash
pnpm -C clients install --frozen-lockfile
pnpm -C clients --filter @rtylr/mcp dev        # run from source with tsx
pnpm -C clients --filter @rtylr/mcp typecheck
pnpm -C clients --filter @rtylr/mcp build      # compiles to dist/
```

To point at a local backend:

```bash
RTYLR_API_URL=http://localhost:9999 pnpm -C clients --filter @rtylr/mcp dev
```

## License

MIT — see the repository [LICENSE](../../LICENSE). Built by [Voxire](https://voxire.com).
