import React from 'react';
import { motion } from 'motion/react';
import { HiCheck } from 'react-icons/hi';
import { useRouter } from 'next/router';
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
  featured?: boolean;
}

const plans: Plan[] = [
  {
    name: 'Free',
    price: '$0',
    priceSuffix: 'forever',
    description: 'Trying rtylr with a small team',
    features: [
      'Up to 5 team members',
      'All 7 apps included',
      'Basic support',
      'No credit card required',
    ],
    ctaText: 'Get Started',
  },
  {
    name: 'Pro',
    price: '$4.99',
    priceSuffix: '/month',
    priceNote: 'billed monthly',
    description: 'Small businesses getting started',
    features: [
      'Up to 20 team members',
      'All 7 apps with advanced features',
      'Priority support',
      'AI assistant',
    ],
    ctaText: 'Start with Pro',
    featured: true,
  },
  {
    name: 'Scale',
    price: '$2.99',
    priceSuffix: '/member/month',
    priceNote: 'for every member beyond the first 20 (Pro price still applies for the first 20)',
    description: 'Growing teams that outgrew Pro',
    features: [
      'For teams of 20+ members',
      'All 7 apps + advanced analytics',
      'Dedicated support',
      'AI assistant',
      'SSO + audit logs',
    ],
    ctaText: 'Contact Sales',
  },
];

const PricingSection = (): JSX.Element => {
  const router = useRouter();

  const handleCTA = () => {
    router.push('/demo');
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
            Simple pricing for teams
            <br />
            of every size.
          </h2>
          <p className={styles.sectionDescription}>
            Start free. Upgrade when you need more. No hidden fees, no surprises.
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
                <span className={styles.priceSuffix}>{plan.priceSuffix}</span>
              </div>
              {plan.priceNote && <p className={styles.priceNote}>{plan.priceNote}</p>}
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
                onClick={handleCTA}
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
