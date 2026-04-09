import React from 'react';
import { motion } from 'framer-motion';
import { fadeInUp, viewportOnce } from '../../../lib/animations';
import styles from './HowItWorksSection.module.scss';

const steps = [
  {
    number: '01',
    title: 'Sign Up',
    description:
      'Create your account in minutes and get instant access to the full rtylr suite. No credit card required.',
  },
  {
    number: '02',
    title: 'Set Up Hub',
    description:
      'Use Hub to configure your locations, team, brand settings, and preferences. Your command center for everything.',
  },
  {
    number: '03',
    title: 'Assign Staff to Apps',
    description:
      'Give each team member access to the apps they need. POS for floor staff, HR for managers, Finance for your accountant.',
  },
  {
    number: '04',
    title: 'Launch',
    description:
      'Go live across every department at once. POS, ERP, HR, CRM, Finance, and Flow all running together from day one.',
  },
];

const HowItWorksSection: React.FC = () => {
  return (
    <section className={styles.howItWorksSection}>
      {/* Animated background orbs */}
      <motion.div
        className={styles.bgOrbOne}
        animate={{
          x: [0, 60, 0],
          y: [0, 40, 0],
          scale: [1, 1.15, 1],
        }}
        transition={{
          duration: 18,
          repeat: Infinity,
          ease: 'easeInOut',
        }}
      />
      <motion.div
        className={styles.bgOrbTwo}
        animate={{
          x: [0, -50, 0],
          y: [0, -30, 0],
          scale: [1, 1.1, 1],
        }}
        transition={{
          duration: 22,
          repeat: Infinity,
          ease: 'easeInOut',
        }}
      />

      <div className={styles.container}>
        <motion.div
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={fadeInUp}
        >
          <h2 className={styles.sectionTitle}>How It Works</h2>
          <p className={styles.sectionDescription}>
            From sign-up to full business operations in four simple steps.
          </p>
        </motion.div>

        <div className={styles.timeline}>
          <div className={styles.timelineLine} />
          {steps.map((step, index) => (
            <motion.div
              key={index}
              className={styles.timelineItem}
              initial={{ opacity: 0, x: -40 }}
              whileInView={{ opacity: 1, x: 0 }}
              viewport={{ once: true, margin: '-80px' }}
              transition={{
                duration: 0.6,
                delay: 0.1,
                ease: [0.25, 0.46, 0.45, 0.94],
              }}
            >
              <motion.div
                className={styles.stepNumberWrap}
                initial={{ scale: 0.5, opacity: 0 }}
                whileInView={{ scale: 1, opacity: 1 }}
                viewport={{ once: true, margin: '-80px' }}
                transition={{
                  duration: 0.5,
                  delay: 0.2,
                  ease: [0.25, 0.46, 0.45, 0.94],
                }}
              >
                <div className={styles.stepNumber}>{step.number}</div>
              </motion.div>
              <div className={styles.stepContent}>
                <h3 className={styles.stepTitle}>{step.title}</h3>
                <p className={styles.stepDescription}>{step.description}</p>
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
};

export default HowItWorksSection;
