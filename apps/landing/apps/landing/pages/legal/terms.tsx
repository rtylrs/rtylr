import React from 'react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { webPageSchema } from '../../lib/structured-data';
import styles from './legal.module.scss';

// ── Table of contents anchors ────────────────────────────────────────────
const SECTIONS = [
  { id: 'acceptance', label: 'Acceptance of Terms' },
  { id: 'account', label: 'Your Account' },
  { id: 'plans-billing', label: 'Plans and Billing' },
  { id: 'seat-proration', label: 'Per-Seat Billing and Proration' },
  { id: 'trial-refunds', label: 'Trials, Cancellation, Refunds' },
  { id: 'acceptable-use', label: 'Acceptable Use' },
  { id: 'data-ownership', label: 'Your Data' },
  { id: 'ip', label: 'Intellectual Property' },
  { id: 'third-party', label: 'Third-Party Services' },
  { id: 'disclaimers', label: 'Disclaimers' },
  { id: 'liability', label: 'Limitation of Liability' },
  { id: 'indemnification', label: 'Indemnification' },
  { id: 'termination', label: 'Termination' },
  { id: 'changes', label: 'Changes to These Terms' },
  { id: 'governing-law', label: 'Governing Law' },
  { id: 'contact', label: 'Contact' },
];

const TermsPage = (): JSX.Element => {
  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title="Terms of Service | rtylr"
          description="The terms that govern your use of rtylr, operated by Voxire."
          canonical="https://rtylr.com/legal/terms/"
          structuredData={webPageSchema({
            name: 'Terms of Service',
            path: '/legal/terms/',
            description: 'The terms that govern your use of rtylr, operated by Voxire.',
          })}
        />

        <main className={styles.page}>
          <div className={styles.container}>
            <header className={styles.header}>
              <span className={styles.eyebrow}>LEGAL · TERMS OF SERVICE</span>
              <h1 className={styles.title}>Terms of Service</h1>
              <p className={styles.meta}>
                Last updated: April 18, 2026 · Effective immediately for new customers, and 30 days
                from notification for existing customers.
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
                <h2 id="acceptance">Acceptance of Terms</h2>
                <p>
                  These Terms of Service (&ldquo;<strong>Terms</strong>&rdquo;) govern your access
                  to and use of the rtylr platform, applications, and related services
                  (collectively, &ldquo;<strong>rtylr</strong>&rdquo; or the &ldquo;
                  <strong>Service</strong>&rdquo;) operated by Voxire (&ldquo;<strong>we</strong>
                  &rdquo;, &ldquo;<strong>us</strong>&rdquo;). By creating an account, clicking
                  &ldquo;Get started&rdquo;, or otherwise using rtylr, you agree to these Terms on
                  behalf of yourself and any business, organization, or entity you represent.
                </p>
                <p>
                  If you do not agree to these Terms, do not use rtylr. If you are accepting these
                  Terms on behalf of a business, you represent that you have authority to bind that
                  business.
                </p>

                <h2 id="account">Your Account</h2>
                <p>
                  To use most features of rtylr you must create an account. You agree to provide
                  accurate, current, and complete information during registration and to keep that
                  information up to date. You are responsible for:
                </p>
                <ul>
                  <li>maintaining the confidentiality of your password and access tokens;</li>
                  <li>
                    any activity that happens under your account, including activity by team members
                    you add;
                  </li>
                  <li>
                    promptly notifying us at{' '}
                    <a href="mailto:security@voxire.com">security@voxire.com</a> of any unauthorized
                    access or breach.
                  </li>
                </ul>
                <p>
                  Accounts must be held by individuals at least 18 years old, or by legal entities
                  acting through an authorized representative.
                </p>

                <h2 id="plans-billing">Plans and Billing</h2>
                <p>
                  rtylr is offered on a Free plan and paid plans (Pro, Scale). Plan features,
                  limits, and prices are described at <a href="/pricing">rtylr.com/pricing</a>. By
                  subscribing to a paid plan, you authorize us (or our payment processor) to charge
                  the payment method on file for recurring fees, taxes, and any overages defined on
                  the pricing page.
                </p>
                <p>
                  Unless otherwise stated, all fees are exclusive of applicable taxes. We use Paddle
                  Inc. as the Merchant of Record for most transactions; Paddle&apos;s payment terms
                  may also apply to your purchase.
                </p>

                <h2 id="seat-proration">Per-Seat Billing and Proration</h2>
                <p>
                  Paid plans are billed per active seat. A &ldquo;seat&rdquo; is any user with login
                  access to your workspace. Seats are billed on a daily-proration basis:
                </p>
                <ul>
                  <li>
                    When you add a seat mid-cycle, you are charged only for the days remaining in
                    that billing period.
                  </li>
                  <li>
                    When you remove a seat mid-cycle, the remaining days of that seat are credited
                    against your next invoice.
                  </li>
                  <li>
                    Billing is summarised at the end of each period so you can see exactly which
                    seats were active on which days.
                  </li>
                </ul>
                <p>
                  On the Scale plan, tiered pricing applies automatically when your workspace
                  exceeds 30 active seats during the billing period.
                </p>

                <h2 id="trial-refunds">Trials, Cancellation, Refunds</h2>
                <p>
                  New workspaces may start on a trial of a paid plan. At trial end, the workspace
                  automatically reverts to the Free plan unless a payment method is added. You may
                  cancel at any time; cancellation takes effect at the end of the current billing
                  period.
                </p>
                <p>
                  Because rtylr is offered on a daily-prorated basis, unused days within the current
                  cycle are credited forward rather than refunded. We will issue a refund only where
                  required by applicable consumer-protection law.
                </p>

                <h2 id="acceptable-use">Acceptable Use</h2>
                <p>You will not:</p>
                <ul>
                  <li>
                    use rtylr to store or transmit unlawful content, malware, or material that
                    infringes third-party rights;
                  </li>
                  <li>
                    probe, scan, or test the vulnerability of rtylr or circumvent its security
                    controls, except under a published responsible-disclosure program;
                  </li>
                  <li>
                    use automated means to extract data from the Service beyond reasonable API usage
                    consistent with your plan;
                  </li>
                  <li>resell, sublicense, or use rtylr to build a competing product;</li>
                  <li>
                    use the Service to send spam, violate consumer privacy laws, or process payment
                    data outside PCI-compliant endpoints we provide.
                  </li>
                </ul>
                <p>
                  We reserve the right to suspend access to workspaces that we reasonably believe
                  are in breach of this Acceptable Use section.
                </p>

                <h2 id="data-ownership">Your Data</h2>
                <p>
                  You retain all rights to the data you upload, store, or generate in rtylr (&ldquo;
                  <strong>Customer Data</strong>&rdquo;). You grant Voxire a limited, worldwide,
                  royalty-free license to host, process, display, and transmit Customer Data solely
                  to provide the Service.
                </p>
                <p>
                  You can export your Customer Data at any time from your workspace settings. On
                  account termination, Customer Data is retained for 30 days to allow recovery, then
                  deleted from live systems within 60 days (backups are purged on their rolling
                  schedule, no later than 90 days).
                </p>
                <p>
                  How we handle personal data is described in our{' '}
                  <a href="/legal/privacy-policy">Privacy Policy</a>.
                </p>

                <h2 id="ip">Intellectual Property</h2>
                <p>
                  rtylr, including all software, designs, marks, and documentation, is owned by
                  Voxire and its licensors. These Terms grant you a limited, non-exclusive,
                  non-transferable right to use the Service as provided. No other rights are
                  granted.
                </p>
                <p>
                  If you provide feedback or suggestions about the Service, you grant us a
                  perpetual, royalty-free license to use that feedback to improve rtylr.
                </p>

                <h2 id="third-party">Third-Party Services</h2>
                <p>
                  rtylr integrates with third-party services (e.g. payment processors, accounting
                  exports, email delivery). Your use of those integrations is governed by their
                  terms. We are not responsible for the availability, content, or practices of
                  third-party services.
                </p>

                <h2 id="disclaimers">Disclaimers</h2>
                <p>
                  rtylr is provided &ldquo;<strong>as is</strong>&rdquo; and &ldquo;
                  <strong>as available</strong>&rdquo;. To the maximum extent permitted by law, we
                  disclaim all warranties, express or implied, including merchantability, fitness
                  for a particular purpose, and non-infringement. We do not warrant that the Service
                  will be uninterrupted, error-free, or that it will meet your specific
                  requirements.
                </p>

                <h2 id="liability">Limitation of Liability</h2>
                <p>
                  To the maximum extent permitted by law, Voxire and its officers, employees, and
                  agents are not liable for any indirect, incidental, special, consequential, or
                  punitive damages, including lost profits, lost revenue, lost data, or replacement
                  costs, arising out of or in connection with your use of rtylr.
                </p>
                <p>
                  Our total aggregate liability to you for any claim arising out of these Terms is
                  limited to the greater of (a) the fees you paid us in the twelve months preceding
                  the event giving rise to the claim, or (b) one hundred US dollars (USD $100).
                </p>

                <h2 id="indemnification">Indemnification</h2>
                <p>
                  You agree to defend, indemnify, and hold harmless Voxire from any claim arising
                  from your Customer Data, your breach of these Terms, or your violation of
                  applicable law, except where such claim arises from our gross negligence or
                  willful misconduct.
                </p>

                <h2 id="termination">Termination</h2>
                <p>
                  You may cancel your account at any time from workspace settings. We may suspend or
                  terminate your access if you materially breach these Terms, if required to comply
                  with law, or if your account is delinquent beyond 30 days. Sections that by their
                  nature should survive termination (payment obligations, limitations of liability,
                  governing law) will survive.
                </p>

                <h2 id="changes">Changes to These Terms</h2>
                <p>
                  We may update these Terms from time to time. When we do, we will update the
                  &ldquo;Last updated&rdquo; date and, for material changes, send notice to the
                  primary owner email on file. Continued use of rtylr after a change takes effect
                  constitutes acceptance of the updated Terms.
                </p>

                <h2 id="governing-law">Governing Law</h2>
                <p>
                  These Terms are governed by the laws of the jurisdiction where Voxire is
                  registered, without regard to conflict-of-laws principles. You and Voxire agree to
                  the exclusive jurisdiction of the competent courts of that jurisdiction for any
                  dispute arising under these Terms, except that either party may seek injunctive
                  relief in any court with proper jurisdiction to protect its intellectual property
                  or security.
                </p>

                <h2 id="contact">Contact</h2>
                <p>
                  Questions about these Terms? Email{' '}
                  <a href="mailto:info@voxire.com">info@voxire.com</a>. For security issues
                  specifically, use <a href="mailto:security@voxire.com">security@voxire.com</a>.
                </p>

                <div className={styles.contactStrip}>
                  <h3>Have a legal or compliance question?</h3>
                  <p>
                    We&apos;re happy to answer. Reach us at{' '}
                    <a href="mailto:info@voxire.com">info@voxire.com</a>, or via WhatsApp at +961 3
                    940 708.
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

export default TermsPage;
