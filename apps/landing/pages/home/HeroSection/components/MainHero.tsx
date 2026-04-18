import React, { useRef } from 'react';
import { motion, useMotionValue, useSpring } from 'motion/react';
import { useRouter } from 'next/router';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../../../lib/navigation';
import styles from '../HeroSection.module.scss';

const PRODUCTS = ['POS', 'ERP', 'HR', 'CRM', 'Finance', 'Flow', 'Insights'];

const MainHero: React.FC = (): React.JSX.Element => {
  const router = useRouter();
  const containerRef = useRef<HTMLDivElement>(null);

  const mouseX = useMotionValue(500);
  const mouseY = useMotionValue(320);
  const springX = useSpring(mouseX, { stiffness: 38, damping: 22 });
  const springY = useSpring(mouseY, { stiffness: 38, damping: 22 });

  const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>): void => {
    const rect = containerRef.current?.getBoundingClientRect();
    if (!rect) return;
    mouseX.set(e.clientX - rect.left);
    mouseY.set(e.clientY - rect.top);
  };

  const handleGetStarted = (): void => {
    navigateExternal(`${APP_URLS.auth}/register`);
  };

  const handleBookDemo = (): void => {
    void router.push('/demo');
  };

  return (
    <div className={styles.mainHero} ref={containerRef} onMouseMove={handleMouseMove}>
      {/* Mouse-follow radial glow */}
      <motion.div
        className={styles.mouseGlow}
        style={{ left: springX, top: springY }}
        aria-hidden="true"
      />

      {/* Ambient background orbs */}
      <div className={styles.ambientOrb1} aria-hidden="true" />
      <div className={styles.ambientOrb2} aria-hidden="true" />

      <div className={styles.heroInner}>
        {/* Eyebrow badge */}
        <motion.div
          className={styles.eyebrow}
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1, duration: 0.7, ease: [0.16, 1, 0.3, 1] }}
        >
          <span className={styles.eyebrowDot} />
          <span>Buy rtylr, get a whole business powered by AI</span>
        </motion.div>

        {/* Main headline */}
        <motion.h1
          className={styles.heroTitle}
          initial={{ opacity: 0, y: 50, filter: 'blur(14px)' }}
          animate={{ opacity: 1, y: 0, filter: 'blur(0px)' }}
          transition={{ delay: 0.22, duration: 1.0, ease: [0.16, 1, 0.3, 1] }}
        >
          Your entire business,
          <br />
          <span className={styles.heroTitleGradient}>one platform.</span>
        </motion.h1>

        {/* Subtitle */}
        <motion.p
          className={styles.heroSubtitle}
          initial={{ opacity: 0, y: 30, filter: 'blur(8px)' }}
          animate={{ opacity: 1, y: 0, filter: 'blur(0px)' }}
          transition={{ delay: 0.44, duration: 0.9, ease: [0.16, 1, 0.3, 1] }}
        >
          POS · ERP · HR · CRM · Finance · Flow · Insights.
          <br />
          Built to work together from day one.
        </motion.p>

        {/* CTA buttons */}
        <motion.div
          className={styles.heroCTAs}
          initial={{ opacity: 0, y: 24 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.62, duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
        >
          <button className={styles.ctaPrimary} onClick={handleGetStarted}>
            Get started free
          </button>
          <button className={styles.ctaSecondary} onClick={handleBookDemo}>
            Book a demo
          </button>
        </motion.div>

        {/* Product pills */}
        <motion.div
          className={styles.productPills}
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.82, duration: 0.7, ease: [0.16, 1, 0.3, 1] }}
        >
          {PRODUCTS.map((p) => (
            <span key={p} className={styles.productPill}>
              {p}
            </span>
          ))}
        </motion.div>
      </div>
    </div>
  );
};

export default MainHero;
