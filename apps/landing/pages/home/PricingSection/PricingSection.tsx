import React from 'react';
import { motion } from 'framer-motion';
import {
  fadeInUp,
  staggerContainer,
  viewportOnce,
} from '../../../lib/animations';
import styles from './PricingSection.module.scss';

const cardVariants = {
  hidden: { opacity: 0, y: 30 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.5, ease: [0.25, 0.46, 0.45, 0.94] },
  },
};

const PricingSection = (): JSX.Element => {
  return (
    <section className={styles.pricingSection}>
      <div className={styles.container}>
        <motion.h1
          className={styles.sectionTitle}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={fadeInUp}
        >
          Pricing
        </motion.h1>
        <motion.div
          className={styles.pricingGrid}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={staggerContainer}
        >
          <motion.div className={styles.pricingCard} variants={cardVariants}>
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
          </motion.div>

          <motion.div className={styles.pricingCard} variants={cardVariants}>
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
          </motion.div>

          <motion.div className={styles.pricingCard} variants={cardVariants}>
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
          </motion.div>
        </motion.div>
      </div>
    </section>
  );
};

export default PricingSection;
