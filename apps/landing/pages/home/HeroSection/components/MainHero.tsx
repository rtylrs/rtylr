import React from 'react';
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
        <div className={styles.heroContent}>
          <h1 className={styles.heroTitle}>
            Stop Juggling Systems.
            <br />
            Start Selling Smarter.
          </h1>
        </div>
        <div className={styles.heroSidebar}>
          <button className={styles.demoButton} onClick={handleFreeDemo}>
            Book a Demo
          </button>
          <p className={styles.description}>
            Your single source of truth for in-store, online, and back-office
            operations. POS · Inventory · eCommerce · Analytics — works offline,
            scales globally.
          </p>
        </div>
      </div>
    </div>
  );
};

export default MainHero;
