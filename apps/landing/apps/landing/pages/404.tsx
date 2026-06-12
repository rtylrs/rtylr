import React from 'react';
import { cn } from '../lib/cn';
import { motion } from 'motion/react';
import Link from 'next/link';
import { useRouter } from 'next/router';
import { APP_URLS } from '@voxire/config';
import AppLayout from '../components/Layout/AppLayout';
import MainLayout from '../layouts/Main/MainLayout';
import HelmetContainer from '../components/HelmetContainer';
import { navigateExternal } from '../lib/navigation';
import { blurFadeIn, staggerBlur } from '../lib/animations';
import { SplitWords } from '../components/motion/SplitWords';
import styles from './404.module.scss';

const SUGGESTED_LINKS = [
  { href: '/', label: 'Homepage', description: 'Back to the start.' },
  { href: '/pricing', label: 'Pricing', description: 'See plans and per-seat pricing.' },
  {
    href: '/pos',
    label: 'Products',
    description: 'POS, ERP, HR, CRM, Finance, Flow.',
  },
  { href: '/demo', label: 'Book a demo', description: 'See rtylr running on your data.' },
  { href: '/contact', label: 'Contact', description: 'Talk to someone on our team.' },
];

const NotFoundPage = (): JSX.Element => {
  const router = useRouter();

  const goHome = (): void => {
    void router.push('/');
  };

  const goStartFree = (): void => {
    navigateExternal(`${APP_URLS.auth}/register`);
  };

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title="Page not found · rtylr"
          description="The page you're looking for doesn't exist. Here's where you can go next."
          robots="noindex, follow"
        />

        <main className={styles.page}>
          {/* Framing rails, match the hero aesthetic */}
          <div className={styles.frame} aria-hidden="true">
            <span className={cn(styles.vLine, styles.vLineLeft)} />
            <span className={cn(styles.vLine, styles.vLineRight)} />
          </div>

          <motion.section
            className={styles.content}
            initial="hidden"
            animate="visible"
            variants={staggerBlur}
          >
            <motion.span className={styles.eyebrow} variants={blurFadeIn}>
              ERROR · 404
            </motion.span>

            <SplitWords
              as="h1"
              className={styles.title}
              text={"This page\ndoesn't exist."}
              delay={0.1}
            />

            <motion.p className={styles.subtitle} variants={blurFadeIn}>
              The link may be broken, renamed, or the page hasn&apos;t been built yet. Try one of
              these instead.
            </motion.p>

            <motion.div className={styles.buttons} variants={blurFadeIn}>
              <button type="button" onClick={goHome} className={styles.primaryBtn}>
                Back to home
              </button>
              <button type="button" onClick={goStartFree} className={styles.outlineBtn}>
                Start free
              </button>
            </motion.div>

            <motion.div className={styles.suggestionsGrid} variants={blurFadeIn}>
              {SUGGESTED_LINKS.map((link) => (
                <Link key={link.href} href={link.href} className={styles.suggestionCell}>
                  <span className={styles.suggestionLabel}>{link.label}</span>
                  <span className={styles.suggestionDesc}>{link.description}</span>
                  <span className={styles.suggestionArrow} aria-hidden="true">
                    →
                  </span>
                </Link>
              ))}
            </motion.div>
          </motion.section>
        </main>
      </MainLayout>
    </AppLayout>
  );
};

export default NotFoundPage;
