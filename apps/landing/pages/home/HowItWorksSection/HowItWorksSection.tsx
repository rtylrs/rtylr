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
      'Create your account in minutes and get instant access to the full rtylr suite — no credit card required.',
  },
  {
    number: '02',
    title: 'Set Up Hub',
    description:
      'Use Hub to configure your locations, menus, QR codes, and brand settings. Your command centre for everything.',
  },
  {
    number: '03',
    title: 'Assign Staff to Apps',
    description:
      'Give each team member access to the apps they need — POS for floor staff, HR for managers, Finance for your accountant.',
  },
  {
    number: '04',
    title: 'Launch',
    description:
      'Go live across every department at once. POS, ERP, HR, CRM, Finance, and Flow — all running together from day one.',
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
            From sign-up to full business operations in four simple steps
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
