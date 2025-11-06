import React from 'react';
import styles from './HowItWorksSection.module.scss';

const steps = [
  {
    number: '01',
    title: 'Sign Up',
    description: 'Create your account in minutes. No credit card required to get started.',
  },
  {
    number: '02',
    title: 'Configure',
    description: 'Set up your menu, inventory, and payment methods. Our team is here to help.',
  },
  {
    number: '03',
    title: 'Start Selling',
    description: 'Begin processing orders and managing your business from day one.',
  },
];

const HowItWorksSection: React.FC = () => {
  return (
    <section className={styles.howItWorksSection}>
      <div className={styles.container}>
        <h2 className={styles.sectionTitle}>How It Works</h2>
        <p className={styles.sectionDescription}>
          Get started with Rtylr in three simple steps
        </p>
        <div className={styles.stepsGrid}>
          {steps.map((step, index) => (
            <div key={index} className={styles.stepItem}>
              <div className={styles.stepNumber}>{step.number}</div>
              <h3 className={styles.stepTitle}>{step.title}</h3>
              <p className={styles.stepDescription}>{step.description}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
};

export default HowItWorksSection;

