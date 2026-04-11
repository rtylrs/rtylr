export const apiOverview = {
  title: 'API Overview',
  body: `
<p>The rtylr API provides programmatic access to your organization's data and operations. Use it to build custom integrations, automate workflows, or extend rtylr's functionality.</p>

<h2 id="base-url">Base URL</h2>
<pre><code>https://api.rtylr.com/v1</code></pre>

<h2 id="authentication">Authentication</h2>
<p>All API requests require a Bearer token in the Authorization header:</p>
<pre><code>Authorization: Bearer YOUR_API_KEY</code></pre>
<p>Generate API keys in <strong>Hub &gt; Settings &gt; API Keys</strong>. Each key is scoped to specific apps and permissions.</p>

<h2 id="rate-limits">Rate Limits</h2>
<p>API requests are rate-limited per organization:</p>
<ul>
  <li><strong>Standard plan:</strong> 100 requests per minute</li>
  <li><strong>Pro plan:</strong> 500 requests per minute</li>
  <li><strong>Enterprise:</strong> Custom limits</li>
</ul>
<p>Rate limit headers are included in every response:</p>
<pre><code>X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1700000000</code></pre>

<h2 id="response-format">Response Format</h2>
<p>All responses are JSON. Successful responses include a <code>data</code> field:</p>
<pre><code>{
  "data": { ... },
  "meta": {
    "page": 1,
    "per_page": 25,
    "total": 150
  }
}</code></pre>

<h2 id="errors">Errors</h2>
<p>Error responses include a <code>code</code> and <code>message</code>:</p>
<pre><code>{
  "error": {
    "code": "not_found",
    "message": "The requested resource was not found."
  }
}</code></pre>

<h2 id="pagination">Pagination</h2>
<p>List endpoints support cursor-based pagination using <code>page</code> and <code>per_page</code> query parameters. The default page size is 25, with a maximum of 100.</p>
  `.trim(),
};
