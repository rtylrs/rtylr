import React from 'react';
import { motion } from 'motion/react';
import Link from 'next/link';
import { fadeInUp, staggerBlur, blurFadeIn, viewportOnce } from '../../../lib/animations';
import styles from './UseCasesSection.module.scss';

const useCases = [
  {
    title: 'Food & Beverage',
    description: 'Restaurants, cafes, and bars',
    href: '/pos',
  },
  {
    title: 'Retail & Omnichannel',
    description: 'Stores with online and in-person sales',
    href: '/pos',
  },
  {
    title: 'Services & Appointments',
    description: 'Salons, spas, and service businesses',
    href: '/crm',
  },
  {
    title: 'Grocery & Convenience',
    description: 'Supermarkets and convenience stores',
    href: '/pos',
  },
];

const UseCasesSection: React.FC = () => {
  return (
    <section className={styles.useCasesSection}>
      <div className={styles.container}>
        <motion.div
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={fadeInUp}
        >
          <h2 className={styles.sectionTitle}>Solutions for Every Business</h2>
          <p className={styles.sectionDescription}>Tailored POS solutions for your industry</p>
        </motion.div>

        <motion.div
          className={styles.useCasesGrid}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={staggerBlur}
        >
          {useCases.map((useCase, index) => (
            <motion.div key={index} variants={blurFadeIn}>
              <Link href={useCase.href} className={styles.useCaseCard}>
                <h3 className={styles.useCaseTitle}>{useCase.title}</h3>
                <p className={styles.useCaseDescription}>{useCase.description}</p>
                <span className={styles.useCaseLink}>Learn more →</span>
              </Link>
            </motion.div>
          ))}
        </motion.div>
      </div>
    </section>
  );
};

export default UseCasesSection;
