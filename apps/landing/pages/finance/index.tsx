import React from 'react';
import { motion } from 'motion/react';
import { useRouter } from 'next/router';
import {
  FileText,
  Receipt,
  PieChart,
  Landmark,
  ArrowUp,
  ArrowDown,
  CheckCircle2,
} from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../lib/navigation';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import styles from './finance.module.scss';

// Monthly P&L rows displayed in the hero mockup
const PL_ROWS = [
  { label: 'Food & beverage', amount: 184_210, delta: 8.4, positive: true },
  { label: 'Catering & events', amount: 42_180, delta: 12.1, positive: true },
  { label: 'Delivery fees', amount: 9_640, delta: -3.2, positive: false },
  { label: 'Subtotal — Revenue', amount: 236_030, delta: 7.8, positive: true, subtotal: true },
  { label: 'COGS', amount: -68_420, delta: 2.1, positive: false },
  { label: 'Labor', amount: -74_890, delta: 0.9, positive: false },
  { label: 'Rent & utilities', amount: -22_400, delta: 0, positive: false },
  { label: 'Operating expenses', amount: -18_720, delta: 5.4, positive: false },
  { label: 'Subtotal — Costs', amount: -184_430, delta: 2.2, positive: false, subtotal: true },
  { label: 'Net income', amount: 51_600, delta: 18.3, positive: true, total: true },
];

const PILLARS = [
  {
    icon: FileText,
    title: 'Invoices that collect themselves',
    copy: 'Recurring schedules, auto-reminders, one-click payment. Paid invoices reconcile automatically against the books.',
  },
  {
    icon: Receipt,
    title: 'Expenses, captured at source',
    copy: 'Snap the receipt. Policy-aware categorisation. Approved-to-paid in the same app. No more Friday shoeboxes.',
  },
  {
    icon: PieChart,
    title: 'Real P&L in real time',
    copy: 'Every POS sale, every invoice, every payroll run rolls into a live profit-and-loss. Not a month-end surprise.',
  },
  {
    icon: Landmark,
    title: 'Tax prep without the scramble',
    copy: 'Jurisdiction-aware sales tax. Quarter-end snapshots. One export your accountant will actually thank you for.',
  },
];

const fmt = (n: number): string => {
  const sign = n < 0 ? '−' : '';
  return `${sign}$${Math.abs(n).toLocaleString('en-US', { maximumFractionDigits: 0 })}`;
};

const FinancePage = (): JSX.Element => {
  const router = useRouter();

  const goRegister = (): void => {
    navigateExternal(`${APP_URLS.auth}/register`);
  };

  const goDemo = (): void => {
    void router.push('/demo');
  };

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title="Finance — A real P&L, in real time. | rtylr"
          description="Invoices, expenses, sales tax, and a live profit-and-loss. Every sale and every payroll roll into the books automatically."
        />

        <div className={styles.page}>
          {/* ── HERO — centered copy, full-width P&L mockup below ────────── */}
          <section className={styles.hero}>
            <motion.div
              className={styles.heroCopy}
              initial="hidden"
              animate="visible"
              variants={staggerBlur}
            >
              <motion.span className={styles.heroEyebrow} variants={blurFadeIn}>
                rtylr Finance
              </motion.span>
              <motion.h1 className={styles.heroTitle} variants={blurFadeIn}>
                A real P&amp;L,
                <br />
                in real time.
              </motion.h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                Invoices, expenses, sales tax, and a live profit-and-loss. Every POS sale, every
                payroll run, every paid invoice rolls into the books automatically — no exports, no
                reconciliation, no month-end surprises.
              </motion.p>
              <motion.div className={styles.heroCtas} variants={blurFadeIn}>
                <button type="button" onClick={goRegister} className={styles.primaryBtn}>
                  Get started
                </button>
                <button type="button" onClick={goDemo} className={styles.outlineBtn}>
                  Book a demo
                </button>
              </motion.div>
            </motion.div>

            {/* Live P&L mockup */}
            <motion.div
              className={styles.plMock}
              initial={{ opacity: 0, y: 40 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.4, duration: 1.0, ease: [0.16, 1, 0.3, 1] }}
              aria-hidden="true"
            >
              <div className={styles.plHeader}>
                <div>
                  <div className={styles.plTitle}>Profit &amp; Loss · April 2026</div>
                  <div className={styles.plSubtitle}>
                    Terrace · monthly view · updated 14:02 EST
                  </div>
                </div>
                <div className={styles.plNetBadge}>
                  <span className={styles.plNetLabel}>Net this month</span>
                  <span className={styles.plNetValue}>$51,600</span>
                  <span className={styles.plNetDelta}>
                    <ArrowUp size={12} strokeWidth={2.5} />
                    +18.3% vs March
                  </span>
                </div>
              </div>

              <div className={styles.plTableWrap}>
                <table className={styles.plTable}>
                  <thead>
                    <tr>
                      <th>Line item</th>
                      <th className={styles.colNum}>Amount</th>
                      <th className={styles.colNum}>vs last month</th>
                    </tr>
                  </thead>
                  <tbody>
                    {PL_ROWS.map((row) => (
                      <tr
                        key={row.label}
                        className={`${row.subtotal ? styles.rowSubtotal : ''} ${
                          row.total ? styles.rowTotal : ''
                        }`}
                      >
                        <td>{row.label}</td>
                        <td className={styles.colNum}>{fmt(row.amount)}</td>
                        <td className={styles.colNum}>
                          {row.delta !== 0 ? (
                            <span
                              className={`${styles.delta} ${
                                row.positive ? styles.deltaUp : styles.deltaDown
                              }`}
                            >
                              {row.positive ? (
                                <ArrowUp size={11} strokeWidth={2.5} />
                              ) : (
                                <ArrowDown size={11} strokeWidth={2.5} />
                              )}
                              {Math.abs(row.delta).toFixed(1)}%
                            </span>
                          ) : (
                            <span className={styles.deltaFlat}>—</span>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              <div className={styles.plFooter}>
                <CheckCircle2 size={13} strokeWidth={2.2} />
                <span>
                  All sources reconciled · 12 POS registers · 186 invoices · 342 expense entries
                </span>
              </div>
            </motion.div>
          </section>

          {/* ── PILLAR GRID — 2×2 shared borders ─────────────────────────── */}
          <section className={styles.pillarsSection}>
            <motion.div
              className={styles.pillarsHeader}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              <motion.span className={styles.sectionEyebrow} variants={blurFadeIn}>
                FOUR THINGS ACCOUNTING SHOULD DO
              </motion.span>
              <motion.h2 className={styles.sectionTitle} variants={blurFadeIn}>
                One app for the
                <br />
                stuff the ledger
                <br />
                should just do.
              </motion.h2>
            </motion.div>

            <motion.div
              className={styles.pillarsGrid}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              {PILLARS.map(({ icon: Icon, title, copy }) => (
                <motion.div key={title} className={styles.pillarCard} variants={blurFadeIn}>
                  <div className={styles.pillarIcon}>
                    <Icon size={18} strokeWidth={1.8} />
                  </div>
                  <h3 className={styles.pillarTitle}>{title}</h3>
                  <p className={styles.pillarCopy}>{copy}</p>
                </motion.div>
              ))}
            </motion.div>
          </section>

          {/* ── FLOWS INTO ACCOUNTANT ───────────────────────────────────── */}
          <motion.section
            className={styles.accountantSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.div className={styles.accountantInner} variants={blurFadeIn}>
              <h2 className={styles.accountantTitle}>
                Your accountant will love it.
                <br />
                <span className={styles.accountantTitleSoft}>Finally.</span>
              </h2>
              <p className={styles.accountantCopy}>
                One click to export the chart of accounts, journal entries, payroll detail, and
                sales-tax filings. GAAP-compliant output. QuickBooks and Xero bridge for the
                hybrid-tool holdouts.
              </p>
              <div className={styles.accountantFormats}>
                <span className={styles.formatPill}>CSV</span>
                <span className={styles.formatPill}>XLSX</span>
                <span className={styles.formatPill}>PDF</span>
                <span className={styles.formatPill}>QuickBooks</span>
                <span className={styles.formatPill}>Xero</span>
                <span className={styles.formatPill}>Sage</span>
              </div>
            </motion.div>
          </motion.section>

          {/* ── CTA ───────────────────────────────────────────────────── */}
          <motion.section
            className={styles.ctaBlock}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.h2 className={styles.ctaTitle} variants={blurFadeIn}>
              Stop closing your books
              <br />
              with a spreadsheet.
            </motion.h2>
            <motion.div className={styles.ctaButtons} variants={blurFadeIn}>
              <button type="button" onClick={goRegister} className={styles.primaryBtn}>
                Get started
              </button>
              <button type="button" onClick={goDemo} className={styles.outlineBtn}>
                Book a demo
              </button>
            </motion.div>
          </motion.section>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default FinancePage;
