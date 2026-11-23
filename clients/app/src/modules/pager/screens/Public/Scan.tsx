import React, { useEffect, useId, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useParams } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { helpers } from '@voxire/utils';
import { BellOff } from 'lucide-react';
import { pagerTicketPath } from '../../pagerPaths';
import { normaliseTicketNumber } from '../../shared/lib/ticketNumber';
import {
  DEPLOY_GAP_ATTEMPTS,
  DEPLOY_GAP_FIRST_DELAY_MS,
  PAGER_ERROR,
  errorCode,
  errorStatus,
  isTransient,
} from '../../shared/lib/apiError';
import { usePagerSandbox, usePublicPagerClient } from '../../shared/lib/pagerClients';
import { openScanTicket, rememberScan } from '../../shared/lib/scanSession';

type ScanState =
  'scanning' | 'askNumber' | 'unavailable' | 'venueBusy' | 'tooManyAttempts' | 'failed';

function stateFor(error: unknown): ScanState {
  if (errorCode(error) === PAGER_ERROR.venueNotFound) return 'unavailable';
  // The venue has customers type their receipt number instead.
  if (errorCode(error) === PAGER_ERROR.typeYourNumber) return 'askNumber';
  // Also a 429, but about the venue's queue rather than this phone's attempts.
  if (errorCode(error) === PAGER_ERROR.tooManyOpenTickets) return 'venueBusy';
  if (errorStatus(error) === 429) return 'tooManyAttempts';
  return 'failed';
}

type JoinError = 'numberTaken' | 'numberInvalid' | 'venueBusy' | 'tooManyAttempts' | 'failed';

function joinErrorFor(error: unknown): JoinError {
  switch (errorCode(error)) {
    case PAGER_ERROR.labelTaken:
      return 'numberTaken';
    case PAGER_ERROR.invalidLabel:
      return 'numberInvalid';
    case PAGER_ERROR.tooManyOpenTickets:
      return 'venueBusy';
  }
  return errorStatus(error) === 429 ? 'tooManyAttempts' : 'failed';
}

interface NumberFormProps {
  slug: string;
  remember: boolean;
  onJoined: (publicToken: string) => void;
}

/** At a venue whose customers type the order number printed on their receipt. */
function NumberForm({ slug, remember, onJoined }: NumberFormProps): React.JSX.Element {
  const { t } = useTranslation('pager');
  const client = usePublicPagerClient();
  const inputId = useId();
  const [value, setValue] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<JoinError | null>(null);

  const join = async (): Promise<void> => {
    setBusy(true);
    setError(null);
    try {
      const { publicToken } = await client.joinTicket({ slug, number: value });
      if (!publicToken) throw new Error('The pager answered without a ticket token');
      if (remember) rememberScan(slug, publicToken);
      onJoined(publicToken);
    } catch (failure) {
      setError(joinErrorFor(failure));
      setBusy(false);
    }
  };

  return (
    <form
      className="pager-number-form"
      onSubmit={(event) => {
        event.preventDefault();
        if (value && !busy) void join();
      }}
    >
      <h1 className="pager-status">{t('pager.scan.numberTitle')}</h1>
      <p className="pager-text">{t('pager.scan.numberHint')}</p>
      <label htmlFor={inputId} className="sr-only">
        {t('pager.scan.numberLabel')}
      </label>
      <input
        id={inputId}
        className="pager-number-input"
        dir="ltr"
        inputMode="numeric"
        enterKeyHint="go"
        autoComplete="off"
        autoCapitalize="characters"
        spellCheck={false}
        maxLength={12}
        value={value}
        aria-invalid={error === 'numberInvalid' || error === 'numberTaken'}
        onChange={(event) => setValue(normaliseTicketNumber(event.target.value))}
      />
      {error && (
        <p className="pager-error" role="alert">
          {t(`pager.scan.${error}`)}
        </p>
      )}
      <button type="submit" className="pager-button" disabled={!value || busy}>
        {t('pager.scan.numberSubmit')}
      </button>
    </form>
  );
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
  const [slow, setSlow] = useState(false);
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
    setSlow(false);
    let tries = 0;
    // A deploy makes the pager unreachable for a few seconds; waiting 1, 2, 4
    // and 8 seconds rides it out before the customer sees an error.
    helpers
      .retry(
        () => {
          tries += 1;
          if (tries > 1 && mounted.current) setSlow(true);
          return openScanTicket(client, slug, { remember: !sandbox });
        },
        DEPLOY_GAP_ATTEMPTS,
        DEPLOY_GAP_FIRST_DELAY_MS,
        (error) => mounted.current && isTransient(error)
      )
      .then(
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

  if (state === 'askNumber') {
    return (
      <div className="pager-center">
        <HelmetContainer title={t('pager.scan.pageTitle')} />
        <NumberForm
          slug={slug}
          remember={!sandbox}
          onJoined={(publicToken) => {
            if (onScanned) onScanned(publicToken);
            else void navigate({ to: pagerTicketPath(publicToken), replace: true });
          }}
        />
      </div>
    );
  }

  if (state === 'scanning') {
    return (
      <div className="pager-center">
        <HelmetContainer title={t('pager.scan.pageTitle')} />
        <p className="pager-text" role="status">
          {t(slow ? 'pager.scan.slow' : 'pager.scan.scanning')}
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
