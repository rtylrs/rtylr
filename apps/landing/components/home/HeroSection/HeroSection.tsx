import React from 'react';
import MainHero from './components/MainHero';
import StatsSection from './components/StatsSection';
import styles from './HeroSection.module.scss';

const HeroSection = (): JSX.Element => {
  return (
    <section className={styles.heroSection}>
      <div className={styles.heroContentWrap}>
        <MainHero />
        <StatsSection />
      </div>
    </section>
  );
};

export default HeroSection;
