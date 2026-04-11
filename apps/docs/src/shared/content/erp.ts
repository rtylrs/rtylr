export const erp = {
  title: 'ERP (Inventory & Supply Chain)',
  body: `
<p>The rtylr ERP app manages products, inventory, suppliers, and the supply chain. It integrates directly with POS (stock deductions on sale), Store (online catalog sync), and Finance (cost of goods sold).</p>

<h2 id="products">Products</h2>
<p>Each product has a name, SKU, barcode, description, category, price, cost, tax class, and unit of measure. Products can be simple (single item) or variable (with variants like size or color). Each variant has its own SKU, barcode, price, and stock level.</p>

<h2 id="categories">Categories</h2>
<p>Products are organized into a hierarchical category tree. Categories can be nested to any depth and are shared across POS, Store, and ERP views. Each category can have an image, description, and display order.</p>

<h2 id="suppliers">Suppliers</h2>
<p>Maintain a directory of suppliers with contact info, payment terms, lead times, and default currency. Each product can be linked to one or more suppliers with supplier-specific pricing and SKU mapping.</p>

<h2 id="purchase-orders">Purchase Orders</h2>
<p>Create purchase orders (POs) from scratch or from reorder suggestions. A PO includes the supplier, expected delivery date, line items with quantities and agreed prices, and notes. POs move through statuses: <code>draft</code>, <code>sent</code>, <code>partially_received</code>, <code>received</code>, and <code>closed</code>.</p>

<h2 id="receiving-stock">Receiving Stock</h2>
<p>When goods arrive, receive them against a purchase order. The receiving screen shows expected vs. actual quantities. Partial receives are supported. On receipt, stock levels update immediately and a cost entry is created in Finance.</p>

<h2 id="stock-takes">Stock Takes</h2>
<p>Perform full or partial stock takes (inventory counts). Create a stock take session, assign staff, and count items by scanning barcodes or entering quantities manually. The system calculates discrepancies between expected and actual counts. Approved adjustments update stock levels and log the variance.</p>

<h2 id="stock-transfers">Stock Transfers</h2>
<p>Transfer stock between warehouses or branches. A transfer order specifies the source location, destination, and items with quantities. The sending location's stock decreases and the receiving location's stock increases when the transfer is confirmed.</p>

<h2 id="low-stock-alerts">Low Stock Alerts</h2>
<p>Set minimum stock thresholds per product and location. When stock falls below the threshold, an alert is triggered. Alerts appear in Hub notifications and can be configured to send email or push notifications.</p>

<h2 id="reorder-suggestions">Reorder Suggestions</h2>
<p>Based on sales velocity, current stock levels, and supplier lead times, ERP generates reorder suggestions. Suggestions can be converted to purchase orders with one click.</p>

<h2 id="price-lists">Price Lists</h2>
<p>Create multiple price lists for different channels (retail, wholesale, online) or customer groups. Each price list can override the base price for any product. Price lists can be assigned to specific branches or customer segments.</p>

<h2 id="expiry-tracking">Expiry Tracking</h2>
<p>Track batch numbers and expiry dates for perishable goods. The system alerts when items are approaching expiry and can be configured to prevent sale of expired stock at the POS.</p>

<h2 id="inventory-valuation">Inventory Valuation</h2>
<p>View the total value of inventory across all locations. Valuation methods supported include FIFO (first in, first out), weighted average cost, and specific identification. Valuation reports are available by product, category, or location.</p>

<h2 id="export-import">Export &amp; Import CSV</h2>
<p>Bulk import products, stock levels, and supplier data from CSV files. Export any product list, stock report, or purchase order history to CSV for external analysis or migration.</p>
  `.trim(),
};
