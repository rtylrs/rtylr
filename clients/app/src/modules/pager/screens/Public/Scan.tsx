import React, { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useParams } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { BellOff } from 'lucide-react';
import { pagerTicketPath } from '../../pagerPaths';
import { PAGER_ERROR, errorCode, errorStatus } from '../../shared/lib/apiError';
import { usePagerSandbox, usePublicPagerClient } from '../../shared/lib/pagerClients';
import { openScanTicket } from '../../shared/lib/scanSession';

type ScanState = 'scanning' | 'unavailable' | 'venueBusy' | 'tooManyAttempts' | 'failed';

function stateFor(error: unknown): ScanState {
  if (errorCode(error) === PAGER_ERROR.venueNotFound) return 'unavailable';
  // Also a 429, but about the venue's queue rather than this phone's attempts.
  if (errorCode(error) === PAGER_ERROR.tooManyOpenTickets) return 'venueBusy';
  if (errorStatus(error) === 429) return 'tooManyAttempts';
  return 'failed';
}

interface ScanProps {
  /** The venue, when not taken from the route. */
  slug?: string;
  /** Replaces opening the ticket page, as in the demo's phone. */
  onScanned?: (publicToken: string) => void;
}

/**
 * Where the venue's QR lands. The customer does nothing here: the page opens
 * this phone's ticket on load, so the cashier sees it even if the customer
 * never taps anything, and moves on to the ticket page with the code.
 */
const Scan = ({ slug: givenSlug, onScanned }: ScanProps = {}): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const { slug: rawSlug } = useParams({ strict: false });
  const slug = (givenSlug ?? rawSlug ?? '').toLowerCase();
  const navigate = useNavigate();
  const client = usePublicPagerClient();
  // The demo starts a new phone each time, and must not leave a token behind.
  const sandbox = usePagerSandbox();
  const [state, setState] = useState<ScanState>(slug ? 'scanning' : 'unavailable');
  const [attempt, setAttempt] = useState(0);
  // React runs effects twice in development; one scan per attempt.
  const started = useRef(-1);
  const mounted = useRef(true);

  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
    };
  }, []);

  useEffect(() => {
    if (!slug || started.current === attempt) return;
    started.current = attempt;
    setState('scanning');
    openScanTicket(client, slug, { remember: !sandbox }).then(
      (publicToken) => {
        if (!mounted.current) return;
        if (onScanned) onScanned(publicToken);
        // A router navigation, not a page load: the unlocked audio lives in
        // this document. Replace, so Back does not land on this page again.
        else void navigate({ to: pagerTicketPath(publicToken), replace: true });
      },
      (error: unknown) => {
        if (mounted.current) setState(stateFor(error));
      }
    );
  }, [slug, attempt, client, sandbox, onScanned, navigate]);

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

  if (state === 'scanning') {
    return (
      <div className="pager-center">
        <HelmetContainer title={t('pager.scan.pageTitle')} />
        <p className="pager-text" role="status">
          {t('pager.scan.scanning')}
        </p>
      </div>
    );
  }

  return (
    <div className="pager-center">
      <HelmetContainer title={t('pager.scan.pageTitle')} />
      <p className="pager-error" role="alert">
        {t(`pager.scan.${state}`)}
      </p>
      <button type="button" className="pager-button" onClick={() => setAttempt((n) => n + 1)}>
        {t('pager.scan.retry')}
      </button>
    </div>
  );
};

export default Scan;
