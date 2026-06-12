import React, { useState } from 'react';
import { cn } from '../../lib/cn';
import { motion, AnimatePresence } from 'motion/react';
import { useRouter } from 'next/router';
import { Check, Minus, ChevronDown } from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { pricingPageSchema } from '../../lib/structured-data';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../lib/navigation';
import { blurFadeIn, staggerBlur, viewportOnce, SPRING_SMOOTH } from '../../lib/animations';
import { SplitWords } from '../../components/motion/SplitWords';
import PricingCards from '../../components/pricing/PricingCards';
import type { BillingCycle, Plan } from '../../components/pricing/pricingData';
import styles from './pricing.module.scss';

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
      "You're prompted to upgrade to Pro before adding seat 11. Nothing breaks. Your data, reports, and configurations all carry over.",
  },
  {
    question: 'How does per-seat billing work?',
    answer:
      "You pay only for the days each seat is active. Add a server mid-month, you're billed for those days. Remove them, the remaining days credit against your next invoice. No full-month minimums.",
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

const PricingPage = (): React.JSX.Element => {
  const router = useRouter();
  const [billing, setBilling] = useState<BillingCycle>('monthly');
  const [openFaq, setOpenFaq] = useState<number | null>(null);

  const proMonthly = 6.99;

  const handleCta = (action: Plan['ctaAction']): void => {
    if (action === 'contact') {
      void router.push('/contact');
      return;
    }
    const plan = action === 'register-pro' ? 'pro' : undefined;
    const url = plan ? `${APP_URLS.auth}/register?plan=${plan}` : `${APP_URLS.auth}/register`;
    navigateExternal(url);
  };

  const toggleFaq = (index: number): void => {
    setOpenFaq(openFaq === index ? null : index);
  };

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title="Pricing | rtylr"
          description="Simple, transparent per-seat pricing for the all-in-one rtylr business platform — POS, ERP, HR, CRM, Finance & Flow. Free to start, no credit card required."
          canonical="https://rtylr.com/pricing/"
          structuredData={pricingPageSchema({
            tiers: [
              {
                name: 'Free',
                price: '0',
                description: 'All 7 apps, up to 10 seats. No credit card required.',
                url: 'https://rtylr.com/pricing/',
              },
              {
                name: 'Pro',
                price: String(proMonthly),
                description:
                  'Per seat, per day, prorated. AI assistant, priority support, seat-based billing.',
                url: 'https://rtylr.com/pricing/',
              },
              {
                name: 'Scale',
                price: 'Custom',
                description:
                  'For teams of 30+ active seats. SSO, audit logs, dedicated onboarding, SLA.',
                url: 'https://rtylr.com/contact/',
              },
            ],
            faqs: FAQ_ITEMS,
          })}
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
              text={'Pay per seat, per day.\nNot per month.'}
              delay={0.1}
            />
            <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
              All 7 apps on every plan. Daily proration: add a seat for 10 days, pay for 10 days. No
              per-location fees, no module upsells.
            </motion.p>

            <motion.div className={styles.billingToggle} variants={blurFadeIn}>
              <button
                className={cn(styles.toggleOption, billing === 'monthly' && styles.toggleActive)}
                onClick={() => setBilling('monthly')}
              >
                {billing === 'monthly' && (
                  <motion.span
                    layoutId="billingPill"
                    className={styles.togglePill}
                    transition={SPRING_SMOOTH}
                  />
                )}
                <span className={styles.toggleLabel}>Monthly</span>
              </button>
              <button
                className={cn(styles.toggleOption, billing === 'annual' && styles.toggleActive)}
                onClick={() => setBilling('annual')}
              >
                {billing === 'annual' && (
                  <motion.span
                    layoutId="billingPill"
                    className={styles.togglePill}
                    transition={SPRING_SMOOTH}
                  />
                )}
                <span className={styles.toggleLabel}>
                  Annual
                  <span className={styles.savingsBadge}>Save 20%</span>
                </span>
              </button>
            </motion.div>
          </motion.section>

          {/* Plan cards */}
          <section className={styles.plansSection}>
            <div className={styles.plansContainer}>
              <PricingCards billing={billing} onCta={handleCta} />
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
                  className={cn(styles.faqItem, openFaq === index && styles.faqItemOpen)}
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
                      className={cn(styles.faqChevron, openFaq === index && styles.faqChevronOpen)}
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
