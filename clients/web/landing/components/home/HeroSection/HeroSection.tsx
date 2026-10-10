import React from 'react';
import MainHero from './components/MainHero';
import styles from './HeroSection.module.scss';

// Scroll motion lives on the product window inside MainHero; the hero itself
// stays put so the window reads clearly while it flattens into view.
const HeroSection = (): React.JSX.Element => (
  <section className={styles.heroSection}>
    <div className={styles.heroContentWrap}>
      <MainHero />
    </div>
  </section>
);

export default HeroSection;
