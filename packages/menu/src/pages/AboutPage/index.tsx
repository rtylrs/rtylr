import React from 'react';
import { BusinessInfo } from '../../types';
import styles from './AboutPage.module.scss';

interface AboutPageProps {
  business: BusinessInfo;
}

const AboutPage: React.FC<AboutPageProps> = ({ business }) => {
  return (
    <div className={styles.about}>
      {business.story && (
        <section className={styles.about__section}>
          <h2 className={styles.about__sectionTitle}>Our Story</h2>
          <p className={styles.about__story}>{business.story}</p>
        </section>
      )}

      <section className={styles.about__section}>
        <h2 className={styles.about__sectionTitle}>Info</h2>
        <div className={styles.about__infoGrid}>
          {business.hours && (
            <div className={styles.about__infoCard}>
              <span className={styles.about__infoIcon}>🕐</span>
              <div>
                <span className={styles.about__infoLabel}>Hours</span>
                <span className={styles.about__infoValue}>{business.hours}</span>
              </div>
            </div>
          )}
          {business.address && (
            <a
              className={styles.about__infoCard}
              href={`https://maps.google.com/?q=${encodeURIComponent(business.address)}`}
              target="_blank"
              rel="noopener noreferrer"
            >
              <span className={styles.about__infoIcon}>📍</span>
              <div>
                <span className={styles.about__infoLabel}>Address</span>
                <span className={styles.about__infoValue}>{business.address}</span>
              </div>
            </a>
          )}
          {business.phone && (
            <a className={styles.about__infoCard} href={`tel:${business.phone}`}>
              <span className={styles.about__infoIcon}>📞</span>
              <div>
                <span className={styles.about__infoLabel}>Phone</span>
                <span className={styles.about__infoValue}>{business.phone}</span>
              </div>
            </a>
          )}
          {business.email && (
            <a className={styles.about__infoCard} href={`mailto:${business.email}`}>
              <span className={styles.about__infoIcon}>✉️</span>
              <div>
                <span className={styles.about__infoLabel}>Email</span>
                <span className={styles.about__infoValue}>{business.email}</span>
              </div>
            </a>
          )}
          {business.website && (
            <a
              className={styles.about__infoCard}
              href={business.website}
              target="_blank"
              rel="noopener noreferrer"
            >
              <span className={styles.about__infoIcon}>🌐</span>
              <div>
                <span className={styles.about__infoLabel}>Website</span>
                <span className={styles.about__infoValue}>{business.website}</span>
              </div>
            </a>
          )}
        </div>
      </section>

      {business.socials && Object.values(business.socials).some(Boolean) && (
        <section className={styles.about__section}>
          <h2 className={styles.about__sectionTitle}>Follow Us</h2>
          <div className={styles.about__socials}>
            {business.socials.instagram && (
              <a
                className={styles.about__socialBtn}
                href={`https://instagram.com/${business.socials.instagram}`}
                target="_blank"
                rel="noopener noreferrer"
              >
                Instagram · @{business.socials.instagram}
              </a>
            )}
            {business.socials.facebook && (
              <a
                className={styles.about__socialBtn}
                href={`https://facebook.com/${business.socials.facebook}`}
                target="_blank"
                rel="noopener noreferrer"
              >
                Facebook · {business.socials.facebook}
              </a>
            )}
            {business.socials.tiktok && (
              <a
                className={styles.about__socialBtn}
                href={`https://tiktok.com/@${business.socials.tiktok}`}
                target="_blank"
                rel="noopener noreferrer"
              >
                TikTok · @{business.socials.tiktok}
              </a>
            )}
          </div>
        </section>
      )}
    </div>
  );
};

export default AboutPage;
