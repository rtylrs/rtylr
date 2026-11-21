import React from 'react';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { useRouter } from 'next/router';
import {
  ArrowRight,
  BatteryCharging,
  BellRing,
  QrCode,
  ReceiptText,
  SearchX,
  SprayCan,
  Wallet,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { SplitWords } from '../../components/motion/SplitWords';
import { productWithBreadcrumbs } from '../../lib/structured-data';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import { track, EV_CTA_CLICK } from '../../lib/analytics';
import { PAGER_BRAND, PAGER_TRY_URL } from '../../lib/pager';
import styles from './pager.module.scss';

interface Card {
  title: string;
  body: string;
}

// Icons only. The copy resolves from pager.how.steps and
// pager.compare.items, matched by position.
const STEP_ICONS: LucideIcon[] = [QrCode, ReceiptText, BellRing];
const COMPARE_ICONS: LucideIcon[] = [Wallet, BatteryCharging, SprayCan, SearchX];

const TryButton: React.FC<{ source: string }> = ({ source }) => {
  const { t } = useTranslation();
  const router = useRouter();
  return (
    <a
      href={PAGER_TRY_URL}
      className={styles.ctaPrimary}
      onClick={() => track(EV_CTA_CLICK, { cta: 'pager-try', source, page: router.pathname })}
    >
      {t('pager.tryCta')}
      <ArrowRight size={15} strokeWidth={2.2} aria-hidden="true" className="rtl-mirror" />
    </a>
  );
};

const CardGrid: React.FC<{ cards: Card[]; icons: LucideIcon[]; numbered?: boolean }> = ({
  cards,
  icons,
  numbered = false,
}) => (
  <motion.ol
    className={numbered ? styles.steps : styles.cards}
    initial="hidden"
    whileInView="visible"
    viewport={viewportOnce}
    variants={staggerBlur}
  >
    {cards.map((card, i) => {
      const Icon = icons[i];
      return (
        <motion.li key={card.title} className={styles.card} variants={blurFadeIn}>
          <span className={styles.cardIcon} aria-hidden="true">
            {numbered ? <span className={styles.stepNumber}>{i + 1}</span> : null}
            {Icon ? <Icon size={22} strokeWidth={1.6} /> : null}
          </span>
          <h3 className={styles.cardTitle}>{card.title}</h3>
          <p className={styles.cardBody}>{card.body}</p>
        </motion.li>
      );
    })}
  </motion.ol>
);

const PagerPage = (): React.JSX.Element => {
  const { t } = useTranslation();
  const brand = { brand: PAGER_BRAND };
  const steps = t('pager.how.steps', { returnObjects: true }) as unknown as Card[];
  const compare = t('pager.compare.items', { returnObjects: true }) as unknown as Card[];

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title={t('pager.meta.title', brand)}
          description={t('pager.meta.description', brand)}
          keywords={t('pager.meta.keywords', brand)}
          canonical="https://rtylr.com/pager/"
          structuredData={productWithBreadcrumbs({
            name: PAGER_BRAND,
            path: '/pager/',
            subCategory: 'Restaurant pager',
            description: t('pager.meta.schemaDescription', brand),
          })}
        />

        <div className={styles.page}>
          <header className={styles.hero}>
            <motion.div initial="hidden" animate="visible" variants={staggerBlur}>
              <motion.span className={styles.eyebrow} variants={blurFadeIn}>
                {t('pager.hero.eyebrow', brand)}
              </motion.span>
              <SplitWords
                as="h1"
                className={styles.heroTitle}
                text={t('pager.hero.title')}
                restLinesClassName={styles.muted}
                delay={0.1}
              />
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                {t('pager.hero.subtitle', brand)}
              </motion.p>
              <motion.div className={styles.heroCtas} variants={blurFadeIn}>
                <TryButton source="hero" />
                <span className={styles.ctaNote}>{t('pager.hero.note')}</span>
              </motion.div>
            </motion.div>
          </header>

          <section className={styles.section} aria-labelledby="pager-how">
            <h2 id="pager-how" className={styles.sectionTitle}>
              {t('pager.how.title')}
            </h2>
            <p className={styles.sectionIntro}>{t('pager.how.intro')}</p>
            <CardGrid cards={steps} icons={STEP_ICONS} numbered />
          </section>

          <section className={styles.section} aria-labelledby="pager-compare">
            <h2 id="pager-compare" className={styles.sectionTitle}>
              {t('pager.compare.title')}
            </h2>
            <p className={styles.sectionIntro}>{t('pager.compare.intro')}</p>
            <CardGrid cards={compare} icons={COMPARE_ICONS} />
          </section>

          <section className={styles.section} aria-labelledby="pager-free">
            <div className={styles.free}>
              <h2 id="pager-free" className={styles.sectionTitle}>
                {t('pager.free.title')}
              </h2>
              <p className={styles.freeBody}>{t('pager.free.body', brand)}</p>
            </div>
          </section>

          <section className={styles.closing} aria-labelledby="pager-cta">
            <h2 id="pager-cta" className={styles.closingTitle}>
              {t('pager.closing.title')}
            </h2>
            <p className={styles.closingSubtitle}>{t('pager.closing.subtitle')}</p>
            <TryButton source="closing" />
          </section>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default PagerPage;
