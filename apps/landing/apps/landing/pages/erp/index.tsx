import React, { useCallback, useState, useEffect, useMemo } from 'react';
import { cn } from '../../lib/cn';
import { motion } from 'motion/react';
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
import FaqSection from '../../components/FaqSection/FaqSection';
import { productWithBreadcrumbs } from '../../lib/structured-data';
import { ERP_FAQS } from '../../lib/faqs';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../lib/navigation';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import { SplitWords } from '../../components/motion/SplitWords';
import styles from './erp.module.scss';

// Initial inventory, stock drains each time "Simulate Friday service" fires.
interface InventoryRow {
  sku: string;
  name: string;
  stock: number;
  min: number;
  // Depletion rate per simulated service cycle
  drain: number;
}

const INITIAL_INVENTORY: InventoryRow[] = [
  { sku: 'PIZ-MARG', name: 'Margherita (8oz dough)', stock: 34, min: 40, drain: 4 },
  { sku: 'WIN-RED-02', name: 'House Red (750ml)', stock: 124, min: 60, drain: 12 },
  { sku: 'OIL-OLIVE', name: 'Olive Oil (5L)', stock: 6, min: 12, drain: 1 },
  { sku: 'COF-BEAN', name: 'Espresso Bean (1kg)', stock: 58, min: 30, drain: 6 },
  { sku: 'FLR-00-3KG', name: 'Flour 00 (3kg)', stock: 19, min: 25, drain: 3 },
];

function statusFor(row: InventoryRow): 'critical' | 'low' | 'ok' {
  if (row.stock <= row.min * 0.5) return 'critical';
  if (row.stock < row.min) return 'low';
  return 'ok';
}

const CAPABILITIES = [
  {
    icon: Boxes,
    title: 'Every SKU, every variant',
    copy: 'Track ingredients down to the gram. Recipe-linked depletion. Real-time counts as sales happen.',
  },
  {
    icon: Truck,
    title: 'Suppliers & purchase orders',
    copy: 'Auto-reorder at threshold. Multi-supplier price comparison. Expected-arrival tracking per line item.',
  },
  {
    icon: Warehouse,
    title: 'Multi-location transfers',
    copy: 'Move stock between sites with full audit trail. Location-level reporting. No spreadsheet reconciliation.',
  },
  {
    icon: AlertTriangle,
    title: 'Low-stock before it hurts',
    copy: 'Smart thresholds based on historical velocity. Alerts route to the right person. Nothing runs out mid-service.',
  },
];

const ERPPage = (): JSX.Element => {
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
          title="ERP · Inventory that counts itself. | rtylr"
          description="Real-time inventory, suppliers, and purchasing. Recipe-linked depletion. Multi-location transfers with full audit trail."
          canonical="https://rtylr.com/erp/"
          structuredData={productWithBreadcrumbs({
            name: 'ERP',
            path: '/erp/',
            subCategory: 'Inventory and Supply Chain Management',
            description:
              'Real-time inventory, suppliers, purchase orders, and warehouse management.',
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
                rtylr ERP · Inventory
              </motion.span>
              <SplitWords
                as="h1"
                className={styles.heroTitle}
                text={'Inventory that\ncounts itself.'}
                delay={0.1}
              />
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                Every margherita sold drops flour, mozzarella, and olive oil. Every PO approved
                moves the warehouse count. Every low-stock alert fires before you actually run out.
                Not after the Saturday rush notices.
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
                  <span className={cn(styles.tableTab, styles.tableTabActive)}>All SKUs</span>
                  <span className={styles.tableTab}>Low stock · {lowStockCount}</span>
                  <span className={styles.tableTab}>Suppliers</span>
                  <span className={styles.tableTab}>Purchase orders</span>
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
                        Service running…
                      </>
                    ) : (
                      <>
                        <Play size={11} strokeWidth={2.5} fill="currentColor" />
                        Simulate Friday service
                      </>
                    )}
                  </button>
                  <button
                    type="button"
                    onClick={resetInventory}
                    className={styles.resetBtn}
                    aria-label="Reset inventory"
                    title="Reset"
                  >
                    <RotateCcw size={12} strokeWidth={2} />
                  </button>
                </div>
              </div>
              <table className={styles.table}>
                <thead>
                  <tr>
                    <th>SKU</th>
                    <th>Item</th>
                    <th className={styles.colNum}>On hand</th>
                    <th className={styles.colNum}>Min</th>
                    <th>Status</th>
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
                        <td>{row.name}</td>
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
                              ? 'Critical'
                              : status === 'low'
                                ? 'Reorder'
                                : 'Healthy'}
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
                  Auto-reorder triggered for {lowStockCount} item{lowStockCount === 1 ? '' : 's'} ·
                  PO #{4821 + lowStockCount} drafted
                </span>
              </div>
            </motion.div>
          </section>

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
                CAPABILITIES
              </motion.span>
              <motion.h2 className={styles.sectionTitle} variants={blurFadeIn}>
                Everything from
                <br />
                the loading dock
                <br />
                to the P&amp;L.
              </motion.h2>
            </motion.div>

            <div className={styles.capabilitiesList}>
              {CAPABILITIES.map(({ icon: Icon, title, copy }, i) => (
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
              ))}
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
              <span className={styles.sectionEyebrow}>LIVE RECIPE DEPLETION</span>
              <h2 className={styles.depletionTitle}>Pizza sold. Flour dropped. Reorder drafted.</h2>
              <p className={styles.depletionCopy}>
                Every POS sale cascades through the recipe tree. Flour, yeast, mozzarella, basil,
                olive oil, each depletes by exactly the gram. When anything crosses its reorder
                threshold, a PO is drafted with your preferred supplier and routed for approval.
              </p>
              <div className={styles.depletionFlow}>
                <div className={styles.flowStep}>
                  <div className={styles.flowStepIcon}>
                    <Package size={16} strokeWidth={2} />
                  </div>
                  <div>
                    <div className={styles.flowStepTitle}>1 Margherita sold</div>
                    <div className={styles.flowStepDesc}>POS fires recipe lookup</div>
                  </div>
                </div>
                <div className={styles.flowArrow}>→</div>
                <div className={styles.flowStep}>
                  <div className={styles.flowStepIcon}>
                    <TrendingDown size={16} strokeWidth={2} />
                  </div>
                  <div>
                    <div className={styles.flowStepTitle}>7 ingredients depleted</div>
                    <div className={styles.flowStepDesc}>
                      Flour, mozz, basil, oil, salt, yeast, dough
                    </div>
                  </div>
                </div>
                <div className={styles.flowArrow}>→</div>
                <div className={styles.flowStep}>
                  <div className={styles.flowStepIcon}>
                    <Truck size={16} strokeWidth={2} />
                  </div>
                  <div>
                    <div className={styles.flowStepTitle}>PO auto-drafted</div>
                    <div className={styles.flowStepDesc}>Routed for manager approval</div>
                  </div>
                </div>
              </div>
            </motion.div>
          </motion.section>

          <FaqSection title="Questions about rtylr ERP" faqs={ERP_FAQS} />

          {/* ── CTA ──────────────────────────────────────────────────────── */}
          <motion.section
            className={styles.ctaBlock}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.h2 className={styles.ctaTitle} variants={blurFadeIn}>
              Close your inventory
              <br />
              books while you sleep.
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

export default ERPPage;
