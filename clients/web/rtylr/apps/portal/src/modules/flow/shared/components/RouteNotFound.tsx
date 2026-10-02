import React from 'react';
import { Compass } from 'lucide-react';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';

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
      <h1 className="text-flow-title font-semibold" style={{ color: 'var(--flow-text)' }}>
        {t('flow.notFound.title')}
      </h1>
      {/* 44ch, per S-24: a sentence this long is unreadable across a wide
          content column. */}
      <p className="mt-1.5 max-w-[44ch] text-flow-body" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flow.notFound.description')}
      </p>
      <Link
        to="/flow/my-issues"
        className="mt-5 inline-flex min-h-11 items-center rounded-md px-4 text-flow-body font-semibold"
        style={{ background: 'var(--flow-action-bg)', color: 'var(--flow-on-accent)' }}
      >
        {t('flow.notFound.cta')}
      </Link>
    </div>
  );
};

export default RouteNotFound;
