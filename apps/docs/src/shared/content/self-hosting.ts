import type { DocContent } from '../content';

export const selfHostingOverview: DocContent = {
  title: 'Self-Hosting Overview',
  body: `
<p>rtylr can run on infrastructure you own. The self-host package is an installer and runtime orchestrator &mdash; it contains no product source code. The source repositories build public Docker images, and the self-host package wires those images together for a customer-owned deployment via two supported install paths: a <strong>Helm chart</strong> (primary) and <strong>Docker Compose</strong>.</p>

<h2 id="runtime-shape">Runtime Shape</h2>
<pre><code>browser
  |
  v
ingress / caddy proxy
  |-- auth | dash | pos | erp | hr | crm | finance      (frontends, :8080)
  |   flow | insights | recruit | menu
  |-- api:9999
  |-- upload:8080

api / worker / upload
  |-- MySQL                 (bundled local OR bring-your-own)
  |-- Redis                 (bundled local OR bring-your-own)
  |-- floci object storage  (bundled local OR bring-your-own S3-compatible)
  |
  +--&gt; https://license.voxire.com   (ONLY external dependency)</code></pre>
<p>The upload service is intentionally separate from the API service. Uploads route to the upload service (the self-host analog of <code>cdn.voxire.com</code>, reached via the configurable <code>UPLOAD_URL</code>); the API only stores and reads media metadata.</p>

<h2 id="self-sustaining-stack">Self-Sustaining Stack</h2>
<p>The stack is self-contained by default. You may bring your own MySQL, Redis, and S3-compatible object storage by pointing <code>DB_HOST</code>, <code>REDIS_URL</code>, and <code>S3_*</code> at your services. If you do not, the stack spins up local MySQL, Redis, and <strong>floci</strong> (local S3-compatible object storage).</p>
<ul>
  <li><strong>Compose</strong>: local infra is gated behind the <code>local-infra</code> profile.</li>
  <li><strong>Helm</strong>: local infra is gated behind <code>localInfra.{mysql,redis,floci}.enabled</code>.</li>
</ul>
<p>The only thing the stack ever reaches on the public internet is the hosted license authority <code>https://license.voxire.com</code>.</p>

<h2 id="configuration">Configuration</h2>
<p>Every app URL is configurable (<code>AUTH_URL</code>, <code>DASH_URL</code>, <code>POS_URL</code>, <code>ERP_URL</code>, <code>HR_URL</code>, <code>CRM_URL</code>, <code>FINANCE_URL</code>, <code>FLOW_URL</code>, <code>INSIGHTS_URL</code>, <code>RECRUIT_URL</code>, <code>MENU_URL</code>, <code>API_URL</code>, <code>UPLOAD_URL</code>). The backend derives <code>CORS_ALLOWED_ORIGINS</code> from the configured frontend URLs. No hosted domains are hardcoded in any image.</p>

<h2 id="further-reading">Further Reading</h2>
<ul>
  <li><a href="/docs/self-hosting-license">License Flow</a> &mdash; how the license key is validated</li>
  <li><a href="/docs/self-hosting-images">Image Contract</a> &mdash; the image set and contract</li>
  <li><a href="/docs/self-hosting-helm">Helm Install</a> &mdash; the primary install path</li>
  <li><a href="/docs/self-hosting-compose">Docker Compose</a> &mdash; the Compose reference</li>
  <li><a href="/docs/self-hosting-dns">DNS Setup</a> &mdash; the records to create</li>
  <li><a href="/docs/self-hosting-env">Backend Env Reference</a> &mdash; backend configuration</li>
  <li><a href="/docs/self-hosting-upgrade">Upgrade Flow</a> &mdash; version pinning and upgrades</li>
</ul>
  `.trim(),
};

export const selfHostingLicense: DocContent = {
  title: 'Self-Hosting: License Flow',
  body: `
<p>Self-hosted rtylr requires a license key issued by Voxire sales.</p>

<h2 id="how-it-works">How it works</h2>
<ol>
  <li>You receive a license key from Voxire sales.</li>
  <li>You set it as <code>RTYLR_LICENSE_KEY</code> (in <code>.env</code> for Compose, or <code>license.key</code> / <code>license.existingSecret</code> for Helm).</li>
  <li>The backend <code>api</code> service validates the key against the hosted authority <code>https://license.voxire.com</code>.</li>
</ol>
<pre><code>api  --(validate RTYLR_LICENSE_KEY)--&gt;  https://license.voxire.com</code></pre>

<h2 id="important">Important</h2>
<ul>
  <li>The authority URL <code>https://license.voxire.com</code> is <strong>hardcoded in the backend</strong> and is <strong>not</strong> customer-configurable. There is no <code>LICENSE_AUTHORITY_URL</code> setting &mdash; it does not exist.</li>
  <li>There is <strong>no local license service</strong> and <strong>no <code>rtylr/license</code> image</strong>. Customers do not run a license runmode.</li>
  <li><code>https://license.voxire.com</code> is the <strong>only</strong> external dependency of the entire self-host stack. Everything else (MySQL, Redis, object storage) can run locally or on your own infrastructure.</li>
</ul>

<h2 id="network-requirements">Network requirements</h2>
<p>The host running the <code>api</code> service must be able to reach <code>https://license.voxire.com</code> over HTTPS (port 443). If your environment uses an egress proxy or firewall, allowlist that host.</p>

<h2 id="setting-the-key">Setting the key</h2>
<p>Compose (<code>.env</code>):</p>
<pre><code>RTYLR_LICENSE_KEY=your-key-here
DEPLOYMENT_MODE=self_hosted</code></pre>
<p>Helm:</p>
<pre><code>helm upgrade --install rtylr ./helm/rtylr \\
  --set-string license.key=your-key-here</code></pre>
<p>Or reference an existing Kubernetes Secret with a <code>RTYLR_LICENSE_KEY</code> key:</p>
<pre><code>helm upgrade --install rtylr ./helm/rtylr \\
  --set license.existingSecret=rtylr-license</code></pre>
  `.trim(),
};

export const selfHostingImages: DocContent = {
  title: 'Self-Hosting: Image Contract',
  body: `
<p>Images are published to Docker Hub under the namespace configured by:</p>
<pre><code>RTYLR_IMAGE_REGISTRY=docker.io
RTYLR_IMAGE_NAMESPACE=rtylr
RTYLR_VERSION=1.0.0</code></pre>
<p>A single <code>RTYLR_VERSION</code> pins ALL images (backend and frontends) to the same tag. Images are tagged by version &mdash; not just <code>latest</code>.</p>

<h2 id="backend-images">Backend images</h2>
<table>
  <thead>
    <tr><th>Image</th><th>Port</th><th>Notes</th></tr>
  </thead>
  <tbody>
    <tr><td><code>rtylr/api</code></td><td>9999</td><td>Validates <code>RTYLR_LICENSE_KEY</code> against the hosted authority</td></tr>
    <tr><td><code>rtylr/worker</code></td><td>&mdash;</td><td>Background worker; no HTTP</td></tr>
    <tr><td><code>rtylr/upload</code></td><td>8080</td><td>Media/CDN service; self-host analog of cdn.voxire.com</td></tr>
  </tbody>
</table>

<h2 id="frontend-images">Frontend images (static nginx on :8080)</h2>
<p>There are 11 frontend images: <code>rtylr/auth</code>, <code>rtylr/dash</code>, <code>rtylr/pos</code>, <code>rtylr/erp</code>, <code>rtylr/hr</code>, <code>rtylr/crm</code>, <code>rtylr/finance</code>, <code>rtylr/flow</code>, <code>rtylr/insights</code>, <code>rtylr/recruit</code>, <code>rtylr/menu</code>.</p>
<p>Each frontend serves the compiled application on port <code>8080</code> and reads runtime configuration (<code>API_URL</code>, <code>UPLOAD_URL</code>, <code>AUTH_URL</code>, <code>DASH_URL</code>) from environment variables at container start, so you can change URLs without rebuilding.</p>

<h2 id="not-shipped">Not part of the self-host image set</h2>
<ul>
  <li>There is <strong>no <code>rtylr/license</code> image</strong>. Customers do not run a license runmode. The backend <code>api</code> validates <code>RTYLR_LICENSE_KEY</code> against the hosted authority <code>https://license.voxire.com</code>.</li>
  <li>There is <strong>no <code>docs</code> image</strong> in the self-host set.</li>
  <li>The marketing <code>landing</code> site is <strong>not</strong> shipped for self-host.</li>
  <li>The shared <code>packages</code> are build-time dependencies only, consumed by the app image build pipelines &mdash; never mounted into this runtime.</li>
</ul>
  `.trim(),
};

export const selfHostingHelm: DocContent = {
  title: 'Self-Hosting: Helm Install',
  body: `
<p>The Helm chart is the primary install path. It lives at <code>helm/rtylr/</code> in the self-host package and deploys the backend (<code>api</code>, <code>worker</code>, <code>upload</code>) and all frontends as Deployments + Services behind a single Ingress, plus a ConfigMap/Secret for environment.</p>

<h2 id="prerequisites">Prerequisites</h2>
<ul>
  <li>A Kubernetes cluster and a <code>kubectl</code> context</li>
  <li>Helm 3.8+ (or Helm 4)</li>
  <li>An ingress controller (e.g. ingress-nginx) and DNS configured</li>
  <li>A license key from Voxire sales</li>
</ul>

<h2 id="minimal-install">Minimal install (bring your own datastores)</h2>
<pre><code>helm upgrade --install rtylr ./helm/rtylr \\
  --set image.tag=1.0.0 \\
  --set-string license.key=YOUR_LICENSE_KEY \\
  --set-string secrets.jwtSecret=$(openssl rand -hex 32) \\
  --set ingress.domain=example.com \\
  --set database.host=mysql.internal --set-string database.password=... \\
  --set redis.url=redis://redis.internal:6379 \\
  --set s3.endpoint=https://s3.amazonaws.com \\
  --set s3.bucket=rtylr --set s3.region=eu-west-1 \\
  --set-string s3.accessKeyId=... --set-string s3.secretAccessKey=...</code></pre>

<h2 id="self-contained-install">Self-contained install (bundled datastores, evaluation)</h2>
<pre><code>helm upgrade --install rtylr ./helm/rtylr \\
  --set image.tag=1.0.0 \\
  --set-string license.key=YOUR_LICENSE_KEY \\
  --set-string secrets.jwtSecret=$(openssl rand -hex 32) \\
  --set ingress.domain=example.com \\
  --set localInfra.mysql.enabled=true --set-string localInfra.mysql.rootPassword=$(openssl rand -hex 16) \\
  --set localInfra.redis.enabled=true \\
  --set localInfra.floci.enabled=true \\
  --set-string database.password=$(openssl rand -hex 16) \\
  --set-string s3.secretAccessKey=$(openssl rand -hex 16)</code></pre>
<p>The bundled datastores are single-replica and intended for evaluation. Use a managed/HA database, cache, and object store for production.</p>

<h2 id="key-values">Key values</h2>
<table>
  <thead>
    <tr><th>Value</th><th>Purpose</th></tr>
  </thead>
  <tbody>
    <tr><td><code>image.registry</code> / <code>image.namespace</code> / <code>image.tag</code></td><td>Image source + global version</td></tr>
    <tr><td><code>license.key</code> / <code>license.existingSecret</code></td><td>License key</td></tr>
    <tr><td><code>deploymentMode</code></td><td><code>self_hosted</code></td></tr>
    <tr><td><code>urls.*</code></td><td>Configurable app URLs</td></tr>
    <tr><td><code>secrets.jwtSecret</code> / <code>secrets.existingSecret</code></td><td>JWT secret (required)</td></tr>
    <tr><td><code>database.*</code> / <code>redis.url</code> / <code>s3.*</code></td><td>External datastore config</td></tr>
    <tr><td><code>extraEnv</code></td><td>Plain passthrough env applied to backend (ConfigMap)</td></tr>
    <tr><td><code>extraSecretEnv</code></td><td>Secret passthrough env applied to backend (Secret)</td></tr>
    <tr><td><code>backend.&lt;svc&gt;.replicas</code> / <code>frontends.&lt;app&gt;.replicas</code></td><td>Replica counts</td></tr>
    <tr><td><code>frontends.&lt;app&gt;.enabled</code></td><td>Toggle individual frontends</td></tr>
    <tr><td><code>ingress.*</code></td><td>Ingress class, domain, annotations, TLS</td></tr>
    <tr><td><code>localInfra.{mysql,redis,floci}.enabled</code></td><td>Deploy bundled datastores</td></tr>
  </tbody>
</table>

<h3 id="passthrough-env-example">Passthrough env example</h3>
<pre><code>extraEnv:
  AXIOM_DATASET: rtylr-prod
  AI_PRIMARY_PROVIDER: openai
extraSecretEnv:
  OPENAI_API_KEY: sk-...
  RESEND_API_KEY: re_...</code></pre>

<h2 id="verify-the-chart">Verify the chart</h2>
<pre><code>helm lint helm/rtylr --set secrets.jwtSecret=$(openssl rand -hex 32)
helm template rtylr helm/rtylr --set secrets.jwtSecret=$(openssl rand -hex 32) | less</code></pre>

<h2 id="upgrade">Upgrade</h2>
<p>Bump <code>image.tag</code> (a single value pins all images) and re-run <code>helm upgrade --install</code>. See the <a href="/docs/self-hosting-upgrade">Upgrade Flow</a>.</p>
  `.trim(),
};

export const selfHostingCompose: DocContent = {
  title: 'Self-Hosting: Docker Compose',
  body: `
<p>Docker Compose is a supported install path (the Helm chart is the primary one). The stack is defined in <code>docker-compose.yaml</code> in the self-host package.</p>

<h2 id="prerequisites">Prerequisites</h2>
<ul>
  <li>Docker Engine / Docker Desktop with Compose v2</li>
  <li>A license key from Voxire sales</li>
</ul>

<h2 id="setup">Setup</h2>
<pre><code>./scripts/setup.sh</code></pre>
<p>This copies <code>.env.example</code> to <code>.env</code>, generates secrets (<code>JWT_SECRET</code>, <code>INTERNAL_TOKEN</code>, <code>CDN_API_KEY</code>, MySQL/S3 credentials), and prompts for <code>RTYLR_LICENSE_KEY</code>.</p>

<h2 id="start">Start</h2>
<p>Self-contained (bundles MySQL, Redis, floci via the <code>local-infra</code> profile):</p>
<pre><code>docker compose --profile local-infra --env-file .env up -d</code></pre>
<p>Bring your own MySQL/Redis/S3 (set <code>DB_HOST</code>, <code>REDIS_URL</code>, <code>S3_*</code> in <code>.env</code>, then omit the profile):</p>
<pre><code>docker compose --env-file .env up -d</code></pre>

<h2 id="profiles">Profiles</h2>
<table>
  <thead>
    <tr><th>Profile</th><th>Services added</th></tr>
  </thead>
  <tbody>
    <tr><td>(none)</td><td>caddy, api, worker, upload, all frontends</td></tr>
    <tr><td><code>local-infra</code></td><td>+ db (MySQL), redis, floci (S3), floci-init</td></tr>
  </tbody>
</table>
<p>When the <code>local-infra</code> profile is off, the backend points at whatever you set in <code>DB_HOST</code> / <code>REDIS_URL</code> / <code>S3_*</code>.</p>

<h2 id="services-and-ports">Services and ports</h2>
<table>
  <thead>
    <tr><th>Service</th><th>Image</th><th>Internal port</th></tr>
  </thead>
  <tbody>
    <tr><td>caddy</td><td>caddy:2-alpine</td><td>80 / 443 (published)</td></tr>
    <tr><td>api</td><td>rtylr/api</td><td>9999</td></tr>
    <tr><td>worker</td><td>rtylr/worker</td><td>&mdash;</td></tr>
    <tr><td>upload</td><td>rtylr/upload</td><td>8080</td></tr>
    <tr><td>auth/dash/pos/erp/hr/crm/finance/flow/insights/recruit/menu</td><td>rtylr/&lt;app&gt;</td><td>8080</td></tr>
    <tr><td>db</td><td>mysql:8.0 (local-infra)</td><td>3306</td></tr>
    <tr><td>redis</td><td>redis:7-alpine (local-infra)</td><td>6379</td></tr>
    <tr><td>floci</td><td>localstack/localstack:3 (local-infra)</td><td>4566</td></tr>
  </tbody>
</table>
<p>Routing is handled by Caddy off <code>RTYLR_DOMAIN</code> &mdash; see the <code>Caddyfile</code> and the <a href="/docs/self-hosting-dns">DNS Setup</a> guide.</p>

<h2 id="common-operations">Common operations</h2>
<pre><code>./scripts/doctor.sh          # validate environment + compose config
./scripts/update.sh 3.300.14 # bump RTYLR_VERSION, pull, restart
./scripts/backup.sh          # dump MySQL + archive floci storage
docker compose --env-file .env logs -f api
docker compose --profile local-infra --env-file .env down</code></pre>

<h2 id="validate-the-compose-file">Validate the compose file</h2>
<pre><code>docker compose --env-file .env config &gt;/dev/null
docker compose --profile local-infra --env-file .env config &gt;/dev/null</code></pre>
  `.trim(),
};

export const selfHostingDns: DocContent = {
  title: 'Self-Hosting: DNS Setup',
  body: `
<p>rtylr serves each app on its own subdomain of your base domain. With Compose the base domain is <code>RTYLR_DOMAIN</code>; with Helm it is <code>ingress.domain</code>.</p>

<h2 id="subdomains">Subdomains</h2>
<p>Create an A/AAAA (or CNAME) record for each of these, all pointing at the host/load balancer that fronts the stack:</p>
<table>
  <thead>
    <tr><th>Subdomain</th><th>Service</th></tr>
  </thead>
  <tbody>
    <tr><td><code>api</code></td><td>backend api</td></tr>
    <tr><td><code>upload</code></td><td>upload / media</td></tr>
    <tr><td><code>auth</code></td><td>auth frontend</td></tr>
    <tr><td><code>dash</code></td><td>dashboard</td></tr>
    <tr><td><code>pos</code></td><td>point of sale</td></tr>
    <tr><td><code>erp</code></td><td>erp</td></tr>
    <tr><td><code>hr</code></td><td>hr</td></tr>
    <tr><td><code>crm</code></td><td>crm</td></tr>
    <tr><td><code>finance</code></td><td>finance</td></tr>
    <tr><td><code>flow</code></td><td>flow</td></tr>
    <tr><td><code>insights</code></td><td>insights</td></tr>
    <tr><td><code>recruit</code></td><td>recruit</td></tr>
    <tr><td><code>menu</code></td><td>menu</td></tr>
  </tbody>
</table>
<p>For base domain <code>example.com</code>, that is <code>api.example.com</code>, <code>dash.example.com</code>, and so on. A wildcard <code>*.example.com</code> record also works and covers every app.</p>

<h2 id="local-development">Local development</h2>
<p>With <code>RTYLR_DOMAIN=localhost</code>, <code>*.localhost</code> resolves to <code>127.0.0.1</code> on most systems with no DNS configuration. Use <code>http://dash.localhost</code>, <code>http://api.localhost</code>, etc.</p>

<h2 id="tls">TLS</h2>
<ul>
  <li><strong>Compose</strong>: the bundled <code>Caddyfile</code> is HTTP-first. Remove <code>auto_https off</code> for automatic Let's Encrypt certificates on a real domain, or terminate TLS in a proxy/load balancer in front of the stack.</li>
  <li><strong>Helm</strong>: set <code>ingress.tls.enabled=true</code> and <code>ingress.tls.secretName</code>, and let cert-manager (or your ingress controller) provision certificates.</li>
</ul>
  `.trim(),
};

export const selfHostingEnv: DocContent = {
  title: 'Self-Hosting: Backend Env Reference',
  body: `
<p>These variables configure the backend services (<code>api</code>, <code>worker</code>, <code>upload</code>). Compose reads them from <code>.env</code>; Helm renders them into a ConfigMap (non-secret) and Secret (secret).</p>

<h2 id="license-and-mode">License &amp; mode</h2>
<table>
  <thead>
    <tr><th>Variable</th><th>Default</th><th>Notes</th></tr>
  </thead>
  <tbody>
    <tr><td><code>RTYLR_LICENSE_KEY</code></td><td>&mdash;</td><td>Sales-issued key, validated against <code>https://license.voxire.com</code></td></tr>
    <tr><td><code>DEPLOYMENT_MODE</code></td><td><code>self_hosted</code></td><td></td></tr>
    <tr><td><code>APP_ENV</code></td><td><code>production</code></td><td></td></tr>
  </tbody>
</table>
<p>There is no <code>LICENSE_AUTHORITY_URL</code> / <code>LICENSE_PUBLIC_KEY</code> / <code>LICENSE_KEY</code> &mdash; the authority is hardcoded. See the <a href="/docs/self-hosting-license">License Flow</a>.</p>

<h2 id="configurable-urls">Configurable URLs</h2>
<p><code>AUTH_URL</code>, <code>DASH_URL</code>, <code>POS_URL</code>, <code>ERP_URL</code>, <code>HR_URL</code>, <code>CRM_URL</code>, <code>FINANCE_URL</code>, <code>FLOW_URL</code>, <code>INSIGHTS_URL</code>, <code>RECRUIT_URL</code>, <code>MENU_URL</code>, <code>API_URL</code>, <code>UPLOAD_URL</code>.</p>
<p><code>CORS_ALLOWED_ORIGINS</code> is derived from the frontend URLs (Compose builds it in the compose file; Helm builds it in the ConfigMap).</p>

<h2 id="secrets">Secrets</h2>
<table>
  <thead>
    <tr><th>Variable</th><th>Notes</th></tr>
  </thead>
  <tbody>
    <tr><td><code>JWT_SECRET</code></td><td>Required; generated by <code>setup.sh</code> / set via <code>secrets.jwtSecret</code></td></tr>
    <tr><td><code>INTERNAL_TOKEN</code></td><td>Internal service-to-service token</td></tr>
    <tr><td><code>CDN_API_KEY</code></td><td>Upload service key</td></tr>
  </tbody>
</table>

<h2 id="database">Database</h2>
<table>
  <thead>
    <tr><th>Variable</th><th>Default</th></tr>
  </thead>
  <tbody>
    <tr><td><code>DB_HOST</code></td><td><code>db</code> (bundled) or your host</td></tr>
    <tr><td><code>DB_PORT</code></td><td><code>3306</code></td></tr>
    <tr><td><code>DB_NAME</code></td><td><code>rtylr_self_host</code></td></tr>
    <tr><td><code>DB_USER</code></td><td><code>rtylr_user</code></td></tr>
    <tr><td><code>DB_PASSWORD</code></td><td>(secret)</td></tr>
  </tbody>
</table>

<h2 id="cache">Cache</h2>
<table>
  <thead>
    <tr><th>Variable</th><th>Default</th></tr>
  </thead>
  <tbody>
    <tr><td><code>REDIS_URL</code></td><td><code>redis://redis:6379</code> (bundled) or your URL</td></tr>
  </tbody>
</table>

<h2 id="object-storage">Object storage (S3-compatible)</h2>
<table>
  <thead>
    <tr><th>Variable</th><th>Default</th><th>Notes</th></tr>
  </thead>
  <tbody>
    <tr><td><code>S3_BUCKET_NAME</code></td><td><code>rtylr-self-host</code></td><td></td></tr>
    <tr><td><code>S3_REGION</code></td><td><code>us-east-1</code></td><td></td></tr>
    <tr><td><code>S3_ENDPOINT</code></td><td><code>http://floci:4566</code></td><td>Your S3 endpoint when BYO</td></tr>
    <tr><td><code>S3_FORCE_PATH_STYLE</code></td><td><code>true</code></td><td>Needed for MinIO/floci</td></tr>
    <tr><td><code>AWS_ACCESS_KEY_ID</code></td><td><code>rtylr</code></td><td></td></tr>
    <tr><td><code>AWS_SECRET_ACCESS_KEY</code></td><td>(secret)</td><td></td></tr>
    <tr><td><code>CDN_MAX_UPLOAD_MB</code></td><td><code>100</code></td><td></td></tr>
  </tbody>
</table>

<h2 id="optional-integrations">Optional passthrough integrations</h2>
<p><code>SENTRY_DSN</code>, <code>AXIOM_TOKEN</code>, <code>AXIOM_DATASET</code>, <code>RESEND_API_KEY</code>, <code>OPENAI_API_KEY</code>, <code>OPENAI_ORG_ID</code>, <code>OPENAI_BASE_URL</code>, <code>GEMINI_API_KEY</code>, <code>DEEPSEEK_API_KEY</code>, <code>NVIDIA_NIM_API_KEY</code>, <code>AI_PRIMARY_PROVIDER</code>, <code>PADDLE_API_KEY</code>, <code>PADDLE_WEBHOOK_SECRET</code>, <code>PADDLE_CLIENT_TOKEN</code>, <code>PADDLE_SANDBOX</code>.</p>
<p>In Helm, add any additional variables through <code>extraEnv</code> (non-secret) and <code>extraSecretEnv</code> (secret) &mdash; both are applied to the backend services.</p>
  `.trim(),
};

export const selfHostingUpgrade: DocContent = {
  title: 'Self-Hosting: Upgrade Flow',
  body: `
<p>All rtylr images share a single version. Bump one value to upgrade the whole stack, Retool-style.</p>

<h2 id="the-version">The version</h2>
<p><code>RTYLR_VERSION</code> (e.g. <code>1.0.0</code>) pins every image &mdash; backend and frontends &mdash; to the same Docker tag. It lives in <code>versions.env</code> (and <code>.env</code> for Compose); the Helm equivalent is <code>image.tag</code>.</p>

<h2 id="compose">Compose</h2>
<pre><code>./scripts/update.sh 3.300.14</code></pre>
<p>This writes <code>RTYLR_VERSION=3.300.14</code> into <code>versions.env</code> and <code>.env</code>, then pulls the new images and restarts. Run without an argument to pull/restart the current pinned version. The script auto-detects whether the <code>local-infra</code> profile is in use.</p>
<p>Always back up first:</p>
<pre><code>./scripts/backup.sh</code></pre>

<h2 id="helm">Helm</h2>
<pre><code>helm upgrade --install rtylr ./helm/rtylr \\
  --reuse-values \\
  --set image.tag=3.300.14</code></pre>
<p>Or update <code>image.tag</code> in your values file and run <code>helm upgrade</code>.</p>

<h2 id="rollback">Rollback</h2>
<ul>
  <li><strong>Helm</strong>: <code>helm rollback rtylr</code> (or to a specific revision).</li>
  <li><strong>Compose</strong>: set <code>RTYLR_VERSION</code> back to the previous tag and run <code>./scripts/update.sh &lt;previous&gt;</code>.</li>
</ul>

<h2 id="notes">Notes</h2>
<ul>
  <li>Pin a specific version in production. Avoid <code>latest</code>.</li>
  <li>Database migrations (if any) are run by the <code>api</code>/<code>worker</code> images on start; take a backup before upgrading.</li>
</ul>
  `.trim(),
};
