export const hub = {
  title: 'Hub (Command Center)',
  body: `
<p>The rtylr Hub is the central command center for your organization. It is the first thing you see after logging in and the place where you manage your team, apps, settings, and billing.</p>

<h2 id="dashboard">Customizable Dashboard</h2>
<p>The Hub dashboard is a drag-and-drop canvas of widgets. Default widgets include recent activity, quick actions, sales summary, and pending tasks. Add, remove, and rearrange widgets to fit your workflow. Widget data is pulled from all connected apps.</p>

<h2 id="team-management">Team Management</h2>
<p>Manage your organization's team from Hub:</p>
<ul>
  <li><strong>Invite</strong> - Send email invitations to new team members</li>
  <li><strong>Roles</strong> - Assign roles that control access across apps (Owner, Admin, Operator, Staff, Viewer)</li>
  <li><strong>Permissions</strong> - Fine-tune permissions per app and per action (e.g., allow refunds in POS but not void orders)</li>
  <li><strong>Deactivate</strong> - Revoke access without deleting the user record</li>
</ul>

<h2 id="apps-management">Apps Management</h2>
<p>Enable or disable rtylr apps for your organization. Each app (POS, ERP, HR, CRM, Finance, Flow, Store, Book, Support, Insights) can be turned on independently. Disabled apps are hidden from the sidebar and inaccessible to all team members.</p>

<h2 id="addons">Addons</h2>
<p>Browse and manage addons that extend rtylr's functionality. Addons can be enabled per branch or organization-wide. See the <a href="/docs/addons">Addon System</a> documentation for details.</p>

<h2 id="activity-log">Activity Log</h2>
<p>The activity log shows a chronological feed of actions taken across all apps: user logins, data changes, settings updates, and system events. Filter by user, app, action type, or date range.</p>

<h2 id="audit-trail">Audit Trail</h2>
<p>The audit trail provides a detailed, immutable record of all significant actions for compliance purposes. It records who did what, when, from which device, and the before/after state of changed data. Audit logs cannot be deleted and are retained according to your plan's retention policy.</p>

<h2 id="developer-tools">Developer Tools</h2>
<ul>
  <li><strong>API Keys</strong> - Generate and manage API keys scoped to specific apps and permissions. See the <a href="/docs/api-overview">API Overview</a> for usage details.</li>
  <li><strong>Webhooks</strong> - Configure webhook endpoints to receive real-time notifications when events occur (e.g., new order, customer created, inventory updated). Each webhook can be filtered by event type.</li>
</ul>

<h2 id="integrations">Integrations</h2>
<p>Connect rtylr to third-party services. Available integrations include payment processors, accounting software, email providers, SMS gateways, and shipping carriers. Integrations are configured in <strong>Hub &gt; Settings &gt; Integrations</strong>.</p>

<h2 id="settings">Settings</h2>
<p>Organization-wide settings managed in Hub:</p>
<ul>
  <li><strong>Business</strong> - Company name, address, timezone, logo, and legal info</li>
  <li><strong>Domain</strong> - Custom domain configuration for Store and public-facing pages</li>
  <li><strong>Email</strong> - Sender address, email templates, and SMTP settings for transactional emails</li>
  <li><strong>Localization</strong> - Default language, date format, and number format</li>
</ul>

<h2 id="billing">Billing</h2>
<p>View your current plan, manage your subscription, update payment methods, and download invoices. Usage-based addon charges are itemized on each billing statement. Upgrade or downgrade plans at any time with prorated billing.</p>
  `.trim(),
};
