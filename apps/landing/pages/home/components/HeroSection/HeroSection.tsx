import React from 'react';
import { APP_URLS, Button } from '@voxire/shared';
import styles from './HeroSection.module.scss';
import { FaArrowRight } from 'react-icons/fa';

const HeroSection = (): JSX.Element => {
  const handleRegisterClick = () => {
    window.location.href = `${APP_URLS.auth}/register`;
  };

  const handleFreeDemo = () => {
    window.location.href = `${APP_URLS.auth}/register`;
  };

  return (
    <section className={styles.heroSection}>
      <div className={styles.topBanner}>
        <div className={styles.bannerContainer}>
          <div className={styles.leftSection}>
            <h3>Try our online</h3>
            <h3>free demo now!</h3>
            <Button
              variant="dark-outline"
              className={styles.registerButton}
              onClick={handleRegisterClick}
              suffix={<FaArrowRight />}
            >
              Register Now
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
                <h3>Super Quick</h3>
                <h3>& Easy Set-up</h3>
              </div>
              <div className={styles.videoPlaceholder}>
                <div className={styles.playButton}>▶</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className={styles.mainHero}>
        <div className={styles.container}>
          <div className={styles.heroContent}>
            <h1 className={styles.heroTitle}>
              Stop Juggling Systems.
              <br />
              Start Selling Smarter.
            </h1>
          </div>
          <div className={styles.heroSidebar}>
            <button className={styles.demoButton} onClick={handleFreeDemo}>
              Free Demo
            </button>
            <p className={styles.description}>
              Cloud-native management software that unifies POS, inventory,
              eCommerce and warehousing — AI-powered and offline sync.
            </p>
          </div>
        </div>
      </div>

      {/* Trusted By Section */}
      <div className={styles.trustedBy}>
        <div className={styles.trustedContainer}>
          <div className={styles.logoScroll}>
            <div className={styles.logoItem}>eBird</div>
            <div className={styles.logoItem}>rarify</div>
            <div className={styles.logoItem}>_zapier</div>
            <div className={styles.logoItem}>SUPERHUMAN</div>
            <div className={styles.logoItem}>LottieFiles</div>
            <div className={styles.logoItem}>dribbble</div>
            <div className={styles.logoItem}>MR MARVIS</div>
            <div className={styles.logoItem}>HOD</div>
          </div>
        </div>
      </div>
    </section>
  );
};

export default HeroSection;
