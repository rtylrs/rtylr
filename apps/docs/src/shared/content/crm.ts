export const crm = {
  title: 'CRM (Customer Relations)',
  body: `
<p>The rtylr CRM app helps you build lasting customer relationships. It combines customer profiles, loyalty programs, marketing campaigns, and feedback into a single view.</p>

<h2 id="customer-profiles">Customer Profiles</h2>
<p>Each customer has a profile with contact information, order history, loyalty status, notes, tags, and a complete timeline of interactions. Profiles are created automatically when a customer makes a purchase at POS or places an order in Store, or they can be created manually.</p>

<h2 id="loyalty-program">Loyalty Program</h2>
<p>Configure a points-based loyalty program with tiers and rewards:</p>
<ul>
  <li><strong>Points</strong> - Customers earn points per currency unit spent. Points earning rate is configurable globally or per product category.</li>
  <li><strong>Tiers</strong> - Define tier levels (e.g., Bronze, Silver, Gold) with point thresholds. Higher tiers unlock better earning rates or exclusive rewards.</li>
  <li><strong>Rewards</strong> - Create a catalog of rewards redeemable with points: discounts, free items, store credit, or custom perks.</li>
</ul>
<p>Loyalty status is visible at POS during checkout. Customers can also view and redeem points through the Store.</p>

<h2 id="campaigns">Campaigns (Email &amp; SMS)</h2>
<p>Create and send marketing campaigns via email or SMS. The campaign builder includes:</p>
<ul>
  <li>Drag-and-drop email editor with templates</li>
  <li>SMS composer with character count and preview</li>
  <li>Audience selection from segments or manual lists</li>
  <li>Scheduling for future delivery</li>
  <li>A/B testing for subject lines and content</li>
  <li>Delivery, open, click, and conversion tracking</li>
</ul>

<h2 id="segments">Segments</h2>
<p>Create customer segments using filters based on purchase history, total spend, loyalty tier, location, tags, last visit date, and custom fields. Segments update dynamically and can be used as campaign audiences or for targeted promotions at POS.</p>

<h2 id="automation-builder">Automation Builder</h2>
<p>Build automated workflows triggered by customer events:</p>
<ul>
  <li>Welcome email after first purchase</li>
  <li>Birthday reward notification</li>
  <li>Win-back campaign after inactivity</li>
  <li>Loyalty tier upgrade notification</li>
  <li>Post-purchase feedback request</li>
</ul>
<p>Automations use a visual builder with triggers, conditions, delays, and actions.</p>

<h2 id="feedback">Feedback</h2>
<p>Collect customer feedback through post-purchase surveys, in-store QR codes, or email links. Feedback is attached to the customer profile and the relevant order. Aggregate feedback scores are tracked over time and visible in Insights dashboards.</p>

<h2 id="templates">Templates</h2>
<p>Create reusable email and SMS templates for campaigns, automations, and transactional messages (order confirmations, receipts, appointment reminders). Templates support dynamic placeholders for customer name, order details, loyalty balance, and more.</p>

<h2 id="customer-timeline">Customer Timeline</h2>
<p>The timeline on each customer profile shows a chronological feed of all interactions: purchases, loyalty point changes, campaign messages received, feedback submitted, support tickets, and manual notes added by staff.</p>

<h2 id="merge-import">Merge &amp; Import</h2>
<p>Merge duplicate customer records into a single profile, combining order history, loyalty points, and notes. Import customers from CSV files with field mapping. Export customer lists for external use.</p>
  `.trim(),
};
