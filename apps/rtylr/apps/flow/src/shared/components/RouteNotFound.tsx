import React from 'react';
import { Compass } from 'lucide-react';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';

/**
 * The 404 screen.
 *
 * `createRouter` carried no `defaultNotFoundComponent`, so an unknown URL got
 * TanStack's raw fallback: the words "Not Found" at the top left of a white
 * page, outside the app frame, untranslated and with nothing from the token
 * set. This is wired in as that component and rendered inside the layout, so
 * a mistyped or stale link leaves the sidebar and navigation in place.
 *
 * The shell follows S-24: a 44px glyph in a neutral circle, a 16px/600
 * heading, one sentence capped at 44ch, exactly one action.
 */
const RouteNotFound: React.FC = () => {
  const { t } = useTranslation();
  return (
    <div className="flex h-full flex-col items-center justify-center px-6 py-16 text-center">
      <div
        className="mb-4 flex h-11 w-11 items-center justify-center rounded-full"
        style={{ background: 'var(--flow-hover)' }}
      >
        <Compass size={20} style={{ color: 'var(--flow-text-muted)' }} aria-hidden />
      </div>
      <h1 className="text-title font-semibold" style={{ color: 'var(--flow-text)' }}>
        {t('flow.notFound.title')}
      </h1>
      {/* 44ch, per S-24: a sentence this long is unreadable across a wide
          content column. */}
      <p className="mt-1.5 max-w-[44ch] text-body" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flow.notFound.description')}
      </p>
      <Link
        to="/my-issues"
        className="mt-5 inline-flex min-h-11 items-center rounded-md px-4 text-body font-semibold"
        style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
      >
        {t('flow.notFound.cta')}
      </Link>
    </div>
  );
};

export default RouteNotFound;
