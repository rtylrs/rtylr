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
    question: 'Can I switch plans at any time?',
    answer:
      'Yes. Upgrade mid-cycle and new features unlock instantly. Downgrade and the change applies at the end of the current period. Your data, team, and settings stay put.',
  },
  {
    question: 'Is there a free trial for Pro?',
    answer:
      'Every new workspace starts with a 14-day Pro trial, no credit card required. At day 14 it reverts to the Free plan automatically unless you add a payment method.',
  },
  {
    question: 'What happens when I hit 10 seats on Free?',
    answer:
      'You\u2019re prompted to upgrade to Pro before adding seat 11. Nothing breaks. Your data, reports, and configurations all carry over.',
  },
  {
    question: 'How does per-seat billing work?',
    answer:
      'You pay only for the days each seat is active. Add a server mid-month, you\u2019re billed for those days. Remove them, the remaining days credit against your next invoice. No full-month minimums.',
  },
  {
    question: 'When does Scale pricing kick in?',
    answer:
      'Scale is designed for teams of 30+ active seats. Once your workspace crosses that threshold for a billing period, Scale pricing applies automatically: cheaper per seat, more locations, SSO, and audit logs.',
  },
  {
    question: 'Any contract or lock-in?',
    answer:
      'None. Monthly plans cancel any time. Annual plans are billed upfront and save 20%, but you can still cancel. Unused days credit against your next bill.',
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

  const proMonthly = 6.99;
  const proAnnual = 4.99;
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
          canonical="https://rtylr.com/pricing/"
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
              Pay per seat, per day.
              <br />
              Not per month.
            </motion.h1>
            <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
              All 7 apps on every plan. Daily proration: add a seat for 10 days, pay for 10 days. No
              per-location fees, no module upsells.
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
                  <p className={styles.planTagline}>
                    Everything you need to run a single location.
                  </p>
                </div>
                <ul className={styles.planFeatures}>
                  <li>Up to 10 team members</li>
                  <li>All 7 apps included</li>
                  <li>1 location</li>
                  <li>0% transaction fees under $10k/mo</li>
                  <li>1,000 AI credits / month</li>
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
                  <p className={styles.planTagline}>
                    Growing teams scaling across multiple locations.
                  </p>
                </div>
                <ul className={styles.planFeatures}>
                  <li>Up to 30 team members</li>
                  <li>Up to 5 locations</li>
                  <li>0% transaction fees under $100k/mo</li>
                  <li>10,000 AI credits / month + 500 per seat</li>
                  <li>Priority email + chat support</li>
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
                    <span className={styles.priceAmount}>$4.99</span>
                    <span className={styles.pricePeriod}>/seat/month</span>
                  </div>
                  {billing === 'annual' && (
                    <p className={styles.annualNote}>
                      $2.99/seat on annual · unlocks at 30+ active seats
                    </p>
                  )}
                  <p className={styles.planTagline}>
                    Unlimited seats. Every feature unlocked. Every app. Every location.
                  </p>
                </div>
                <ul className={styles.planFeatures}>
                  <li>Unlimited team members</li>
                  <li>Unlimited locations</li>
                  <li>0% transaction fees (no cap)</li>
                  <li>50,000 AI credits / month + 1,000 per seat</li>
                  <li>SSO, audit logs, dedicated success manager</li>
                </ul>
                <button className={styles.planCta} onClick={handleContactSales}>
                  Talk to sales
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
              Start on Free. Pay only when you grow.
            </motion.h2>
            <motion.p className={styles.bottomCtaSubtitle} variants={blurFadeIn}>
              10 seats and all 7 apps are yours from day one. No credit card until you need seat 11.
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
