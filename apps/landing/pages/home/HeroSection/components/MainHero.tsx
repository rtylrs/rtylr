import React from 'react';
import { motion } from 'framer-motion';
import { useRouter } from 'next/router';
import styles from '../HeroSection.module.scss';

const MainHero: React.FC = () => {
  const router = useRouter();

  const handleFreeDemo = () => {
    router.push('/demo');
  };

  return (
    <div className={styles.mainHero}>
      <div className={styles.container}>
        <motion.div
          className={styles.heroContent}
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{
            delay: 0.3,
            duration: 0.7,
            ease: [0.25, 0.46, 0.45, 0.94],
          }}
        >
          <h1 className={styles.heroTitle}>
            Stop Juggling Systems.
            <br />
            Start Selling Smarter.
          </h1>
        </motion.div>
        <motion.div
          className={styles.heroSidebar}
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{
            delay: 0.5,
            duration: 0.7,
            ease: [0.25, 0.46, 0.45, 0.94],
          }}
        >
          <button className={styles.demoButton} onClick={handleFreeDemo}>
            Book a Demo
          </button>
          <p className={styles.description}>
            Your single source of truth for in-store, online, and back-office
            operations. POS · Inventory · eCommerce · Analytics works offline,
            scales globally.
          </p>
        </motion.div>
      </div>
    </div>
  );
};

export default MainHero;
