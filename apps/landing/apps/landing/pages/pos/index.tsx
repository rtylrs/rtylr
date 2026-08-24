import React from 'react';
import { cn } from '../../lib/cn';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { useRouter } from 'next/router';
import { Receipt, CreditCard, Split, Wifi, WifiOff } from 'lucide-react';
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
import { formatMoney } from '../../lib/formatMoney';
import styles from './pos.module.scss';

// ── Static receipt items shown in the hero mockup ──────────────────────────
// itemIndex points at pos.hero.receipt.items[i] for the display name.
const RECEIPT_ITEMS = [
  { itemIndex: 0, qty: 2, price: 28.0 },
  { itemIndex: 1, qty: 1, price: 9.5 },
  { itemIndex: 2, qty: 2, price: 6.0 },
  { itemIndex: 3, qty: 2, price: 7.0 },
];

const FEATURE_ICONS = [Receipt, CreditCard, Split, WifiOff];

// Which product supports what, as codes. Column order is rtylr, Toast,
// Square, Shopify POS. Row order matches pos.compare.rows in the locale
// files, so the two arrays must stay the same length: the i18n gate checks
// that a translated list keeps its source's length for exactly this reason.
//
// A "Per-location monthly fee" row used to sit second, reading $0 for rtylr
// against $69 to $165 and $89/register for two competitors. It came out
// with the rest of the placeholder pricing: rtylr publishes no price, so $0
// was the same false claim being removed everywhere else, and quoting a
// competitor's rate undated and uncited is the one thing the comparison
// rules here forbid outright.
const COMPARE_VALUES: string[][] = [
  ['yes', 'no', 'partial', 'partial'],
  ['no', 'yes', 'yes', 'yes'],
  ['yes', 'no', 'partial', 'partial'],
  ['yes', 'add-on', 'add-on', 'no'],
  ['yes', 'add-on', 'no', 'no'],
  ['yes', 'add-on', 'partial', 'no'],
  ['yes', 'partial', 'yes', 'yes'],
];

const POSPage = (): React.JSX.Element => {
  const faq = useFaq('pos');
  const router = useRouter();
  const { t, i18n } = useTranslation();
  const receiptItemNames = t('pos.hero.receipt.items', { returnObjects: true }) as string[];
  const pct = (n: number): string => `${n}${i18n.language === 'fr' ? ' %' : '%'}`;

  const features = t('pos.features.items', { returnObjects: true }) as unknown as {
    title: string;
    copy: string;
  }[];
  const compareRows = t('pos.compare.rows', { returnObjects: true }) as unknown as string[];
  const cell = {
    yes: t('pos.compare.cell.included'),
    no: t('pos.compare.cell.excluded'),
    partial: t('pos.compare.cell.partial'),
    addon: t('pos.compare.cell.paidAddon'),
  };

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
          title={t('pos.meta.title')}
          description={t('pos.meta.description')}
          canonical="https://rtylr.com/pos/"
          structuredData={productWithBreadcrumbs({
            name: 'POS',
            path: '/pos/',
            subCategory: 'Point of Sale',
            description: t('pos.meta.schemaDescription'),
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
                {t('pos.hero.eyebrow')}
              </motion.span>
              <SplitWords
                as="h1"
                className={styles.heroTitle}
                text={t('pos.hero.title')}
                delay={0.1}
              />
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                {t('pos.hero.subtitle')}
              </motion.p>
              <motion.div className={styles.heroCtas} variants={blurFadeIn}>
                <button type="button" onClick={goRegister} className={styles.primaryBtn}>
                  {t('common.cta.getStarted')}
                </button>
                <button type="button" onClick={goDemo} className={styles.outlineBtn}>
                  {t('common.cta.seeItLive')}
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
                <span className={styles.receiptOrder} dir="ltr">
                  {t('pos.hero.receipt.order', { orderId: 'A-2048', table: 12 })}
                </span>
                <span className={styles.receiptLive}>
                  <span className={styles.livePulse} />
                  {t('pos.hero.receipt.live')}
                </span>
              </div>
              <div className={styles.receiptBody}>
                <div className={styles.receiptTitle}>{t('pos.hero.receipt.venue')}</div>
                <div className={styles.receiptDate}>
                  {t('pos.hero.receipt.date')} · <span dir="ltr">19:42</span> ·{' '}
                  {t('pos.hero.receipt.server', { name: 'Maya' })}
                </div>
                <div className={styles.receiptDivider} />
                <ul className={styles.receiptList}>
                  {RECEIPT_ITEMS.map((item) => (
                    <li key={item.itemIndex} className={styles.receiptItem}>
                      <span className={styles.receiptQty} dir="ltr">
                        {item.qty}×
                      </span>
                      <span className={styles.receiptName}>{receiptItemNames[item.itemIndex]}</span>
                      <span className={styles.receiptPrice}>
                        {formatMoney(item.price, i18n.language)}
                      </span>
                    </li>
                  ))}
                </ul>
                <div className={styles.receiptDivider} />
                <div className={styles.receiptSumRow}>
                  <span>{t('pos.hero.receipt.subtotal')}</span>
                  <span>{formatMoney(50.5, i18n.language)}</span>
                </div>
                <div className={styles.receiptSumRow}>
                  <span>{t('pos.hero.receipt.tax', { pct: pct(10) })}</span>
                  <span>{formatMoney(5.05, i18n.language)}</span>
                </div>
                <div className={styles.receiptSumRow}>
                  <span>{t('pos.hero.receipt.tip', { pct: pct(18) })}</span>
                  <span>{formatMoney(9.99, i18n.language)}</span>
                </div>
                <div className={styles.receiptDivider} />
                <div className={cn(styles.receiptSumRow, styles.receiptTotal)}>
                  <span>{t('pos.hero.receipt.total')}</span>
                  <span>{formatMoney(65.54, i18n.language)}</span>
                </div>
                <div className={styles.receiptPaid}>
                  <Wifi size={12} strokeWidth={2.5} />
                  {t('pos.hero.receipt.paid')}
                </div>
              </div>
              <div className={styles.receiptTear} aria-hidden="true" />
            </motion.div>
          </section>

          <DirectAnswer label={t('pos.directAnswer.label')}>
            {t('pos.directAnswer.body')}
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
                {t('pos.features.eyebrow')}
              </motion.span>
              <motion.h2 className={styles.sectionTitle} variants={blurFadeIn}>
                <Lines text={t('pos.features.title')} />
              </motion.h2>
            </motion.div>

            <motion.div
              className={styles.featureGrid}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              {features.map((feature, i) => {
                const Icon = FEATURE_ICONS[i];
                return (
                  <motion.div
                    key={feature.title}
                    className={styles.featureCard}
                    variants={blurFadeIn}
                  >
                    <div className={styles.featureIcon}>
                      <Icon size={20} strokeWidth={1.8} />
                    </div>
                    <h3 className={styles.featureTitle}>{feature.title}</h3>
                    <p className={styles.featureCopy}>{feature.copy}</p>
                  </motion.div>
                );
              })}
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
              <span className={styles.sectionEyebrow}>{t('pos.compare.eyebrow')}</span>
              <h2 className={styles.sectionTitle}>
                <Lines text={t('pos.compare.title')} />
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
                  {COMPARE_VALUES.map((values, rowIndex) => (
                    <tr key={compareRows[rowIndex]}>
                      <th scope="row">{compareRows[rowIndex]}</th>
                      {values.map((v, i) => {
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
                            {variant === 'text' ? v : cell[variant]}
                          </td>
                        );
                      })}
                    </tr>
                  ))}
                </tbody>
              </table>
              <p className={styles.compareFootnote}>
                {t('pos.compare.footnote')}{' '}
                <a href="mailto:hello@voxire.com">{t('pos.compare.footnoteLink')}</a>.
              </p>
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
              <Lines text={t('pos.cta.title')} />
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

export default POSPage;
