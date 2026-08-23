import React from 'react';
import { motion } from 'motion/react';
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
import styles from './about.module.scss';

const PRODUCT_CLUSTERS = [
  {
    label: 'Front of House',
    apps: ['POS', 'CRM'],
    description: 'Manage every customer touchpoint, point of sale, loyalty, campaigns, and beyond.',
  },
  {
    label: 'Operations',
    apps: ['ERP', 'Finance'],
    description: 'Keep inventory, purchasing, invoices, and P&L tight across every location.',
  },
  {
    label: 'People',
    apps: ['HR', 'Flow', 'Insights'],
    description: 'Staff scheduling, project management, and analytics all unified in one place.',
  },
];

const VALUES = [
  {
    title: 'Everything in one place',
    body: 'No more juggling seven tools. One login, one workspace, one source of truth for your business.',
  },
  {
    title: 'Built for teams',
    body: 'From a two-person shop to a 500-seat enterprise, rtylr scales to fit how your team actually works.',
  },
  {
    title: 'AI-first by design',
    body: 'Every app ships with an embedded AI assistant, not a bolt-on, but part of the core experience.',
  },
  {
    title: 'Always available offline',
    body: 'POS and critical workflows work without internet. Sync happens automatically when you reconnect.',
  },
];

const AboutPage = (): JSX.Element => {
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
          title="About rtylr · The All-in-One Business OS for Modern Teams"
          description="We're building the operating system for commerce. One platform for POS, ERP, HR, CRM, Finance, Flow, and Insights."
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
                ABOUT VOXIRE &amp; RTYLR
              </motion.span>
              <motion.h1 className={styles.heroTitle} variants={blurFadeIn}>
                An operating system
                <br />
                for real businesses.
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
                  <span className={styles.sectionLabel}>OUR MISSION</span>
                  <h2 className={styles.missionTitle}>Stop the CSV life.</h2>
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
                THE STORY
              </motion.span>
              <motion.h2 className={styles.storyTitle} variants={blurFadeIn}>
                We got tired of the duct tape.
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
                The Platform
              </motion.span>
              <motion.h2 className={styles.sectionTitle} variants={blurFadeIn}>
                Seven apps. One platform.
              </motion.h2>
              <div className={styles.productsGrid}>
                {PRODUCT_CLUSTERS.map((cluster) => (
                  <motion.div
                    key={cluster.label}
                    className={styles.productCard}
                    variants={blurFadeIn}
                  >
                    <div className={styles.productCardTop}>
                      <h3 className={styles.productCardLabel}>{cluster.label}</h3>
                      <div className={styles.productPills}>
                        {cluster.apps.map((app) => (
                          <span key={app} className={styles.appPill}>
                            {app}
                          </span>
                        ))}
                      </div>
                    </div>
                    <p className={styles.productCardDescription}>{cluster.description}</p>
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
                What we stand for.
              </motion.h2>
              <div className={styles.valuesGrid}>
                {VALUES.map((value) => (
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
                <h2 className={styles.ctaTitle}>Join the businesses already using rtylr.</h2>
                <p className={styles.ctaSubtitle}>
                  Free forever for up to 5 members. No credit card required.
                </p>
                <div className={styles.ctaButtons}>
                  <button className={styles.ctaPrimary} onClick={handleGetStarted}>
                    Get started
                  </button>
                  <button className={styles.ctaSecondary} onClick={handleBookDemo}>
                    Book a demo
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
