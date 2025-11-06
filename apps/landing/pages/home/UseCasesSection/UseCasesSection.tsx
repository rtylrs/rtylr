import React from 'react';
import Link from 'next/link';
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
        <h2 className={styles.sectionTitle}>Solutions for Every Business</h2>
        <p className={styles.sectionDescription}>
          Tailored POS solutions for your industry
        </p>
        <div className={styles.useCasesGrid}>
          {useCases.map((useCase, index) => (
            <Link key={index} href={useCase.href} className={styles.useCaseCard}>
              <h3 className={styles.useCaseTitle}>{useCase.title}</h3>
              <p className={styles.useCaseDescription}>
                {useCase.description}
              </p>
              <span className={styles.useCaseLink}>Learn more →</span>
            </Link>
          ))}
        </div>
      </div>
    </section>
  );
};

export default UseCasesSection;

