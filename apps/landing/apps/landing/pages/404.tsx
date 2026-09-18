import React from 'react';
import { cn } from '../lib/cn';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
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

// href only. The label and description resolve from notFound.links,
// matched by position.
const SUGGESTED_HREFS = ['/', '/pricing', '/pos', '/demo', '/contact'];

const NotFoundPage = (): React.JSX.Element => {
  const { t } = useTranslation();
  const links = t('notFound.links', { returnObjects: true }) as unknown as {
    label: string;
    description: string;
  }[];
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
          title={t('notFound.meta.title')}
          description={t('notFound.meta.description')}
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
                {t('notFound.eyebrow')}
              </motion.span>

            <SplitWords
              as="h1"
              className={styles.title}
              text={t('notFound.title')}
              delay={0.1}
            />

            <motion.p className={styles.subtitle} variants={blurFadeIn}>
              The link may be broken, renamed, or the page hasn&apos;t been built yet. Try one of
              these instead.
            </motion.p>

            <motion.div className={styles.buttons} variants={blurFadeIn}>
              <button type="button" onClick={goHome} className={styles.primaryBtn}>
                {t('notFound.backHome')}
              </button>
              <button type="button" onClick={goStartFree} className={styles.outlineBtn}>
                {t('notFound.startFree')}
              </button>
            </motion.div>

            <motion.div className={styles.suggestionsGrid} variants={blurFadeIn}>
              {SUGGESTED_HREFS.map((href, i) => (
                <Link key={href} href={href} className={styles.suggestionCell}>
                  <span className={styles.suggestionLabel}>{links[i].label}</span>
                  <span className={styles.suggestionDesc}>{links[i].description}</span>
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
