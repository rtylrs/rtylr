import React from 'react';
import { motion } from 'framer-motion';
import styles from './HeroSection.module.scss';
import TopBanner from './components/TopBanner';
import MainHero from './components/MainHero';
import StatsSection from './components/StatsSection';

const HeroSection = (): JSX.Element => {
  return (
    <section className={styles.heroSection}>
      {/* Floating background orbs */}
      <motion.div
        className={styles.heroOrbOne}
        animate={{
          x: [0, 100, 0],
          y: [0, 50, 0],
          scale: [1, 1.2, 1],
        }}
        transition={{
          duration: 20,
          repeat: Infinity,
          ease: 'easeInOut',
        }}
      />
      <motion.div
        className={styles.heroOrbTwo}
        animate={{
          x: [0, -80, 0],
          y: [0, 60, 0],
          scale: [1, 1.15, 1],
        }}
        transition={{
          duration: 24,
          repeat: Infinity,
          ease: 'easeInOut',
        }}
      />
      <motion.div
        className={styles.heroOrbThree}
        animate={{
          x: [0, 60, 0],
          y: [0, -40, 0],
          scale: [1, 1.1, 1],
        }}
        transition={{
          duration: 28,
          repeat: Infinity,
          ease: 'easeInOut',
        }}
      />

      <div className={styles.heroContentWrap}>
        <TopBanner />
        <MainHero />
        <StatsSection />
      </div>
    </section>
  );
};

export default HeroSection;
