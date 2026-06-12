import React from 'react';
import { motion } from 'motion/react';
import { CountUp } from '../../../motion/CountUp';
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
    <motion.div
      className={styles.trustedBy}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{
        delay: 0.7,
        duration: 0.6,
        ease: [0.25, 0.46, 0.45, 0.94],
      }}
    >
      <div className={styles.trustedContainer}>
        <div className={styles.logoScroll}>
          {stats.map((stat, index) => (
            <motion.div
              key={index}
              className={styles.logoItem}
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: 0.8 + index * 0.06, duration: 0.4 }}
            >
              <CountUp value={stat.number} className={styles.statNumber} />
              <div className={styles.statLabel}>{stat.label}</div>
            </motion.div>
          ))}
        </div>
      </div>
    </motion.div>
  );
};

export default StatsSection;
