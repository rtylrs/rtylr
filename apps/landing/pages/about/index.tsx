import React from 'react';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { useRouter } from 'next/router';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../lib/navigation';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { aboutPageSchema } from '../../lib/structured-data';
import {
  blurFadeIn,
  staggerBlur,
  viewportOnce,
  slideInLeft,
  slideInRight,
} from '../../lib/animations';
import Lines from '../../components/text/Lines';
import styles from './about.module.scss';

// The app lists are product names, not copy. Label and description
// resolve from about.clusters, matched by position.
const CLUSTER_APPS: string[][] = [
  ['POS', 'CRM'],
  ['ERP', 'Finance'],
  ['HR', 'Flow', 'Insights'],
];



const AboutPage = (): React.JSX.Element => {
  const { t } = useTranslation();
  const clusters = t('about.clusters', { returnObjects: true }) as unknown as {
    label: string;
    description: string;
  }[];
  const values = t('about.values', { returnObjects: true }) as unknown as {
    title: string;
    body: string;
  }[];
  const router = useRouter();

  const handleGetStarted = (): void => {
    navigateExternal(`${APP_URLS.auth}/register`);
  };

  const handleBookDemo = (): void => {
    void router.push('/demo');
  };

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title={t('about.meta.title')}
          description={t('about.meta.description')}
          canonical="https://rtylr.com/about/"
          structuredData={aboutPageSchema({
            description:
              "We're building the operating system for commerce. One platform for POS, ERP, HR, CRM, Finance, Flow and Insights.",
          })}
        />

        <div className={styles.aboutPage}>
          {/* Hero */}
          <motion.section
            className={styles.hero}
            initial="hidden"
            animate="visible"
            variants={staggerBlur}
          >
            <div className={styles.heroInner}>
              <motion.span className={styles.heroEyebrow} variants={blurFadeIn}>
                <span className={styles.eyebrowDot} />
                {t('about.hero.eyebrow')}
              </motion.span>
              <motion.h1 className={styles.heroTitle} variants={blurFadeIn}>
                <Lines text={t('about.hero.title')} />
              </motion.h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                Most small businesses run on six to eight disconnected apps. They export CSVs all
                Friday. They reconcile on Sundays. They pay for integrations they hate. We think
                that&apos;s a bug in the industry. rtylr is our fix.
              </motion.p>
            </div>
          </motion.section>

          {/* Mission */}
          <motion.section
            className={styles.missionSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <div className={styles.sectionContainer}>
              <div className={styles.missionGrid}>
                <motion.div className={styles.missionLeft} variants={slideInLeft}>
                  <span className={styles.sectionLabel}>{t('about.mission.eyebrow')}</span>
                  <h2 className={styles.missionTitle}>{t('about.mission.title')}</h2>
                  <p className={styles.missionBody}>
                    Commerce software should work the way commerce does, one sale, one inventory
                    drop, one payroll line, one customer record. Not seven apps pretending they talk
                    to each other. rtylr replaces the whole stack with a single platform where
                    everything updates at once, because it&apos;s all the same database underneath.
                  </p>
                </motion.div>
              </div>
            </div>
          </motion.section>

          {/* Story */}
          <motion.section
            className={styles.storySection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <div className={styles.sectionContainer}>
              <motion.span className={styles.sectionLabel} variants={blurFadeIn}>
                {t('about.story.eyebrow')}
              </motion.span>
              <motion.h2 className={styles.storyTitle} variants={blurFadeIn}>
                {t('about.story.title')}
              </motion.h2>
              <motion.div className={styles.storyBody} variants={blurFadeIn}>
                <p>
                  We kept watching businesses in Lebanon and across the region stitch together
                  QuickBooks for the books, a standalone POS for the register, a separate HR portal
                  for payroll, and a CRM that never talked to any of them. Every week the owner
                  spent Friday exporting CSVs and Sunday reconciling them.
                </p>
                <p>
                  The tools were fine. The problem was that they were separate. Seven subscriptions,
                  seven UIs, seven places your data could disagree with itself. And somehow, the
                  customer still got billed for &ldquo;integrations.&rdquo;
                </p>
                <p>
                  So in 2024 Abed and Mohammad left their corporate roles and started building what
                  should have existed all along, a single platform for the whole business, with AI
                  wired in from the ground floor.
                </p>
              </motion.div>
            </div>
          </motion.section>

          {/* Product clusters */}
          <motion.section
            className={styles.productsSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <div className={styles.sectionContainer}>
              <motion.span className={styles.sectionLabel} variants={blurFadeIn}>
                {t('about.platform.eyebrow')}
              </motion.span>
              <motion.h2 className={styles.sectionTitle} variants={blurFadeIn}>
                {t('about.platform.title')}
              </motion.h2>
              <div className={styles.productsGrid}>
                {CLUSTER_APPS.map((apps, ci) => (
                  <motion.div
                    key={clusters[ci].label}
                    className={styles.productCard}
                    variants={blurFadeIn}
                  >
                    <div className={styles.productCardTop}>
                      <h3 className={styles.productCardLabel}>{clusters[ci].label}</h3>
                      <div className={styles.productPills}>
                        {apps.map((app) => (
                          <span key={app} className={styles.appPill}>
                            {app}
                          </span>
                        ))}
                      </div>
                    </div>
                    <p className={styles.productCardDescription}>{clusters[ci].description}</p>
                  </motion.div>
                ))}
              </div>
            </div>
          </motion.section>

          {/* Values */}
          <motion.section
            className={styles.valuesSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <div className={styles.sectionContainer}>
              <motion.span className={styles.sectionLabel} variants={blurFadeIn}>
                Values
              </motion.span>
              <motion.h2 className={styles.sectionTitle} variants={blurFadeIn}>
                {t('about.valuesTitle')}
              </motion.h2>
              <div className={styles.valuesGrid}>
                {values.map((value) => (
                  <motion.div key={value.title} className={styles.valueCard} variants={blurFadeIn}>
                    <div className={styles.valueDot} aria-hidden="true" />
                    <h3 className={styles.valueTitle}>{value.title}</h3>
                    <p className={styles.valueBody}>{value.body}</p>
                  </motion.div>
                ))}
              </div>
            </div>
          </motion.section>

          {/* CTA */}
          <motion.section
            className={styles.ctaSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <div className={styles.sectionContainer}>
              <motion.div className={styles.ctaInner} variants={blurFadeIn}>
                <h2 className={styles.ctaTitle}>{t('about.cta.title')}</h2>
                <p className={styles.ctaSubtitle}>
                {t('about.cta.subtitle')}
              </p>
                <div className={styles.ctaButtons}>
                  <button className={styles.ctaPrimary} onClick={handleGetStarted}>
                {t('common.cta.getStarted')}
              </button>
                  <button className={styles.ctaSecondary} onClick={handleBookDemo}>
                {t('common.cta.bookDemo')}
              </button>
                </div>
              </motion.div>
            </div>
          </motion.section>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default AboutPage;
