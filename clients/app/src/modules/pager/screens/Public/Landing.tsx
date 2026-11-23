import React, { useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { MotionConfig } from 'motion/react';
import { HelmetContainer } from '@voxire/components';
import { RENNILI_URL } from '@voxire/config';
import { PAGER_LANDING_PATH } from '../../pagerPaths';
import { FaqSection } from './landing/FaqSection';
import { LandingFooter } from './landing/LandingFooter';
import { LandingHero } from './landing/LandingHero';
import { LandingNav } from './landing/LandingNav';
import { CompareTable, FreeBand, HighlightsBand, StepsSection } from './landing/LandingStory';
import { useScrolledPast } from './landing/useLandingObservers';

/**
 * Rennili's landing page, the root of rennili.rtylr.com. Each part of the
 * story is told once; the longer versions live on the How it works, Features
 * and FAQ pages.
 */
const Landing = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const heroRef = useRef<HTMLElement>(null);
  const scrolledPastHero = useScrolledPast(heroRef);

  return (
    <MotionConfig reducedMotion="user">
      <div className="rennili-landing">
        <HelmetContainer
          title={t('pager.landing.meta.title')}
          description={t('pager.landing.meta.description')}
          keywords={t('pager.landing.meta.keywords')}
          image={`${RENNILI_URL}/og-image.png`}
          url={`${RENNILI_URL}/`}
        />

        <LandingNav
          currentPath={PAGER_LANDING_PATH}
          scrolled={scrolledPastHero}
          showMobileBar={scrolledPastHero}
        />

        <main>
          <LandingHero heroRef={heroRef} />
          <HighlightsBand />
          <StepsSection />
          <CompareTable />
          <FreeBand />
          <FaqSection />
        </main>

        <LandingFooter onLanding />
      </div>
    </MotionConfig>
  );
};

export default Landing;
