import React from 'react';
import { motion } from 'framer-motion';
import { Button } from '@voxire/components';
import { APP_URLS } from '@voxire/config';
import { useRouter } from 'next/router';
import { fadeInUp, fadeIn, viewportOnce } from '../../../lib/animations';
import { navigateExternal } from '../../../lib/navigation';
import styles from './CTASection.module.scss';

const CTASection: React.FC = () => {
  const router = useRouter();

  const handleBookDemo = () => {
    router.push('/demo');
  };

  const handleGetStarted = () => {
    navigateExternal(`${APP_URLS.auth}/register`);
  };

  return (
    <section className={styles.ctaSection}>
      {/* Animated gradient background */}
      <motion.div
        className={styles.gradientBg}
        animate={{
          backgroundPosition: ['0% 50%', '100% 50%', '0% 50%'],
        }}
        transition={{
          duration: 18,
          repeat: Infinity,
          ease: 'linear',
        }}
      />

      {/* Floating orbs */}
      <motion.div
        className={styles.orbOne}
        animate={{
          x: [0, 80, 0],
          y: [0, -40, 0],
          scale: [1, 1.2, 1],
        }}
        transition={{
          duration: 14,
          repeat: Infinity,
          ease: 'easeInOut',
        }}
      />
      <motion.div
        className={styles.orbTwo}
        animate={{
          x: [0, -60, 0],
          y: [0, 50, 0],
          scale: [1.1, 1, 1.1],
        }}
        transition={{
          duration: 16,
          repeat: Infinity,
          ease: 'easeInOut',
        }}
      />
      <motion.div
        className={styles.orbThree}
        animate={{
          x: [0, 40, 0],
          y: [0, 30, 0],
          scale: [1, 1.15, 1],
        }}
        transition={{
          duration: 20,
          repeat: Infinity,
          ease: 'easeInOut',
        }}
      />

      <div className={styles.container}>
        <motion.div
          className={styles.content}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={fadeIn}
        >
          <motion.h2 className={styles.ctaTitle} variants={fadeInUp}>
            Ready to run your entire
            <br />
            business from one place?
          </motion.h2>
          <motion.p className={styles.ctaDescription} variants={fadeInUp}>
            Free forever for up to 5 members. No credit card required.
          </motion.p>
          <motion.div className={styles.ctaButtons} variants={fadeInUp}>
            <Button
              variant="primary"
              onClick={handleGetStarted}
              width="220px"
              height="56px"
              color="#4583C4"
            >
              Get started free
            </Button>
            <Button
              variant="dark-outline"
              onClick={handleBookDemo}
              width="220px"
              height="56px"
            >
              Book a demo
            </Button>
          </motion.div>
        </motion.div>
      </div>
    </section>
  );
};

export default CTASection;
