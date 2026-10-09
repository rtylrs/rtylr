import React, { useRef } from 'react';
import { motion, useScroll, useTransform } from 'motion/react';
import MainHero from './components/MainHero';
import styles from './HeroSection.module.scss';

const HeroSection = (): React.JSX.Element => {
  const sectionRef = useRef<HTMLElement>(null);

  // Hero recedes as the next section scrolls over it - scales down,
  // dims, and drifts up slightly slower than the scroll.
  const { scrollYProgress } = useScroll({
    target: sectionRef,
    offset: ['start start', 'end start'],
  });
  const scale = useTransform(scrollYProgress, [0, 1], [1, 0.94]);
  const opacity = useTransform(scrollYProgress, [0, 0.85], [1, 0.25]);
  const y = useTransform(scrollYProgress, [0, 1], ['0%', '8%']);

  return (
    <section ref={sectionRef} className={styles.heroSection}>
      <motion.div className={styles.heroContentWrap} style={{ scale, opacity, y }}>
        <MainHero />
      </motion.div>
    </section>
  );
};

export default HeroSection;
