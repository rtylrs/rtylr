import React from 'react';
import { motion } from 'motion/react';
import { useRouter } from 'next/router';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../../../lib/navigation';
import styles from '../HeroSection.module.scss';

const APPS = [
  { id: 'pos', label: 'POS', color: '#4485c4' },
  { id: 'erp', label: 'Inventory', color: '#03aed2' },
  { id: 'hr', label: 'HR', color: '#4485c4' },
  { id: 'crm', label: 'CRM', color: '#03aed2' },
  { id: 'finance', label: 'Finance', color: '#4485c4' },
  { id: 'flow', label: 'Flow', color: '#03aed2' },
  { id: 'insights', label: 'Insights', color: '#4485c4' },
];

const MainHero: React.FC = (): React.JSX.Element => {
  const router = useRouter();

  const handleGetStarted = (): void => {
    navigateExternal(`${APP_URLS.auth}/register`);
  };

  const handleBookDemo = (): void => {
    void router.push('/demo');
  };

  return (
    <div className={styles.mainHero}>
      {/* Announcement badge */}
      <motion.div
        className={styles.announceBadge}
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.05, duration: 0.6, ease: [0.16, 1, 0.3, 1] }}
      >
        <span className={styles.announceBadgePill}>NEW</span>
        <span className={styles.announceBadgeText}>
          AI built into every app — launch your business in minutes
        </span>
        <span className={styles.announceBadgeArrow}>→</span>
      </motion.div>

      {/* Headline — full width, massive */}
      <motion.h1
        className={styles.heroTitle}
        initial={{ opacity: 0, y: 40, filter: 'blur(12px)' }}
        animate={{ opacity: 1, y: 0, filter: 'blur(0px)' }}
        transition={{ delay: 0.18, duration: 1.0, ease: [0.16, 1, 0.3, 1] }}
      >
        Your entire business,
        <br />
        <span className={styles.heroTitleAccent}>one platform.</span>
      </motion.h1>

      {/* Subtitle */}
      <motion.p
        className={styles.heroSubtitle}
        initial={{ opacity: 0, y: 24 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.36, duration: 0.85, ease: [0.16, 1, 0.3, 1] }}
      >
        POS · ERP · HR · CRM · Finance · Flow · Insights — all talking to each other,
        <br />
        all powered by AI, all from one login.
      </motion.p>

      {/* Dark preview card — Retool-style */}
      <motion.div
        className={styles.previewCard}
        initial={{ opacity: 0, y: 48, scale: 0.97 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        transition={{ delay: 0.5, duration: 1.0, ease: [0.16, 1, 0.3, 1] }}
      >
        {/* Card top bar */}
        <div className={styles.previewCardBar}>
          <div className={styles.previewCardDots}>
            <span />
            <span />
            <span />
          </div>
          <div className={styles.previewCardUrl}>rtylr.com · hub</div>
          <div className={styles.previewCardActions}>
            <button className={styles.previewCardBtn} onClick={handleGetStarted} type="button">
              Get started free
            </button>
          </div>
        </div>

        {/* App grid inside card */}
        <div className={styles.previewCardBody}>
          <div className={styles.previewAppsGrid}>
            {APPS.map((app, i) => (
              <motion.div
                key={app.id}
                className={styles.previewAppItem}
                initial={{ opacity: 0, y: 12 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{
                  delay: 0.65 + i * 0.07,
                  duration: 0.55,
                  ease: [0.16, 1, 0.3, 1],
                }}
              >
                <div className={styles.previewAppDot} style={{ background: app.color }} />
                <span className={styles.previewAppLabel}>{app.label}</span>
                <div className={styles.previewAppBar}>
                  <div
                    className={styles.previewAppBarFill}
                    style={{
                      width: `${55 + ((i * 17) % 40)}%`,
                      background: app.color,
                    }}
                  />
                </div>
                <span className={styles.previewAppStatus}>Active</span>
              </motion.div>
            ))}
          </div>

          {/* Fake metric row */}
          <div className={styles.previewMetrics}>
            {[
              { label: 'Revenue today', value: '$12,480' },
              { label: 'Active orders', value: '34' },
              { label: 'Staff clocked in', value: '12' },
              { label: 'Inventory alerts', value: '3' },
            ].map((m) => (
              <div key={m.label} className={styles.previewMetric}>
                <span className={styles.previewMetricValue}>{m.value}</span>
                <span className={styles.previewMetricLabel}>{m.label}</span>
              </div>
            ))}
          </div>
        </div>
      </motion.div>

      {/* "Talk to an expert" link — below card like Retool */}
      <motion.div
        className={styles.expertLink}
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 1.0, duration: 0.7 }}
      >
        <button className={styles.expertLinkBtn} onClick={handleBookDemo} type="button">
          Talk to an expert and book a demo →
        </button>
      </motion.div>
    </div>
  );
};

export default MainHero;
