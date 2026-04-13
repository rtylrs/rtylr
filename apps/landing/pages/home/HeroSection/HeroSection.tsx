import React from 'react';
import { motion } from 'motion/react';
import styles from './HeroSection.module.scss';
import TopBanner from './components/TopBanner';
import MainHero from './components/MainHero';
import StatsSection from './components/StatsSection';

const HeroSection = (): JSX.Element => {
  return (
    <section className={styles.heroSection}>
      {/* Animated gradient orbs */}
      <motion.div
        style={{
          position: 'absolute',
          width: 600,
          height: 600,
          borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(126,68,150,0.09) 0%, transparent 70%)',
          bottom: -100,
          right: -150,
          pointerEvents: 'none',
          zIndex: 0,
          filter: 'blur(70px)',
        }}
        animate={{ x: [0, -60, 30, 0], y: [0, -40, -20, 0], scale: [1, 1.1, 1.18, 1] }}
        transition={{ duration: 30, repeat: Infinity, ease: 'easeInOut' }}
      />
      <motion.div
        style={{
          position: 'absolute',
          width: 400,
          height: 400,
          borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(100,197,185,0.07) 0%, transparent 70%)',
          top: '40%',
          right: '20%',
          pointerEvents: 'none',
          zIndex: 0,
          filter: 'blur(50px)',
        }}
        animate={{ x: [0, 40, -50, 0], y: [0, -50, 20, 0], scale: [1, 0.9, 1.08, 1] }}
        transition={{ duration: 20, repeat: Infinity, ease: 'easeInOut' }}
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
