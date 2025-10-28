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
          <h3>Start Your Free Trial</h3>
          <p className={styles.subtitle}>No credit card required</p>
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
          <div className={styles.qrSection}>
            <div className={styles.qrCodePlaceholder}>
              <div className={styles.qrBox}>
                <img
                  src="/images/home/hero-banner-qr-code.png"
                  alt="QR Code"
                  style={{
                    width: '100%',
                    height: '100%',
                    objectFit: 'contain',
                  }}
                />
              </div>
            </div>
          </div>

          <div className={styles.videoSection}>
            <div className={styles.videoText}>
              <h3>Quick & Easy</h3>
              <h3>Setup</h3>
            </div>
            <div className={styles.videoPlaceholder}>
              <div className={styles.playButton}>▶</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default TopBanner;
