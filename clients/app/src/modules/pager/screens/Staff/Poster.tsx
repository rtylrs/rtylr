import React from 'react';
import { createPortal } from 'react-dom';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { Alert } from '@voxire/ui';
import { ArrowLeft, Printer } from 'lucide-react';
import { Button } from '../../../../shared/components/ui';
import { PAGER_BOARD_PATH } from '../../pagerPaths';
import { usePagerVenue } from '../../shared/lib/usePagerVenue';
import { CounterPoster } from './CounterPoster';

const Poster = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const { name, logoUrl, joinUrl, loading } = usePagerVenue();

  return (
    <>
      <HelmetContainer title={t('pager.poster.pageTitle')} />
      <div className="w-full min-w-0 space-y-5">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="min-w-0">
            <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
              {t('pager.poster.heading')}
            </h1>
            <p className="mt-1 text-xs" style={{ color: 'var(--dash-text-dim)' }}>
              {t('pager.poster.intro')}
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            {/* The Portal's Button renders a <button> only, so the link wears its classes. */}
            <Link
              to={PAGER_BOARD_PATH}
              className="portal-button portal-button--outline portal-button--md"
            >
              <ArrowLeft size={16} aria-hidden="true" className="rtl:rotate-180" />
              {t('pager.poster.back')}
            </Link>
            <Button disabled={!joinUrl} onClick={() => window.print()}>
              <Printer size={16} aria-hidden="true" />
              {t('pager.poster.print')}
            </Button>
          </div>
        </div>

        {joinUrl ? (
          <>
            <CounterPoster venueName={name} logoUrl={logoUrl} joinUrl={joinUrl} qrSize={220} />
            {/* The printed copy sits outside the Portal shell, so the print
                rules in pager.css can hide everything else on the page. */}
            {createPortal(
              <div className="pager-poster-print">
                <CounterPoster venueName={name} logoUrl={logoUrl} joinUrl={joinUrl} qrSize={360} />
              </div>,
              document.body
            )}
          </>
        ) : loading ? (
          <p className="text-sm" role="status" style={{ color: 'var(--dash-text-dim)' }}>
            {t('pager.poster.loading')}
          </p>
        ) : (
          <Alert variant="warning">{t('pager.poster.noLink')}</Alert>
        )}
      </div>
    </>
  );
};

export default Poster;
