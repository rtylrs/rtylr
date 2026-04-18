import React from 'react';
import { motion } from 'motion/react';
import { HiCheck } from 'react-icons/hi';
import { useRouter } from 'next/router';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../../lib/navigation';
import { blurFadeIn, blurFadeInSlow, staggerBlur, viewportOnce } from '../../../lib/animations';
import styles from './PricingSection.module.scss';

interface Plan {
  name: string;
  price: string;
  priceSuffix: string;
  priceNote?: string;
  description: string;
  features: string[];
  ctaText: string;
  ctaHref: 'register' | 'register-pro' | 'register-scale' | 'contact';
  featured?: boolean;
}

const plans: Plan[] = [
  {
    name: 'Free',
    price: '$0',
    priceSuffix: 'forever',
    priceNote: 'No credit card required · cancel anytime',
    description: 'Everything you need to run a single location.',
    features: [
      'Up to 10 team members',
      'All 7 apps included',
      '1 location',
      '0% transaction fees under $10k/mo',
      '1,000 AI credits / month',
    ],
    ctaText: 'Start free',
    ctaHref: 'register',
  },
  {
    name: 'Pro',
    price: '$6.99',
    priceSuffix: '/seat/month',
    priceNote: '$4.99/seat billed annually · up to 30 seats',
    description: 'Growing teams scaling across multiple locations.',
    features: [
      'Up to 30 team members',
      'Up to 5 locations',
      '0% transaction fees under $100k/mo',
      '10,000 AI credits / month + 500 per seat',
      'Priority email + chat support',
    ],
    ctaText: 'Start with Pro',
    ctaHref: 'register-pro',
    featured: true,
  },
  {
    name: 'Scale',
    price: '$4.99',
    priceSuffix: '/seat/month',
    priceNote: '$2.99/seat billed annually · unlocks at 30+ active seats',
    description: 'Unlimited seats and every feature unlocked.',
    features: [
      'Unlimited team members',
      'Unlimited locations',
      '0% transaction fees (no cap)',
      '50,000 AI credits / month + 1,000 per seat',
      'SSO, audit logs, dedicated success manager',
    ],
    ctaText: 'Start with Scale',
    ctaHref: 'register-scale',
  },
];

const PricingSection = (): JSX.Element => {
  const router = useRouter();

  const handleCTA = (href: Plan['ctaHref']): void => {
    if (href === 'contact') {
      void router.push('/contact');
      return;
    }
    const planParam =
      href === 'register-pro' ? 'pro' : href === 'register-scale' ? 'scale' : undefined;
    const url = planParam
      ? `${APP_URLS.auth}/register?plan=${planParam}`
      : `${APP_URLS.auth}/register`;
    navigateExternal(url);
  };

  return (
    <section className={styles.pricingSection}>
      <div className={styles.container}>
        <motion.div
          className={styles.header}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={blurFadeInSlow}
        >
          <span className={styles.badge}>Pricing</span>
          <h2 className={styles.sectionTitle}>
            Simple pricing.
            <br />
            All 7 apps, one price.
          </h2>
          <p className={styles.sectionDescription}>
            Start free. Upgrade when you need more. No per-location fees, no module upsells.
          </p>
        </motion.div>

        <motion.div
          className={styles.pricingGrid}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={staggerBlur}
        >
          {plans.map((plan) => (
            <motion.div
              key={plan.name}
              className={`${styles.pricingCard} ${plan.featured ? styles.featured : ''}`}
              variants={blurFadeIn}
              whileHover={{
                y: -8,
                transition: { duration: 0.25, ease: 'easeOut' },
              }}
            >
              {plan.featured && <div className={styles.popularBadge}>Most Popular</div>}
              <div className={styles.planHeader}>
                <h3 className={styles.planName}>{plan.name}</h3>
                <p className={styles.planDescription}>{plan.description}</p>
              </div>
              <div className={styles.priceContainer}>
                <span className={styles.amount}>{plan.price}</span>
                {plan.priceSuffix && <span className={styles.priceSuffix}>{plan.priceSuffix}</span>}
              </div>
              <p className={styles.priceNote}>{plan.priceNote ?? '\u00A0'}</p>
              <ul className={styles.featuresList}>
                {plan.features.map((feature, i) => (
                  <li key={i} className={styles.featureItem}>
                    <HiCheck className={styles.checkIcon} />
                    <span>{feature}</span>
                  </li>
                ))}
              </ul>
              <button
                className={`${styles.ctaButton} ${plan.featured ? styles.ctaFeatured : ''}`}
                onClick={() => handleCTA(plan.ctaHref)}
              >
                {plan.ctaText}
              </button>
            </motion.div>
          ))}
        </motion.div>
      </div>
    </section>
  );
};

export default PricingSection;
