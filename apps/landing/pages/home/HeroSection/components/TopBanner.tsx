import React from 'react';
import { APP_URLS, Button } from '@voxire/shared';
import { navigateExternal } from '../../../../lib/navigation';
import { FaArrowRight } from 'react-icons/fa';
import styles from '../HeroSection.module.scss';

const TopBanner: React.FC = () => {
  const handleRegisterClick = () => {
    navigateExternal(`${APP_URLS.auth}/register`);
  };

  return (
    <div className={styles.topBanner}>
      <div className={styles.bannerContainer}>
        <div className={styles.leftSection}>
          <div className={styles.badgeLine}>
            <span className={styles.alphaBadge}>ALPHA</span>
            <span className={styles.badgeText}>The all-in-one POS that grows with you</span>
          </div>
          <h3 className={styles.mainHeading}>
            Run your business with Rtylr
          </h3>
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
            src="/images/home/hero-banner-top-bar-bg.png"
            alt="Banner"
            className={styles.bannerImage}
          />
        </div>
      </div>
    </div>
  );
};

export default TopBanner;
