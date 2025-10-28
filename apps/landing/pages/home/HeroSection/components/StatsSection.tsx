import React from 'react';
import styles from '../HeroSection.module.scss';

const StatsSection: React.FC = () => {
  const stats = [
    { number: '50K+', label: 'Daily Transactions' },
    { number: '99.9%', label: 'Uptime' },
    { number: '20+', label: 'Countries' },
    { number: '24/7', label: 'Support' },
    { number: 'Offline', label: 'Always Works' },
    { number: 'AI', label: 'Powered' },
    { number: 'Multi', label: 'Language' },
    { number: 'Cloud', label: 'Native' },
  ];

  return (
    <div className={styles.trustedBy}>
      <div className={styles.trustedContainer}>
        <div className={styles.logoScroll}>
          {stats.map((stat, index) => (
            <div key={index} className={styles.logoItem}>
              <div className={styles.statNumber}>{stat.number}</div>
              <div className={styles.statLabel}>{stat.label}</div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default StatsSection;
