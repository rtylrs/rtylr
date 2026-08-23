import React from 'react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import styles from './legal.module.scss';
import { webPageSchema } from '../../lib/structured-data';

const SECTIONS = [
  { id: 'overview', label: 'Overview' },
  { id: 'data-we-collect', label: 'Data We Collect' },
  { id: 'how-we-use', label: 'How We Use Data' },
  { id: 'legal-bases', label: 'Legal Bases (GDPR)' },
  { id: 'sharing', label: 'Sharing and Sub-Processors' },
  { id: 'international', label: 'International Transfers' },
  { id: 'retention', label: 'Retention and Deletion' },
  { id: 'your-rights', label: 'Your Rights' },
  { id: 'cookies', label: 'Cookies' },
  { id: 'children', label: 'Children' },
  { id: 'changes', label: 'Changes to This Policy' },
  { id: 'contact', label: 'Contact the DPO' },
];

const PrivacyPage = (): JSX.Element => {
  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title="Privacy Policy | rtylr"
          description="How rtylr and Voxire handle your personal data, under GDPR, CCPA, and equivalent laws."
          canonical="https://rtylr.com/legal/privacy-policy/"
          structuredData={webPageSchema({
            name: 'Privacy Policy',
            path: '/legal/privacy-policy/',
            description: 'How rtylr and Voxire handle your personal data.',
          })}
        />

        <main className={styles.page}>
          <div className={styles.container}>
            <header className={styles.header}>
              <span className={styles.eyebrow}>LEGAL · PRIVACY POLICY</span>
              <h1 className={styles.title}>Privacy Policy</h1>
              <p className={styles.meta}>
                Last updated: April 18, 2026 · Applies to rtylr.com and the rtylr platform.
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
                <h2 id="overview">Overview</h2>
                <p>
                  This Privacy Policy describes how Voxire (&ldquo;<strong>we</strong>&rdquo;)
                  collects, uses, shares, and protects personal data in connection with the rtylr
                  platform. We act as a <strong>data controller</strong> for personal data about our
                  customers and website visitors, and as a <strong>data processor</strong> for
                  Customer Data that our customers process in rtylr on behalf of their own
                  end-users.
                </p>
                <p>
                  If you are an end-user of a business that uses rtylr (for example, a diner whose
                  email is captured in a restaurant&apos;s CRM), the rtylr customer is the
                  controller of your data. Please contact that business first to exercise your
                  rights; we will help them do so.
                </p>

                <h2 id="data-we-collect">Data We Collect</h2>

                <h3>Account and billing data</h3>
                <p>
                  When you create a workspace we collect name, email, business name, role, country,
                  and a password hash. When you subscribe, our payment processor (Paddle) collects
                  billing address, tax identifiers, and card tokens. We do not store raw card
                  numbers on our servers.
                </p>

                <h3>Customer Data</h3>
                <p>
                  This is the data you upload or generate inside rtylr, menu items, invoices,
                  schedules, customer profiles, tasks. You control this data; we process it only to
                  provide the Service and as instructed by you.
                </p>

                <h3>Usage and device data</h3>
                <p>
                  We log IP address, browser/device type, referring URL, pages accessed, and
                  timestamps. We use this to operate, secure, and improve rtylr, and to generate
                  aggregate analytics.
                </p>

                <h3>Cookies and similar technologies</h3>
                <p>
                  See the <a href="#cookies">Cookies</a> section below for details.
                </p>

                <h2 id="how-we-use">How We Use Data</h2>
                <p>We use personal data to:</p>
                <ul>
                  <li>provide, secure, and maintain the Service;</li>
                  <li>process billing and send transactional emails;</li>
                  <li>respond to support requests and improve product quality;</li>
                  <li>detect and prevent fraud, abuse, and security incidents;</li>
                  <li>
                    send product announcements to the primary owner on a workspace (you can opt out
                    from your profile);
                  </li>
                  <li>comply with legal obligations and enforce our Terms.</li>
                </ul>

                <h2 id="legal-bases">Legal Bases (GDPR)</h2>
                <p>For EEA/UK residents, we rely on the following legal bases:</p>
                <ul>
                  <li>
                    <strong>Contract</strong>, to provide the Service and manage billing.
                  </li>
                  <li>
                    <strong>Legitimate interests</strong>, to secure the Service, detect abuse, and
                    improve features. We balance these interests against your rights.
                  </li>
                  <li>
                    <strong>Consent</strong>, for non-essential cookies and certain marketing
                    communications. You can withdraw consent at any time.
                  </li>
                  <li>
                    <strong>Legal obligation</strong>, when we are required to retain or disclose
                    data by law.
                  </li>
                </ul>

                <h2 id="sharing">Sharing and Sub-Processors</h2>
                <p>
                  We share personal data only with service providers (&ldquo;sub-processors&rdquo;)
                  that help us operate rtylr, under contracts that require them to protect the data.
                  Our primary sub-processors include:
                </p>
                <ul>
                  <li>
                    <strong>Cloudflare</strong>, CDN, DNS, DDoS protection.
                  </li>
                  <li>
                    <strong>Paddle</strong>, payment processing and merchant of record.
                  </li>
                  <li>
                    <strong>Firebase / Google Cloud</strong>, authentication infrastructure.
                  </li>
                  <li>
                    <strong>Resend / Postmark</strong>, transactional email delivery.
                  </li>
                  <li>
                    <strong>PostHog / Axiom</strong>, product analytics and error tracking (when
                    enabled).
                  </li>
                </ul>
                <p>
                  A current list is available on request at{' '}
                  <a href="mailto:privacy@voxire.com">privacy@voxire.com</a>. We notify customers
                  before adding a new sub-processor that has access to Customer Data.
                </p>
                <p>We do not sell personal data. We do not share personal data with advertisers.</p>

                <h2 id="international">International Transfers</h2>
                <p>
                  rtylr may process data outside your country of residence. Where we transfer data
                  from the EEA, UK, or Switzerland, we rely on the European Commission&apos;s
                  Standard Contractual Clauses (2021) or an equivalent mechanism.
                </p>

                <h2 id="retention">Retention and Deletion</h2>
                <ul>
                  <li>
                    Account data is retained for the life of the account plus 30 days for recovery,
                    then deleted within 60 days.
                  </li>
                  <li>
                    Billing records are retained for 7 years to comply with tax and accounting
                    rules.
                  </li>
                  <li>Security logs are retained for 90 days; audit logs for 12 months.</li>
                  <li>Encrypted backups are on a rolling schedule and purged within 90 days.</li>
                </ul>

                <h2 id="your-rights">Your Rights</h2>
                <p>Depending on where you live, you may have the right to:</p>
                <ul>
                  <li>access the personal data we hold about you;</li>
                  <li>correct inaccurate data;</li>
                  <li>request deletion of your data;</li>
                  <li>object to or restrict processing;</li>
                  <li>port your data to another service;</li>
                  <li>withdraw consent at any time (where processing relies on consent);</li>
                  <li>lodge a complaint with a supervisory authority.</li>
                </ul>
                <p>
                  To exercise any of these rights, email{' '}
                  <a href="mailto:privacy@voxire.com">privacy@voxire.com</a>. We respond within 30
                  days.
                </p>

                <h2 id="cookies">Cookies</h2>
                <p>We use three categories of cookies:</p>
                <ul>
                  <li>
                    <strong>Strictly necessary</strong>, authentication, load balancing, CSRF
                    protection. These are always on.
                  </li>
                  <li>
                    <strong>Preference</strong>, remember your language, theme, and workspace
                    selection.
                  </li>
                  <li>
                    <strong>Analytics</strong>, help us understand which pages work and which
                    don&apos;t. These only run with your consent (visible on first visit).
                  </li>
                </ul>
                <p>
                  You can clear cookies at any time from your browser. Doing so may log you out of
                  rtylr.
                </p>

                <h2 id="children">Children</h2>
                <p>
                  rtylr is not directed to children under 16. We do not knowingly collect personal
                  data from children. If you believe a child has registered, contact{' '}
                  <a href="mailto:privacy@voxire.com">privacy@voxire.com</a> and we will delete the
                  account.
                </p>

                <h2 id="changes">Changes to This Policy</h2>
                <p>
                  We may update this Privacy Policy to reflect changes to the Service or the law.
                  When we do, we update the &ldquo;Last updated&rdquo; date and, for material
                  changes, notify primary owners by email.
                </p>

                <h2 id="contact">Contact the DPO</h2>
                <p>
                  Our Data Protection contact is reachable at{' '}
                  <a href="mailto:privacy@voxire.com">privacy@voxire.com</a>. Postal mail: Voxire,
                  c/o Data Protection, Beirut, Lebanon.
                </p>

                <div className={styles.contactStrip}>
                  <h3>Data request, complaint, or just a question?</h3>
                  <p>
                    Email <a href="mailto:privacy@voxire.com">privacy@voxire.com</a>. We aim to
                    respond within 3 business days and fully resolve within 30 days.
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

export default PrivacyPage;
