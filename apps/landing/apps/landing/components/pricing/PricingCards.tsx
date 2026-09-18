import React from 'react';
import { cn } from '../../lib/cn';
import { motion } from 'motion/react';
import { HiCheck } from 'react-icons/hi';
import { useTranslation } from 'react-i18next';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import { PLANS, type Plan } from './pricingData';
import styles from './PricingCards.module.scss';

interface PricingCardsProps {
  onCta: (action: Plan['ctaAction']) => void;
}

// No price block. The cards carry the tier name, who it is for, and what it
// includes, which is everything rtylr can currently state as fact. See
// pricingData.ts for why the numbers came out.
//
// Copy resolves per plan id from `pricing.plans.<id>`. The feature list
// comes back as an array via returnObjects, so the locale file holds a real
// list: a translator translates a list, and reordering it cannot silently
// reassign one tier's feature to another the way numbered keys could.
const PricingCards = ({ onCta }: PricingCardsProps): React.JSX.Element => {
  const { t } = useTranslation();

  return (
    <motion.div
      className={styles.pricingGrid}
      initial="hidden"
      whileInView="visible"
      viewport={viewportOnce}
      variants={staggerBlur}
    >
      {PLANS.map((plan) => {
        const features = t(`pricing.plans.${plan.id}.features`, {
          returnObjects: true,
        }) as unknown as string[];

        return (
          <motion.div
            key={plan.id}
            className={cn(styles.pricingCard, plan.featured && styles.featured)}
            variants={blurFadeIn}
          >
            {plan.featured && (
              <div className={styles.popularBadge}>{t('pricing.plans.popular')}</div>
            )}
            <div className={styles.planHeader}>
              <h3 className={styles.planName}>{t(`pricing.plans.${plan.id}.name`)}</h3>
              <p className={styles.planDescription}>
                {t(`pricing.plans.${plan.id}.description`)}
              </p>
            </div>
            <ul className={styles.featuresList}>
              {features.map((feature) => (
                <li key={feature} className={styles.featureItem}>
                  <HiCheck className={styles.checkIcon} />
                  <span>{feature}</span>
                </li>
              ))}
            </ul>
            <button className={styles.ctaButton} onClick={() => onCta(plan.ctaAction)}>
              {t(`pricing.plans.${plan.id}.cta`)}
            </button>
          </motion.div>
        );
      })}
    </motion.div>
  );
};

export default PricingCards;
