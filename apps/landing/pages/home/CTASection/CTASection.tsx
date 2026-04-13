import React from 'react';
import { motion } from 'motion/react';
import { Button } from '@voxire/components';
import { APP_URLS } from '@voxire/config';
import { useRouter } from 'next/router';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../../lib/animations';
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
      {/* Ambient orbs */}
      <motion.div
        style={{
          position: 'absolute',
          width: 700,
          height: 700,
          borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(69,131,196,0.13) 0%, transparent 70%)',
          top: -200,
          left: -200,
          pointerEvents: 'none',
          zIndex: 0,
          filter: 'blur(70px)',
        }}
        animate={{ x: [0, 60, -30, 0], y: [0, 40, 20, 0], scale: [1, 1.1, 0.95, 1] }}
        transition={{ duration: 22, repeat: Infinity, ease: 'easeInOut' }}
      />
      <motion.div
        style={{
          position: 'absolute',
          width: 500,
          height: 500,
          borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(126,68,150,0.1) 0%, transparent 70%)',
          bottom: -150,
          right: -100,
          pointerEvents: 'none',
          zIndex: 0,
          filter: 'blur(60px)',
        }}
        animate={{ x: [0, -50, 20, 0], y: [0, -30, -15, 0], scale: [1, 1.12, 1.05, 1] }}
        transition={{ duration: 28, repeat: Infinity, ease: 'easeInOut' }}
      />

      <div className={styles.container}>
        <motion.div
          className={styles.content}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={staggerBlur}
        >
          <motion.h2 className={styles.ctaTitle} variants={blurFadeIn}>
            Ready to run your entire
            <br />
            business from one place?
          </motion.h2>
          <motion.p className={styles.ctaDescription} variants={blurFadeIn}>
            Free forever for up to 5 members. No credit card required.
          </motion.p>
          <motion.div className={styles.ctaButtons} variants={blurFadeIn}>
            <Button variant="soft" onClick={handleGetStarted} width="200px" height="44px">
              Get started free
            </Button>
            <Button variant="dark-outline" onClick={handleBookDemo} width="200px" height="44px">
              Book a demo
            </Button>
          </motion.div>
        </motion.div>
      </div>
    </section>
  );
};

export default CTASection;
