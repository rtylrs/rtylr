import React from 'react';
import { cn } from '../../lib/cn';
import { motion } from 'motion/react';
import { HiCheck } from 'react-icons/hi';
import { useTranslation } from 'react-i18next';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import { PLANS, type Plan } from './pricingData';
import { formatRegionalPrice, type RegionalPrices } from '../../lib/regionalPricing';
import styles from './PricingCards.module.scss';

interface PricingCardsProps {
  onCta: (action: Plan['ctaAction']) => void;
  /**
   * Prices to show, by plan, when the visitor's region has any. Without them
   * the cards carry no price at all, which is what every other visitor sees.
   */
  prices?: RegionalPrices;
}

// No price block. The cards carry the tier name, who it is for, and what it
// includes, which is everything rtylr can currently state as fact. See
// pricingData.ts for why the numbers came out.
//
// Copy resolves per plan id from `pricing.plans.<id>`. The feature list
// comes back as an array via returnObjects, so the locale file holds a real
// list: a translator translates a list, and reordering it cannot silently
// reassign one tier's feature to another the way numbered keys could.
const PricingCards = ({ onCta, prices }: PricingCardsProps): React.JSX.Element => {
  const { t, i18n } = useTranslation();
  const showPrices = Object.keys(prices ?? {}).length > 0;

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
        const price = prices?.[plan.id];

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
            {showPrices && (
              <div className={styles.priceBlock} data-testid={`price-${plan.id}`}>
                {price && price.monthlyCents > 0 ? (
                  <>
                    <p className={styles.price}>
                      <span className={styles.priceAmount}>
                        {formatRegionalPrice(price.monthlyCents, price, i18n.language)}
                      </span>{' '}
                      <span className={styles.priceUnit}>{t('pricing.regional.perSeatMonth')}</span>
                    </p>
                    {price.annualCents > 0 && price.annualCents < price.monthlyCents && (
                      <p className={styles.priceNote}>
                        {t('pricing.regional.annualNote', {
                          price: formatRegionalPrice(price.annualCents, price, i18n.language),
                        })}
                      </p>
                    )}
                  </>
                ) : (
                  // Free is free in every region. Any other plan without a
                  // regional price stays what the page says it is everywhere:
                  // quoted.
                  <p className={styles.price}>
                    {plan.id === 'free' ? (
                      <span className={styles.priceAmount}>{t('pricing.regional.free')}</span>
                    ) : (
                      <span className={styles.priceQuoted}>{t('pricing.regional.quoted')}</span>
                    )}
                  </p>
                )}
              </div>
            )}
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
