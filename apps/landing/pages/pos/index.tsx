import React from 'react';
import { motion } from 'motion/react';
import { useRouter } from 'next/router';
import { Receipt, CreditCard, Split, Wifi, WifiOff, Users, Clock, TrendingUp } from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../lib/navigation';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import styles from './pos.module.scss';

// ── Static receipt items shown in the hero mockup ──────────────────────────
const RECEIPT_ITEMS = [
  { name: 'Margherita Pizza', qty: 2, price: 28.0 },
  { name: 'House Salad', qty: 1, price: 9.5 },
  { name: 'Sparkling Water', qty: 2, price: 6.0 },
  { name: 'Espresso', qty: 2, price: 7.0 },
];

const FEATURES = [
  {
    icon: Receipt,
    title: 'Tickets that keep up',
    copy: 'Split, merge, combine. Send courses to the right station. Fire when you say fire.',
  },
  {
    icon: CreditCard,
    title: 'Every payment method',
    copy: 'Card, Apple Pay, Google Pay, cash, gift, comp, house account. One register handles all of it.',
  },
  {
    icon: Split,
    title: 'Tips, tabs, and splits',
    copy: 'Seat-level splits. Auto-gratuity. Pre-auth tabs. Tips hit payroll the next morning.',
  },
  {
    icon: WifiOff,
    title: 'Works when the internet doesn\u2019t',
    copy: 'Full offline mode. Queues transactions locally and syncs the moment you\u2019re back online.',
  },
];

const STATS = [
  { value: '< 200ms', label: 'Time per checkout' },
  { value: '99.99%', label: 'Offline reliability' },
  { value: '12+', label: 'Payment methods' },
  { value: '7', label: 'Apps connected out of the box' },
];

const POSPage = (): JSX.Element => {
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
          title="POS — The register that keeps up. | rtylr"
          description="Full-service POS for restaurants and retail. Tickets, splits, tips, and a real offline mode. Built into the rtylr platform."
        />

        <div className={styles.page}>
          {/* ── HERO — split layout: copy left, receipt mockup right ───────── */}
          <section className={styles.hero}>
            <div className={styles.heroFrame} aria-hidden="true">
              <span className={`${styles.vLine} ${styles.vLineLeft}`} />
              <span className={`${styles.vLine} ${styles.vLineRight}`} />
            </div>

            <motion.div
              className={styles.heroCopy}
              initial="hidden"
              animate="visible"
              variants={staggerBlur}
            >
              <motion.span className={styles.heroEyebrow} variants={blurFadeIn}>
                rtylr POS
              </motion.span>
              <motion.h1 className={styles.heroTitle} variants={blurFadeIn}>
                The register
                <br />
                that keeps up.
              </motion.h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                Fire courses, split checks, take every payment method, and keep running when the
                internet drops. The only POS designed to match the pace of real service.
              </motion.p>
              <motion.div className={styles.heroCtas} variants={blurFadeIn}>
                <button type="button" onClick={goRegister} className={styles.primaryBtn}>
                  Get started
                </button>
                <button type="button" onClick={goDemo} className={styles.outlineBtn}>
                  See it live
                </button>
              </motion.div>
            </motion.div>

            {/* Receipt mockup — thermal-printer aesthetic */}
            <motion.div
              className={styles.receiptMock}
              initial={{ opacity: 0, y: 40 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.4, duration: 1.0, ease: [0.16, 1, 0.3, 1] }}
              aria-hidden="true"
            >
              <div className={styles.receiptHeader}>
                <span className={styles.receiptDots}>
                  <span />
                  <span />
                  <span />
                </span>
                <span className={styles.receiptOrder}>Order #A-2048 · Table 12</span>
                <span className={styles.receiptLive}>
                  <span className={styles.livePulse} />
                  LIVE
                </span>
              </div>
              <div className={styles.receiptBody}>
                <div className={styles.receiptTitle}>rtylr · Terrace</div>
                <div className={styles.receiptDate}>Apr 18 · 7:42 PM · Server: Maya</div>
                <div className={styles.receiptDivider} />
                <ul className={styles.receiptList}>
                  {RECEIPT_ITEMS.map((item) => (
                    <li key={item.name} className={styles.receiptItem}>
                      <span className={styles.receiptQty}>{item.qty}×</span>
                      <span className={styles.receiptName}>{item.name}</span>
                      <span className={styles.receiptPrice}>${item.price.toFixed(2)}</span>
                    </li>
                  ))}
                </ul>
                <div className={styles.receiptDivider} />
                <div className={styles.receiptSumRow}>
                  <span>Subtotal</span>
                  <span>$50.50</span>
                </div>
                <div className={styles.receiptSumRow}>
                  <span>Tax (10%)</span>
                  <span>$5.05</span>
                </div>
                <div className={styles.receiptSumRow}>
                  <span>Tip (18%)</span>
                  <span>$9.99</span>
                </div>
                <div className={styles.receiptDivider} />
                <div className={`${styles.receiptSumRow} ${styles.receiptTotal}`}>
                  <span>Total</span>
                  <span>$65.54</span>
                </div>
                <div className={styles.receiptPaid}>
                  <Wifi size={12} strokeWidth={2.5} />
                  Paid · Visa ••4891
                </div>
              </div>
              <div className={styles.receiptTear} aria-hidden="true" />
            </motion.div>
          </section>

          {/* ── FEATURE QUAD — 2×2, asymmetric typography ─────────────────── */}
          <section className={styles.featureSection}>
            <motion.div
              className={styles.featureHeader}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              <motion.span className={styles.sectionEyebrow} variants={blurFadeIn}>
                WHAT IT DOES
              </motion.span>
              <motion.h2 className={styles.sectionTitle} variants={blurFadeIn}>
                Four things every
                <br />
                register should do.
              </motion.h2>
            </motion.div>

            <motion.div
              className={styles.featureGrid}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              {FEATURES.map(({ icon: Icon, title, copy }) => (
                <motion.div key={title} className={styles.featureCard} variants={blurFadeIn}>
                  <div className={styles.featureIcon}>
                    <Icon size={20} strokeWidth={1.8} />
                  </div>
                  <h3 className={styles.featureTitle}>{title}</h3>
                  <p className={styles.featureCopy}>{copy}</p>
                </motion.div>
              ))}
            </motion.div>
          </section>

          {/* ── STATS STRIP — matching stats pattern from home ────────────── */}
          <motion.section
            className={styles.statsSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <div className={styles.statsGrid}>
              {STATS.map((stat) => (
                <motion.div key={stat.label} className={styles.statCell} variants={blurFadeIn}>
                  <span className={styles.statValue}>{stat.value}</span>
                  <span className={styles.statLabel}>{stat.label}</span>
                </motion.div>
              ))}
            </div>
          </motion.section>

          {/* ── INTEGRATED SECTION — "it talks to everything else" ─────────── */}
          <motion.section
            className={styles.integratedSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <div className={styles.integratedLeft}>
              <motion.span className={styles.sectionEyebrow} variants={blurFadeIn}>
                BUILT INTO THE REST
              </motion.span>
              <motion.h2 className={styles.sectionTitle} variants={blurFadeIn}>
                Every sale hits
                <br />
                every other app.
              </motion.h2>
              <motion.p className={styles.integratedCopy} variants={blurFadeIn}>
                Tips flow into HR payroll. Inventory counts down live in ERP. Customer purchases
                power CRM segments. Revenue rolls up into Finance. No CSV exports, no nightly syncs,
                no second source of truth.
              </motion.p>
            </div>
            <motion.ul className={styles.integratedList} variants={blurFadeIn}>
              <li>
                <Users size={16} strokeWidth={2} />
                <span>
                  <strong>HR</strong> — tips, hours, and shift close-out flow into payroll
                </span>
              </li>
              <li>
                <TrendingUp size={16} strokeWidth={2} />
                <span>
                  <strong>ERP</strong> — inventory depletes per menu item in real time
                </span>
              </li>
              <li>
                <Clock size={16} strokeWidth={2} />
                <span>
                  <strong>CRM</strong> — every receipt enriches the customer profile
                </span>
              </li>
              <li>
                <Receipt size={16} strokeWidth={2} />
                <span>
                  <strong>Finance</strong> — cash drawer reconciled against deposits nightly
                </span>
              </li>
            </motion.ul>
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
              Your register should
              <br />
              keep up. This one does.
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

export default POSPage;
