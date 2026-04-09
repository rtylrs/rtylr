import React from 'react';
import { motion } from 'framer-motion';
import { Button } from '@voxire/components';
import { useRouter } from 'next/router';
import { fadeInUp, fadeIn, viewportOnce } from '../../../lib/animations';
import styles from './CTASection.module.scss';

const CTASection: React.FC = () => {
  const router = useRouter();

  const handleBookDemo = () => {
    router.push('/demo');
  };

  const handleGetStarted = () => {
    router.push('/demo');
  };

  return (
    <section className={styles.ctaSection}>
      {/* Animated background orbs */}
      <motion.div
        className={styles.orbOne}
        animate={{
          scale: [1, 1.2, 1],
          opacity: [0.3, 0.5, 0.3],
        }}
        transition={{
          duration: 8,
          repeat: Infinity,
          ease: 'easeInOut',
        }}
      />
      <motion.div
        className={styles.orbTwo}
        animate={{
          scale: [1.2, 1, 1.2],
          opacity: [0.2, 0.4, 0.2],
        }}
        transition={{
          duration: 10,
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
          <motion.span className={styles.eyebrow} variants={fadeInUp}>
            One platform. Every department.
          </motion.span>
          <motion.h2 className={styles.ctaTitle} variants={fadeInUp}>
            Buy rtylr,
            <br />
            buy a whole business.
          </motion.h2>
          <motion.p className={styles.ctaDescription} variants={fadeInUp}>
            POS, ERP, HR, CRM, Finance, and Flow — all working together out of
            the box. Powered by AI. Free to start, no credit card required.
          </motion.p>
          <motion.div className={styles.ctaButtons} variants={fadeInUp}>
            <Button
              variant="primary"
              onClick={handleGetStarted}
              width="200px"
              height="52px"
            >
              Get Started
            </Button>
            <Button
              variant="dark-outline"
              onClick={handleBookDemo}
              width="200px"
              height="52px"
            >
              Book a Demo
            </Button>
          </motion.div>
        </motion.div>
      </div>
    </section>
  );
};

export default CTASection;
