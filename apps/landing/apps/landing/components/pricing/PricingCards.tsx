import React from 'react';
import { cn } from '../../lib/cn';
import { motion } from 'motion/react';
import { HiCheck } from 'react-icons/hi';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import { PLANS, type BillingCycle, type Plan } from './pricingData';
import styles from './PricingCards.module.scss';

interface PricingCardsProps {
  billing?: BillingCycle;
  onCta: (action: Plan['ctaAction']) => void;
}

const PricingCards = ({ billing = 'monthly', onCta }: PricingCardsProps): React.JSX.Element => (
  <motion.div
    className={styles.pricingGrid}
    initial="hidden"
    whileInView="visible"
    viewport={viewportOnce}
    variants={staggerBlur}
  >
    {PLANS.map((plan) => {
      const price = plan.price[billing];
      const priceNote = plan.priceNote[billing];
      return (
        <motion.div
          key={plan.name}
          className={cn(styles.pricingCard, plan.featured && styles.featured)}
          variants={blurFadeIn}
          whileHover={{ y: -8, transition: { duration: 0.25, ease: 'easeOut' } }}
        >
          {plan.featured && <div className={styles.popularBadge}>Most Popular</div>}
          <div className={styles.planHeader}>
            <h3 className={styles.planName}>{plan.name}</h3>
            <p className={styles.planDescription}>{plan.description}</p>
          </div>
          <div className={styles.priceContainer}>
            <span className={styles.amount}>{price}</span>
            <span className={styles.priceSuffix}>{plan.priceSuffix}</span>
          </div>
          <p className={styles.priceNote}>{priceNote ?? ' '}</p>
          <ul className={styles.featuresList}>
            {plan.features.map((feature, i) => (
              <li key={i} className={styles.featureItem}>
                <HiCheck className={styles.checkIcon} />
                <span>{feature}</span>
              </li>
            ))}
          </ul>
          <button className={styles.ctaButton} onClick={() => onCta(plan.ctaAction)}>
            {plan.ctaText}
          </button>
        </motion.div>
      );
    })}
  </motion.div>
);

export default PricingCards;
