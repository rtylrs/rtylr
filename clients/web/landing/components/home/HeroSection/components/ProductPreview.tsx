import React, { useRef } from 'react';
import { motion, useReducedMotion, useScroll, useTransform } from 'motion/react';
import { EASE_OUT_EXPO } from '../../../../lib/animations';
import FloatingCards from './FloatingCards';
import HeroPortal from './HeroPortal';
import styles from '../HeroSection.module.scss';

/**
 * The product window under the hero prompt: a live, clickable rtylr
 * workspace on mock data. It starts tilted back and a little small, then
 * flattens and settles as the reader scrolls toward it.
 */
const ProductPreview: React.FC<{ alt: string }> = ({ alt }) => {
  const ref = useRef<HTMLDivElement>(null);
  const reduceMotion = useReducedMotion();

  const { scrollYProgress } = useScroll({
    target: ref,
    offset: ['start end', 'start 0.35'],
  });
  const rotateX = useTransform(scrollYProgress, [0, 1], [22, 0]);
  const scale = useTransform(scrollYProgress, [0, 1], [0.9, 1]);
  const glowOpacity = useTransform(scrollYProgress, [0, 1], [0.35, 0.9]);

  // Longer range for the floating cards' parallax: from the stage entering
  // the viewport until it has scrolled fully past.
  const { scrollYProgress: passProgress } = useScroll({
    target: ref,
    offset: ['start end', 'end start'],
  });

  return (
    <motion.div
      ref={ref}
      className={styles.productStage}
      initial={{ opacity: 0, y: 48 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: 0.7, duration: 1.1, ease: EASE_OUT_EXPO }}
    >
      <motion.div
        className={styles.productGlow}
        style={reduceMotion ? undefined : { opacity: glowOpacity }}
        aria-hidden="true"
      />
      <FloatingCards progress={passProgress} />
      <motion.div
        className={styles.productWindow}
        style={reduceMotion ? undefined : { rotateX, scale }}
      >
        <div className={styles.productChrome} aria-hidden="true">
          <span className={styles.productDot} />
          <span className={styles.productDot} />
          <span className={styles.productDot} />
          <span className={styles.productUrl}>portal.rtylr.com</span>
        </div>
        <section className={styles.productBody} aria-label={alt}>
          <HeroPortal />
        </section>
      </motion.div>
    </motion.div>
  );
};

export default ProductPreview;
