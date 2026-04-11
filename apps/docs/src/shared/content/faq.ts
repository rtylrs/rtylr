export const faq = {
  title: 'Frequently Asked Questions',
  body: `
<p>Answers to common questions about rtylr.</p>

<h2 id="pricing">What are the pricing plans?</h2>
<p>rtylr offers tiered pricing based on features and usage. All plans include core apps (Hub, POS, ERP). Higher tiers unlock additional apps, more team members, and premium addons. Visit <strong>rtylr.com/pricing</strong> for current plans and pricing. All plans include a free trial period.</p>

<h2 id="data-privacy">How is my data protected?</h2>
<p>All data is encrypted in transit (TLS 1.3) and at rest (AES-256). Each organization's data is isolated at the database level. rtylr follows industry-standard security practices including regular penetration testing, SOC 2 compliance, and automated backups. You own your data and can export it at any time.</p>

<h2 id="payment-methods">What payment methods does POS support?</h2>
<p>POS supports cash, credit/debit cards (via integrated payment terminals), gift cards, store credit, and split payments. Specific payment terminal hardware compatibility depends on your region. Contact support for a list of supported devices.</p>

<h2 id="offline-mode">Does rtylr work offline?</h2>
<p>Yes. The POS app is designed to work offline. When the internet connection drops, orders and payments are stored locally and automatically sync when connectivity is restored. Other apps require an active connection but will gracefully handle intermittent connectivity.</p>

<h2 id="multi-language">Does rtylr support multiple languages?</h2>
<p>The admin interface is available in English with more languages in development. For customer-facing content (Store, booking pages, emails), the Multi-Language addon lets you translate content into any supported language. Right-to-left (RTL) layouts are supported.</p>

<h2 id="custom-domain">Can I use my own domain?</h2>
<p>Yes. With the Custom Domain addon, you can connect your own domain to your Store and booking pages. rtylr handles SSL certificate provisioning automatically. DNS configuration instructions are provided during setup.</p>

<h2 id="api-access">Is there an API?</h2>
<p>Yes. rtylr provides a RESTful API for programmatic access to your data. The API supports all core operations and uses bearer token authentication. See the <a href="/docs/api-overview">API Overview</a> for details. API access is available through the API Access addon or included in higher-tier plans.</p>

<h2 id="data-export">Can I export my data?</h2>
<p>Yes. Most apps support CSV and PDF export for individual reports and lists. The Data Export addon provides bulk export of all your organization's data in standard formats (CSV, JSON) for compliance or migration purposes.</p>

<h2 id="mobile-support">Is there a mobile app?</h2>
<p>rtylr is a responsive web application that works on tablets and mobile devices. The POS app is optimized for tablet-size screens. A dedicated mobile app for staff (shift clock-in, schedule viewing, task management) is on the roadmap.</p>

<h2 id="integrations">What third-party integrations are available?</h2>
<p>rtylr integrates with payment processors, accounting software, email providers, SMS gateways, and shipping carriers. Integrations are managed in <strong>Hub &gt; Settings &gt; Integrations</strong>. For custom integrations, use the API and webhook system.</p>

<h2 id="support">How do I get support?</h2>
<p>All plans include email support. The Priority Support addon provides faster response times and a dedicated account manager. Documentation, guides, and FAQs are available at <strong>docs.rtylr.com</strong>.</p>
  `.trim(),
};
