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
              variant="soft"
              onClick={handleGetStarted}
              width="200px"
              height="44px"
            >
              Get started free
            </Button>
            <Button
              variant="dark-outline"
              onClick={handleBookDemo}
              width="200px"
              height="44px"
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
