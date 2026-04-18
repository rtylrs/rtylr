import React from 'react';
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
} from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../lib/navigation';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import styles from './erp.module.scss';

// Mock inventory rows shown in the hero table
const INVENTORY_ROWS = [
  {
    sku: 'PIZ-MARG',
    name: 'Margherita (8oz dough)',
    stock: 34,
    min: 40,
    status: 'low' as const,
  },
  {
    sku: 'WIN-RED-02',
    name: 'House Red (750ml)',
    stock: 124,
    min: 60,
    status: 'ok' as const,
  },
  {
    sku: 'OIL-OLIVE',
    name: 'Olive Oil (5L)',
    stock: 6,
    min: 12,
    status: 'critical' as const,
  },
  {
    sku: 'COF-BEAN',
    name: 'Espresso Bean (1kg)',
    stock: 58,
    min: 30,
    status: 'ok' as const,
  },
  {
    sku: 'FLR-00-3KG',
    name: 'Flour 00 (3kg)',
    stock: 19,
    min: 25,
    status: 'low' as const,
  },
];

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
          title="ERP — Inventory that counts itself. | rtylr"
          description="Real-time inventory, suppliers, and purchasing. Recipe-linked depletion. Multi-location transfers with full audit trail."
        />

        <div className={styles.page}>
          {/* ── HERO — full-width inventory table as the focal element ────── */}
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
              <motion.h1 className={styles.heroTitle} variants={blurFadeIn}>
                Inventory that
                <br />
                counts itself.
              </motion.h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                Every sale depletes the right recipe. Every purchase order updates the right
                warehouse. Every low-stock alert arrives before the shelf is actually empty.
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

            {/* Inventory table mockup — full-width under the copy */}
            <motion.div
              className={styles.tableMock}
              initial={{ opacity: 0, y: 30 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.35, duration: 1.0, ease: [0.16, 1, 0.3, 1] }}
              aria-hidden="true"
            >
              <div className={styles.tableHeader}>
                <div className={styles.tableTabs}>
                  <span className={`${styles.tableTab} ${styles.tableTabActive}`}>All SKUs</span>
                  <span className={styles.tableTab}>Low stock · 2</span>
                  <span className={styles.tableTab}>Suppliers</span>
                  <span className={styles.tableTab}>Purchase orders</span>
                </div>
                <div className={styles.tableSync}>
                  <span className={styles.tableSyncDot} />
                  Synced 2s ago
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
                  {INVENTORY_ROWS.map((row) => (
                    <tr key={row.sku}>
                      <td className={styles.cellSku}>{row.sku}</td>
                      <td>{row.name}</td>
                      <td className={styles.colNum}>
                        <span className={styles.stockValue}>{row.stock}</span>
                      </td>
                      <td className={styles.colNum}>{row.min}</td>
                      <td>
                        <span
                          className={`${styles.badge} ${
                            row.status === 'critical'
                              ? styles.badgeCritical
                              : row.status === 'low'
                                ? styles.badgeLow
                                : styles.badgeOk
                          }`}
                        >
                          {row.status === 'critical'
                            ? 'Critical'
                            : row.status === 'low'
                              ? 'Reorder'
                              : 'Healthy'}
                        </span>
                      </td>
                      <td className={styles.colAction}>
                        <ArrowUpRight size={14} />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              <div className={styles.tableFooter}>
                <span>
                  <Zap size={12} strokeWidth={2.5} />
                  Auto-reorder triggered for 2 items · PO #4821 drafted
                </span>
              </div>
            </motion.div>
          </section>

          {/* ── CAPABILITIES — alternating left/right layout ──────────────── */}
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
                  className={`${styles.capabilityRow} ${i % 2 === 1 ? styles.capabilityRowReverse : ''}`}
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

          {/* ── DEPLETION STRIP — "a sale depletes recipes live" ─────────── */}
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
                olive oil — each depletes by exactly the gram. When anything crosses its reorder
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
