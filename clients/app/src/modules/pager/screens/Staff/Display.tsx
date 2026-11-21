import React, { useEffect } from 'react';
import { createPortal } from 'react-dom';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { Alert } from '@voxire/ui';
import { X } from 'lucide-react';
import { readyAlarm } from '../../../../shared/alarm/readyAlarm';
import { PAGER_BOARD_PATH } from '../../pagerPaths';
import { usePagerVenue } from '../../shared/lib/usePagerVenue';
import { RenniliLogo } from '../../shared/components/RenniliLogo';
import { VenueLogo } from '../../shared/components/VenueLogo';
import { VenueQr } from './VenueQr';

const DISPLAY_LANGUAGES = ['en', 'ar'] as const;

function ScanLine({ lng }: { lng: (typeof DISPLAY_LANGUAGES)[number] }): React.JSX.Element {
  const { t } = useTranslation('pager', { lng });
  return (
    <p lang={lng} dir={lng === 'ar' ? 'rtl' : 'ltr'} className="pager-display__line">
      {t('pager.display.scan')}
    </p>
  );
}

/**
 * The counter display: a phone or tablet on a stand facing the customers,
 * showing the venue's QR all day. It covers the Portal shell and keeps the
 * screen awake while it is open.
 */
const Display = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const { name, logoUrl, joinUrl, loading } = usePagerVenue();

  // The alarm's wake lock, with nothing to ring.
  useEffect(() => {
    readyAlarm.sync({ keepAwake: true, ringKey: null });
    return () => readyAlarm.release();
  }, []);

  return (
    <>
      <HelmetContainer title={t('pager.display.pageTitle')} />
      {createPortal(
        <div className="pager-display">
          <Link
            to={PAGER_BOARD_PATH}
            className="pager-display__close"
            aria-label={t('pager.display.close')}
          >
            <X size={22} aria-hidden="true" />
          </Link>
          <VenueLogo src={logoUrl} className="pager-display__logo" />
          {name && <p className="pager-display__venue">{name}</p>}
          {joinUrl ? (
            <div className="pager-display__qr">
              <VenueQr url={joinUrl} size={640} />
            </div>
          ) : loading ? (
            <p role="status">{t('pager.poster.loading')}</p>
          ) : (
            <Alert variant="warning">{t('pager.poster.noLink')}</Alert>
          )}
          {DISPLAY_LANGUAGES.map((lng) => (
            <ScanLine key={lng} lng={lng} />
          ))}
          <div className="pager-display__brand">
            <RenniliLogo size={28} />
          </div>
        </div>,
        document.body
      )}
    </>
  );
};

export default Display;
