export const finance = {
  title: 'Finance',
  body: `
<p>The rtylr Finance app provides full accounting capabilities. It connects to POS sales, ERP purchases, and HR payroll to give you a complete financial picture without manual data entry.</p>

<h2 id="invoices">Invoices</h2>
<p>Create professional invoices and send them to customers via email. Invoices support line items, taxes, discounts, and notes. Track invoice status: <code>draft</code>, <code>sent</code>, <code>viewed</code>, <code>paid</code>, <code>overdue</code>, and <code>void</code>. Customers can pay invoices online through a payment link.</p>

<h2 id="recurring-invoices">Recurring Invoices</h2>
<p>Set up invoices that generate automatically on a schedule (weekly, monthly, quarterly, or custom). Recurring invoices use a template and create a new invoice each period. They can be configured to auto-send on creation.</p>

<h2 id="quotes">Quotes</h2>
<p>Create and send quotes (estimates) to customers. Quotes can be converted to invoices with one click when the customer accepts. Track quote status: <code>draft</code>, <code>sent</code>, <code>accepted</code>, <code>declined</code>, and <code>expired</code>.</p>

<h2 id="expenses">Expenses</h2>
<p>Record business expenses with categories, vendors, dates, and receipt uploads. Expenses can be created manually, imported from bank feeds, or generated from ERP purchase orders. Recurring expenses can be automated.</p>

<h2 id="bills">Bills</h2>
<p>Track bills from suppliers and vendors. Bills are linked to purchase orders from ERP when applicable. Schedule bill payments and track aging (current, 30 days, 60 days, 90+ days overdue).</p>

<h2 id="chart-of-accounts">Chart of Accounts</h2>
<p>Maintain a customizable chart of accounts organized by type: assets, liabilities, equity, revenue, cost of goods sold, and expenses. Default accounts are created automatically and mapped to POS, ERP, and HR transactions.</p>

<h2 id="journal-entries">Journal Entries</h2>
<p>Create manual journal entries for adjustments, corrections, or transactions that fall outside automated flows. Each entry must balance (debits equal credits). Journal entries support attachments and notes for audit purposes.</p>

<h2 id="profit-and-loss">Profit &amp; Loss (P&amp;L)</h2>
<p>View your income statement for any date range. The P&amp;L report breaks down revenue by source (POS, Store, invoices), cost of goods sold (from ERP), and operating expenses. Compare periods side by side to track trends.</p>

<h2 id="balance-sheet">Balance Sheet</h2>
<p>View your financial position at any point in time. The balance sheet shows assets (cash, inventory, receivables), liabilities (payables, loans), and equity. Inventory valuation is pulled from ERP.</p>

<h2 id="cash-flow">Cash Flow</h2>
<p>The cash flow statement tracks money in and out of the business across operating, investing, and financing activities. It helps forecast future cash needs and identify trends in cash generation.</p>

<h2 id="tax-reports">Tax Reports</h2>
<p>Generate tax reports based on your jurisdiction. Sales tax reports aggregate tax collected at POS and Store. VAT reports show input and output tax for filing. Reports can be exported for submission to tax authorities.</p>

<h2 id="bank-feeds">Bank Feeds</h2>
<p>Connect your bank accounts to automatically import transactions. Bank feeds support major banking providers and update daily. Imported transactions appear in an inbox for categorization and matching.</p>

<h2 id="reconciliation">Reconciliation</h2>
<p>Match bank feed transactions to recorded entries in Finance. The reconciliation view shows unmatched transactions and suggests matches based on amount, date, and description. Mark transactions as reconciled to verify your books against the bank statement.</p>
  `.trim(),
};
