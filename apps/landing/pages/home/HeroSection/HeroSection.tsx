import React from 'react';
import MainHero from './components/MainHero';
import StatsSection from './components/StatsSection';
import styles from './HeroSection.module.scss';

const HeroSection = (): JSX.Element => {
  return (
    <section className={styles.heroSection}>
      {/* Accent glow orbs — no image, just subtle brand-colored radials */}
      <div className={styles.glowOrb1} aria-hidden="true" />
      <div className={styles.glowOrb2} aria-hidden="true" />

      <div className={styles.heroContentWrap}>
        <MainHero />
        <StatsSection />
      </div>
    </section>
  );
};

export default HeroSection;
