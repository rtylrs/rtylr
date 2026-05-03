# rtylr MCP Server

An [MCP (Model Context Protocol)](https://modelcontextprotocol.io) server that gives AI assistants like Claude access to your rtylr workspace — tasks, projects, customers, invoices, inventory, and more.

## What it does

Once connected, your AI assistant can:

- **Flow** — list workspaces and projects, create/update/delete tasks, view your assigned tasks
- **CRM** — look up customers, view campaigns, check loyalty tiers
- **Finance** — list invoices and expenses, generate P&L reports, create invoices
- **ERP** — browse inventory, look up products and suppliers, view purchase orders

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
1. Your browser opens to `auth.rtylr.com` for login + consent
2. You approve the requested permissions
3. The token is saved to `~/.rtylr-mcp/config.json` (readable only by you)
4. Tokens refresh automatically

To sign out and re-authorize, delete `~/.rtylr-mcp/config.json`.

## Available tools

| Tool | Description |
|------|-------------|
| `flow_list_workspaces` | List all your workspaces |
| `flow_list_projects` | List projects in a workspace |
| `flow_list_tasks` | List tasks (filter by project, status) |
| `flow_get_task` | Get a single task by ID |
| `flow_create_task` | Create a new task |
| `flow_update_task` | Update title, status, priority, assignee, due date |
| `flow_delete_task` | Delete a task |
| `flow_create_project` | Create a new project |
| `flow_my_tasks` | Tasks assigned to you |
| `crm_list_customers` | List CRM customers |
| `crm_get_customer` | Get a customer by ID |
| `crm_list_campaigns` | List marketing campaigns |
| `crm_list_loyalty_tiers` | List loyalty program tiers |
| `finance_list_invoices` | List invoices (filter by status) |
| `finance_list_expenses` | List expenses |
| `finance_profit_loss` | P&L report for a date range |
| `finance_create_invoice` | Create a new invoice |
| `erp_list_products` | List inventory products |
| `erp_get_product` | Get a product by ID |
| `erp_list_suppliers` | List suppliers |
| `erp_list_purchase_orders` | List purchase orders |
| `erp_inventory_report` | Current stock levels |

## Configuration

| Environment variable | Default | Description |
|---------------------|---------|-------------|
| `RTYLR_API_URL` | `https://api.voxire.dev` | Override the API base URL (for self-hosted or dev) |

## Development

```bash
git clone https://github.com/voxire/rtylr-mcp
cd rtylr-mcp
npm install
npm run dev
```

To point at a local backend:

```bash
RTYLR_API_URL=http://localhost:9999 npm run dev
```

## License

MIT — built by [Voxire](https://voxire.com)
