# CRM model and flow

How rtylr's CRM is meant to work end to end, and the order it is being rebuilt
in. It follows the shape the established CRMs converge on (HubSpot, Pipedrive,
Salesforce, Odoo, Close): records for who you know, separate work objects for
the selling you do with them, and activities that tie the two together.

## Records: who you know

| Record   | Table            | What it is |
|----------|------------------|------------|
| Person   | `crm_customers`  | A human the business has a relationship with: a buyer from POS, the store or bookings, or a converted sales contact. Loyalty, segments, campaigns and insights all read this table. |
| Company  | `crm_companies`  | An organisation people belong to. Deals are usually with a company. |

`crm_customers` keeps its name because twenty modules read it; the UI calls it
People/Customers. Prospects are **not** people yet: a cold list of 80,000 rows
must not inflate customer counts, KPIs or campaign audiences, and nobody on it
has consented to marketing.

## Work objects: the selling you do

### Leads: who to qualify

A lead is a prospect to call or email: a person's name, title, phone and email,
plus the company as text. It lives in `crm_leads` until it is qualified or
dropped.

| Status         | Meaning |
|----------------|---------|
| `new`          | Not contacted yet. |
| `attempting`   | Tried, no conversation yet (no answer, voicemail). |
| `connected`    | Had a conversation; qualifying. |
| `nurture`      | Interested but bad timing; revisit later. |
| `qualified`    | Converted: a person, a company and usually a deal now exist. |
| `disqualified` | Not a fit; a reason is required. |

Logging a call moves `new` to `attempting` or `connected` by its outcome, so
the status reflects the work without bookkeeping.

**Conversion** is the single hand-off from prospecting to selling. It publishes
the lead's contact as a `customer_contact_captured` event (subject
`crm_lead`); the customers module matches it to an existing person by email or
phone or creates one, and links it back. It also finds or creates the company
by name, and opens a deal in the chosen pipeline stage, prefilled with the
lead's estimated value.

### Deals: what you are trying to sell

A deal is one opportunity, in `crm_deals`: a title, a company and primary
person, a value billed once, monthly or yearly (the same terms a contract
uses), an expected close date and an owner. It sits in a stage of a pipeline.

Pipelines and stages are per business (`crm_pipelines`, `crm_pipeline_stages`)
so a business can run several sales motions. Each stage has a win probability,
which gives the weighted pipeline value. The default pipeline is created on
first use:

| Stage       | Probability | Kind |
|-------------|-------------|------|
| Qualified   | 10%         | open |
| Meeting     | 25%         | open |
| Proposal    | 50%         | open |
| Negotiation | 75%         | open |
| Won         | 100%        | won  |
| Lost        | 0%          | lost |

Deals can move between open stages in either direction; real deals slip.
Losing requires a reason. A closed deal can be reopened.

**Winning** creates a draft contract prefilled from the deal (value, billing,
person, company). Signing the contract is what Finance invoices from, and what
makes the person a customer in reporting.

## Activities and tasks: the work itself

Every touch is an activity in `crm_touchpoints`, attached to a lead, a deal or
a person: a call with its outcome and duration, an email, a meeting, an SMS or
a note. Call outcomes are the dispositions cold calling needs: connected, no
answer, voicemail, busy, wrong number, call back, meeting booked, not
interested.

Follow-ups are tasks in `crm_tasks` with a due date and an assignee. A lead's
and a deal's next follow-up is its earliest open task, stored on the row so a
list of 80,000 leads can filter and sort by it. The Tasks page is the daily
worklist: overdue, today, upcoming.

## Lists at scale

Every list is server-side: filters, search, sorting and cursor pagination
(`(sort value, id)` keysets), with counts and stats from separate aggregate
queries. Nothing loads a whole table into the browser.

Leads filter on reps (or unassigned), source, import, industry, size,
country, city, billing, follow-up window, whether phone, email, website,
LinkedIn, company or value are filled in, whether the lead was ever worked,
call attempts, value, date added and last activity. Every lead list query and
the per-status count share one filter block in `leads.sql` (a test keeps them
identical), so the tab counts always match the list. Text sets travel as JSON
arrays, since values like "Food, Beverage" contain commas.

Imports run in batches recorded in `crm_import_batches`. The browser parses the
file, the user maps columns, and rows go up in chunks of 500. The server
normalises email and phone, dedupes against the business's leads by email,
then phone, then company plus name, and returns a result per row. Failed rows
are kept with their reason for download. An import can be undone while its
leads are untouched.

## Dashboard

The dashboard reads aggregates only. Pipeline value, weighted value, win rate
and cycle time come from the deal stats; the lead funnel, follow-ups due and
today's calls from the lead stats; open tasks from the task counts. Activity
over a 7, 30 or 90-day window comes from `getSalesActivity`
(`dashboard.sql`): calls and reached calls, leads added and deals won per day,
plus each rep's calls, qualified leads and won and lost deals. Days are the
caller's own: timestamps shift by the caller's UTC offset before the day is
taken, and quiet days are filled in as zeros. Calls count for whoever logged
them, leads and deals for whoever they are assigned to.

## Navigation

| Group     | Pages |
|-----------|-------|
| Overview  | Dashboard |
| Sales     | Leads, Deals, Tasks, Contracts |
| Customers | Customers, Companies |
| Marketing | Segments, Campaigns, Templates |
| Retention | Loyalty, Feedback |

## Rebuild order

1. **Sales**: leads with contact details and statuses, call logging and tasks,
   lead import, conversion, deals in configurable pipelines, won to contract,
   and server-paginated lists with stats.
2. **Customers and Companies**: server-paginated people list, company records,
   one timeline per person and company across orders, deals and activities.
3. **Marketing**: segments, campaigns and templates over consented people,
   with lifecycle and deal filters.
4. **Dashboard and retention**: sales reporting (pipeline, weighted value, win
   rate, cycle time, activity per rep) next to the customer metrics; loyalty
   and feedback reworked on the same list patterns.
