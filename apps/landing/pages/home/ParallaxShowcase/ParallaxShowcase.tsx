import React, { useRef } from 'react';
import { motion, useScroll, useTransform } from 'motion/react';
import { blurFadeIn, viewportOnce } from '../../../lib/animations';
import styles from './ParallaxShowcase.module.scss';

const ParallaxShowcase: React.FC = (): React.JSX.Element => {
  const sectionRef = useRef<HTMLElement>(null);

  const { scrollYProgress } = useScroll({
    target: sectionRef,
    offset: ['start end', 'end start'],
  });

  // Image moves opposite to scroll — element enters from below, image drifts upward
  // Big pixel range = obvious cinematic parallax depth
  const imgY = useTransform(scrollYProgress, [0, 1], ['120px', '-420px']);
  const textY = useTransform(scrollYProgress, [0, 1], ['60px', '-60px']);

  return (
    <section ref={sectionRef} className={styles.section}>
      {/* Parallax image fills the section */}
      <motion.div className={styles.imgWrap} style={{ y: imgY }} aria-hidden="true">
        <img src="/images/home/hero-banner-top-bar-bg-old.png" alt="" className={styles.img} />
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

        <motion.h2 className={styles.title} variants={blurFadeIn}>
          Every team. Every workflow.
          <br />
          <span className={styles.titleAccent}>One platform.</span>
        </motion.h2>

        <motion.p className={styles.subtitle} variants={blurFadeIn}>
          rtylr isn&apos;t seven tools glued together. It&apos;s one platform where a POS sale
          depletes inventory, flows tips into payroll, and enriches the customer profile — from the
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
              <span className={styles.statNumber}>{stat.number}</span>
              <span className={styles.statLabel}>{stat.label}</span>
            </div>
          ))}
        </motion.div>
      </motion.div>
    </section>
  );
};

export default ParallaxShowcase;
