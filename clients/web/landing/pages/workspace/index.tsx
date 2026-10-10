import React, { useEffect, useRef, useState } from 'react';
import { motion, useInView } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { useRouter } from 'next/router';
import { APP_URLS } from '@voxire/config';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import FaqSection from '../../components/FaqSection/FaqSection';
import { SplitWords } from '../../components/motion/SplitWords';
import ConnectedSection from '../../components/home/ConnectedSection/ConnectedSection';
import CTASection from '../../components/home/CTASection/CTASection';
import { FlowDemo } from '../../components/home/FeaturesSection/demos/FlowDemo';
import { FinanceDemo } from '../../components/home/FeaturesSection/demos/FinanceDemo';
import { CRMDemo } from '../../components/home/FeaturesSection/demos/CRMDemo';
import { HRDemo } from '../../components/home/FeaturesSection/demos/HRDemo';
import { ERPDemo } from '../../components/home/FeaturesSection/demos/ERPDemo';
import { POSDemo } from '../../components/home/FeaturesSection/demos/POSDemo';
import { InsightsDemo } from '../../components/home/FeaturesSection/demos/InsightsDemo';
import { cn } from '../../lib/cn';
import { useFaq } from '../../lib/faqs';
import { productWithBreadcrumbs } from '../../lib/structured-data';
import { navigateExternal } from '../../lib/navigation';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import { track, EV_CTA_CLICK, EV_SIGNUP_CLICK } from '../../lib/analytics';
import styles from './workspace.module.scss';

type AreaId = 'work' | 'finance' | 'customers' | 'team' | 'operations' | 'sales' | 'insights';

// One section per area, in the same order as the homepage. The id is the
// URL fragment the old module pages redirect to (see cloudflare/_worker.js),
// so renaming one breaks inbound links.
const AREAS: { id: AreaId; color: string; Demo: React.ComponentType<{ isActive: boolean }> }[] = [
  { id: 'work', color: '#ef4444', Demo: FlowDemo },
  { id: 'finance', color: '#f59e0b', Demo: FinanceDemo },
  { id: 'customers', color: '#8b5cf6', Demo: CRMDemo },
  { id: 'team', color: '#10b981', Demo: HRDemo },
  { id: 'operations', color: '#64c5b9', Demo: ERPDemo },
  { id: 'sales', color: '#3b82f6', Demo: POSDemo },
  { id: 'insights', color: '#6fa3db', Demo: InsightsDemo },
];

interface AreaCopy {
  label: string;
  title: string;
  body: string;
  features: { title: string; body: string }[];
}

const AreaSection: React.FC<{
  id: AreaId;
  color: string;
  Demo: React.ComponentType<{ isActive: boolean }>;
  flip: boolean;
}> = ({ id, color, Demo, flip }) => {
  const { t } = useTranslation();
  const copy = t(`workspace.areas.${id}`, { returnObjects: true }) as unknown as AreaCopy;
  const stageRef = useRef<HTMLDivElement>(null);
  const inView = useInView(stageRef, { margin: '-15% 0px' });

  return (
    <section
      id={id}
      className={cn(styles.area, flip && styles.areaFlip)}
      style={{ '--area': color } as React.CSSProperties}
    >
      <motion.div
        className={styles.areaCopy}
        initial="hidden"
        whileInView="visible"
        viewport={viewportOnce}
        variants={staggerBlur}
      >
        <motion.span className={styles.areaLabel} variants={blurFadeIn}>
          <span className={styles.areaDot} />
          {copy.label}
        </motion.span>
        <SplitWords
          as="h2"
          className={styles.areaTitle}
          text={t(`workspace.areas.${id}.title`)}
          restLinesClassName={styles.muted}
          trigger="view"
        />
        <motion.p className={styles.areaBody} variants={blurFadeIn}>
          {copy.body}
        </motion.p>
        <motion.ul className={styles.features} variants={blurFadeIn}>
          {copy.features.map((f) => (
            <li key={f.title} className={styles.feature}>
              <strong>{f.title}</strong> {f.body}
            </li>
          ))}
        </motion.ul>
      </motion.div>

      <div className={styles.stage} ref={stageRef} aria-hidden="true">
        <div className={styles.demo}>
          <Demo isActive={inView} />
        </div>
      </div>
    </section>
  );
};

/** Sticky pills that jump between areas and track which one is on screen. */
const AreaNav: React.FC = () => {
  const { t } = useTranslation();
  const [active, setActive] = useState<AreaId>('work');

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries.filter((e) => e.isIntersecting);
        if (visible.length) setActive(visible[0].target.id as AreaId);
      },
      // A thin band a third of the way down the viewport: whichever
      // section crosses it is the one being read.
      { rootMargin: '-33% 0px -66% 0px' },
    );
    AREAS.forEach(({ id }) => {
      const el = document.getElementById(id);
      if (el) observer.observe(el);
    });
    return () => observer.disconnect();
  }, []);

  return (
    <nav className={styles.nav} aria-label={t('workspace.hero.navLabel')}>
      <div className={styles.navInner}>
        {AREAS.map(({ id, color }) => (
          <a
            key={id}
            href={`#${id}`}
            className={cn(styles.navPill, active === id && styles.navPillActive)}
            style={{ '--area': color } as React.CSSProperties}
            aria-current={active === id ? 'true' : undefined}
          >
            <span className={styles.areaDot} />
            {t(`workspace.areas.${id}.label`)}
          </a>
        ))}
      </div>
    </nav>
  );
};

const WorkspacePage = (): React.JSX.Element => {
  const { t } = useTranslation();
  const router = useRouter();
  const faq = useFaq('workspace');

  const goRegister = (): void => {
    track(EV_SIGNUP_CLICK, { source: 'workspace-hero' });
    navigateExternal(`${APP_URLS.auth}/register`);
  };

  const goDemo = (): void => {
    track(EV_CTA_CLICK, { cta: 'demo', page: router.pathname });
    void router.push('/demo');
  };

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title={t('workspace.meta.title')}
          description={t('workspace.meta.description')}
          canonical="https://rtylr.com/workspace/"
          structuredData={productWithBreadcrumbs({
            name: 'rtylr workspace',
            path: '/workspace/',
            subCategory: 'All-in-one Business Software',
            description: t('workspace.meta.schemaDescription'),
          })}
        />

        <div className={styles.page}>
          <header className={styles.hero}>
            <motion.div initial="hidden" animate="visible" variants={staggerBlur}>
              <motion.span className={styles.eyebrow} variants={blurFadeIn}>
                {t('workspace.hero.eyebrow')}
              </motion.span>
              <SplitWords
                as="h1"
                className={styles.heroTitle}
                text={t('workspace.hero.title')}
                restLinesClassName={styles.muted}
                delay={0.1}
              />
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                {t('workspace.hero.subtitle')}
              </motion.p>
              <motion.div className={styles.heroCtas} variants={blurFadeIn}>
                <button type="button" onClick={goRegister} className={styles.ctaPrimary}>
                  {t('common.cta.getStarted')}
                </button>
                <button type="button" onClick={goDemo} className={styles.ctaSecondary}>
                  {t('common.cta.bookDemo')}
                </button>
              </motion.div>
            </motion.div>
          </header>

          <AreaNav />

          <div className={styles.areas}>
            {AREAS.map((area, i) => (
              <AreaSection key={area.id} {...area} flip={i % 2 === 1} />
            ))}
          </div>

          <ConnectedSection />
          <FaqSection title={faq.title} intro={faq.intro} faqs={faq.items} />
          <CTASection />
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default WorkspacePage;
