# POS (Point of Sale)

The rtylr POS app is a full-featured point of sale system built for restaurants, retail, and service businesses. It supports multi-branch and multi-terminal deployments, offline operation, and deep integration with ERP, Finance, and CRM.

## Multi-Branch & Multi-Terminal Setup

Each branch (location) can have multiple terminals. Terminals are registered in **Hub > Branches** and assigned to a specific branch and floor. Each terminal maintains its own session state, cash drawer, and receipt printer configuration.

- Terminals can be tablets, desktops, or dedicated POS hardware
- Floor plans let you map tables to physical layouts for dine-in
- Terminal settings (printer, drawer, default order type) are configured independently

## Employee PIN Login

Staff authenticate at the terminal using a numeric PIN. PINs are assigned per employee in HR and can be changed by the employee or a manager. The PIN screen supports fast user switching without logging out of the terminal session.

## Order Flow

POS supports three order types:

- **Dine-in** — Select a table from the floor plan, add items, send to kitchen, and close when the customer pays
- **Takeaway** — Create an order without a table assignment, print a pickup receipt
- **Delivery** — Attach a customer address, assign a driver, and track delivery status

Orders move through statuses: `draft`, `sent`, `preparing`, `ready`, `served`, `paid`, and `closed`.

## Modifiers & Combos

Products can have modifier groups (e.g., size, toppings, cooking preference). Each modifier can carry an additional price. Combo meals bundle multiple items at a discounted price and can include modifier selections per component.

## Split Payments

An order can be split across multiple payment methods or divided evenly among guests. Supported split modes:

- Split by amount (each guest pays a custom amount)
- Split evenly (total divided by number of guests)
- Split by item (assign items to individual guests)

## Payment Methods

POS accepts the following payment methods:

- **Cash** — With automatic change calculation
- **Card** — Integrated with payment terminal hardware
- **Gift Card** — Redeemable cards sold and managed within rtylr
- **Store Credit** — Applied from CRM customer balance

## Tips

Tips can be added before or after payment. Configurable tip presets (e.g., 10%, 15%, 20%) display on the payment screen. Tips are tracked per employee and appear in shift and payroll reports.

## Hold & Park Orders

Orders can be placed on hold (parked) and resumed later. Parked orders appear in a dedicated list and retain all items, modifiers, and customer info. This is useful for customers who need to step away or when juggling multiple orders.

## Kitchen Display System (KDS)

Orders sent from POS appear on the KDS in real time. Kitchen staff can view, prioritize, and mark items as prepared. The KDS groups items by station (e.g., grill, bar, dessert) and shows elapsed time for each order. Bump actions move orders to the next status.

## Cash Drawer Management

Each terminal tracks its cash drawer independently. Supported operations:

- **Open drawer** — Manual open with reason logging
- **Cash in / Cash out** — Record petty cash additions or removals
- **Expected vs. actual** — End-of-shift reconciliation

## Shift Open & Close

Each employee opens a shift when they start and closes it when they finish. Opening a shift records the starting cash float. Closing a shift requires counting the drawer and produces a summary with total sales, payment breakdown, tips, discounts, refunds, and cash variance.

## Day Summary Reports

The day summary aggregates all shifts for a given day at a branch. It includes gross sales, net sales, tax collected, discounts, refunds, payment method totals, top-selling items, and average order value.

## Offline Mode

POS continues to operate when the internet connection drops. Orders, payments, and inventory changes are stored locally using Dexie (IndexedDB) and automatically sync to the server when connectivity is restored. Conflict resolution uses a last-write-wins strategy with server-side validation.

## Barcode Scanning

POS supports USB and Bluetooth barcode scanners. Scanning a barcode instantly adds the matching product to the current order. If no product matches, the cashier is prompted to search manually or create a new product.

## Refunds & Voids

- **Void** — Cancel an item or entire order before payment. Voided items are removed from the order and inventory is restored.
- **Refund** — Return a paid item. Refunds can be full or partial and are issued to the original payment method or as store credit.

## Manager PIN Override

Certain actions require manager authorization, including applying discounts above a threshold, processing refunds, voiding orders, opening the cash drawer without a sale, and reprinting closed receipts. A manager enters their PIN to approve the action, and the override is logged in the audit trail.

## Audit Trail

Every action in POS is logged: order creation, item additions, price overrides, discounts, voids, refunds, payment captures, drawer opens, and shift events. The audit trail records the employee, timestamp, terminal, and action details. Logs are immutable and accessible to owners and admins in Hub.
