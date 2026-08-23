import React from 'react';
import { cn } from '../../lib/cn';
import { motion } from 'motion/react';
import { useRouter } from 'next/router';
import { Receipt, CreditCard, Split, Wifi, WifiOff } from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import DirectAnswer from '../../components/DirectAnswer/DirectAnswer';
import FaqSection from '../../components/FaqSection/FaqSection';
import { productWithBreadcrumbs } from '../../lib/structured-data';
import { POS_FAQS } from '../../lib/faqs';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../lib/navigation';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import { SplitWords } from '../../components/motion/SplitWords';
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

const POSPage = (): JSX.Element => {
  const router = useRouter();

  const goRegister = (): void => {
    navigateExternal(`${APP_URLS.auth}/register?redirect=pos`);
  };

  const goDemo = (): void => {
    void router.push('/demo');
  };

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title="POS · The register that keeps up. | rtylr"
          description="Full-service POS for restaurants and retail. Tickets, splits, tips, and a real offline mode. Built into the rtylr platform."
          canonical="https://rtylr.com/pos/"
          structuredData={productWithBreadcrumbs({
            name: 'POS',
            path: '/pos/',
            subCategory: 'Point of Sale',
            description:
              'Full-service POS for restaurants and retail. Tickets, splits, tips, and a real offline mode.',
          })}
        />

        <div className={styles.page}>
          {/* ── HERO, split layout: copy left, receipt mockup right ───────── */}
          <section className={styles.hero}>
            <div className={styles.heroFrame} aria-hidden="true">
              <span className={cn(styles.vLine, styles.vLineLeft)} />
              <span className={cn(styles.vLine, styles.vLineRight)} />
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
              <SplitWords
                as="h1"
                className={styles.heroTitle}
                text={'The register\nthat keeps up.'}
                delay={0.1}
              />
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                Fire courses to the right station. Split a check seven ways. Take Apple Pay, gift
                cards, house accounts. Keep running when the WiFi goes down. Built for the pace of
                real service, not a demo video.
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

            {/* Receipt mockup, thermal-printer aesthetic */}
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
                <div className={cn(styles.receiptSumRow, styles.receiptTotal)}>
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

          <DirectAnswer label="What is rtylr POS?">
            rtylr POS is the point-of-sale app inside rtylr, built for restaurants and retail businesses
            running one location or twenty. It handles tickets, course firing, check splits, tips,
            modifiers and card payments, and it keeps taking orders when the connection drops. Because it
            writes to the same database as inventory, staff and accounting, a sale moves stock and revenue
            without an export.
          </DirectAnswer>

          {/* ── FEATURE QUAD, 2×2, asymmetric typography ─────────────────── */}
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

          {/* ── COMPARISON TABLE, vs Toast / Square / Shopify POS ────────── */}
          <motion.section
            className={styles.compareSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.div className={styles.compareHeader} variants={blurFadeIn}>
              <span className={styles.sectionEyebrow}>VS THE OTHERS</span>
              <h2 className={styles.sectionTitle}>
                What every other register
                <br />
                quietly charges for.
              </h2>
            </motion.div>

            <motion.div className={styles.compareTableWrap} variants={blurFadeIn}>
              <table className={styles.compareTable}>
                <thead>
                  <tr>
                    <th scope="col" />
                    <th scope="col" className={styles.compareUs}>
                      <span className={styles.compareHeadLogo}>rtylr POS</span>
                    </th>
                    <th scope="col">Toast</th>
                    <th scope="col">Square</th>
                    <th scope="col">Shopify POS</th>
                  </tr>
                </thead>
                <tbody>
                  {[
                    {
                      feature: 'Offline mode (full service, no internet)',
                      values: ['yes', 'no', 'partial', 'partial'],
                    },
                    // A "Per-location monthly fee" row used to sit here,
                    // reading $0 for rtylr against $69 to $165 and
                    // $89/register for two competitors. It came out with the
                    // rest of the placeholder pricing: rtylr publishes no
                    // price, so $0 was the same false claim being removed
                    // everywhere else, and quoting a competitor's rate
                    // undated and uncited is the one thing the comparison
                    // rules here forbid outright. A priced comparison needs
                    // their published pricing page, cited, with the date read.
                    {
                      feature: 'Payment processing locked in',
                      values: ['no', 'yes', 'yes', 'yes'],
                    },
                    {
                      feature: 'Hardware you can buy elsewhere',
                      values: ['yes', 'no', 'partial', 'partial'],
                    },
                    {
                      feature: 'HR + payroll included',
                      values: ['yes', 'add-on', 'add-on', 'no'],
                    },
                    {
                      feature: 'Inventory depletes per recipe',
                      values: ['yes', 'add-on', 'no', 'no'],
                    },
                    {
                      feature: 'Real-time CRM tied to receipts',
                      values: ['yes', 'add-on', 'partial', 'no'],
                    },
                    {
                      feature: 'Public REST API v1',
                      values: ['yes', 'partial', 'yes', 'yes'],
                    },
                  ].map((row) => (
                    <tr key={row.feature}>
                      <th scope="row">{row.feature}</th>
                      {row.values.map((v, i) => {
                        const isUs = i === 0;
                        const variant =
                          v === 'yes'
                            ? 'yes'
                            : v === 'no'
                              ? 'no'
                              : v === 'partial'
                                ? 'partial'
                                : v === 'add-on'
                                  ? 'addon'
                                  : 'text';
                        return (
                          <td
                            key={i}
                            className={cn(
                              isUs && styles.compareUs,
                              variant === 'yes'
                                ? styles.cellYes
                                : variant === 'no'
                                  ? styles.cellNo
                                  : variant === 'partial'
                                    ? styles.cellPartial
                                    : variant === 'addon'
                                      ? styles.cellAddon
                                      : styles.cellText
                            )}
                          >
                            {variant === 'yes'
                              ? 'Included'
                              : variant === 'no'
                                ? '-'
                                : variant === 'partial'
                                  ? 'Partial'
                                  : variant === 'addon'
                                    ? 'Paid add-on'
                                    : v}
                          </td>
                        );
                      })}
                    </tr>
                  ))}
                </tbody>
              </table>
              <p className={styles.compareFootnote}>
                Comparison based on publicly available pricing and product documentation, April
                2026. Something out of date? <a href="mailto:hello@voxire.com">Let us know</a>.
              </p>
            </motion.div>
          </motion.section>

          <FaqSection title="Questions about rtylr POS" faqs={POS_FAQS} />

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
