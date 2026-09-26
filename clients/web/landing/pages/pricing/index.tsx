import React from 'react';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { useRouter } from 'next/router';
import { Check, Minus } from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { pricingPageSchema } from '../../lib/structured-data';
import DirectAnswer from '../../components/DirectAnswer/DirectAnswer';
import FaqSection from '../../components/FaqSection/FaqSection';
import { useFaq } from '../../lib/faqs';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../lib/navigation';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import { SplitWords } from '../../components/motion/SplitWords';
import PricingCards from '../../components/pricing/PricingCards';
import type { Plan } from '../../components/pricing/pricingData';
import styles from './pricing.module.scss';

interface PlanFeature {
  free: boolean | string;
  pro: boolean | string;
  scale: boolean | string;
}

// Which tier includes what. The row LABELS live in the locale files under
// pricing.compare.rows and are matched to these by index, so the array here
// and the array there must stay the same length and order. A truth table is
// not copy, which is why it did not move.
//
// "Seat-based billing" used to be a row. It went out with the prices,
// because it describes how rtylr charges and rtylr does not publish that.
const PLAN_FEATURES: PlanFeature[] = [
  { free: true, pro: true, scale: true },
  { free: false, pro: true, scale: true },
  { free: false, pro: true, scale: true },
  { free: false, pro: false, scale: true },
  { free: false, pro: false, scale: true },
  { free: false, pro: false, scale: true },
  { free: false, pro: false, scale: true },
];

// These render on the page and are not marked up. There was an FAQPage
// block built from this same array; it went out in Phase 1 because Google
// removed the FAQ rich result in May 2026, and because the schema had
// become the only place the answers existed in the served HTML.
//
// Two questions were removed and two rewritten when the placeholder
// prices came out: "How does per-seat billing work?" described a billing
// model rtylr has not committed to, and the answers about annual
// discounts and cheaper Scale seats were price claims with no price on
// the page to back them. The first question below replaces them, because
// a visitor who came here for a number deserves a straight answer about
// why there isn't one.
// aria-label is read aloud, so "Included" is copy like any other.
const FeatureCell = ({
  value,
  labels,
}: {
  value: boolean | string;
  labels: { included: string; notIncluded: string };
}): React.JSX.Element => {
  if (value === true) {
    return (
      <span className={styles.checkIcon} role="img" aria-label={labels.included}>
        <Check size={16} strokeWidth={2.5} />
      </span>
    );
  }
  if (value === false) {
    return (
      <span className={styles.dashIcon} role="img" aria-label={labels.notIncluded}>
        <Minus size={16} strokeWidth={2} />
      </span>
    );
  }
  return <span className={styles.featureText}>{value}</span>;
};

const PricingPage = (): React.JSX.Element => {
  const router = useRouter();
  const { t } = useTranslation();

  const compareRows = t('pricing.compare.rows', { returnObjects: true }) as unknown as string[];
  const cellLabels = {
    included: t('pricing.compare.included'),
    notIncluded: t('pricing.compare.notIncluded'),
  };
  const faq = useFaq('pricing');

  const handleCta = (action: Plan['ctaAction']): void => {
    if (action === 'contact') {
      void router.push('/contact');
      return;
    }
    const plan = action === 'register-pro' ? 'pro' : undefined;
    const url = plan ? `${APP_URLS.auth}/register?plan=${plan}` : `${APP_URLS.auth}/register`;
    navigateExternal(url);
  };

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title={t('pricing.meta.title')}
          description={t('pricing.meta.description')}
          canonical="https://rtylr.com/pricing/"
          structuredData={pricingPageSchema()}
        />

        <div className={styles.pricingPage}>
          {/* Hero */}
          <motion.section
            className={styles.hero}
            initial="hidden"
            animate="visible"
            variants={staggerBlur}
          >
            <SplitWords
              as="h1"
              className={styles.heroTitle}
              text={t('pricing.hero.title')}
              delay={0.1}
            />
            <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
              {t('pricing.hero.subtitle')}
            </motion.p>
          </motion.section>

          <DirectAnswer label={t('pricing.directAnswer.label')}>
            {t('pricing.directAnswer.body')}
          </DirectAnswer>

          {/* Plan cards */}
          <section className={styles.plansSection}>
            <div className={styles.plansContainer}>
              {/* The plan names inside PricingCards are h3. Without an h2
                  here the page jumped h1 to h3. */}
              <h2 className={styles.plansTitle}>{t('pricing.plans.title')}</h2>
              <PricingCards onCta={handleCta} />
            </div>
          </section>

          {/* Feature comparison table */}
          <motion.section
            className={styles.tableSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.h2 className={styles.sectionTitle} variants={blurFadeIn}>
              {t('pricing.compare.title')}
            </motion.h2>
            <motion.div className={styles.tableWrapper} variants={blurFadeIn}>
              <table className={styles.comparisonTable}>
                <thead>
                  <tr>
                    <th className={styles.featureCol}>{t('pricing.compare.feature')}</th>
                    <th>{t('pricing.plans.free.name')}</th>
                    <th className={styles.proCol}>{t('pricing.plans.pro.name')}</th>
                    <th>{t('pricing.plans.scale.name')}</th>
                  </tr>
                </thead>
                <tbody>
                  {PLAN_FEATURES.map((feature, i) => (
                    <tr key={compareRows[i]}>
                      <td className={styles.featureName}>{compareRows[i]}</td>
                      <td>
                        <FeatureCell value={feature.free} labels={cellLabels} />
                      </td>
                      <td className={styles.proCol}>
                        <FeatureCell value={feature.pro} labels={cellLabels} />
                      </td>
                      <td>
                        <FeatureCell value={feature.scale} labels={cellLabels} />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </motion.div>
          </motion.section>

          {/*
            The hand-rolled accordion this replaces mounted an answer only
            while it was open, so the answers reached no crawler, no
            markdown twin and no reader without JavaScript. The single copy
            of that text in the served bytes was inside the FAQPage JSON-LD,
            which is now gone, so the visible section is the delivery
            mechanism. FaqSection uses <details>: every answer is in the
            DOM, collapsed by CSS.
          */}
          <FaqSection title={faq.title} faqs={faq.items} />

          {/* Bottom CTA */}
          <motion.section
            className={styles.bottomCta}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.h2 className={styles.bottomCtaTitle} variants={blurFadeIn}>
              {t('pricing.bottom.title')}
            </motion.h2>
            <motion.p className={styles.bottomCtaSubtitle} variants={blurFadeIn}>
              {t('pricing.bottom.subtitle')}
            </motion.p>
            <motion.div className={styles.bottomCtaButtons} variants={blurFadeIn}>
              <button className={styles.ctaPrimary} onClick={() => handleCta('register')}>
                {t('pricing.plans.free.cta')}
              </button>
              <button className={styles.ctaSecondary} onClick={() => void router.push('/demo')}>
                {t('common.cta.bookDemo')}
              </button>
            </motion.div>
          </motion.section>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default PricingPage;
