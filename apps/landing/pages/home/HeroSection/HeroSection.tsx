import React from 'react';
import { APP_URLS, Button } from '@voxire/shared';
import { navigateExternal } from '../../../lib/navigation';
import styles from './HeroSection.module.scss';
import { FaArrowRight } from 'react-icons/fa';

const HeroSection = (): JSX.Element => {
  const handleRegisterClick = () => {
    navigateExternal(`${APP_URLS.auth}/register`);
  };

  const handleFreeDemo = () => {
    navigateExternal(`${APP_URLS.auth}/register`);
  };

  return (
    <section className={styles.heroSection}>
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
              Book a Demo
            </button>
            <p className={styles.description}>
              Your single source of truth for in-store, online, and back-office
              operations. POS · Inventory · eCommerce · Analytics — works
              offline, scales globally.
            </p>
          </div>
        </div>
      </div>

      {/* Stats Section */}
      <div className={styles.trustedBy}>
        <div className={styles.trustedContainer}>
          <div className={styles.logoScroll}>
            <div className={styles.logoItem}>
              <div className={styles.statNumber}>50K+</div>
              <div className={styles.statLabel}>Daily Transactions</div>
            </div>
            <div className={styles.logoItem}>
              <div className={styles.statNumber}>99.9%</div>
              <div className={styles.statLabel}>Uptime</div>
            </div>
            <div className={styles.logoItem}>
              <div className={styles.statNumber}>20+</div>
              <div className={styles.statLabel}>Countries</div>
            </div>
            <div className={styles.logoItem}>
              <div className={styles.statNumber}>24/7</div>
              <div className={styles.statLabel}>Support</div>
            </div>
            <div className={styles.logoItem}>
              <div className={styles.statNumber}>Offline</div>
              <div className={styles.statLabel}>Always Works</div>
            </div>
            <div className={styles.logoItem}>
              <div className={styles.statNumber}>AI</div>
              <div className={styles.statLabel}>Powered</div>
            </div>
            <div className={styles.logoItem}>
              <div className={styles.statNumber}>Multi</div>
              <div className={styles.statLabel}>Language</div>
            </div>
            <div className={styles.logoItem}>
              <div className={styles.statNumber}>Cloud</div>
              <div className={styles.statLabel}>Native</div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};

export default HeroSection;
