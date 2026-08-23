// Shared FAQ content. Each block answers a question a real buyer types
// into Google, ChatGPT or Perplexity. The questions are phrased the way
// people ask them, not the way marketing writes them: "how much does X
// cost" beats "our pricing explained" every time.
//
// These are rendered on the page, by components/FaqSection. That is the
// whole point of them and it was not true until Phase 1: every array
// here used to be imported solely to build FAQPage JSON-LD, so the
// questions were asserted to Google and appeared nowhere a reader could
// see them. There is no FAQPage markup any more (Google removed the rich
// result in May 2026 and the documentation in June), so the visible copy
// is now the only delivery mechanism. If you add an entry, render it.
//
// Three rules for anything written here:
//
//   1. **No prices.** rtylr does not publish any. Tier *names* and what
//      a tier includes are fine; a rate, a discount or a currency amount
//      is not. Answers here previously claimed per-seat daily proration
//      and a free tier at 3 seats, neither of which matched /pricing/.
//   2. **No uncited competitor claim.** "Toast is POS-first and
//      everything else is bolted on" needs a dated citation to Toast's
//      own documentation. Until a comparison page exists that carries
//      one, answer these questions with what rtylr does and let the
//      reader draw the comparison. Competitors read these pages and they
//      complain to Google.
//   3. **No em dashes, and no hyphen standing in for one.** Restructure,
//      or use a colon.
//
// Answers stay factual and roughly 40 to 80 words, which is the length
// that survives snippet extraction intact.

import type { FAQEntry } from './structured-data';

/**
 * Site-wide FAQ, for the home page and /pricing/. Covers the buyer
 * questions that do not belong to one app.
 */
export const GLOBAL_FAQS: FAQEntry[] = [
  {
    question: 'What is rtylr?',
    answer:
      'rtylr is an all-in-one business operating system for small and mid-sized businesses. It bundles point of sale, ERP, HR, CRM, finance and project management into one platform, so a team stops stitching a register, a spreadsheet, an accounting tool and a task tracker together. Built for restaurants, retail and services businesses running 2 to 200 staff.',
  },
  {
    question: 'How much does rtylr cost?',
    answer:
      'rtylr has not published its prices yet. The Free plan is genuinely free, needs no card, and covers 10 team members at one location with all 7 apps, so you can run a location on it today. Pro and Scale are quoted: tell us your team size and how many locations you run. See rtylr.com/pricing for what each plan includes.',
  },
  {
    question: 'Can I use only one app from rtylr, or do I need all of them?',
    answer:
      'You can use any single app on its own. Flow, Finance or HR each work standalone, with their own login, settings and team. The reason to run more than one is shared data: customers from CRM reach the POS, staff from HR reach the schedule, and POS revenue reaches Finance without an export. It is never required.',
  },
  {
    question: 'Does rtylr replace my point-of-sale system?',
    answer:
      'Yes. rtylr POS is a full-service register for restaurants and retail. It handles tickets, splits, tips, modifiers, inventory counts, offline mode and receipt printing on certified hardware. It runs on iPad, Android tablet and desktop, and it syncs in real time with the rest of the platform, so a sale moves stock without an export.',
  },
  {
    question: 'Is my data mine?',
    answer:
      'Yes. Every rtylr account exports clean CSVs of customers, orders, staff, invoices and inventory at any time from the settings page, with no support ticket. Data lives in the EU or US region, your choice. See rtylr.com/trust for compliance detail and rtylr.com/legal/security for the security architecture.',
  },
  {
    question: 'What does rtylr replace?',
    answer:
      'For most businesses that adopt it: a register, an inventory sheet, a scheduling tool, a payroll export, a customer list and an accounting package. The claim rtylr makes is not that each of those is worse on its own. It is that running them separately means the same number is entered more than once, and the versions disagree.',
  },
];

/**
 * Per-app FAQs, rendered on each product page.
 */

export const POS_FAQS: FAQEntry[] = [
  {
    question: 'Does rtylr POS work offline?',
    answer:
      'Yes. rtylr POS keeps a full local copy of the menu, modifiers and active orders. When the internet drops, the register keeps taking orders, printing tickets and processing cash. Card transactions queue and settle automatically once the connection returns, so a Saturday night outage does not stop the floor.',
  },
  {
    question: 'What hardware works with rtylr POS?',
    answer:
      'rtylr POS runs on iPad, Android tablet and Windows or macOS desktop. It pairs with Epson and Star receipt printers, standard barcode scanners, Zebra and Symbol scanners for retail, and Stripe Terminal, Clover and Square card readers. See rtylr.com/hardware for the tested list.',
  },
  {
    question: 'Can rtylr POS handle multi-location restaurants?',
    answer:
      'Yes, on the Pro plan and above. Each location gets its own menu, pricing, staff and reporting, and the owner gets one consolidated dashboard across every site. A menu change can be pushed to all locations at once or tuned per location.',
  },
  {
    question: 'What does rtylr POS replace?',
    answer:
      'The register, and the work that usually sits behind it: the stock count someone redoes by hand, the tips spreadsheet that feeds payroll, and the end-of-day export into accounting. Because the POS writes to the same database as ERP, HR and Finance, a sale updates stock and revenue as it happens rather than overnight.',
  },
];

export const FLOW_FAQS: FAQEntry[] = [
  {
    question: 'What is rtylr Flow?',
    answer:
      "Flow is rtylr's project and task management app. It has kanban boards, lists, timeline views, cycles, milestones, custom fields and automations. What makes it different from a standalone tracker is that it reads the rest of your business: who is on shift from HR, which customer a project belongs to from CRM, and which work is billable from Finance.",
  },
  {
    question: 'What can Flow do that a standalone task tracker cannot?',
    answer:
      'Assign a task to a staff member who already exists in HR, attach it to a customer who already exists in CRM, and push billable hours onto an invoice in Finance, without copying anything between tools. The tasks, boards, comments and automations you would expect are all there; the connection to real business records is the part a separate tracker cannot have.',
  },
  {
    question: 'Does Flow have a free tier?',
    answer:
      'Yes. Flow is included in the Free plan along with the other six apps, covering 10 team members at one location, with unlimited projects, tasks, comments and attachments. Pro and Scale raise the team and location limits and add the controls a larger team needs. rtylr does not publish prices yet, so Pro and Scale are quoted.',
  },
];

export const HR_FAQS: FAQEntry[] = [
  {
    question: 'What does rtylr HR do?',
    answer:
      'rtylr HR handles employee records, scheduling, shift swaps, time-off requests, payroll runs and compliance documents. It supports restaurant-specific workflows including tip pooling, minor work-hour limits and split shifts, alongside general retail and services HR.',
  },
  {
    question: 'Does rtylr HR run payroll for the US?',
    answer:
      'rtylr HR calculates gross-to-net pay, taxes and withholdings, and exports pay files compatible with Gusto, ADP and Rippling for filing. Direct filing through rtylr is on the roadmap. For now rtylr does the calculation and a third-party provider does the IRS filing.',
  },
  {
    question: 'Can staff see their schedule on their phone?',
    answer:
      'Yes. Every staff member gets a mobile view of their schedule, time-off balance and shift swap requests. They can request a swap and submit time off from their phone without an admin account.',
  },
  {
    question: 'What does rtylr HR replace?',
    answer:
      'The scheduling tool, the time-off inbox, and the double entry that happens when a new hire is created in HR, again in payroll, again in the CRM and again in access control. In rtylr the person is created once and the schedule, the pay run and the POS login all read the same record.',
  },
];

export const CRM_FAQS: FAQEntry[] = [
  {
    question: 'What does rtylr CRM track?',
    answer:
      'rtylr CRM tracks customers, visits, purchases, loyalty points and campaign touchpoints. For restaurants and retail it captures who ordered what straight from the POS, with no import. For services businesses you can log contacts, notes and deal stages by hand.',
  },
  {
    question: 'Does rtylr CRM have email marketing?',
    answer:
      'Yes. Segment customers by purchase history, loyalty tier or custom tags, then send branded email campaigns. Open and click tracking are built in. SMS campaigns are available on the Pro plan through the Twilio integration.',
  },
  {
    question: 'What does rtylr CRM replace?',
    answer:
      'The loyalty app, the mailing list, and the export that connects them to what people actually bought. Because CRM shares a database with the POS, a segment like "spent above average, has not visited in 60 days" stays current on its own instead of being rebuilt from a CSV each month.',
  },
];

export const FINANCE_FAQS: FAQEntry[] = [
  {
    question: 'What does rtylr Finance handle?',
    answer:
      'rtylr Finance handles invoicing, expenses, bill payments, bank reconciliation, and profit and loss, balance sheet and cash flow reports. Revenue from the POS, Stripe and other integrations arrives automatically. Expenses can come from a bank feed or be entered by hand.',
  },
  {
    question: 'Can rtylr Finance replace QuickBooks?',
    answer:
      'For many small businesses, yes: invoicing, expenses, reconciliation, P&L and tax reports are all covered. A business with complex audit requirements or multi-entity consolidation may still want a dedicated ledger, and rtylr Finance exports to QuickBooks and Xero for exactly that case. That limit is worth knowing before you switch.',
  },
  {
    question: 'Does rtylr Finance support multiple currencies?',
    answer:
      'Yes. Invoice in any currency, and rtylr converts using the exchange rate at invoice time. P&L reports can be read in any reporting currency, with FX lines shown explicitly rather than folded into the totals.',
  },
  {
    question: 'What does rtylr Finance replace?',
    answer:
      'The month-end export. Revenue, refunds, tips, purchase orders and payroll already live in the same database as the ledger, so the close does not begin with reconciling four systems against each other. That is the difference the page is really claiming, and it is the one a standalone accounting package cannot make.',
  },
];

export const ERP_FAQS: FAQEntry[] = [
  {
    question: 'What does rtylr ERP manage?',
    answer:
      'rtylr ERP covers inventory, suppliers, purchase orders, stock transfers between locations, low-stock alerts and cost tracking. It reads item sales from the POS as they happen, so on-hand stock stays accurate without a nightly sync.',
  },
  {
    question: 'Does rtylr ERP support multi-warehouse inventory?',
    answer:
      'Yes. Every item can be stocked at multiple locations with per-location counts, reorder thresholds and transfer tracking. A purchase order can be split across warehouses, and staff see only their assigned location by default.',
  },
  {
    question: 'What does rtylr ERP replace?',
    answer:
      'The stock spreadsheet, and the CSV that was supposed to keep it in step with sales. A sale in the POS depletes the recipe components in ERP at the moment it happens, so "how many units of one component remain across three locations" is a question you can answer without opening a second tool.',
  },
];
