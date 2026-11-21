import React, { useEffect, useId, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { useNavigate, useParams } from '@tanstack/react-router';
import { FullScreenLoader, HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { BellOff } from 'lucide-react';
import { readyAlarm } from '../../../../shared/alarm/readyAlarm';
import { pagerTicketPath } from '../../pagerPaths';
import LanguageSwitch from '../../shared/components/LanguageSwitch';
import { PAGER_ERROR, errorCode, errorStatus } from '../../shared/lib/apiError';
import { joinWithRetry } from '../../shared/lib/joinWithRetry';
import { usePublicPagerClient } from '../../shared/lib/pagerClients';
import { askPushPermission, usePushKey } from '../../shared/lib/pushAlerts';
import { normaliseTicketNumber } from '../../shared/lib/ticketNumber';

type Phase = 'idle' | 'joining' | 'waitingForCashier' | 'notYet';
type JoinError = 'invalidNumber' | 'venueBusy' | 'tooManyAttempts' | 'failed';

const isNotYet = (error: unknown): boolean => errorCode(error) === PAGER_ERROR.ticketNotFound;

function joinErrorFor(error: unknown): JoinError {
  if (errorCode(error) === PAGER_ERROR.invalidNumber) return 'invalidNumber';
  // Also a 429, but about the venue's queue rather than this phone's attempts.
  if (errorCode(error) === PAGER_ERROR.tooManyOpenTickets) return 'venueBusy';
  if (errorStatus(error) === 429) return 'tooManyAttempts';
  return 'failed';
}

interface JoinProps {
  /** The venue, when not taken from the route. */
  slug?: string;
  /** Replaces opening the ticket page, as in the demo's phone. */
  onJoined?: (publicToken: string) => void;
}

const Join = ({ slug: givenSlug, onJoined }: JoinProps = {}): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const { slug: rawSlug } = useParams({ strict: false });
  const slug = (givenSlug ?? rawSlug ?? '').toLowerCase();
  const navigate = useNavigate();
  const publicPagerClient = usePublicPagerClient();
  const pushKey = usePushKey();
  const inputId = useId();
  const errorId = useId();

  const venue = useQuery({
    queryKey: ['pager', 'venue', slug],
    queryFn: () => publicPagerClient.getPublicVenue({ slug }),
    enabled: slug !== '',
    retry: (failures, error) => errorStatus(error) !== 404 && failures < 2,
  });

  const [number, setNumber] = useState('');
  const [phase, setPhase] = useState<Phase>('idle');
  const [joinError, setJoinError] = useState<JoinError | null>(null);
  const [venueGone, setVenueGone] = useState(false);
  const attempt = useRef<AbortController | null>(null);

  useEffect(() => () => attempt.current?.abort(), []);

  const join = async (): Promise<void> => {
    if (!number || phase === 'joining' || phase === 'waitingForCashier') return;
    const controller = new AbortController();
    attempt.current = controller;
    setJoinError(null);
    setPhase('joining');
    try {
      const ticket = await joinWithRetry({
        join: () => publicPagerClient.joinTicket({ slug, number }, { signal: controller.signal }),
        isNotYet,
        onWaiting: () => setPhase('waitingForCashier'),
        signal: controller.signal,
      });
      if (controller.signal.aborted) return;
      if (!ticket.publicToken) throw new Error('The pager answered without a ticket token');
      if (onJoined) {
        onJoined(ticket.publicToken);
        return;
      }
      // A router navigation, not a page load: the unlocked audio lives in this document.
      void navigate({ to: pagerTicketPath(ticket.publicToken) });
    } catch (error) {
      if (controller.signal.aborted) return;
      if (isNotYet(error)) {
        setPhase('notYet');
      } else if (errorCode(error) === PAGER_ERROR.venueNotFound) {
        setVenueGone(true);
      } else {
        setPhase('idle');
        setJoinError(joinErrorFor(error));
      }
    }
  };

  const changeNumber = (): void => {
    attempt.current?.abort();
    setPhase('idle');
  };

  if (slug !== '' && venue.isPending) return <FullScreenLoader />;

  const unavailable =
    slug === '' ||
    venueGone ||
    errorStatus(venue.error) === 404 ||
    (venue.data !== undefined && !venue.data.enabled);
  if (unavailable) {
    return (
      <div className="pager-center">
        <HelmetContainer title={t('pager.join.pageTitle')} />
        <EmptyState
          icon={BellOff}
          title={t('pager.join.venueNotFoundTitle')}
          description={t('pager.join.venueNotFoundDescription')}
        />
      </div>
    );
  }

  if (venue.isError) {
    return (
      <div className="pager-center">
        <HelmetContainer title={t('pager.join.pageTitle')} />
        <p className="pager-error" role="alert">
          {t('pager.join.loadFailed')}
        </p>
        <button type="button" className="pager-button" onClick={() => void venue.refetch()}>
          {t('pager.join.retry')}
        </button>
      </div>
    );
  }

  const joining = phase === 'joining';

  return (
    <div className="pager-screen">
      <HelmetContainer title={t('pager.join.pageTitle')} />
      <main className="pager-page">
        <div className="pager-topbar">
          <p className="pager-venue">
            {venue.data?.logoUrl && <img src={venue.data.logoUrl} alt="" />}
            <span>{venue.data?.name}</span>
          </p>
          <LanguageSwitch />
        </div>

        <h1 className="pager-heading">{t('pager.join.heading')}</h1>

        {phase === 'waitingForCashier' && (
          <div className="pager-panel" role="status">
            <p className="pager-panel-title">{t('pager.join.waitingForCashierTitle')}</p>
            <p className="pager-text">{t('pager.join.waitingForCashierBody', { number })}</p>
            <button
              type="button"
              className="pager-button pager-button--secondary"
              onClick={changeNumber}
            >
              {t('pager.join.changeNumber')}
            </button>
          </div>
        )}

        {phase === 'notYet' && (
          <div className="pager-panel" role="alert">
            <p className="pager-panel-title">{t('pager.join.notYetTitle', { number })}</p>
            <p className="pager-text">{t('pager.join.notYetBody')}</p>
            <button
              type="button"
              className="pager-button"
              onClick={() => {
                readyAlarm.unlock();
                askPushPermission(pushKey);
                void join();
              }}
            >
              {t('pager.join.retry')}
            </button>
            <button
              type="button"
              className="pager-button pager-button--secondary"
              onClick={changeNumber}
            >
              {t('pager.join.changeNumber')}
            </button>
          </div>
        )}

        {(phase === 'idle' || joining) && (
          <form
            className="pager-form"
            noValidate
            onSubmit={(event) => {
              event.preventDefault();
              void join();
            }}
          >
            <p className="pager-text">{t('pager.join.intro')}</p>
            <label className="pager-label" htmlFor={inputId}>
              {t('pager.join.numberLabel')}
            </label>
            <input
              id={inputId}
              className="pager-number-input"
              dir="ltr"
              inputMode="numeric"
              enterKeyHint="go"
              autoComplete="off"
              autoCapitalize="characters"
              autoCorrect="off"
              spellCheck={false}
              value={number}
              readOnly={joining}
              aria-invalid={joinError === 'invalidNumber'}
              aria-describedby={joinError ? errorId : undefined}
              onChange={(event) => {
                setNumber(normaliseTicketNumber(event.target.value));
                setJoinError(null);
              }}
            />
            {joinError && (
              <p className="pager-error" id={errorId} role="alert">
                {t(`pager.join.${joinError}`)}
              </p>
            )}
            {/* The browser only lets a page ring, or ask to notify, after a tap,
                and this is the one tap there is. Both run in the click itself,
                ahead of the submit and of every await. Pressing "go" on the
                keyboard clicks this button too. */}
            <button
              type="submit"
              className="pager-button"
              disabled={!number || joining}
              onClick={() => {
                readyAlarm.unlock();
                askPushPermission(pushKey);
              }}
            >
              {joining ? t('pager.join.submitting') : t('pager.join.submit')}
            </button>
            <p className="pager-text">{t('pager.join.soundHint')}</p>
          </form>
        )}
      </main>
    </div>
  );
};

export default Join;
