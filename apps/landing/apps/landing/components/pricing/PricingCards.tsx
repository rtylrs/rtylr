import React from 'react';
import { cn } from '../../lib/cn';
import { AnimatePresence, motion } from 'motion/react';
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
        >
          {plan.featured && <div className={styles.popularBadge}>Most Popular</div>}
          <div className={styles.planHeader}>
            <h3 className={styles.planName}>{plan.name}</h3>
            <p className={styles.planDescription}>{plan.description}</p>
          </div>
          <div className={styles.priceContainer}>
            <span className={styles.amount}>
              <AnimatePresence mode="popLayout" initial={false}>
                <motion.span
                  key={price}
                  style={{ display: 'inline-block' }}
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -10 }}
                  transition={{ duration: 0.3, ease: [0.32, 0.72, 0, 1] }}
                >
                  {price}
                </motion.span>
              </AnimatePresence>
            </span>
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
