import React from 'react';
import { APP_URLS, Button } from '@voxire/shared';
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
    <div className={styles.topBanner}>
      {/* Desktop Layout */}
      <div className={styles.bannerContainer}>
        <div className={styles.leftSection}>
          <div className={styles.badgeLine}>
            <span className={styles.alphaBadge}>ALPHA</span>
            <span className={styles.badgeText}>
              The all-in-one POS that grows with you
            </span>
          </div>
          <h3 className={styles.mainHeading}>Run your business with Rtylr</h3>
          <Button
            variant="dark-outline"
            className={styles.registerButton}
            onClick={handleRegisterClick}
            suffix={<FaArrowRight />}
          >
            Get Started Free
          </Button>
        </div>

        <div className={styles.rightSection}>
          <img
            src="/images/home/hero-banner-top-bar-bg.webp"
            alt="Banner"
            className={styles.bannerImage}
          />
        </div>
      </div>

      {/* Mobile Layout */}
      <div className={styles.mobileBanner}>
        <div className={styles.badgeLine}>
          <span className={styles.alphaBadge}>ALPHA</span>
          <div className={styles.badgeTextContainer}>
            <div className={styles.badgeText}>
              <span className={styles.badgeTextContent}>
                The all-in-one POS that grows with you&nbsp;&nbsp;&nbsp;&nbsp;
              </span>
              <span className={styles.badgeTextContent} aria-hidden="true">
                The all-in-one POS that grows with you&nbsp;&nbsp;&nbsp;&nbsp;
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
    </div>
  );
};

export default TopBanner;
