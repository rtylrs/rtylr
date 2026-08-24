import React, { useCallback, useState, useEffect, useMemo } from 'react';
import { cn } from '../../lib/cn';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { useRouter } from 'next/router';
import {
  Package,
  TrendingDown,
  Truck,
  Warehouse,
  ArrowUpRight,
  Boxes,
  Zap,
  AlertTriangle,
  Play,
  RotateCcw,
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
import styles from './erp.module.scss';

// Initial inventory, stock drains each time "Simulate Friday service" fires.
// nameIndex points at erp.hero.inventory.items[i] for the display name.
interface InventoryRow {
  sku: string;
  nameIndex: number;
  stock: number;
  min: number;
  // Depletion rate per simulated service cycle
  drain: number;
}

const INITIAL_INVENTORY: InventoryRow[] = [
  { sku: 'PIZ-MARG', nameIndex: 0, stock: 34, min: 40, drain: 4 },
  { sku: 'WIN-RED-02', nameIndex: 1, stock: 124, min: 60, drain: 12 },
  { sku: 'OIL-OLIVE', nameIndex: 2, stock: 6, min: 12, drain: 1 },
  { sku: 'COF-BEAN', nameIndex: 3, stock: 58, min: 30, drain: 6 },
  { sku: 'FLR-00-3KG', nameIndex: 4, stock: 19, min: 25, drain: 3 },
];

function statusFor(row: InventoryRow): 'critical' | 'low' | 'ok' {
  if (row.stock <= row.min * 0.5) return 'critical';
  if (row.stock < row.min) return 'low';
  return 'ok';
}

const CAPABILITY_ICONS = [Boxes, Truck, Warehouse, AlertTriangle];
const FLOW_ICONS = [Package, TrendingDown, Truck];

const ERPPage = (): React.JSX.Element => {
  const faq = useFaq('erp');
  const { t } = useTranslation();

  const capabilities = t('erp.capabilities.items', { returnObjects: true }) as unknown as {
    title: string;
    copy: string;
  }[];
  const flowSteps = t('erp.flow.steps', { returnObjects: true }) as unknown as {
    title: string;
    desc: string;
  }[];
  const inventoryItemNames = t('erp.hero.inventory.items', { returnObjects: true }) as string[];
  const router = useRouter();
  const [inventory, setInventory] = useState<InventoryRow[]>(INITIAL_INVENTORY);
  const [isSimulating, setIsSimulating] = useState(false);

  // Low-stock count drives the tab label
  const lowStockCount = useMemo(
    () => inventory.filter((r) => statusFor(r) !== 'ok').length,
    [inventory]
  );

  const simulateService = useCallback((): void => {
    if (isSimulating) return;
    setIsSimulating(true);
    // Two ticks of depletion, enough to visibly move the bars
    setTimeout(() => {
      setInventory((prev) =>
        prev.map((row) => ({
          ...row,
          stock: Math.max(0, row.stock - Math.round(row.drain * 0.6)),
        }))
      );
    }, 200);
    setTimeout(() => {
      setInventory((prev) =>
        prev.map((row) => ({
          ...row,
          stock: Math.max(0, row.stock - Math.round(row.drain * 0.4)),
        }))
      );
      setIsSimulating(false);
    }, 900);
  }, [isSimulating]);

  const resetInventory = (): void => {
    setInventory(INITIAL_INVENTORY);
  };

  // Auto-run once on first view so the effect is visible without clicking
  useEffect(() => {
    const t = setTimeout(() => simulateService(), 1800);
    return () => clearTimeout(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const goRegister = (): void => {
    navigateExternal(`${APP_URLS.auth}/register?redirect=erp`);
  };

  const goDemo = (): void => {
    void router.push('/demo');
  };

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title={t('erp.meta.title')}
          description={t('erp.meta.description')}
          canonical="https://rtylr.com/erp/"
          structuredData={productWithBreadcrumbs({
            name: 'ERP',
            path: '/erp/',
            subCategory: 'Inventory and Supply Chain Management',
            description: t('erp.meta.schemaDescription'),
          })}
        />

        <div className={styles.page}>
          {/* ── HERO, full-width inventory table as the focal element ────── */}
          <section className={styles.hero}>
            <motion.div
              className={styles.heroCopy}
              initial="hidden"
              animate="visible"
              variants={staggerBlur}
            >
              <motion.span className={styles.heroEyebrow} variants={blurFadeIn}>
                {t('erp.hero.eyebrow')}
              </motion.span>
              <SplitWords
                as="h1"
                className={styles.heroTitle}
                text={t('erp.hero.title')}
                delay={0.1}
              />
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                {t('erp.hero.subtitle')}
              </motion.p>
              <motion.div className={styles.heroCtas} variants={blurFadeIn}>
                <button type="button" onClick={goRegister} className={styles.primaryBtn}>
                  {t('common.cta.getStarted')}
                </button>
                <button type="button" onClick={goDemo} className={styles.outlineBtn}>
                  {t('common.cta.bookDemo')}
                </button>
              </motion.div>
            </motion.div>

            {/* Inventory table mockup, full-width under the copy */}
            <motion.div
              className={styles.tableMock}
              initial={{ opacity: 0, y: 30 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.35, duration: 1.0, ease: [0.16, 1, 0.3, 1] }}
              aria-hidden="true"
            >
              <div className={styles.tableHeader}>
                <div className={styles.tableTabs}>
                  <span className={cn(styles.tableTab, styles.tableTabActive)}>
                    {t('erp.mock.tabs.allSkus')}
                  </span>
                  <span className={styles.tableTab}>
                    {t('erp.mock.tabs.lowStock', { count: lowStockCount })}
                  </span>
                  <span className={styles.tableTab}>{t('erp.mock.tabs.suppliers')}</span>
                  <span className={styles.tableTab}>{t('erp.mock.tabs.purchaseOrders')}</span>
                </div>
                <div className={styles.tableActions}>
                  <button
                    type="button"
                    onClick={simulateService}
                    className={styles.simulateBtn}
                    disabled={isSimulating}
                  >
                    {isSimulating ? (
                      <>
                        <span className={styles.simulatePulse} />
                        {t('erp.mock.simulateRunning')}
                      </>
                    ) : (
                      <>
                        <Play size={11} strokeWidth={2.5} fill="currentColor" />
                        {t('erp.mock.simulate')}
                      </>
                    )}
                  </button>
                  <button
                    type="button"
                    onClick={resetInventory}
                    className={styles.resetBtn}
                    aria-label={t('erp.mock.resetLabel')}
                    title={t('erp.mock.reset')}
                  >
                    <RotateCcw size={12} strokeWidth={2} />
                  </button>
                </div>
              </div>
              <table className={styles.table}>
                <thead>
                  <tr>
                    <th>{t('erp.mock.table.sku')}</th>
                    <th>{t('erp.mock.table.item')}</th>
                    <th className={styles.colNum}>{t('erp.mock.table.onHand')}</th>
                    <th className={styles.colNum}>{t('erp.mock.table.min')}</th>
                    <th>{t('erp.mock.table.status')}</th>
                    <th className={styles.colAction} />
                  </tr>
                </thead>
                <tbody>
                  {inventory.map((row) => {
                    const status = statusFor(row);
                    const fillPct = Math.min(100, (row.stock / (row.min * 1.8)) * 100);
                    return (
                      <tr key={row.sku}>
                        <td className={styles.cellSku}>{row.sku}</td>
                        <td>{inventoryItemNames[row.nameIndex]}</td>
                        <td className={styles.colNum}>
                          <motion.span
                            key={row.stock}
                            className={styles.stockValue}
                            initial={{ opacity: 0, y: -3 }}
                            animate={{ opacity: 1, y: 0 }}
                            transition={{ duration: 0.22 }}
                          >
                            {row.stock}
                          </motion.span>
                          <div className={styles.stockBar} aria-hidden="true">
                            <motion.div
                              className={cn(
                                styles.stockBarFill,
                                status === 'critical'
                                  ? styles.stockBarCritical
                                  : status === 'low'
                                    ? styles.stockBarLow
                                    : styles.stockBarOk
                              )}
                              animate={{ width: `${fillPct}%` }}
                              transition={{ duration: 0.6, ease: [0.16, 1, 0.3, 1] }}
                            />
                          </div>
                        </td>
                        <td className={styles.colNum}>{row.min}</td>
                        <td>
                          <motion.span
                            key={status}
                            className={cn(
                              styles.badge,
                              status === 'critical'
                                ? styles.badgeCritical
                                : status === 'low'
                                  ? styles.badgeLow
                                  : styles.badgeOk
                            )}
                            initial={{ opacity: 0, scale: 0.9 }}
                            animate={{ opacity: 1, scale: 1 }}
                            transition={{ duration: 0.2 }}
                          >
                            {status === 'critical'
                              ? t('erp.mock.status.critical')
                              : status === 'low'
                                ? t('erp.mock.status.reorder')
                                : t('erp.mock.status.healthy')}
                          </motion.span>
                        </td>
                        <td className={styles.colAction}>
                          <ArrowUpRight size={14} />
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
              <div className={styles.tableFooter}>
                <span>
                  <Zap size={12} strokeWidth={2.5} />
                  {t('erp.mock.footer', { count: lowStockCount, po: 4821 + lowStockCount })}
                </span>
              </div>
            </motion.div>
          </section>

          <DirectAnswer label={t('erp.directAnswer.label')}>
            {t('erp.directAnswer.body')}
          </DirectAnswer>

          {/* ── CAPABILITIES, alternating left/right layout ──────────────── */}
          <section className={styles.capabilitiesSection}>
            <motion.div
              className={styles.capabilitiesHeader}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              <motion.span className={styles.sectionEyebrow} variants={blurFadeIn}>
                {t('erp.capabilities.eyebrow')}
              </motion.span>
              <motion.h2 className={styles.sectionTitle} variants={blurFadeIn}>
                <Lines text={t('erp.capabilities.title')} />
              </motion.h2>
            </motion.div>

            <div className={styles.capabilitiesList}>
              {capabilities.map(({ title, copy }, i) => {
                const Icon = CAPABILITY_ICONS[i];
                return (
                  <motion.div
                    key={title}
                    className={cn(styles.capabilityRow, i % 2 === 1 && styles.capabilityRowReverse)}
                    initial={{ opacity: 0, x: i % 2 === 0 ? -20 : 20 }}
                    whileInView={{ opacity: 1, x: 0 }}
                    viewport={{ once: true, margin: '-80px' }}
                    transition={{ duration: 0.7, ease: [0.16, 1, 0.3, 1] }}
                  >
                    <div className={styles.capabilityNumber}>{String(i + 1).padStart(2, '0')}</div>
                    <div className={styles.capabilityContent}>
                      <div className={styles.capabilityIcon}>
                        <Icon size={20} strokeWidth={1.8} />
                      </div>
                      <h3 className={styles.capabilityTitle}>{title}</h3>
                      <p className={styles.capabilityCopy}>{copy}</p>
                    </div>
                  </motion.div>
                );
              })}
            </div>
          </section>

          {/* ── DEPLETION STRIP, "a sale depletes recipes live" ─────────── */}
          <motion.section
            className={styles.depletionSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.div className={styles.depletionInner} variants={blurFadeIn}>
              <span className={styles.sectionEyebrow}>{t('erp.depletion.eyebrow')}</span>
              <h2 className={styles.depletionTitle}>{t('erp.depletion.title')}</h2>
              <p className={styles.depletionCopy}>{t('erp.depletion.copy')}</p>
              <div className={styles.depletionFlow}>
                {flowSteps.map(({ title, desc }, i) => {
                  const Icon = FLOW_ICONS[i];
                  return (
                    <React.Fragment key={title}>
                      {i > 0 && <div className={cn(styles.flowArrow, 'rtl-mirror')}>→</div>}
                      <div className={styles.flowStep}>
                        <div className={styles.flowStepIcon}>
                          <Icon size={16} strokeWidth={2} />
                        </div>
                        <div>
                          <div className={styles.flowStepTitle}>{title}</div>
                          <div className={styles.flowStepDesc}>{desc}</div>
                        </div>
                      </div>
                    </React.Fragment>
                  );
                })}
              </div>
            </motion.div>
          </motion.section>

          <FaqSection title={faq.title} faqs={faq.items} />

          {/* ── CTA ──────────────────────────────────────────────────────── */}
          <motion.section
            className={styles.ctaBlock}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.h2 className={styles.ctaTitle} variants={blurFadeIn}>
              <Lines text={t('erp.cta.title')} />
            </motion.h2>
            <motion.div className={styles.ctaButtons} variants={blurFadeIn}>
              <button type="button" onClick={goRegister} className={styles.primaryBtn}>
                {t('common.cta.getStarted')}
              </button>
              <button type="button" onClick={goDemo} className={styles.outlineBtn}>
                {t('common.cta.bookDemo')}
              </button>
            </motion.div>
          </motion.section>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default ERPPage;
