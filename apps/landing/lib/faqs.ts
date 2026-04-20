// Shared FAQ content for AI-SEO extraction. Each block answers a
// question a real buyer types into ChatGPT / Google AI Overviews /
// Perplexity. The questions are phrased the way people ask them, not
// the way marketing writes them — "how much does X cost" beats "our
// pricing explained" every time.
//
// The admin of this file is the marketing lead. Answers should stay
// factual, ~40-80 words (optimal for snippet extraction), and
// current. When pricing or features change, edit here.

import type { FAQEntry } from './structured-data';

/**
 * Global FAQ — used on the homepage and /pricing. Covers the top
 * buyer questions that don't belong to a specific app.
 */
export const GLOBAL_FAQS: FAQEntry[] = [
  {
    question: 'What is rtylr?',
    answer:
      'rtylr is an all-in-one business operating system for small and mid-sized businesses. It bundles point of sale, ERP, HR, CRM, finance, and project management into one platform so teams stop stitching Shopify, Toast, QuickBooks, Asana, and a dozen other tools together. Built for restaurants, retail, and services businesses with 2–200 staff.',
  },
  {
    question: 'How much does rtylr cost?',
    answer:
      'rtylr is priced per seat, per day, with daily proration — you pay only for the seats you actually used. There is a free tier for businesses under 3 active seats. Paid tiers (Pro and Scale) add multi-location, advanced analytics, and priority support. See rtylr.com/pricing for current per-seat rates.',
  },
  {
    question: 'Can I use only one app from rtylr, or do I need all of them?',
    answer:
      'You can buy and use any single app. rtylr Flow, Recruit, or Finance each work standalone with their own login, settings, and team. The upside of using multiple rtylr apps is shared data — customers from CRM flow into POS, staff from HR onto schedules, invoices from POS into Finance — but it is never required.',
  },
  {
    question: 'Does rtylr replace my point-of-sale system?',
    answer:
      'Yes. rtylr POS is a full-service register for restaurants and retail. It handles tickets, splits, tips, modifiers, inventory counts, offline mode, and receipt printing on certified hardware. It runs on iPad, Android tablet, and desktop, and syncs in real time with the rest of the platform.',
  },
  {
    question: 'Is my data mine?',
    answer:
      'Yes. Every rtylr account exports clean CSVs of customers, orders, staff, invoices, and inventory at any time from the settings page, no support ticket required. Data lives in EU or US regions (your choice). See rtylr.com/trust for compliance details and rtylr.com/legal/security for the security architecture.',
  },
  {
    question: 'What makes rtylr different from Toast, Square, or Lightspeed?',
    answer:
      'Toast, Square, and Lightspeed are POS-first — everything else (HR, finance, inventory) is a bolted-on module or a third-party integration. rtylr is platform-first: POS, HR, CRM, finance, ERP, and project management are peer apps that share a single database. Per-seat daily proration means you are not locked into month-over-month waste on unused seats.',
  },
];

/**
 * Per-app FAQs. Emitted on each product page so "how does rtylr
 * Flow compare to Asana" or "does rtylr POS work offline" queries
 * get answered with markup Google and ChatGPT can extract directly.
 */

export const POS_FAQS: FAQEntry[] = [
  {
    question: 'Does rtylr POS work offline?',
    answer:
      'Yes. rtylr POS keeps a full local copy of the menu, modifiers, and active orders. When the internet drops, the register keeps taking orders, printing tickets, and processing cash. Card transactions queue and settle automatically when the connection returns.',
  },
  {
    question: 'What hardware works with rtylr POS?',
    answer:
      'rtylr POS runs on iPad, Android tablet, and Windows / macOS desktop. It pairs with Epson and Star receipt printers, standard barcode scanners, Zebra and Symbol scanners for retail, and Stripe Terminal, Clover, and Square card readers.',
  },
  {
    question: 'Can rtylr POS handle multi-location restaurants?',
    answer:
      'Yes, on the Pro plan and above. Each location has its own menu, pricing, staff, and reporting, and the owner gets a consolidated dashboard across all sites. Menu changes can be pushed to every location at once or tuned per-location.',
  },
  {
    question: 'How does rtylr POS compare to Toast or Square?',
    answer:
      'rtylr POS has feature parity with Toast and Square on the register side — splits, tips, modifiers, offline mode, multi-location. The difference is the platform: inventory, staff, accounting, and customer loyalty are built in, not separate paid add-ons.',
  },
];

export const FLOW_FAQS: FAQEntry[] = [
  {
    question: 'What is rtylr Flow?',
    answer:
      "Flow is rtylr's project and task management app. It has kanban boards, lists, timeline views, cycles, milestones, custom fields, and automations. It is modelled on Linear and Asana but plugs directly into your business data from HR (who is assigned), CRM (customer projects), and Finance (billable work).",
  },
  {
    question: 'How is Flow different from Asana or Linear?',
    answer:
      'Flow does what Asana and Linear do — tasks, boards, projects, comments, automations — plus it reads from the rest of rtylr. You can assign a task to a staff member from HR, attach it to a customer from CRM, and track billable hours into an invoice in Finance without copy-pasting.',
  },
  {
    question: 'Does Flow have a free tier?',
    answer:
      'Yes. Teams under 3 active seats use Flow for free, including unlimited projects, tasks, comments, and attachments. Paid tiers add timeline view, custom fields, automations, and multi-project dashboards.',
  },
];

export const HR_FAQS: FAQEntry[] = [
  {
    question: 'What does rtylr HR do?',
    answer:
      'rtylr HR handles employee records, scheduling, shift swaps, time-off requests, payroll runs, and compliance documents. It supports restaurant-specific workflows (tip pooling, minor-work-hour limits, split shifts) alongside general retail and services HR.',
  },
  {
    question: 'Does rtylr HR run payroll for the US?',
    answer:
      'rtylr HR calculates gross-to-net pay, taxes, and withholdings, and exports pay files compatible with Gusto, ADP, and Rippling for filing. Direct filing through rtylr is on the roadmap; for now rtylr handles the calculations and the third-party provider handles the IRS filings.',
  },
  {
    question: 'Can staff see their schedule on their phone?',
    answer:
      'Yes. Every staff member gets a mobile-friendly view of their schedule, time-off balance, and shift swap requests. They can request swaps and submit time-off from their phone without needing a staff account in the admin.',
  },
];

export const CRM_FAQS: FAQEntry[] = [
  {
    question: 'What does rtylr CRM track?',
    answer:
      'rtylr CRM tracks customers, visits, purchases, loyalty points, and campaign touchpoints. For restaurants and retail, it automatically captures who ordered what from the POS. For services businesses, you can log contacts, notes, and deal stages manually.',
  },
  {
    question: 'Does rtylr CRM have email marketing?',
    answer:
      'Yes. You can segment customers by purchase history, loyalty tier, or custom tags, then send branded email campaigns. Open and click tracking are built in. SMS campaigns are available on the Pro plan via Twilio integration.',
  },
];

export const FINANCE_FAQS: FAQEntry[] = [
  {
    question: 'What does rtylr Finance handle?',
    answer:
      'rtylr Finance handles invoicing, expenses, bill payments, bank reconciliation, and P&L / balance sheet / cash flow reports. Revenue from POS, Stripe, and other integrations flows in automatically; expenses can be imported from bank feeds or typed in manually.',
  },
  {
    question: 'Can rtylr Finance replace QuickBooks?',
    answer:
      'For most small businesses, yes. rtylr Finance covers invoicing, expenses, accounting, P&L, and tax reports. Businesses with complex audit or multi-entity consolidation needs may still prefer QuickBooks or Xero; rtylr Finance can export to both.',
  },
  {
    question: 'Does rtylr Finance support multiple currencies?',
    answer:
      'Yes. You can invoice in any currency, and rtylr handles the conversion using the exchange rate at invoice time. P&L reports can be viewed in any reporting currency with FX lines shown explicitly.',
  },
];

export const ERP_FAQS: FAQEntry[] = [
  {
    question: 'What does rtylr ERP manage?',
    answer:
      'rtylr ERP covers inventory, suppliers, purchase orders, stock transfers across locations, low-stock alerts, and cost tracking. It reads item sales from POS automatically, so on-hand stock stays accurate without a nightly sync.',
  },
  {
    question: 'Does rtylr ERP support multi-warehouse inventory?',
    answer:
      'Yes. Every item can be stocked at multiple locations with per-location counts, reorder thresholds, and transfer tracking. Purchase orders can be split across warehouses, and staff see only their assigned location by default.',
  },
];
