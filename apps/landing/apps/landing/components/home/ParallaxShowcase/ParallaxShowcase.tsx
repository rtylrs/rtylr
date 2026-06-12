import React, { useRef } from 'react';
import { motion, useScroll, useSpring, useTransform } from 'motion/react';
import { blurFadeIn, viewportOnce } from '../../../lib/animations';
import { SplitWords } from '../../motion/SplitWords';
import { CountUp } from '../../motion/CountUp';
import styles from './ParallaxShowcase.module.scss';

const ParallaxShowcase: React.FC = (): React.JSX.Element => {
  const sectionRef = useRef<HTMLElement>(null);

  const { scrollYProgress } = useScroll({
    target: sectionRef,
    offset: ['start end', 'end start'],
  });

  // Spring-smoothed scroll progress so the parallax glides instead of
  // tracking the wheel 1:1.
  const smooth = useSpring(scrollYProgress, { stiffness: 110, damping: 28, mass: 0.6 });

  // Image drifts against the scroll while easing out of a zoomed crop -
  // the zoom-out settle is what sells the depth.
  const imgY = useTransform(smooth, [0, 1], ['120px', '-420px']);
  const imgScale = useTransform(smooth, [0, 0.5, 1], [1.18, 1.05, 1]);
  const textY = useTransform(smooth, [0, 1], ['60px', '-60px']);

  return (
    <section ref={sectionRef} className={styles.section}>
      {/* Parallax image fills the section */}
      <motion.div className={styles.imgWrap} style={{ y: imgY, scale: imgScale }} aria-hidden="true">
        <img
          src="/images/home/hero-banner-top-bar-bg.webp"
          alt=""
          className={styles.img}
          loading="lazy"
          decoding="async"
        />
        <div className={styles.imgOverlay} />
      </motion.div>

      {/* Content floats over the image */}
      <motion.div
        className={styles.content}
        style={{ y: textY }}
        initial="hidden"
        whileInView="visible"
        viewport={viewportOnce}
        variants={{
          hidden: { opacity: 0 },
          visible: { opacity: 1, transition: { staggerChildren: 0.14, delayChildren: 0.1 } },
        }}
      >
        <motion.span className={styles.badge} variants={blurFadeIn}>
          BUILT DIFFERENT
        </motion.span>

        <h2 className={styles.title}>
          <SplitWords as="span" text="Every team. Every workflow." trigger="view" />
          <br />
          <span className={styles.titleAccent}>
            <SplitWords as="span" text="One platform." trigger="view" delay={0.3} />
          </span>
        </h2>

        <motion.p className={styles.subtitle} variants={blurFadeIn}>
          rtylr isn&apos;t seven tools glued together. It&apos;s one platform where a POS sale
          depletes inventory, flows tips into payroll, and enriches the customer profile, from the
          same database, in real time.
        </motion.p>

        <motion.div className={styles.stats} variants={blurFadeIn}>
          {[
            { number: '7', label: 'Apps in one platform' },
            { number: '99.9%', label: 'Uptime guaranteed' },
            { number: '24/7', label: 'Support included' },
            { number: 'AI', label: 'Built into every app' },
          ].map((stat) => (
            <div key={stat.label} className={styles.stat}>
              <CountUp value={stat.number} className={styles.statNumber} />
              <span className={styles.statLabel}>{stat.label}</span>
            </div>
          ))}
        </motion.div>
      </motion.div>
    </section>
  );
};

export default ParallaxShowcase;
