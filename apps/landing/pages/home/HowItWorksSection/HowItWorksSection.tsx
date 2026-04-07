import React from 'react';
import { motion } from 'framer-motion';
import {
  fadeInUp,
  staggerContainer,
  viewportOnce,
} from '../../../lib/animations';
import styles from './HowItWorksSection.module.scss';

const steps = [
  {
    number: '01',
    title: 'Sign Up',
    description:
      'Create your account in minutes. No credit card required to get started.',
  },
  {
    number: '02',
    title: 'Configure',
    description:
      'Set up your menu, inventory, and payment methods. Our team is here to help.',
  },
  {
    number: '03',
    title: 'Start Selling',
    description:
      'Begin processing orders and managing your business from day one.',
  },
];

const HowItWorksSection: React.FC = () => {
  return (
    <section className={styles.howItWorksSection}>
      <div className={styles.container}>
        <motion.div
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={fadeInUp}
        >
          <h2 className={styles.sectionTitle}>How It Works</h2>
          <p className={styles.sectionDescription}>
            Get started with Rtylr in three simple steps
          </p>
        </motion.div>

        <motion.div
          className={styles.stepsGrid}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={staggerContainer}
        >
          {steps.map((step, index) => (
            <motion.div
              key={index}
              className={styles.stepItem}
              variants={fadeInUp}
            >
              <div className={styles.stepNumber}>{step.number}</div>
              <h3 className={styles.stepTitle}>{step.title}</h3>
              <p className={styles.stepDescription}>{step.description}</p>
            </motion.div>
          ))}
        </motion.div>
      </div>
    </section>
  );
};

export default HowItWorksSection;
