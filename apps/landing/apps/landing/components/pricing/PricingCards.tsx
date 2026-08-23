import React from 'react';
import { cn } from '../../lib/cn';
import { motion } from 'motion/react';
import { HiCheck } from 'react-icons/hi';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import { PLANS, type Plan } from './pricingData';
import styles from './PricingCards.module.scss';

interface PricingCardsProps {
  onCta: (action: Plan['ctaAction']) => void;
}

// No price block. The cards carry the tier name, who it is for, and
// what it includes, which is everything rtylr can currently state as
// fact. See pricingData.ts for why the numbers came out.
const PricingCards = ({ onCta }: PricingCardsProps): React.JSX.Element => (
  <motion.div
    className={styles.pricingGrid}
    initial="hidden"
    whileInView="visible"
    viewport={viewportOnce}
    variants={staggerBlur}
  >
    {PLANS.map((plan) => (
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
        <ul className={styles.featuresList}>
          {plan.features.map((feature) => (
            <li key={feature} className={styles.featureItem}>
              <HiCheck className={styles.checkIcon} />
              <span>{feature}</span>
            </li>
          ))}
        </ul>
        <button className={styles.ctaButton} onClick={() => onCta(plan.ctaAction)}>
          {plan.ctaText}
        </button>
      </motion.div>
    ))}
  </motion.div>
);

export default PricingCards;
