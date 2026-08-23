import React from 'react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import styles from './legal.module.scss';
import { webPageSchema } from '../../lib/structured-data';

const SECTIONS = [
  { id: 'program', label: 'Security Program' },
  { id: 'data-protection', label: 'Data Protection' },
  { id: 'infrastructure', label: 'Infrastructure' },
  { id: 'access-control', label: 'Access Control' },
  { id: 'application', label: 'Application Security' },
  { id: 'vulnerability', label: 'Vulnerability Management' },
  { id: 'monitoring', label: 'Monitoring and Incident Response' },
  { id: 'compliance', label: 'Compliance and Certifications' },
  { id: 'vendor', label: 'Vendor Management' },
  { id: 'disclosure', label: 'Responsible Disclosure' },
  { id: 'contact', label: 'Contact Security' },
];

const SecurityPage = (): JSX.Element => {
  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title="Security | rtylr"
          description="How rtylr protects your data, encryption, infrastructure, access controls, and incident response."
          canonical="https://rtylr.com/legal/security/"
          structuredData={webPageSchema({
            name: 'Security',
            path: '/legal/security/',
            description:
              'How rtylr protects your data, encryption, infrastructure, and incident response.',
          })}
        />

        <main className={styles.page}>
          <div className={styles.container}>
            <header className={styles.header}>
              <span className={styles.eyebrow}>LEGAL · SECURITY</span>
              <h1 className={styles.title}>Security at rtylr</h1>
              <p className={styles.meta}>
                Last updated: April 18, 2026 · Our security program is reviewed quarterly.
              </p>
            </header>

            <div className={styles.layout}>
              <nav className={styles.toc} aria-label="Section navigation">
                {/* Label for the nav, not a section of the document: as an
                    h4 it jumped straight from the h1. */}
                <p className={styles.tocTitle}>On this page</p>
                <ol>
                  {SECTIONS.map((s) => (
                    <li key={s.id}>
                      <a href={`#${s.id}`}>{s.label}</a>
                    </li>
                  ))}
                </ol>
              </nav>

              <article className={styles.body}>
                <h2 id="program">Security Program</h2>
                <p>
                  Voxire runs a formal information-security program aligned with the principles of
                  ISO 27001 and SOC 2 Type II. It is maintained by our engineering leadership and
                  reviewed quarterly. The program covers people, process, and technology controls
                  across development, operations, and customer support.
                </p>

                <h2 id="data-protection">Data Protection</h2>
                <h3>Encryption in transit</h3>
                <p>
                  All traffic between customer devices and rtylr is encrypted using TLS 1.2+ (strong
                  cipher suites, perfect forward secrecy, HSTS enabled). We reject connections from
                  legacy TLS versions.
                </p>

                <h3>Encryption at rest</h3>
                <p>
                  Customer Data is stored on encrypted volumes (AES-256). Database-level encryption
                  is enabled for all workspaces. Backups are also encrypted and stored in separate
                  geographical regions.
                </p>

                <h3>Secrets management</h3>
                <p>
                  API keys, tokens, and credentials are stored in a managed secrets vault, rotated
                  on a regular schedule, and scoped to the minimum permissions required.
                </p>

                <h2 id="infrastructure">Infrastructure</h2>
                <p>
                  rtylr runs on reputable cloud infrastructure with ISO 27001, SOC 2, PCI DSS, and
                  GDPR attestations. Our primary providers include Cloudflare (edge), Firebase /
                  Google Cloud (auth and managed services), and a managed database provider with
                  automated failover.
                </p>
                <ul>
                  <li>Production environments are isolated from staging and development.</li>
                  <li>Multi-region backups with point-in-time recovery up to 7 days.</li>
                  <li>Capacity planning and DDoS mitigation at the edge.</li>
                </ul>

                <h2 id="access-control">Access Control</h2>
                <p>
                  Access to production systems is restricted to a named group of engineers, gated
                  behind SSO with mandatory multi-factor authentication. We follow the principle of
                  least privilege: engineers receive only the access they need, and access is
                  reviewed at least quarterly.
                </p>
                <p>
                  All production access is logged with actor, action, and timestamp. Audit logs are
                  retained for 12 months.
                </p>

                <h2 id="application">Application Security</h2>
                <ul>
                  <li>
                    <strong>Authentication.</strong> Passwords are hashed with industry-standard
                    algorithms; MFA is available for all accounts; sessions use httpOnly, Secure,
                    SameSite cookies with short lifetimes.
                  </li>
                  <li>
                    <strong>Authorization.</strong> Role-based permissions enforced server-side on
                    every API call. No trust-the-client logic for sensitive actions.
                  </li>
                  <li>
                    <strong>Input validation.</strong> All external input is validated with
                    schema-level rules; dangerous sinks are avoided.
                  </li>
                  <li>
                    <strong>CSRF, XSS, SQLi.</strong> Prevented through standard framework
                    protections, parameterised queries, and strict content-security-policy headers.
                  </li>
                  <li>
                    <strong>Rate limiting.</strong> Per-IP and per-user rate limits on public
                    endpoints.
                  </li>
                </ul>

                <h2 id="vulnerability">Vulnerability Management</h2>
                <ul>
                  <li>Dependencies scanned on every build; critical CVEs patched within 7 days.</li>
                  <li>Container and image scanning in CI.</li>
                  <li>
                    Annual third-party penetration test; results and remediation evidence available
                    under NDA for prospective Scale customers.
                  </li>
                  <li>Static analysis (SAST) on the main codebase.</li>
                </ul>

                <h2 id="monitoring">Monitoring and Incident Response</h2>
                <p>
                  We maintain 24/7 monitoring and alerting across infrastructure, applications, and
                  authentication flows. Suspected incidents are triaged against a documented
                  incident-response runbook. Customers affected by a security incident are notified
                  in accordance with applicable law and, at minimum, within 72 hours of confirmation
                  for material breaches.
                </p>

                <h2 id="compliance">Compliance and Certifications</h2>
                <ul>
                  <li>
                    <strong>GDPR &amp; UK GDPR</strong>, we act as processor for Customer Data, with
                    DPA available on request.
                  </li>
                  <li>
                    <strong>CCPA/CPRA</strong>, honored for California residents; opt-out mechanisms
                    available.
                  </li>
                  <li>
                    <strong>PCI DSS</strong>, payment processing is offloaded to Paddle, a Level-1
                    certified merchant of record.
                  </li>
                  <li>
                    <strong>SOC 2 Type II</strong>, in progress; target completion Q4 2026.
                  </li>
                </ul>

                <h2 id="vendor">Vendor Management</h2>
                <p>
                  We vet sub-processors for security posture before onboarding, require
                  data-protection agreements, and review major vendors annually. A current list of
                  sub-processors is maintained and available via{' '}
                  <a href="mailto:security@voxire.com">security@voxire.com</a>.
                </p>

                <h2 id="disclosure">Responsible Disclosure</h2>
                <p>
                  We welcome reports from security researchers. If you believe you&apos;ve found a
                  vulnerability, please email{' '}
                  <a href="mailto:security@voxire.com">security@voxire.com</a> with:
                </p>
                <ul>
                  <li>a clear description of the issue and its impact;</li>
                  <li>steps to reproduce;</li>
                  <li>
                    any proof-of-concept material (please do not exfiltrate data beyond what is
                    needed to demonstrate the issue).
                  </li>
                </ul>
                <p>
                  We commit to acknowledging receipt within 48 hours, working in good faith on a
                  fix, and crediting you publicly once the issue is resolved (unless you prefer to
                  remain anonymous). We do not take legal action against researchers acting in good
                  faith under this policy.
                </p>

                <h2 id="contact">Contact Security</h2>
                <p>
                  For security questions, reports, or compliance requests, email{' '}
                  <a href="mailto:security@voxire.com">security@voxire.com</a>. For urgent issues,
                  WhatsApp +961 3 940 708.
                </p>

                <div className={styles.contactStrip}>
                  <h3>Need our SOC 2 letter, SIG, or DPA?</h3>
                  <p>
                    Email <a href="mailto:security@voxire.com">security@voxire.com</a> and
                    we&apos;ll send the current packet under NDA within 7 business days.
                  </p>
                </div>
              </article>
            </div>
          </div>
        </main>
      </MainLayout>
    </AppLayout>
  );
};

export default SecurityPage;
