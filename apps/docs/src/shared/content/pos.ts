export const pos = {
  title: 'POS (Point of Sale)',
  body: `
<p>The rtylr POS app is a full-featured point of sale system built for restaurants, retail, and service businesses. It supports multi-branch and multi-terminal deployments, offline operation, and deep integration with ERP, Finance, and CRM.</p>

<h2 id="multi-branch-multi-terminal">Multi-Branch &amp; Multi-Terminal Setup</h2>
<p>Each branch (location) can have multiple terminals. Terminals are registered in <strong>Hub &gt; Branches</strong> and assigned to a specific branch and floor. Each terminal maintains its own session state, cash drawer, and receipt printer configuration.</p>
<ul>
  <li>Terminals can be tablets, desktops, or dedicated POS hardware</li>
  <li>Floor plans let you map tables to physical layouts for dine-in</li>
  <li>Terminal settings (printer, drawer, default order type) are configured independently</li>
</ul>

<h2 id="employee-pin-login">Employee PIN Login</h2>
<p>Staff authenticate at the terminal using a numeric PIN. PINs are assigned per employee in HR and can be changed by the employee or a manager. The PIN screen supports fast user switching without logging out of the terminal session.</p>

<h2 id="order-flow">Order Flow</h2>
<p>POS supports three order types:</p>
<ul>
  <li><strong>Dine-in</strong> - Select a table from the floor plan, add items, send to kitchen, and close when the customer pays</li>
  <li><strong>Takeaway</strong> - Create an order without a table assignment, print a pickup receipt</li>
  <li><strong>Delivery</strong> - Attach a customer address, assign a driver, and track delivery status</li>
</ul>
<p>Orders move through statuses: <code>draft</code>, <code>sent</code>, <code>preparing</code>, <code>ready</code>, <code>served</code>, <code>paid</code>, and <code>closed</code>.</p>

<h2 id="modifiers-and-combos">Modifiers &amp; Combos</h2>
<p>Products can have modifier groups (e.g., size, toppings, cooking preference). Each modifier can carry an additional price. Combo meals bundle multiple items at a discounted price and can include modifier selections per component.</p>

<h2 id="split-payments">Split Payments</h2>
<p>An order can be split across multiple payment methods or divided evenly among guests. Supported split modes:</p>
<ul>
  <li>Split by amount (each guest pays a custom amount)</li>
  <li>Split evenly (total divided by number of guests)</li>
  <li>Split by item (assign items to individual guests)</li>
</ul>

<h2 id="payment-methods">Payment Methods</h2>
<p>POS accepts the following payment methods:</p>
<ul>
  <li><strong>Cash</strong> - With automatic change calculation</li>
  <li><strong>Card</strong> - Integrated with payment terminal hardware</li>
  <li><strong>Gift Card</strong> - Redeemable cards sold and managed within rtylr</li>
  <li><strong>Store Credit</strong> - Applied from CRM customer balance</li>
</ul>

<h2 id="tips">Tips</h2>
<p>Tips can be added before or after payment. Configurable tip presets (e.g., 10%, 15%, 20%) display on the payment screen. Tips are tracked per employee and appear in shift and payroll reports.</p>

<h2 id="hold-park-orders">Hold &amp; Park Orders</h2>
<p>Orders can be placed on hold (parked) and resumed later. Parked orders appear in a dedicated list and retain all items, modifiers, and customer info. This is useful for customers who need to step away or when juggling multiple orders.</p>

<h2 id="kds">Kitchen Display System (KDS)</h2>
<p>Orders sent from POS appear on the KDS in real time. Kitchen staff can view, prioritize, and mark items as prepared. The KDS groups items by station (e.g., grill, bar, dessert) and shows elapsed time for each order. Bump actions move orders to the next status.</p>

<h2 id="cash-drawer">Cash Drawer Management</h2>
<p>Each terminal tracks its cash drawer independently. Supported operations:</p>
<ul>
  <li><strong>Open drawer</strong> - Manual open with reason logging</li>
  <li><strong>Cash in / Cash out</strong> - Record petty cash additions or removals</li>
  <li><strong>Expected vs. actual</strong> - End-of-shift reconciliation</li>
</ul>

<h2 id="shift-management">Shift Open &amp; Close</h2>
<p>Each employee opens a shift when they start and closes it when they finish. Opening a shift records the starting cash float. Closing a shift requires counting the drawer and produces a summary with total sales, payment breakdown, tips, discounts, refunds, and cash variance.</p>

<h2 id="day-summary">Day Summary Reports</h2>
<p>The day summary aggregates all shifts for a given day at a branch. It includes gross sales, net sales, tax collected, discounts, refunds, payment method totals, top-selling items, and average order value.</p>

<h2 id="offline-mode">Offline Mode</h2>
<p>POS continues to operate when the internet connection drops. Orders, payments, and inventory changes are stored locally using Dexie (IndexedDB) and automatically sync to the server when connectivity is restored. Conflict resolution uses a last-write-wins strategy with server-side validation.</p>

<h2 id="barcode-scanning">Barcode Scanning</h2>
<p>POS supports USB and Bluetooth barcode scanners. Scanning a barcode instantly adds the matching product to the current order. If no product matches, the cashier is prompted to search manually or create a new product.</p>

<h2 id="refunds-voids">Refunds &amp; Voids</h2>
<ul>
  <li><strong>Void</strong> - Cancel an item or entire order before payment. Voided items are removed from the order and inventory is restored.</li>
  <li><strong>Refund</strong> - Return a paid item. Refunds can be full or partial and are issued to the original payment method or as store credit.</li>
</ul>

<h2 id="manager-override">Manager PIN Override</h2>
<p>Certain actions require manager authorization, including applying discounts above a threshold, processing refunds, voiding orders, opening the cash drawer without a sale, and reprinting closed receipts. A manager enters their PIN to approve the action, and the override is logged in the audit trail.</p>

<h2 id="audit-trail">Audit Trail</h2>
<p>Every action in POS is logged: order creation, item additions, price overrides, discounts, voids, refunds, payment captures, drawer opens, and shift events. The audit trail records the employee, timestamp, terminal, and action details. Logs are immutable and accessible to owners and admins in Hub.</p>
  `.trim(),
};
