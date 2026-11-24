import React, { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useParams } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { helpers } from '@voxire/utils';
import { BellOff, BellRing } from 'lucide-react';
import { readyAlarm } from '../../../../shared/alarm/readyAlarm';
import { pagerTicketPath } from '../../pagerPaths';
import LanguageSwitch from '../../shared/components/LanguageSwitch';
import { ByRtylr, RenniliLogo, RenniliMark } from '../../shared/components/RenniliLogo';
import { VenueLogo } from '../../shared/components/VenueLogo';
import {
  DEPLOY_GAP_ATTEMPTS,
  DEPLOY_GAP_FIRST_DELAY_MS,
  PAGER_ERROR,
  errorCode,
  errorStatus,
  isTransient,
} from '../../shared/lib/apiError';
import { usePagerSandbox, usePublicPagerClient } from '../../shared/lib/pagerClients';
import { askPushPermission, askPushPermissionNow, usePushKey } from '../../shared/lib/pushAlerts';
import { openScanTicket, rememberedScan } from '../../shared/lib/scanSession';

type ScanState =
  'loading' | 'ready' | 'scanning' | 'unavailable' | 'venueBusy' | 'tooManyAttempts' | 'failed';

function stateFor(error: unknown): ScanState {
  if (errorCode(error) === PAGER_ERROR.venueNotFound) return 'unavailable';
  // Also a 429, but about the venue's queue rather than this phone's attempts.
  if (errorCode(error) === PAGER_ERROR.tooManyOpenTickets) return 'venueBusy';
  if (errorStatus(error) === 429) return 'tooManyAttempts';
  return 'failed';
}

interface VenuePreview {
  name: string;
  logoUrl?: string;
}

interface ScanProps {
  /** The venue, when not taken from the route. */
  slug?: string;
  /** Replaces opening the ticket page, as in the demo's phone. */
  onScanned?: (publicToken: string) => void;
}

/**
 * Where the venue's QR lands. Shows a single upfront tap so the browser grants
 * audio and notification permissions before the customer enters the queue, then
 * opens (or resumes) the phone's ticket and navigates to the ticket screen.
 */
const Scan = ({ slug: givenSlug, onScanned }: ScanProps = {}): React.JSX.Element => {
  const { t, i18n } = useTranslation('pager');
  const { slug: rawSlug } = useParams({ strict: false });
  const slug = (givenSlug ?? rawSlug ?? '').toLowerCase();
  const navigate = useNavigate();
  const client = usePublicPagerClient();
  // The demo starts a new phone each time, and must not leave a token behind.
  const sandbox = usePagerSandbox();
  const pushKey = usePushKey();
  const [venue, setVenue] = useState<VenuePreview | null>(null);
  const [state, setState] = useState<ScanState>(slug ? 'loading' : 'unavailable');
  const [slow, setSlow] = useState(false);
  const mounted = useRef(true);
  const scanning = useRef(false);
  const queryWantsNew =
    typeof window !== 'undefined' && new URLSearchParams(window.location.search).get('new') === '1';
  const hasRemembered = !sandbox && !queryWantsNew && Boolean(slug && rememberedScan(slug));

  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
    };
  }, []);

  useEffect(() => {
    if (!slug) {
      setState('unavailable');
      return;
    }
    let active = true;
    setState('loading');
    setSlow(false);
    let tries = 0;
    helpers
      .retry(
        () => {
          tries += 1;
          if (tries > 1 && active) setSlow(true);
          return client.getPublicVenue({ slug });
        },
        DEPLOY_GAP_ATTEMPTS,
        DEPLOY_GAP_FIRST_DELAY_MS,
        (error) => active && isTransient(error)
      )
      .then(
        (info) => {
          if (!active) return;
          setVenue({ name: info.name ?? '', logoUrl: info.logoUrl });
          setState('ready');
        },
        (error: unknown) => {
          if (!active) return;
          setState(stateFor(error));
        }
      );
    return () => {
      active = false;
    };
  }, [slug, client, i18n.language]);

  const startScan = (forceNew = false): void => {
    if (!slug || scanning.current) return;
    scanning.current = true;
    // Unlock sound, play a short confirmation chime, and ask notification
    // permission synchronously inside the tap handler before any await.
    readyAlarm.unlock();
    readyAlarm.blip();
    if (!sandbox) {
      if (pushKey) askPushPermission(pushKey);
      else void askPushPermissionNow();
    }

    setState('scanning');
    setSlow(false);
    let tries = 0;
    helpers
      .retry(
        () => {
          tries += 1;
          if (tries > 1 && mounted.current) setSlow(true);
          return openScanTicket(client, slug, {
            remember: !sandbox,
            forceNew: forceNew || queryWantsNew,
          });
        },
        DEPLOY_GAP_ATTEMPTS,
        DEPLOY_GAP_FIRST_DELAY_MS,
        (error) => mounted.current && isTransient(error)
      )
      .then(
        (publicToken) => {
          scanning.current = false;
          if (!mounted.current) return;
          if (onScanned) onScanned(publicToken);
          // Router navigation preserves unlocked audio in this document.
          else void navigate({ to: pagerTicketPath(publicToken), replace: true });
        },
        (error: unknown) => {
          scanning.current = false;
          if (mounted.current) setState(stateFor(error));
        }
      );
  };

  if (state === 'unavailable') {
    return (
      <div className="pager-center">
        <HelmetContainer title={t('pager.scan.pageTitle')} />
        <EmptyState
          icon={BellOff}
          title={t('pager.scan.venueNotFoundTitle')}
          description={t('pager.scan.venueNotFoundDescription')}
        />
      </div>
    );
  }

  if (state === 'loading' || state === 'scanning') {
    return (
      <div className="pager-center">
        <HelmetContainer title={t('pager.scan.pageTitle')} />
        <RenniliMark size={44} className="pager-scan-mark" />
        <p className="pager-text" role="status">
          {t(slow ? 'pager.scan.slow' : 'pager.scan.scanning')}
        </p>
      </div>
    );
  }

  if (state === 'ready') {
    return (
      <div className="pager-screen">
        <HelmetContainer title={t('pager.scan.pageTitle')} />
        <main className="pager-page">
          <div className="pager-topbar">
            <p className="pager-venue">
              <VenueLogo src={venue?.logoUrl} />
              <span>{venue?.name}</span>
            </p>
            {!sandbox && <LanguageSwitch />}
          </div>

          <section className="pager-gate">
            <div className="pager-gate-icon" aria-hidden="true">
              <BellRing size={28} strokeWidth={2.2} />
            </div>
            <h1 className="pager-gate-title">{t('pager.scan.readyTitle')}</h1>
            <p className="pager-gate-hint">{t('pager.scan.readyHint')}</p>
            <p className="pager-gate-pill">{t('pager.scan.permissionsNote')}</p>
            <button
              type="button"
              className="pager-button pager-gate-cta"
              onClick={() => startScan(false)}
            >
              <BellRing size={19} strokeWidth={2.4} aria-hidden="true" />
              <span>{t('pager.scan.startButton')}</span>
            </button>
            {hasRemembered && (
              <button type="button" className="pager-text-action" onClick={() => startScan(true)}>
                {t('pager.scan.newOrderButton')}
              </button>
            )}
          </section>

          <footer className="pager-footer">
            <RenniliLogo size={20} />
            <ByRtylr />
          </footer>
        </main>
      </div>
    );
  }

  return (
    <div className="pager-center">
      <HelmetContainer title={t('pager.scan.pageTitle')} />
      <p className="pager-error" role="alert">
        {t(`pager.scan.${state}`)}
      </p>
      <button type="button" className="pager-button" onClick={() => startScan(false)}>
        {t('pager.scan.retry')}
      </button>
    </div>
  );
};

export default Scan;
