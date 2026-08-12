# Store (E-Commerce)

The rtylr Store app lets you create and manage an online storefront. Products, inventory, and customers are shared with POS, ERP, and CRM so everything stays in sync.

## Product Catalog

Your online product catalog is populated from ERP. Choose which products are published to the storefront, set online-specific pricing (or use a price list), write product descriptions, and upload images. Product variants (size, color) are displayed with option selectors.

## Categories

Organize your storefront with categories and subcategories. Categories from ERP can be reused or you can create store-specific collections. Each category has a name, description, image, and display order. Featured categories can be highlighted on the homepage.

## Orders

Online orders flow into a unified order management view. Each order shows the customer, items, shipping method, payment status, and fulfillment status. Orders progress through statuses: `pending`, `confirmed`, `processing`, `shipped`, `delivered`, `cancelled`, and `refunded`. Order data syncs to Finance for revenue tracking and to ERP for inventory deductions.

## Shipping Zones & Methods

Define shipping zones by country or region. Each zone can have multiple shipping methods (e.g., standard, express, free over a threshold). Shipping rates can be flat, weight-based, or price-based. Local pickup is also supported as a delivery method.

## Payment Methods

Configure payment methods for your storefront. Supported options include credit/debit cards (via Stripe or similar providers), bank transfers, cash on delivery, and digital wallets. Payment settings are managed in **Store > Settings > Payments**.

## Storefront Customization

Customize the look and feel of your online store:

- **Theme** — Choose from built-in themes or customize colors, fonts, and layout
- **Branding** — Upload your logo, favicon, and set brand colors that apply across the storefront
- **Domain** — Use a free `yourstore.rtylr.com` subdomain or connect a custom domain
- **Pages** — Create custom pages (About, Contact, Policies) with a rich text editor
- **Navigation** — Configure header and footer menus with links to categories, pages, or external URLs

## Settings

Store settings cover operational configuration:

- **Currency** — Set the display currency for your storefront (multi-currency support available as an addon)
- **Tax** — Configure tax rates per region or use automatic tax calculation
- **Inventory** — Choose whether to track stock, allow backorders, and set low stock thresholds for online display
- **Checkout** — Enable guest checkout, require account creation, or offer both
- **Notifications** — Configure order confirmation, shipping, and delivery notification emails
