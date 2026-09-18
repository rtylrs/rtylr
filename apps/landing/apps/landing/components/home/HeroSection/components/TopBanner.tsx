import React from 'react';
import { motion } from 'motion/react';
import { Button } from '@voxire/components';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../../../lib/navigation';
import { FaArrowRight } from 'react-icons/fa';
import { HiArrowUpRight } from 'react-icons/hi2';
import styles from '../HeroSection.module.scss';

const TopBanner: React.FC = () => {
  const handleRegisterClick = () => {
    navigateExternal(`${APP_URLS.auth}/register`);
  };

  const handleMobileLinkClick = () => {
    navigateExternal(`${APP_URLS.auth}/register`);
  };

  return (
    <motion.div
      className={styles.topBanner}
      initial={{ opacity: 0, y: -20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.6, ease: [0.25, 0.46, 0.45, 0.94] }}
    >
      {/* Desktop Layout */}
      <div className={styles.bannerContainer}>
        <div className={styles.leftSection}>
          <motion.div
            className={styles.badgeLine}
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: 0.3, duration: 0.5 }}
          >
            <span className={styles.alphaBadge}>BETA</span>
            <span className={styles.badgeText}>The all-in-one platform for your business</span>
          </motion.div>
          <motion.h3
            className={styles.mainHeading}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.4, duration: 0.6 }}
          >
            Run your business with Rtylr
          </motion.h3>
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.5, duration: 0.5 }}
          >
            <Button
              variant="dark-outline"
              className={styles.registerButton}
              onClick={handleRegisterClick}
              suffix={<FaArrowRight />}
            >
              Get Started Free
            </Button>
          </motion.div>
        </div>

        <motion.div
          className={styles.rightSection}
          initial={{ opacity: 0, scale: 1.05 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ delay: 0.2, duration: 0.8, ease: 'easeOut' }}
        >
          <img
            src="/images/home/hero-banner-top-bar-bg.webp"
            alt="Banner"
            className={styles.bannerImage}
          />
        </motion.div>
      </div>

      {/* Mobile Layout */}
      <div className={styles.mobileBanner}>
        <div className={styles.badgeLine}>
          <span className={styles.alphaBadge}>BETA</span>
          <div className={styles.badgeTextContainer}>
            <div className={styles.badgeText}>
              <span className={styles.badgeTextContent}>
                The all-in-one platform for your business&nbsp;&nbsp;&nbsp;&nbsp;
              </span>
              <span className={styles.badgeTextContent} aria-hidden="true">
                The all-in-one platform for your business&nbsp;&nbsp;&nbsp;&nbsp;
              </span>
            </div>
          </div>
        </div>
        <button
          className={styles.mobileLink}
          onClick={handleMobileLinkClick}
          aria-label="Learn more"
        >
          <HiArrowUpRight />
        </button>
      </div>
    </motion.div>
  );
};

export default TopBanner;
