import React, { useEffect, useRef } from 'react';
import { MotionConfig } from 'motion/react';
import { HelmetContainer } from '@voxire/components';
import { RENNILI_URL } from '@voxire/config';
import { LandingFooter } from './LandingFooter';
import { LandingNav } from './LandingNav';
import { findScrollRoot, useScrolledPast } from './useLandingObservers';

interface PublicPageShellProps {
  path: string;
  title: string;
  description: string;
  keywords?: string;
  children: React.ReactNode;
}

/** Shared shell for dedicated SEO pages on rennili.rtylr.com. */
export const PublicPageShell = ({
  path,
  title,
  description,
  keywords,
  children,
}: PublicPageShellProps): React.JSX.Element => {
  const sentinelRef = useRef<HTMLDivElement>(null);
  const scrolled = useScrolledPast(sentinelRef);

  useEffect(() => {
    const node = sentinelRef.current;
    if (!node) return;
    const root = findScrollRoot(node);
    if (root) {
      root.scrollTo({ top: 0, behavior: 'auto' });
    } else {
      window.scrollTo({ top: 0, behavior: 'auto' });
    }
  }, [path]);

  return (
    <MotionConfig reducedMotion="user">
      <div className="rennili-landing">
        <HelmetContainer
          title={title}
          description={description}
          keywords={keywords}
          image={`${RENNILI_URL}/og-image.png`}
          url={`${RENNILI_URL}${path}`}
        />

        <LandingNav currentPath={path} scrolled={scrolled} showMobileBar={scrolled} />

        <div ref={sentinelRef} className="rl-page-sentinel" aria-hidden="true" />

        <main>{children}</main>

        <LandingFooter />
      </div>
    </MotionConfig>
  );
};
