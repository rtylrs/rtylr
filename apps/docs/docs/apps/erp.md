# ERP (Inventory & Supply Chain)

The rtylr ERP app manages products, inventory, suppliers, and the supply chain. It integrates directly with POS (stock deductions on sale), Store (online catalog sync), and Finance (cost of goods sold).

## Products

Each product has a name, SKU, barcode, description, category, price, cost, tax class, and unit of measure. Products can be simple (single item) or variable (with variants like size or color). Each variant has its own SKU, barcode, price, and stock level.

## Categories

Products are organized into a hierarchical category tree. Categories can be nested to any depth and are shared across POS, Store, and ERP views. Each category can have an image, description, and display order.

## Suppliers

Maintain a directory of suppliers with contact info, payment terms, lead times, and default currency. Each product can be linked to one or more suppliers with supplier-specific pricing and SKU mapping.

## Purchase Orders

Create purchase orders (POs) from scratch or from reorder suggestions. A PO includes the supplier, expected delivery date, line items with quantities and agreed prices, and notes. POs move through statuses: `draft`, `sent`, `partially_received`, `received`, and `closed`.

## Receiving Stock

When goods arrive, receive them against a purchase order. The receiving screen shows expected vs. actual quantities. Partial receives are supported. On receipt, stock levels update immediately and a cost entry is created in Finance.

## Stock Takes

Perform full or partial stock takes (inventory counts). Create a stock take session, assign staff, and count items by scanning barcodes or entering quantities manually. The system calculates discrepancies between expected and actual counts. Approved adjustments update stock levels and log the variance.

## Stock Transfers

Transfer stock between warehouses or branches. A transfer order specifies the source location, destination, and items with quantities. The sending location's stock decreases and the receiving location's stock increases when the transfer is confirmed.

## Low Stock Alerts

Set minimum stock thresholds per product and location. When stock falls below the threshold, an alert is triggered. Alerts appear in Hub notifications and can be configured to send email or push notifications.

## Reorder Suggestions

Based on sales velocity, current stock levels, and supplier lead times, ERP generates reorder suggestions. Suggestions can be converted to purchase orders with one click.

## Price Lists

Create multiple price lists for different channels (retail, wholesale, online) or customer groups. Each price list can override the base price for any product. Price lists can be assigned to specific branches or customer segments.

## Expiry Tracking

Track batch numbers and expiry dates for perishable goods. The system alerts when items are approaching expiry and can be configured to prevent sale of expired stock at the POS.

## Inventory Valuation

View the total value of inventory across all locations. Valuation methods supported include FIFO (first in, first out), weighted average cost, and specific identification. Valuation reports are available by product, category, or location.

## Export & Import CSV

Bulk import products, stock levels, and supplier data from CSV files. Export any product list, stock report, or purchase order history to CSV for external analysis or migration.
