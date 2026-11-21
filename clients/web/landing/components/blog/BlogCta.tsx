import Link from 'next/link';
import React from 'react';
import { useTranslation } from 'react-i18next';
import { APP_URLS } from '@voxire/config';

import { EV_CTA_CLICK, EV_SIGNUP_CLICK, track } from '../../lib/analytics';
import { cn } from '../../lib/cn';
import type { SegmentId } from '../../lib/blog/types';
import styles from './Blog.module.scss';

type Action = 'startFree' | 'bookDemo' | 'talkToSales' | 'talkToFounders';

const HREF: Record<Action, string> = {
  startFree: `${APP_URLS.auth}/register`,
  bookDemo: '/demo/',
  talkToSales: '/contact/?topic=sales',
  talkToFounders: '/contact/?topic=investors',
};

/**
 * The first action matches who the post is written for. "Book a demo" is
 * always offered too, as the second action, except on mid-size posts where
 * it is already the first and Start free takes its place.
 */
const ACTIONS: Record<SegmentId, [Action, Action]> = {
  small: ['startFree', 'bookDemo'],
  mid: ['bookDemo', 'startFree'],
  large: ['talkToSales', 'bookDemo'],
  vision: ['talkToFounders', 'bookDemo'],
};

export function BlogCta({
  segment,
  placement,
  slug,
}: {
  segment: SegmentId;
  placement: 'inline' | 'end';
  slug: string;
}): React.JSX.Element {
  const { t } = useTranslation();
  const [primary, secondary] = ACTIONS[segment];

  const onClick = (action: Action) => (): void => {
    const params = { location: `blog_${placement}`, cta: action, segment, post: slug };
    track(action === 'startFree' ? EV_SIGNUP_CLICK : EV_CTA_CLICK, params);
  };

  const button = (action: Action, variant: 'primary' | 'secondary'): React.JSX.Element => {
    const dark = placement === 'end';
    const tone =
      variant === 'primary'
        ? dark
          ? styles.onDark
          : styles.onLight
        : dark
          ? styles.ghostOnDark
          : styles.ghostOnLight;
    const className = cn(styles.action, tone);
    const label = t(`blog.cta.${action}`);
    return action === 'startFree' ? (
      <a href={HREF[action]} className={className} onClick={onClick(action)}>
        {label}
      </a>
    ) : (
      <Link href={HREF[action]} className={className} onClick={onClick(action)}>
        {label}
      </Link>
    );
  };

  return (
    <aside className={cn(styles.cta, placement === 'inline' && styles.ctaInline)}>
      <div className={styles.ctaText}>
        <p className={styles.ctaTitle}>{t(`blog.cta.${segment}.title`)}</p>
        <p className={styles.ctaBody}>{t(`blog.cta.${segment}.body`)}</p>
      </div>
      <div className={styles.actions}>
        {button(primary, 'primary')}
        {button(secondary, 'secondary')}
      </div>
    </aside>
  );
}

export default BlogCta;
