export const support = {
  title: 'Support (Help Desk)',
  body: `
<p>The rtylr Support app provides a shared inbox and ticketing system for handling customer inquiries. It integrates with CRM for customer context and with Hub for team assignment.</p>

<h2 id="shared-inbox">Shared Inbox</h2>
<p>The shared inbox aggregates customer messages from email, contact forms, and in-app chat into a single view. Team members can see all conversations, claim unassigned tickets, and collaborate on replies. Internal notes let agents discuss a ticket privately before responding.</p>

<h2 id="tickets">Tickets</h2>
<p>Tickets represent customer support requests. Each ticket has:</p>
<ul>
  <li>Subject and description</li>
  <li>Customer (linked to CRM profile)</li>
  <li>Assignee and team</li>
  <li>Priority (Low, Medium, High, Urgent)</li>
  <li>Status: <code>open</code>, <code>pending</code>, <code>resolved</code>, <code>closed</code></li>
  <li>Tags and category</li>
  <li>Full conversation thread</li>
</ul>
<p>Tickets can be created manually by agents, generated from incoming emails, or submitted by customers through a support portal.</p>

<h2 id="customer-profiles">Customer Profiles</h2>
<p>When viewing a ticket, the agent sees the customer's CRM profile in a sidebar: contact info, order history, loyalty status, and previous tickets. This context helps agents resolve issues faster without switching between apps.</p>

<h2 id="knowledge-base">Knowledge Base</h2>
<p>Create and publish help articles organized into categories. Articles support rich text formatting, images, and embedded videos. The knowledge base is searchable by customers and agents. Articles can be linked in ticket replies to deflect common questions.</p>

<h2 id="canned-responses">Canned Responses</h2>
<p>Save frequently used replies as canned responses. Agents can insert a canned response into a ticket reply with a shortcut, then customize it before sending. Canned responses support dynamic placeholders for customer name, order details, and ticket info.</p>

<h2 id="sla">SLA Settings</h2>
<p>Define service level agreements with target response and resolution times based on ticket priority. SLA timers are visible on each ticket and in reporting. Breached SLAs trigger escalation notifications to managers. SLA tracking pauses when a ticket is in <code>pending</code> status (waiting on customer).</p>
  `.trim(),
};
