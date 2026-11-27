import React from 'react';
import { BusinessInfo } from '../../types';
import styles from './AboutPage.module.scss';

interface AboutPageProps {
  business: BusinessInfo;
}

function safeWebsiteHref(raw?: string): string | null {
  if (!raw) return null;
  const trimmed = raw.trim();
  if (!trimmed) return null;
  if (/^https?:\/\//i.test(trimmed)) return trimmed;
  if (/^[a-z0-9.-]+\.[a-z]{2,}/i.test(trimmed)) return `https://${trimmed}`;
  return null;
}

function normalizeHandle(raw: string, baseUrl: string, prefixAt = false): { href: string; label: string } {
  const trimmed = raw.trim();
  if (/^https?:\/\//i.test(trimmed)) {
    return { href: trimmed, label: trimmed.replace(/^https?:\/\/(www\.)?/i, '') };
  }
  const clean = trimmed.replace(/^@+/, '');
  return {
    href: `${baseUrl}${prefixAt ? '@' : ''}${encodeURIComponent(clean)}`,
    label: prefixAt ? `@${clean}` : clean,
  };
}

const AboutPage: React.FC<AboutPageProps> = ({ business }) => {
  const hasStory = Boolean(business.story?.trim());
  const websiteHref = safeWebsiteHref(business.website);
  const hasInfo = Boolean(
    business.hours?.trim() ||
      business.address?.trim() ||
      business.phone?.trim() ||
      business.email?.trim() ||
      websiteHref
  );
  const hasSocials = Boolean(
    business.socials &&
      (business.socials.instagram?.trim() ||
        business.socials.facebook?.trim() ||
        business.socials.tiktok?.trim())
  );

  if (!hasStory && !hasInfo && !hasSocials) {
    return (
      <div className={styles.about}>
        <section className={styles.about__section}>
          <p className={styles.about__story}>
            Business details haven&apos;t been added yet.
          </p>
        </section>
      </div>
    );
  }

  return (
    <div className={styles.about}>
      {hasStory && (
        <section className={styles.about__section}>
          <h2 className={styles.about__sectionTitle}>Our Story</h2>
          <p className={styles.about__story}>{business.story}</p>
        </section>
      )}

      {hasInfo && (
        <section className={styles.about__section}>
          <h2 className={styles.about__sectionTitle}>Info</h2>
          <div className={styles.about__infoGrid}>
            {business.hours?.trim() && (
              <div className={styles.about__infoCard}>
                <span className={styles.about__infoIcon}>🕐</span>
                <div>
                  <span className={styles.about__infoLabel}>Hours</span>
                  <span className={styles.about__infoValue}>{business.hours}</span>
                </div>
              </div>
            )}
            {business.address?.trim() && (
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
            {business.phone?.trim() && (
              <a className={styles.about__infoCard} href={`tel:${business.phone}`}>
                <span className={styles.about__infoIcon}>📞</span>
                <div>
                  <span className={styles.about__infoLabel}>Phone</span>
                  <span className={styles.about__infoValue}>{business.phone}</span>
                </div>
              </a>
            )}
            {business.email?.trim() && (
              <a className={styles.about__infoCard} href={`mailto:${business.email}`}>
                <span className={styles.about__infoIcon}>✉️</span>
                <div>
                  <span className={styles.about__infoLabel}>Email</span>
                  <span className={styles.about__infoValue}>{business.email}</span>
                </div>
              </a>
            )}
            {websiteHref && (
              <a
                className={styles.about__infoCard}
                href={websiteHref}
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
      )}

      {hasSocials && business.socials && (
        <section className={styles.about__section}>
          <h2 className={styles.about__sectionTitle}>Follow Us</h2>
          <div className={styles.about__socials}>
            {business.socials.instagram?.trim() && (() => {
              const ig = normalizeHandle(business.socials!.instagram!, 'https://instagram.com/', false);
              return (
                <a
                  className={styles.about__socialBtn}
                  href={ig.href}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  Instagram · @{ig.label.replace(/^@+/, '')}
                </a>
              );
            })()}
            {business.socials.facebook?.trim() && (() => {
              const fb = normalizeHandle(business.socials!.facebook!, 'https://facebook.com/', false);
              return (
                <a
                  className={styles.about__socialBtn}
                  href={fb.href}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  Facebook · {fb.label}
                </a>
              );
            })()}
            {business.socials.tiktok?.trim() && (() => {
              const tt = normalizeHandle(business.socials!.tiktok!, 'https://tiktok.com/', true);
              return (
                <a
                  className={styles.about__socialBtn}
                  href={tt.href}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  TikTok · {tt.label}
                </a>
              );
            })()}
          </div>
        </section>
      )}
    </div>
  );
};

export default AboutPage;
