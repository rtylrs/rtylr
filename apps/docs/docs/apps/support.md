# Support (Help Desk)

The rtylr Support app provides a shared inbox and ticketing system for handling customer inquiries. It integrates with CRM for customer context and with Hub for team assignment.

## Shared Inbox

The shared inbox aggregates customer messages from email, contact forms, and in-app chat into a single view. Team members can see all conversations, claim unassigned tickets, and collaborate on replies. Internal notes let agents discuss a ticket privately before responding.

## Tickets

Tickets represent customer support requests. Each ticket has:

- Subject and description
- Customer (linked to CRM profile)
- Assignee and team
- Priority (Low, Medium, High, Urgent)
- Status: `open`, `pending`, `resolved`, `closed`
- Tags and category
- Full conversation thread

Tickets can be created manually by agents, generated from incoming emails, or submitted by customers through a support portal.

## Customer Profiles

When viewing a ticket, the agent sees the customer's CRM profile in a sidebar: contact info, order history, loyalty status, and previous tickets. This context helps agents resolve issues faster without switching between apps.

## Knowledge Base

Create and publish help articles organized into categories. Articles support rich text formatting, images, and embedded videos. The knowledge base is searchable by customers and agents. Articles can be linked in ticket replies to deflect common questions.

## Canned Responses

Save frequently used replies as canned responses. Agents can insert a canned response into a ticket reply with a shortcut, then customize it before sending. Canned responses support dynamic placeholders for customer name, order details, and ticket info.

## SLA Settings

Define service level agreements with target response and resolution times based on ticket priority. SLA timers are visible on each ticket and in reporting. Breached SLAs trigger escalation notifications to managers. SLA tracking pauses when a ticket is in `pending` status (waiting on customer).
