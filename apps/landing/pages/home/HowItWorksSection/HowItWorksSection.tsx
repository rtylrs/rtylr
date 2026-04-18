import React from 'react';
import { motion } from 'motion/react';
import { blurFadeIn, blurFadeInSlow, staggerBlur, viewportOnce } from '../../../lib/animations';
import styles from './HowItWorksSection.module.scss';

const STEPS = [
  {
    number: '01',
    title: 'Sign up',
    description:
      'Create an account in minutes. No credit card required. Full rtylr suite — unlocked from day one.',
  },
  {
    number: '02',
    title: 'Configure Hub',
    description:
      'Set your locations, team, brand, and preferences in one place. Hub becomes your command center.',
  },
  {
    number: '03',
    title: 'Assign apps',
    description:
      'Give each team member access to what they need. POS for floor staff, HR for managers, Finance for your accountant.',
  },
  {
    number: '04',
    title: 'Go live',
    description:
      'Launch across every department at once. Seven apps, one platform — already talking to each other.',
  },
];

const HowItWorksSection: React.FC = (): React.JSX.Element => {
  return (
    <section className={styles.howItWorksSection}>
      <div className={styles.container}>
        <motion.div
          className={styles.header}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={blurFadeInSlow}
        >
          <span className={styles.eyebrow}>HOW IT WORKS</span>
          <h2 className={styles.sectionTitle}>
            From sign-up to live
            <br />
            in four steps.
          </h2>
        </motion.div>

        <motion.div
          className={styles.grid}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={staggerBlur}
        >
          {STEPS.map((step) => (
            <motion.div key={step.number} className={styles.step} variants={blurFadeIn}>
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
