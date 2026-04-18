import React, { useState } from 'react';
import { motion, AnimatePresence } from 'motion/react';
import { useRouter } from 'next/router';
import { Check, Minus, ChevronDown } from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../lib/navigation';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import styles from './pricing.module.scss';

type BillingCycle = 'monthly' | 'annual';

interface PlanFeature {
  name: string;
  free: boolean | string;
  pro: boolean | string;
  scale: boolean | string;
}

const PLAN_FEATURES: PlanFeature[] = [
  { name: 'All 7 apps', free: true, pro: true, scale: true },
  { name: 'AI assistant', free: false, pro: true, scale: true },
  { name: 'Seat-based billing', free: false, pro: true, scale: true },
  { name: 'Priority support', free: false, pro: true, scale: true },
  { name: 'SSO', free: false, pro: false, scale: true },
  { name: 'Audit logs', free: false, pro: false, scale: true },
  { name: 'Dedicated onboarding', free: false, pro: false, scale: true },
  { name: 'SLA guarantee', free: false, pro: false, scale: true },
];

const FAQ_ITEMS = [
  {
    question: 'Can I switch plans?',
    answer:
      'Yes. You can upgrade or downgrade at any time. When you upgrade, you get immediate access to the new features. When you downgrade, your plan changes at the next billing cycle.',
  },
  {
    question: 'Is there a free trial for Pro?',
    answer:
      'Yes — every new workspace starts with a 14-day Pro trial, no credit card required. After 14 days it automatically switches to the Free plan unless you add a payment method.',
  },
  {
    question: 'What happens when I reach 5+ members on Free?',
    answer:
      'The Free plan supports up to 5 seats. Once you need a 6th member, you will be prompted to upgrade to the Pro plan. All your data and settings are preserved.',
  },
  {
    question: 'Do I pay per seat?',
    answer:
      'Yes, the Pro plan is seat-based. You pay for each active user in your workspace. You can add or remove seats at any time and your billing adjusts automatically.',
  },
  {
    question: 'Is there a contract or lock-in?',
    answer:
      'No contracts, no lock-in. Monthly plans can be cancelled any time. Annual plans are billed upfront for the year and you save 20% compared to monthly billing.',
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

const PricingPage = (): JSX.Element => {
  const router = useRouter();
  const [billing, setBilling] = useState<BillingCycle>('monthly');
  const [openFaq, setOpenFaq] = useState<number | null>(null);

  const proMonthly = 4.99;
  const proAnnual = 3.99;
  const proPrice = billing === 'annual' ? proAnnual : proMonthly;

  const handleGetStarted = (): void => {
    navigateExternal(`${APP_URLS.auth}/register`);
  };

  const handleStartPro = (): void => {
    navigateExternal(`${APP_URLS.auth}/register?plan=pro`);
  };

  const handleContactSales = (): void => {
    void router.push('/contact');
  };

  const handleBookDemo = (): void => {
    void router.push('/demo');
  };

  const toggleFaq = (index: number): void => {
    setOpenFaq(openFaq === index ? null : index);
  };

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title="Pricing | rtylr"
          description="Simple, transparent pricing for every team. Free to start. No credit card required."
        />

        <div className={styles.pricingPage}>
          {/* Hero */}
          <motion.section
            className={styles.hero}
            initial="hidden"
            animate="visible"
            variants={staggerBlur}
          >
            <motion.h1 className={styles.heroTitle} variants={blurFadeIn}>
              Simple pricing for every team
            </motion.h1>
            <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
              Free to start. No credit card required. Scale as you grow.
            </motion.p>

            {/* Billing toggle */}
            <motion.div className={styles.billingToggle} variants={blurFadeIn}>
              <button
                className={`${styles.toggleOption} ${billing === 'monthly' ? styles.toggleActive : ''}`}
                onClick={() => setBilling('monthly')}
              >
                Monthly
              </button>
              <button
                className={`${styles.toggleOption} ${billing === 'annual' ? styles.toggleActive : ''}`}
                onClick={() => setBilling('annual')}
              >
                Annual
                <span className={styles.savingsBadge}>Save 20%</span>
              </button>
            </motion.div>
          </motion.section>

          {/* Plan cards */}
          <section className={styles.plansSection}>
            <div className={styles.plansContainer}>
              {/* Free */}
              <motion.div
                className={styles.planCard}
                initial="hidden"
                whileInView="visible"
                viewport={viewportOnce}
                variants={blurFadeIn}
              >
                <div className={styles.planHeader}>
                  <span className={styles.planName}>Free</span>
                  <div className={styles.planPrice}>
                    <span className={styles.priceAmount}>$0</span>
                    <span className={styles.pricePeriod}>/forever</span>
                  </div>
                  <p className={styles.planTagline}>Perfect for small teams getting started.</p>
                </div>
                <ul className={styles.planFeatures}>
                  <li>Up to 5 members</li>
                  <li>All 7 apps included</li>
                  <li>Basic support</li>
                  <li>No credit card required</li>
                </ul>
                <button className={styles.planCta} onClick={handleGetStarted}>
                  Get started
                </button>
              </motion.div>

              {/* Pro */}
              <motion.div
                className={`${styles.planCard} ${styles.planCardPro}`}
                initial="hidden"
                whileInView="visible"
                viewport={viewportOnce}
                variants={blurFadeIn}
              >
                <div className={styles.popularBadge}>Most Popular</div>
                <div className={styles.planHeader}>
                  <span className={styles.planName}>Pro</span>
                  <div className={styles.planPrice}>
                    <span className={styles.priceAmount}>${proPrice.toFixed(2)}</span>
                    <span className={styles.pricePeriod}>/seat/month</span>
                  </div>
                  {billing === 'annual' && (
                    <p className={styles.annualNote}>
                      Billed annually · ${(proAnnual * 12).toFixed(2)}/seat/year
                    </p>
                  )}
                  <p className={styles.planTagline}>For growing teams that need more.</p>
                </div>
                <ul className={styles.planFeatures}>
                  <li>Teams of any size</li>
                  <li>All 7 apps + AI assistant</li>
                  <li>Priority support</li>
                  <li>Seat-based billing</li>
                </ul>
                <button
                  className={`${styles.planCta} ${styles.planCtaPro}`}
                  onClick={handleStartPro}
                >
                  Start with Pro
                </button>
              </motion.div>

              {/* Scale */}
              <motion.div
                className={styles.planCard}
                initial="hidden"
                whileInView="visible"
                viewport={viewportOnce}
                variants={blurFadeIn}
              >
                <div className={styles.planHeader}>
                  <span className={styles.planName}>Scale</span>
                  <div className={styles.planPrice}>
                    <span className={styles.priceAmount}>Custom</span>
                  </div>
                  <p className={styles.planTagline}>For enterprises with advanced needs.</p>
                </div>
                <ul className={styles.planFeatures}>
                  <li>100+ seats</li>
                  <li>Enterprise features</li>
                  <li>SSO &amp; Audit logs</li>
                  <li>Dedicated support &amp; SLA</li>
                </ul>
                <button className={styles.planCta} onClick={handleContactSales}>
                  Contact sales
                </button>
              </motion.div>
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
              Compare all features
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

          {/* FAQ */}
          <motion.section
            className={styles.faqSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.h2 className={styles.sectionTitle} variants={blurFadeIn}>
              Frequently asked questions
            </motion.h2>
            <motion.div className={styles.faqList} variants={staggerBlur}>
              {FAQ_ITEMS.map((item, index) => (
                <motion.div
                  key={index}
                  className={`${styles.faqItem} ${openFaq === index ? styles.faqItemOpen : ''}`}
                  variants={blurFadeIn}
                >
                  <button
                    className={styles.faqQuestion}
                    onClick={() => toggleFaq(index)}
                    aria-expanded={openFaq === index}
                  >
                    <span>{item.question}</span>
                    <ChevronDown
                      size={18}
                      className={`${styles.faqChevron} ${openFaq === index ? styles.faqChevronOpen : ''}`}
                    />
                  </button>
                  <AnimatePresence>
                    {openFaq === index && (
                      <motion.div
                        className={styles.faqAnswer}
                        initial={{ height: 0, opacity: 0 }}
                        animate={{ height: 'auto', opacity: 1 }}
                        exit={{ height: 0, opacity: 0 }}
                        transition={{ duration: 0.3, ease: [0.25, 0.46, 0.45, 0.94] }}
                      >
                        <p>{item.answer}</p>
                      </motion.div>
                    )}
                  </AnimatePresence>
                </motion.div>
              ))}
            </motion.div>
          </motion.section>

          {/* Bottom CTA */}
          <motion.section
            className={styles.bottomCta}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.h2 className={styles.bottomCtaTitle} variants={blurFadeIn}>
              Ready to get started?
            </motion.h2>
            <motion.p className={styles.bottomCtaSubtitle} variants={blurFadeIn}>
              Join thousands of businesses already running on rtylr.
            </motion.p>
            <motion.div className={styles.bottomCtaButtons} variants={blurFadeIn}>
              <button className={styles.ctaPrimary} onClick={handleGetStarted}>
                Start free
              </button>
              <button className={styles.ctaSecondary} onClick={handleBookDemo}>
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
