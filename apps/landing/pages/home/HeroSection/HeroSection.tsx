import React from 'react';
import styles from './HeroSection.module.scss';
import TopBanner from './components/TopBanner';
import MainHero from './components/MainHero';
import StatsSection from './components/StatsSection';

const HeroSection = (): JSX.Element => {
  return (
    <section className={styles.heroSection}>
      <div className={styles.heroContentWrap}>
        <TopBanner />
        <MainHero />
        <StatsSection />
      </div>
    </section>
  );
};

export default HeroSection;
