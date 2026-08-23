import React from 'react';
import { motion } from 'motion/react';
import { useRouter } from 'next/router';
import { Check, Minus } from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { pricingPageSchema } from '../../lib/structured-data';
import DirectAnswer from '../../components/DirectAnswer/DirectAnswer';
import FaqSection from '../../components/FaqSection/FaqSection';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../lib/navigation';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import { SplitWords } from '../../components/motion/SplitWords';
import PricingCards from '../../components/pricing/PricingCards';
import type { Plan } from '../../components/pricing/pricingData';
import styles from './pricing.module.scss';

interface PlanFeature {
  name: string;
  free: boolean | string;
  pro: boolean | string;
  scale: boolean | string;
}

// What each tier includes. "Seat-based billing" used to be a row here;
// it went out with the prices, because it describes how rtylr charges
// and rtylr does not publish that yet.
const PLAN_FEATURES: PlanFeature[] = [
  { name: 'All 7 apps', free: true, pro: true, scale: true },
  { name: 'AI assistant', free: false, pro: true, scale: true },
  { name: 'Priority support', free: false, pro: true, scale: true },
  { name: 'SSO', free: false, pro: false, scale: true },
  { name: 'Audit logs', free: false, pro: false, scale: true },
  { name: 'Dedicated onboarding', free: false, pro: false, scale: true },
  { name: 'SLA guarantee', free: false, pro: false, scale: true },
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
const FAQ_ITEMS = [
  {
    question: 'How much does rtylr cost?',
    answer:
      'rtylr has not published its prices yet. The Free plan is genuinely free and needs no card, so you can run a location on it today. For Pro or Scale, contact us and we will talk through your team size and locations. We would rather quote you honestly than put a number here we might change.',
  },
  {
    question: 'Can I switch plans at any time?',
    answer:
      'Yes. Upgrade and the new features unlock instantly. Downgrade and the change applies at the end of the current period. Your data, team, and settings stay put.',
  },
  {
    question: 'Is there a free trial for Pro?',
    answer:
      'Every new workspace starts with a 14-day Pro trial, and it needs no card. At day 14 it reverts to the Free plan automatically unless you tell us otherwise.',
  },
  {
    question: 'What happens when I hit 10 seats on Free?',
    answer:
      "You're prompted to move to Pro before adding seat 11. Nothing breaks. Your data, reports, and configurations all carry over.",
  },
  {
    question: 'When does Scale apply?',
    answer:
      'Scale is built for teams past 30 active seats. It removes the seat and location limits and adds SSO, audit logs, a dedicated success manager and an SLA.',
  },
  {
    question: 'Any contract or lock-in?',
    answer:
      'None. You can cancel at any time, and your data is exportable while the account is open.',
  },
];

const FeatureCell = ({ value }: { value: boolean | string }): React.JSX.Element => {
  if (value === true) {
    return (
      <span className={styles.checkIcon} aria-label="Included">
        <Check size={16} strokeWidth={2.5} />
      </span>
    );
  }
  if (value === false) {
    return (
      <span className={styles.dashIcon} aria-label="Not included">
        <Minus size={16} strokeWidth={2} />
      </span>
    );
  }
  return <span className={styles.featureText}>{value}</span>;
};

const PricingPage = (): React.JSX.Element => {
  const router = useRouter();

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
          title="Pricing | rtylr"
          description="rtylr plans and what each one includes: POS, ERP, HR, CRM, Finance and Flow on every tier. Start free with no card. Prices are not published yet, so talk to us for a quote."
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
              text={'All 7 apps.\nEvery plan.'}
              delay={0.1}
            />
            <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
              No module upsells and no per-app licence. What changes between plans is scale.
            </motion.p>
          </motion.section>

          <DirectAnswer label="What do rtylr plans include?">
            Every rtylr plan includes all seven apps: POS, ERP, CRM, HR, Finance, Flow and the
            hardware integrations. Plans differ by scale rather than by feature, so what changes is
            how many people and locations you run and the controls a larger team needs. Free covers
            10 people at one location and needs no card. rtylr has not published prices for Pro and
            Scale yet, so those are quoted.
          </DirectAnswer>

          {/* Plan cards */}
          <section className={styles.plansSection}>
            <div className={styles.plansContainer}>
              {/* The plan names inside PricingCards are h3. Without an h2
                  here the page jumped h1 to h3. */}
              <h2 className={styles.plansTitle}>What is in each plan?</h2>
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
              What changes between plans?
            </motion.h2>
            <motion.div className={styles.tableWrapper} variants={blurFadeIn}>
              <table className={styles.comparisonTable}>
                <thead>
                  <tr>
                    <th className={styles.featureCol}>Feature</th>
                    <th>Free</th>
                    <th className={styles.proCol}>Pro</th>
                    <th>Scale</th>
                  </tr>
                </thead>
                <tbody>
                  {PLAN_FEATURES.map((feature) => (
                    <tr key={feature.name}>
                      <td className={styles.featureName}>{feature.name}</td>
                      <td>
                        <FeatureCell value={feature.free} />
                      </td>
                      <td className={styles.proCol}>
                        <FeatureCell value={feature.pro} />
                      </td>
                      <td>
                        <FeatureCell value={feature.scale} />
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
          <FaqSection title="Questions about rtylr pricing" faqs={FAQ_ITEMS} />

          {/* Bottom CTA */}
          <motion.section
            className={styles.bottomCta}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.h2 className={styles.bottomCtaTitle} variants={blurFadeIn}>
              Start on Free. Move up when you outgrow it.
            </motion.h2>
            <motion.p className={styles.bottomCtaSubtitle} variants={blurFadeIn}>
              10 seats and all 7 apps are yours from day one, with no card. Book a demo if you would
              rather see it running on your own numbers first.
            </motion.p>
            <motion.div className={styles.bottomCtaButtons} variants={blurFadeIn}>
              <button className={styles.ctaPrimary} onClick={() => handleCta('register')}>
                Start free
              </button>
              <button className={styles.ctaSecondary} onClick={() => void router.push('/demo')}>
                Book a demo
              </button>
            </motion.div>
          </motion.section>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default PricingPage;
