export const multiTenant = {
  title: 'Multi-Tenant Architecture',
  body: `
<p>rtylr is built as a multi-tenant platform where each organization operates in complete isolation. This document covers the organizational hierarchy, roles, permissions, and addon gating system.</p>

<h2 id="businesses">Businesses</h2>
<p>A business (organization) is the top-level tenant in rtylr. Each business has its own data, team, settings, billing, and addons. Businesses are fully isolated from each other at the database level.</p>

<h2 id="branches">Branches &amp; Locations</h2>
<p>A business can have one or more branches (physical locations). Each branch has its own address, operating hours, tax settings, and inventory. Branches share the same team directory and product catalog but can have branch-specific pricing, stock levels, and settings.</p>

<h2 id="floors">Floors</h2>
<p>Within a branch, floors represent physical areas (e.g., Main Floor, Patio, Rooftop). Floors are used in POS for table mapping and in Book for resource assignment. Each floor can have its own layout with tables, stations, or rooms.</p>

<h2 id="terminals">Terminals</h2>
<p>Terminals are POS devices registered to a branch and floor. Each terminal has a unique identifier and tracks its own session, cash drawer, and hardware configuration (receipt printer, barcode scanner, payment terminal).</p>

<h2 id="roles">Roles</h2>
<p>rtylr has a layered role system. Organization-level roles control access to apps and settings:</p>
<ul>
  <li><strong>Owner</strong> - Full access to everything, including billing and destructive actions. Cannot be removed.</li>
  <li><strong>Admin</strong> - Full access to all apps and settings except billing and ownership transfer.</li>
  <li><strong>Operator</strong> - Access to day-to-day operational apps (POS, ERP, Book) with limited settings access.</li>
  <li><strong>Staff</strong> - Basic access to assigned apps. Cannot modify settings or view financial data.</li>
  <li><strong>Viewer</strong> - Read-only access to assigned apps. Cannot create, update, or delete anything.</li>
</ul>

<h3 id="pos-roles">POS-Specific Roles</h3>
<p>In addition to organization roles, POS has its own role layer:</p>
<ul>
  <li><strong>Cashier</strong> - Can create orders, process payments, and handle basic operations.</li>
  <li><strong>Shift Lead</strong> - Cashier abilities plus shift management, discount approval, and limited refunds.</li>
  <li><strong>POS Manager</strong> - Full POS access including refunds, voids, price overrides, cash drawer management, and day-end reports.</li>
</ul>

<h2 id="permissions">Permissions</h2>
<p>Permissions are granular controls within each app. They define what a user can do: view, create, update, delete, export, and app-specific actions (e.g., approve refunds, run payroll). Permissions are assigned through roles but can be customized per user when needed.</p>

<h2 id="team-invite-flow">Team Invite Flow</h2>
<p>The team invite process works as follows:</p>
<ol>
  <li>An Admin or Owner goes to <strong>Hub &gt; Team</strong> and clicks Invite.</li>
  <li>Enter the invitee's email address and select a role.</li>
  <li>Choose which apps the invitee should have access to.</li>
  <li>Optionally assign to a specific branch.</li>
  <li>The invitee receives an email with a link to accept the invitation and create their account.</li>
  <li>Once accepted, the user appears in the team list with their assigned role and app access.</li>
</ol>

<h2 id="addon-gating">Addon Gating</h2>
<p>Addons are gated at the organization or branch level. When an addon is disabled, its features are hidden from the interface and its API endpoints return <code>403 Forbidden</code>. This ensures that teams only see and interact with the features their organization has enabled. Addon state is checked on every request, so changes take effect immediately.</p>
  `.trim(),
};
