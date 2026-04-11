export const store = {
  title: 'Store (E-Commerce)',
  body: `
<p>The rtylr Store app lets you create and manage an online storefront. Products, inventory, and customers are shared with POS, ERP, and CRM so everything stays in sync.</p>

<h2 id="product-catalog">Product Catalog</h2>
<p>Your online product catalog is populated from ERP. Choose which products are published to the storefront, set online-specific pricing (or use a price list), write product descriptions, and upload images. Product variants (size, color) are displayed with option selectors.</p>

<h2 id="categories">Categories</h2>
<p>Organize your storefront with categories and subcategories. Categories from ERP can be reused or you can create store-specific collections. Each category has a name, description, image, and display order. Featured categories can be highlighted on the homepage.</p>

<h2 id="orders">Orders</h2>
<p>Online orders flow into a unified order management view. Each order shows the customer, items, shipping method, payment status, and fulfillment status. Orders progress through statuses: <code>pending</code>, <code>confirmed</code>, <code>processing</code>, <code>shipped</code>, <code>delivered</code>, <code>cancelled</code>, and <code>refunded</code>. Order data syncs to Finance for revenue tracking and to ERP for inventory deductions.</p>

<h2 id="shipping">Shipping Zones &amp; Methods</h2>
<p>Define shipping zones by country or region. Each zone can have multiple shipping methods (e.g., standard, express, free over a threshold). Shipping rates can be flat, weight-based, or price-based. Local pickup is also supported as a delivery method.</p>

<h2 id="payment-methods">Payment Methods</h2>
<p>Configure payment methods for your storefront. Supported options include credit/debit cards (via Stripe or similar providers), bank transfers, cash on delivery, and digital wallets. Payment settings are managed in <strong>Store &gt; Settings &gt; Payments</strong>.</p>

<h2 id="storefront-customization">Storefront Customization</h2>
<p>Customize the look and feel of your online store:</p>
<ul>
  <li><strong>Theme</strong> - Choose from built-in themes or customize colors, fonts, and layout</li>
  <li><strong>Branding</strong> - Upload your logo, favicon, and set brand colors that apply across the storefront</li>
  <li><strong>Domain</strong> - Use a free <code>yourstore.rtylr.com</code> subdomain or connect a custom domain</li>
  <li><strong>Pages</strong> - Create custom pages (About, Contact, Policies) with a rich text editor</li>
  <li><strong>Navigation</strong> - Configure header and footer menus with links to categories, pages, or external URLs</li>
</ul>

<h2 id="settings">Settings</h2>
<p>Store settings cover operational configuration:</p>
<ul>
  <li><strong>Currency</strong> - Set the display currency for your storefront (multi-currency support available as an addon)</li>
  <li><strong>Tax</strong> - Configure tax rates per region or use automatic tax calculation</li>
  <li><strong>Inventory</strong> - Choose whether to track stock, allow backorders, and set low stock thresholds for online display</li>
  <li><strong>Checkout</strong> - Enable guest checkout, require account creation, or offer both</li>
  <li><strong>Notifications</strong> - Configure order confirmation, shipping, and delivery notification emails</li>
</ul>
  `.trim(),
};
