import React from 'react';
import { motion } from 'framer-motion';
import Link from 'next/link';
import {
  fadeInUp,
  staggerContainer,
  scaleIn,
  viewportOnce,
} from '../../../lib/animations';
import styles from './UseCasesSection.module.scss';

const useCases = [
  {
    title: 'Food & Beverage',
    description: 'Restaurants, cafes, and bars',
    href: '/solutions/food-beverage',
  },
  {
    title: 'Retail & Omnichannel',
    description: 'Stores with online and in-person sales',
    href: '/solutions/retail',
  },
  {
    title: 'Services & Appointments',
    description: 'Salons, spas, and service businesses',
    href: '/solutions/services',
  },
  {
    title: 'Grocery & Convenience',
    description: 'Supermarkets and convenience stores',
    href: '/solutions/grocery',
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
          <p className={styles.sectionDescription}>
            Tailored POS solutions for your industry
          </p>
        </motion.div>

        <motion.div
          className={styles.useCasesGrid}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={staggerContainer}
        >
          {useCases.map((useCase, index) => (
            <motion.div key={index} variants={scaleIn}>
              <Link href={useCase.href} className={styles.useCaseCard}>
                <h3 className={styles.useCaseTitle}>{useCase.title}</h3>
                <p className={styles.useCaseDescription}>
                  {useCase.description}
                </p>
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
