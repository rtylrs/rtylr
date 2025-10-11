import React from 'react';
import styles from './PricingSection.module.scss';

const PricingSection = (): JSX.Element => {
  return (
    <section className={styles.pricingSection}>
      <div className={styles.container}>
        <h1 className={styles.sectionTitle}>Pricing</h1>
        <div className={styles.pricingGrid}>
          <div className={styles.pricingCard}>
            <h3 className={styles.planName}>Free</h3>
            <div className={styles.priceContainer}>
              <span className={styles.currency}>$</span>
              <span className={styles.amount}>0</span>
            </div>
            <p className={styles.subtitle}>Single outlet</p>
            <p className={styles.description}>
              Core POS, basic inventory, standard reports, offline mode
            </p>
            <button className={styles.getStartedButton}>
              Get Started <span className={styles.arrow}>→</span>
            </button>
          </div>

          <div className={styles.pricingCard}>
            <h3 className={styles.planName}>Pro</h3>
            <div className={styles.priceContainer}>
              <span className={styles.currency}>$</span>
              <span className={styles.amount}>30</span>
            </div>
            <p className={styles.subtitle}>Per location/month</p>
            <p className={styles.description}>
              Multi-location, eCommerce, KDS, analytics, loyalty & CRM
            </p>
            <button className={styles.getStartedButton}>
              Get Started <span className={styles.arrow}>→</span>
            </button>
          </div>

          <div className={styles.pricingCard}>
            <h3 className={styles.planName}>Enterprise</h3>
            <div className={styles.priceContainer}>
              <span className={styles.customPricing}>Custom</span>
            </div>
            <p className={styles.subtitle}>Unlimited scale</p>
            <p className={styles.description}>
              AI suite, API & integrations, priority support, compliance & SLA
            </p>
            <button className={styles.getStartedButton}>
              Contact Sales <span className={styles.arrow}>→</span>
            </button>
          </div>
        </div>
      </div>
    </section>
  );
};

export default PricingSection;
