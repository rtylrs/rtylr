import React, { useState } from 'react';
import { cn } from '../../lib/cn';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
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
import DirectAnswer from '../../components/DirectAnswer/DirectAnswer';
import FaqSection from '../../components/FaqSection/FaqSection';
import { productWithBreadcrumbs } from '../../lib/structured-data';
import { useFaq } from '../../lib/faqs';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../lib/navigation';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import { SplitWords } from '../../components/motion/SplitWords';
import Lines from '../../components/text/Lines';
import styles from './finance.module.scss';

// ── P&L data shaped per time range ────────────────────────────────────────
// Each range defines its own amounts + comparison period label.

type Range = 'week' | 'month' | 'quarter' | 'year';

interface PLRow {
  label: string;
  amount: number;
  delta: number;
  positive: boolean;
  subtotal?: boolean;
  total?: boolean;
}

interface PLData {
  label: string;
  compareLabel: string;
  net: number;
  netDelta: number;
  rows: PLRow[];
}

const PL_DATA: Record<Range, PLData> = {
  week: {
    label: 'Week of April 13 to April 19, 2026',
    compareLabel: 'vs last week',
    net: 11_800,
    netDelta: 9.2,
    rows: [
      { label: 'Food & beverage', amount: 42_380, delta: 6.1, positive: true },
      { label: 'Catering & events', amount: 9_040, delta: 4.8, positive: true },
      { label: 'Delivery fees', amount: 2_180, delta: -1.4, positive: false },
      { label: 'Subtotal · Revenue', amount: 53_600, delta: 5.4, positive: true, subtotal: true },
      { label: 'COGS', amount: -15_620, delta: 1.9, positive: false },
      { label: 'Labor', amount: -17_140, delta: 0.8, positive: false },
      { label: 'Rent & utilities', amount: -5_180, delta: 0, positive: false },
      { label: 'Operating expenses', amount: -3_860, delta: 2.4, positive: false },
      { label: 'Subtotal · Costs', amount: -41_800, delta: 1.6, positive: false, subtotal: true },
      { label: 'Net income', amount: 11_800, delta: 9.2, positive: true, total: true },
    ],
  },
  month: {
    label: 'April 2026',
    compareLabel: 'vs March',
    net: 51_600,
    netDelta: 18.3,
    rows: [
      { label: 'Food & beverage', amount: 184_210, delta: 8.4, positive: true },
      { label: 'Catering & events', amount: 42_180, delta: 12.1, positive: true },
      { label: 'Delivery fees', amount: 9_640, delta: -3.2, positive: false },
      { label: 'Subtotal · Revenue', amount: 236_030, delta: 7.8, positive: true, subtotal: true },
      { label: 'COGS', amount: -68_420, delta: 2.1, positive: false },
      { label: 'Labor', amount: -74_890, delta: 0.9, positive: false },
      { label: 'Rent & utilities', amount: -22_400, delta: 0, positive: false },
      { label: 'Operating expenses', amount: -18_720, delta: 5.4, positive: false },
      { label: 'Subtotal · Costs', amount: -184_430, delta: 2.2, positive: false, subtotal: true },
      { label: 'Net income', amount: 51_600, delta: 18.3, positive: true, total: true },
    ],
  },
  quarter: {
    label: 'Q2 2026 (Apr–Jun to-date)',
    compareLabel: 'vs Q1',
    net: 162_400,
    netDelta: 22.4,
    rows: [
      { label: 'Food & beverage', amount: 548_920, delta: 11.6, positive: true },
      { label: 'Catering & events', amount: 138_440, delta: 18.2, positive: true },
      { label: 'Delivery fees', amount: 29_180, delta: -5.1, positive: false },
      { label: 'Subtotal · Revenue', amount: 716_540, delta: 10.9, positive: true, subtotal: true },
      { label: 'COGS', amount: -206_880, delta: 2.9, positive: false },
      { label: 'Labor', amount: -226_160, delta: 1.3, positive: false },
      { label: 'Rent & utilities', amount: -67_200, delta: 0, positive: false },
      { label: 'Operating expenses', amount: -53_900, delta: 4.8, positive: false },
      { label: 'Subtotal · Costs', amount: -554_140, delta: 2.4, positive: false, subtotal: true },
      { label: 'Net income', amount: 162_400, delta: 22.4, positive: true, total: true },
    ],
  },
  year: {
    label: '2026 year-to-date',
    compareLabel: 'vs 2025 YTD',
    net: 428_000,
    netDelta: 31.8,
    rows: [
      { label: 'Food & beverage', amount: 1_420_680, delta: 14.2, positive: true },
      { label: 'Catering & events', amount: 362_400, delta: 22.6, positive: true },
      { label: 'Delivery fees', amount: 82_940, delta: -2.8, positive: false },
      {
        label: 'Subtotal · Revenue',
        amount: 1_866_020,
        delta: 13.9,
        positive: true,
        subtotal: true,
      },
      { label: 'COGS', amount: -538_100, delta: 3.4, positive: false },
      { label: 'Labor', amount: -584_820, delta: 1.6, positive: false },
      { label: 'Rent & utilities', amount: -179_200, delta: 0, positive: false },
      { label: 'Operating expenses', amount: -135_900, delta: 6.1, positive: false },
      {
        label: 'Subtotal · Costs',
        amount: -1_438_020,
        delta: 2.8,
        positive: false,
        subtotal: true,
      },
      { label: 'Net income', amount: 428_000, delta: 31.8, positive: true, total: true },
    ],
  },
};

const RANGE_TABS: { id: Range; label: string }[] = [
  { id: 'week', label: 'Week' },
  { id: 'month', label: 'Month' },
  { id: 'quarter', label: 'Quarter' },
  { id: 'year', label: 'Year' },
];

// ── Trailing-12-month chart data (seeded to look real) ────────────────────

type ChartMetric = 'revenue' | 'costs' | 'net';

interface ChartMetricDef {
  id: ChartMetric;
  label: string;
  color: string;
}

const CHART_METRICS: ChartMetricDef[] = [
  { id: 'revenue', label: 'Revenue', color: '#6bcf7f' },
  { id: 'costs', label: 'Costs', color: '#e88a8a' },
  { id: 'net', label: 'Net', color: '#d6b88e' },
];

const CHART_LABELS = [
  'May',
  'Jun',
  'Jul',
  'Aug',
  'Sep',
  'Oct',
  'Nov',
  'Dec',
  'Jan',
  'Feb',
  'Mar',
  'Apr',
];

// Revenue that grows with some seasonality; costs shadow at ~78%.
const CHART_REVENUE = [
  168_400, 182_900, 194_200, 208_600, 196_200, 214_500, 228_900, 252_400, 201_800, 218_600, 223_400,
  236_030,
];

const CHART_COSTS = CHART_REVENUE.map((r) => Math.round(r * 0.78));
const CHART_NET = CHART_REVENUE.map((r, i) => r - CHART_COSTS[i]);

const CHART_DATA: Record<ChartMetric, number[]> = {
  revenue: CHART_REVENUE,
  costs: CHART_COSTS,
  net: CHART_NET,
};

const PILLARS_ICONS = [FileText, Receipt, PieChart, Landmark];

const fmt = (n: number): string => {
  const sign = n < 0 ? '−' : '';
  return `${sign}$${Math.abs(n).toLocaleString('en-US', { maximumFractionDigits: 0 })}`;
};

const FinancePage = (): React.JSX.Element => {
  const faq = useFaq('finance');
  const { t } = useTranslation();
  const pillars = t('finance.pillars.items', { returnObjects: true }) as unknown as {
    title: string;
    copy: string;
  }[];
  const router = useRouter();
  const [range, setRange] = useState<Range>('month');
  const data = PL_DATA[range];

  // Chart state, independent from P&L range so both can be played with
  const [chartMetric, setChartMetric] = useState<ChartMetric>('revenue');
  const activeChartMetric = CHART_METRICS.find((m) => m.id === chartMetric) ?? CHART_METRICS[0];
  const chartSeries = CHART_DATA[chartMetric].map((value, i) => ({
    label: CHART_LABELS[i],
    value,
  }));
  const chartMax = Math.max(...CHART_DATA[chartMetric]);
  const chartSum = CHART_DATA[chartMetric].reduce((acc, n) => acc + n, 0);

  const goRegister = (): void => {
    navigateExternal(`${APP_URLS.auth}/register?redirect=finance`);
  };

  const goDemo = (): void => {
    void router.push('/demo');
  };

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title={t('finance.meta.title')}
          description={t('finance.meta.description')}
          canonical="https://rtylr.com/finance/"
          structuredData={productWithBreadcrumbs({
            name: 'Finance',
            path: '/finance/',
            subCategory: 'Accounting and Finance',
            description: t('finance.meta.schemaDescription'),
          })}
        />

        <div className={styles.page}>
          {/* ── HERO, centered copy, full-width P&L mockup below ────────── */}
          <section className={styles.hero}>
            <motion.div
              className={styles.heroCopy}
              initial="hidden"
              animate="visible"
              variants={staggerBlur}
            >
              <motion.span className={styles.heroEyebrow} variants={blurFadeIn}>
                {t('finance.hero.eyebrow')}
              </motion.span>
              <SplitWords
                as="h1"
                className={styles.heroTitle}
                text={'A real P&L,\nin real time.'}
                delay={0.1}
              />
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                {t('finance.hero.subtitle')}
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
                  <div className={styles.plTitle}>Profit &amp; Loss · {data.label}</div>
                  <div className={styles.plSubtitle}>
                    Terrace · updated live · click a range to recalculate
                  </div>
                </div>
                <div className={styles.plNetBadge}>
                  <span className={styles.plNetLabel}>Net {range}</span>
                  <motion.span
                    key={`net-${range}`}
                    className={styles.plNetValue}
                    initial={{ opacity: 0, y: -4 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.24 }}
                  >
                    {fmt(data.net)}
                  </motion.span>
                  <span className={styles.plNetDelta}>
                    <ArrowUp size={12} strokeWidth={2.5} />+{data.netDelta.toFixed(1)}%{' '}
                    {data.compareLabel}
                  </span>
                </div>
              </div>

              {/* Range tabs, live re-computation */}
              <div className={styles.rangeTabs} role="tablist" aria-label={t('finance.chart.rangeLabel')}>
                {RANGE_TABS.map((t) => (
                  <button
                    key={t.id}
                    type="button"
                    role="tab"
                    aria-selected={range === t.id}
                    onClick={() => setRange(t.id)}
                    className={cn(styles.rangeTab, range === t.id && styles.rangeTabActive)}
                  >
                    {t.label}
                  </button>
                ))}
              </div>

              <div className={styles.plTableWrap}>
                <table className={styles.plTable}>
                  <thead>
                    <tr>
                      <th>Line item</th>
                      <th className={styles.colNum}>Amount</th>
                      <th className={styles.colNum}>{data.compareLabel}</th>
                    </tr>
                  </thead>
                  <motion.tbody
                    key={range}
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    transition={{ duration: 0.28 }}
                  >
                    {data.rows.map((row) => (
                      <tr
                        key={row.label}
                        className={cn(
                          row.subtotal && styles.rowSubtotal,
                          row.total && styles.rowTotal
                        )}
                      >
                        <td>{row.label}</td>
                        <td className={styles.colNum}>{fmt(row.amount)}</td>
                        <td className={styles.colNum}>
                          {row.delta !== 0 ? (
                            <span
                              className={cn(
                                styles.delta,
                                row.positive ? styles.deltaUp : styles.deltaDown
                              )}
                            >
                              {row.positive ? (
                                <ArrowUp size={11} strokeWidth={2.5} />
                              ) : (
                                <ArrowDown size={11} strokeWidth={2.5} />
                              )}
                              {Math.abs(row.delta).toFixed(1)}%
                            </span>
                          ) : (
                            <span className={styles.deltaFlat}>-</span>
                          )}
                        </td>
                      </tr>
                    ))}
                  </motion.tbody>
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

          <DirectAnswer label={t('finance.directAnswer.label')}>
            {t('finance.directAnswer.body')}
          </DirectAnswer>

          {/* ── LIVE 12-MONTH CHART + PILLAR STRIP ───────────────────────── */}
          <section className={styles.chartSection}>
            <motion.div
              className={styles.chartHeader}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              <motion.span className={styles.sectionEyebrow} variants={blurFadeIn}>
                LIVE, TRAILING 12 MONTHS
              </motion.span>
              <motion.h2 className={styles.sectionTitle} variants={blurFadeIn}>
                What can you see
                <br />
                without a close?
              </motion.h2>
              <motion.p className={styles.chartCopy} variants={blurFadeIn}>
                Toggle the metric to see revenue, costs, or net for every month of the last year.
                These numbers stream in from every POS sale, every invoice, every payroll run.
              </motion.p>
            </motion.div>

            <motion.div
              className={styles.chartPanel}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={viewportOnce}
              transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
            >
              <div className={styles.chartTabs} role="tablist" aria-label={t('finance.chart.metricLabel')}>
                {CHART_METRICS.map((m) => (
                  <button
                    key={m.id}
                    type="button"
                    role="tab"
                    aria-selected={chartMetric === m.id}
                    onClick={() => setChartMetric(m.id)}
                    className={cn(styles.chartTab, chartMetric === m.id && styles.chartTabActive)}
                  >
                    <span className={styles.chartTabDot} style={{ background: m.color }} />
                    {m.label}
                  </button>
                ))}
                <div className={styles.chartTotal}>
                  <span className={styles.chartTotalLabel}>Sum</span>
                  <motion.span
                    key={`total-${chartMetric}`}
                    className={styles.chartTotalValue}
                    initial={{ opacity: 0, y: -3 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.22 }}
                  >
                    {fmt(chartSum)}
                  </motion.span>
                </div>
              </div>

              <div className={styles.chartBars}>
                {chartSeries.map((month, i) => {
                  const heightPct = chartMax > 0 ? (month.value / chartMax) * 100 : 0;
                  const active = activeChartMetric.color;
                  return (
                    <div key={month.label} className={styles.chartBarCol}>
                      <div className={styles.chartBarWrap}>
                        <motion.div
                          className={styles.chartBar}
                          style={{ background: `linear-gradient(180deg, ${active}, ${active}66)` }}
                          initial={{ height: '0%' }}
                          animate={{ height: `${heightPct}%` }}
                          transition={{
                            delay: i * 0.035,
                            duration: 0.6,
                            ease: [0.16, 1, 0.3, 1],
                          }}
                        />
                        <div className={styles.chartBarTooltip}>{fmt(month.value)}</div>
                      </div>
                      <span className={styles.chartBarLabel}>{month.label}</span>
                    </div>
                  );
                })}
              </div>

              <div className={styles.chartFooter}>
                <span className={styles.chartFooterDot} />
                Updating live · last sync 14:02 EST
              </div>
            </motion.div>

            {/* Compact pillar strip, horizontal, not a grid */}
            <motion.div
              className={styles.pillarStrip}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              {pillars.map(({ title, copy }, i) => {
                const Icon = PILLARS_ICONS[i];
                return (
                <motion.div key={title} className={styles.pillarStripItem} variants={blurFadeIn}>
                  <div className={styles.pillarIcon}>
                    <Icon size={16} strokeWidth={1.8} />
                  </div>
                  <div>
                    <h3 className={styles.pillarTitle}>{title}</h3>
                    <p className={styles.pillarCopy}>{copy}</p>
                  </div>
                </motion.div>
              );
              })}
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
                What can you hand
                <br />
                <span className={styles.accountantTitleSoft}>your accountant?</span>
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

          <FaqSection title={faq.title} faqs={faq.items} />

          {/* ── CTA ───────────────────────────────────────────────────── */}
          <motion.section
            className={styles.ctaBlock}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.h2 className={styles.ctaTitle} variants={blurFadeIn}>
              <Lines text={t('finance.cta.title')} />
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
