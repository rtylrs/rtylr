import React from 'react';
import styles from './PricingSection.module.scss';

const PricingSection = (): JSX.Element => {
  return (
    <section className={styles.pricingSection}>
      <div className={styles.container}>
        <h1 className={styles.sectionTitle}>Pricing</h1>
        <div className={styles.pricingGrid}>
          <div className={styles.pricingCard}>
            <h3 className={styles.planName}>Basic</h3>
            <div className={styles.priceContainer}>
              <span className={styles.currency}>$</span>
              <span className={styles.amount}>10.00</span>
            </div>
            <p className={styles.subtitle}>Single outlet</p>
            <p className={styles.description}>
              core POS, basic inventory, standard reports, offline mode
            </p>
            <button className={styles.getStartedButton}>
              Get Started <span className={styles.arrow}>→</span>
            </button>
          </div>

          <div className={styles.pricingCard}>
            <h3 className={styles.planName}>Standard</h3>
            <div className={styles.priceContainer}>
              <span className={styles.currency}>$</span>
              <span className={styles.amount}>19.99</span>
            </div>
            <p className={styles.subtitle}>2-5 outlets</p>
            <p className={styles.description}>
              multi-location, warehouse module, online storefront
            </p>
            <button className={styles.getStartedButton}>
              Get Started <span className={styles.arrow}>→</span>
            </button>
          </div>

          <div className={styles.pricingCard}>
            <h3 className={styles.planName}>Enterprise</h3>
            <div className={styles.priceContainer}>
              <span className={styles.currency}>$</span>
              <span className={styles.amount}>29.99</span>
            </div>
            <p className={styles.subtitle}>6+ or chains</p>
            <p className={styles.description}>
              AI suite, API & integrations, priority support, compliance setup
            </p>
            <button className={styles.getStartedButton}>
              Get Started <span className={styles.arrow}>→</span>
            </button>
          </div>
        </div>
      </div>
    </section>
  );
};

export default PricingSection;