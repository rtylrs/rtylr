import React from 'react';
import { motion } from 'motion/react';
import Link from 'next/link';
import { Rss } from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import styles from './changelog.module.scss';

// ── Changelog entries ────────────────────────────────────────────────────

type ChangeKind = 'new' | 'improved' | 'fixed' | 'breaking';

interface Change {
  kind: ChangeKind;
  text: string;
}

interface Entry {
  version: string;
  date: string;
  title: string;
  changes: Change[];
}

const KIND_LABEL: Record<ChangeKind, string> = {
  new: 'NEW',
  improved: 'IMPROVED',
  fixed: 'FIXED',
  breaking: 'BREAKING',
};

const ENTRIES: Entry[] = [
  {
    version: 'v4.12',
    date: 'Apr 18, 2026',
    title: 'Scale tier, tiered seat pricing, live seat timelines',
    changes: [
      {
        kind: 'new',
        text: 'Scale plan now auto-unlocks cheaper per-seat pricing when your workspace crosses 30 active seats in a billing period.',
      },
      {
        kind: 'new',
        text: 'Billing settings now shows a Google-style seat timeline so you can see exactly which seats were active on which days.',
      },
      {
        kind: 'improved',
        text: 'Invoice PDFs now break out per-seat proration line items for easier accountant review.',
      },
      {
        kind: 'fixed',
        text: 'Mid-cycle seat removals no longer briefly show the full-month charge.',
      },
    ],
  },
  {
    version: 'v4.11',
    date: 'Apr 04, 2026',
    title: 'Draggable Flow boards, CRM segment builder, rtylr AI in beta',
    changes: [
      {
        kind: 'new',
        text: 'Kanban cards in Flow are now drag-and-drop. Columns can be reordered. Touch support on iPad and Android.',
      },
      {
        kind: 'new',
        text: 'CRM segment builder with 24 attributes and saved-segment triggers for campaigns.',
      },
      {
        kind: 'new',
        text: 'rtylr AI, natural-language business setup. In beta for new workspaces.',
      },
      {
        kind: 'improved',
        text: 'POS offline mode now syncs 3× faster on reconnect with parallel batch uploads.',
      },
    ],
  },
  {
    version: 'v4.10',
    date: 'Mar 21, 2026',
    title: 'Parallel-locations HR payroll, finance exports for Sage',
    changes: [
      {
        kind: 'new',
        text: 'HR payroll can now run multi-location in parallel, cutting run time for chains from hours to minutes.',
      },
      { kind: 'new', text: 'Finance: Sage Intacct export alongside QuickBooks and Xero.' },
      {
        kind: 'improved',
        text: 'Schedule publishing now stages drafts before SMS notifications ship.',
      },
      { kind: 'fixed', text: 'Tip reconciliation rounding error for split checks on Safari iOS.' },
    ],
  },
  {
    version: 'v4.9',
    date: 'Mar 07, 2026',
    title: 'Integrations directory, webhook retry policies, audit-log export',
    changes: [
      {
        kind: 'new',
        text: 'New integrations directory, one place to see every service rtylr talks to.',
      },
      {
        kind: 'new',
        text: 'Webhook retries now follow exponential backoff up to 24h. Status badge on each endpoint.',
      },
      {
        kind: 'improved',
        text: 'Audit logs can be exported to CSV or pushed to an S3 bucket on a schedule.',
      },
    ],
  },
  {
    version: 'v4.8',
    date: 'Feb 21, 2026',
    title: 'ERP recipe trees, multi-currency for Finance',
    changes: [
      {
        kind: 'new',
        text: 'Recipe trees in ERP: ingredients deplete automatically on POS sale down to nested sub-recipes (sauces, doughs, mixes).',
      },
      {
        kind: 'new',
        text: 'Multi-currency invoicing in Finance with live FX rates (ECB + fallback providers).',
      },
      {
        kind: 'improved',
        text: 'Dashboard widgets now load in parallel; initial paint 40% faster.',
      },
    ],
  },
  {
    version: 'v4.7',
    date: 'Feb 07, 2026',
    title: 'SSO for Scale, role-level app access, API rate-limit headers',
    changes: [
      { kind: 'new', text: 'SSO via SAML and OIDC now available on the Scale plan.' },
      {
        kind: 'new',
        text: 'Role-level granular app access, "can view HR but not edit payroll", etc.',
      },
      {
        kind: 'improved',
        text: 'All API responses now include X-RateLimit-Limit and X-RateLimit-Remaining headers.',
      },
      {
        kind: 'breaking',
        text: 'Legacy /v3/orders API sunset today. Migrate to /v1/orders (v1 is the current major).',
      },
    ],
  },
  {
    version: 'v4.6',
    date: 'Jan 24, 2026',
    title: 'Mobile app for managers, delivery marketplace ingestion',
    changes: [
      {
        kind: 'new',
        text: 'Native mobile app (iOS + Android) for managers, approvals, dashboards, shift adjustments on the go.',
      },
      {
        kind: 'new',
        text: 'Direct ingestion from Uber Eats and DoorDash into POS tickets with auto-route to kitchen.',
      },
      {
        kind: 'improved',
        text: 'POS checkout latency reduced from ~340ms to ~180ms on slower networks.',
      },
    ],
  },
  {
    version: 'v4.5',
    date: 'Jan 10, 2026',
    title: 'Public API v1, docs.rtylr.com, Zapier app',
    changes: [
      {
        kind: 'new',
        text: 'Public REST API v1 generally available. Full reference at docs.rtylr.com.',
      },
      { kind: 'new', text: 'Official Zapier app with 36 triggers and 22 actions.' },
      {
        kind: 'improved',
        text: 'Session cookies now strict-SameSite + short-lived refresh pattern.',
      },
    ],
  },
];

const ChangelogPage = (): JSX.Element => {
  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title="Changelog | rtylr"
          description="What we shipped, when we shipped it. New features, improvements, fixes, and breaking changes across the rtylr platform."
          canonical="https://rtylr.com/changelog/"
        />

        <div className={styles.page}>
          {/* HERO */}
          <section className={styles.hero}>
            <motion.div
              className={styles.heroInner}
              initial="hidden"
              animate="visible"
              variants={staggerBlur}
            >
              <motion.span className={styles.heroEyebrow} variants={blurFadeIn}>
                CHANGELOG
              </motion.span>
              <motion.h1 className={styles.heroTitle} variants={blurFadeIn}>
                What shipped,
                <br />
                when it shipped.
              </motion.h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                A running log of every release. New features, improvements, fixes, and the
                occasional breaking change. Follow along via{' '}
                <a href="https://rtylr.com/changelog.rss">RSS</a> or subscribe in your workspace to
                get release notes by email.
              </motion.p>
              <motion.a
                href="https://rtylr.com/changelog.rss"
                className={styles.rssLink}
                variants={blurFadeIn}
              >
                <Rss size={13} strokeWidth={2} />
                Subscribe to the RSS
              </motion.a>
            </motion.div>
          </section>

          {/* TIMELINE */}
          <section className={styles.timelineSection}>
            <motion.ol
              className={styles.timeline}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              {ENTRIES.map((entry) => (
                <motion.li key={entry.version} className={styles.entry} variants={blurFadeIn}>
                  <div className={styles.entryHeader}>
                    <div className={styles.entryMeta}>
                      <span className={styles.entryVersion}>{entry.version}</span>
                      <span className={styles.entryDate}>{entry.date}</span>
                    </div>
                    <h2 className={styles.entryTitle}>{entry.title}</h2>
                  </div>
                  <ul className={styles.changeList}>
                    {entry.changes.map((change, i) => (
                      <li key={i} className={styles.changeItem}>
                        <span className={`${styles.changeKind} ${styles[`kind_${change.kind}`]}`}>
                          {KIND_LABEL[change.kind]}
                        </span>
                        <span className={styles.changeText}>{change.text}</span>
                      </li>
                    ))}
                  </ul>
                </motion.li>
              ))}
            </motion.ol>
          </section>

          {/* CTA */}
          <motion.section
            className={styles.ctaBlock}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.h2 className={styles.ctaTitle} variants={blurFadeIn}>
              Have something we should build?
            </motion.h2>
            <motion.p className={styles.ctaCopy} variants={blurFadeIn}>
              We triage every feature request. Tell us what you need.
            </motion.p>
            <motion.div className={styles.ctaButtons} variants={blurFadeIn}>
              <Link href="/contact" className={styles.ctaPrimary}>
                Send us a request
              </Link>
              <Link href="/support" className={styles.ctaSecondary}>
                Browse help articles
              </Link>
            </motion.div>
          </motion.section>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default ChangelogPage;
