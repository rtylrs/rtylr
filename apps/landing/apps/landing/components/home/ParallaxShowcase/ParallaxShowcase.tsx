import React, { useRef } from 'react';
import { motion, useScroll, useTransform } from 'motion/react';
import styles from './ParallaxShowcase.module.scss';

// Pinned scroll scene: the image starts as a rounded inset card, expands
// to full bleed as you scroll, then the message scrubs in over it. The
// section is deliberately tall; the inner viewport stays pinned while each
// phase gets enough physical scroll distance for trackpads and mouse wheels.
const ParallaxShowcase: React.FC = (): React.JSX.Element => {
  const sectionRef = useRef<HTMLElement>(null);

  const { scrollYProgress } = useScroll({
    target: sectionRef,
    offset: ['start start', 'end end'],
  });

  // Direct scrub (no spring) keeps the scene responsive; the section height
  // controls pacing without introducing input lag. Border radius stays constant (paint
  // happens once); the final 1.045 scale pushes the rounded corners just
  // past the viewport edge, so only cheap composited transforms animate.
  const frameScale = useTransform(scrollYProgress, [0, 0.55], [0.62, 1.045]);
  const imgScale = useTransform(scrollYProgress, [0, 1], [1.25, 1]);

  // Phase 2 (0.5 - 0.8): the message scrubs in over the settled image.
  const contentOpacity = useTransform(scrollYProgress, [0.5, 0.8], [0, 1]);
  const contentY = useTransform(scrollYProgress, [0.5, 0.8], [48, 0]);

  return (
    <section ref={sectionRef} className={styles.section}>
      <div className={styles.sticky}>
        <motion.div className={styles.frame} style={{ scale: frameScale }} aria-hidden="true">
          <motion.img
            src="/images/home/hero-banner-top-bar-bg.webp"
            alt=""
            className={styles.img}
            style={{ scale: imgScale }}
            loading="lazy"
            decoding="async"
          />
          <div className={styles.imgOverlay} />
        </motion.div>

        <motion.div className={styles.content} style={{ opacity: contentOpacity, y: contentY }}>
          <span className={styles.badge}>BUILT DIFFERENT</span>

          <h2 className={styles.title}>
            Every team. Every workflow.
            <br />
            <span className={styles.titleAccent}>One platform.</span>
          </h2>

          <p className={styles.subtitle}>
            rtylr isn&apos;t seven tools glued together. It&apos;s one platform where a POS sale
            depletes inventory, flows tips into payroll, and enriches the customer profile, from the
            same database, in real time.
          </p>

          <div className={styles.stats}>
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
          </div>
        </motion.div>
      </div>
    </section>
  );
};

export default ParallaxShowcase;
