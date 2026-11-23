import React, { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useParams } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import type { GetPublicTicketResponse } from '@voxire/voxie/pager/v1/types';
import { Lock, SearchX } from 'lucide-react';
import { readyAlarm } from '../../../../shared/alarm/readyAlarm';
import { useAlarm } from '../../../../shared/alarm/useAlarm';
import LanguageSwitch from '../../shared/components/LanguageSwitch';
import { ByRtylr, RenniliLogo, RenniliMark } from '../../shared/components/RenniliLogo';
import { VenueLogo } from '../../shared/components/VenueLogo';
import MatchCode from '../../shared/components/MatchCode';
import { DEPLOY_GAP_PATIENCE_MS } from '../../shared/lib/apiError';
import { usePagerSandbox, usePublicPagerClient } from '../../shared/lib/pagerClients';
import {
  askPushPermission,
  followTicketByPush,
  iosTabWithoutPush,
  usePushKey,
} from '../../shared/lib/pushAlerts';
import { ticketAlarmState } from '../../shared/lib/ticketAlarm';
import { createTicketPoller, isOpenStatus, type TicketPoller } from '../../shared/lib/ticketPoller';

// Shorter than the client's default so one stuck request on a weak signal
// cannot hold up the polls behind it for half a minute.
const POLL_TIMEOUT_MS = 10_000;

const Ticket = ({ token: givenToken }: { token?: string } = {}): React.JSX.Element => {
  const { t, i18n } = useTranslation('pager');
  const { token: rawToken } = useParams({ strict: false });
  const token = givenToken ?? rawToken ?? '';
  const publicPagerClient = usePublicPagerClient();
  const sandbox = usePagerSandbox();
  // The demo phone is not the visitor's own, so it gets no device warning.
  const [iosTab] = useState(() => !sandbox && iosTabWithoutPush());
  const alarm = useAlarm(readyAlarm);
  const pushKey = usePushKey();
  // Bumped by the "alert me" tap so a permission granted there subscribes.
  const [pushAsked, setPushAsked] = useState(0);

  const [ticket, setTicket] = useState<GetPublicTicketResponse | null>(null);
  const [notFound, setNotFound] = useState(token === '');
  const [pollFailed, setPollFailed] = useState(false);
  // Until the first answer, failures read as "still getting your code" for
  // as long as a deploy takes; the poller keeps asking every few seconds.
  const [patienceOver, setPatienceOver] = useState(false);
  const [acknowledging, setAcknowledging] = useState(false);
  const [acknowledgeFailed, setAcknowledgeFailed] = useState(false);
  // The ring this phone answered, so its buttons go away before the next poll
  // confirms it and come back if the staff call again.
  const [answeredRing, setAnsweredRing] = useState<string | null>(null);
  const poller = useRef<TicketPoller | null>(null);

  useEffect(() => {
    if (!token) return undefined;
    const created = createTicketPoller<GetPublicTicketResponse>({
      fetchTicket: () =>
        publicPagerClient.getPublicTicket({ publicToken: token }, { timeoutMs: POLL_TIMEOUT_MS }),
      onTicket: (next) => {
        setTicket(next);
        setPollFailed(false);
      },
      onNotFound: () => setNotFound(true),
      onError: () => setPollFailed(true),
    });
    poller.current = created;
    created.start();
    setPatienceOver(false);
    const patience = setTimeout(() => setPatienceOver(true), DEPLOY_GAP_PATIENCE_MS);
    return () => {
      created.stop();
      clearTimeout(patience);
    };
  }, [token, publicPagerClient]);

  const label = ticket?.label ?? '';
  // A name reads "Sara, your order is ready"; a number "Order 42 is ready".
  const byKind = (key: string): string =>
    t(ticket?.labelKind === 'name' ? `${key}Name` : key, { label });
  const readyTitle = byKind('pager.ticket.readyTitle');
  const readyPushTitle = byKind('pager.push.readyTitle');
  const remindPushTitle = byKind('pager.push.remindTitle');
  useEffect(() => {
    readyAlarm.sync(ticketAlarmState(notFound ? null : ticket, readyTitle));
  }, [ticket, notFound, readyTitle]);
  useEffect(() => () => readyAlarm.release(), []);

  const open = isOpenStatus(ticket?.status);
  const venueName = ticket?.venueName ?? '';
  useEffect(() => {
    if (!open || !pushKey) return;
    const text = {
      ready: {
        title: readyPushTitle,
        body: t('pager.push.readyBody', { venue: venueName }),
      },
      remind: {
        title: remindPushTitle,
        body: t('pager.push.remindBody', { venue: venueName }),
      },
    };
    // The bell on the page does not depend on this, so a failure stays quiet.
    followTicketByPush({ publicKey: pushKey, token, text }).catch(() => undefined);
    // i18n.language: a language switch stores the notification text again.
  }, [
    open,
    pushKey,
    token,
    readyPushTitle,
    remindPushTitle,
    venueName,
    pushAsked,
    i18n.language,
    t,
  ]);

  if (notFound) {
    return (
      <div className="pager-center">
        <HelmetContainer title={t('pager.ticket.pageTitleLoading')} />
        <EmptyState
          icon={SearchX}
          title={t('pager.ticket.notFoundTitle')}
          description={t('pager.ticket.notFoundDescription')}
        />
      </div>
    );
  }

  if (!ticket) {
    return (
      <div className="pager-center">
        <HelmetContainer title={t('pager.ticket.pageTitleLoading')} />
        {pollFailed && patienceOver ? (
          <>
            <p className="pager-error" role="alert">
              {t('pager.ticket.loadFailed')}
            </p>
            <button
              type="button"
              className="pager-button"
              onClick={() => poller.current?.pollNow()}
            >
              {t('pager.ticket.retry')}
            </button>
          </>
        ) : (
          <p className="pager-text" role="status">
            {t(pollFailed ? 'pager.scan.slow' : 'pager.ticket.loading')}
          </p>
        )}
      </div>
    );
  }

  const status = ticket.status ?? '';
  const ready = status === 'ready';
  const ringMark = `${ticket.readyAt ?? ''}|${ticket.remindedAt ?? ''}`;
  const acknowledged = Boolean(ticket.acknowledgedAt) || answeredRing === ringMark;
  const ahead = ticket.ahead ?? 0;
  const labelled = Boolean(ticket.labelKind);
  const labelSize: React.CSSProperties & { '--pager-label-chars': number } = {
    '--pager-label-chars': Math.max(Array.from(label).length, 1),
  };
  const code = ticket.code ?? '';

  const acknowledge = async (): Promise<void> => {
    readyAlarm.silence();
    setAcknowledging(true);
    setAcknowledgeFailed(false);
    try {
      await publicPagerClient.acknowledgeTicket({ publicToken: token });
      setAnsweredRing(ringMark);
      poller.current?.pollNow();
    } catch {
      setAcknowledgeFailed(true);
    } finally {
      setAcknowledging(false);
    }
  };

  const screenClass = [
    'pager-screen',
    ready ? 'pager-screen--ready' : '',
    ready && alarm.ringing ? 'pager-screen--ringing' : '',
  ]
    .filter(Boolean)
    .join(' ');

  return (
    <div className={screenClass}>
      <HelmetContainer
        title={
          labelled ? byKind('pager.ticket.pageTitle') : t('pager.ticket.pageTitleCode', { code })
        }
      />
      <main className="pager-page">
        <div className="pager-topbar">
          <p className="pager-venue">
            <VenueLogo src={ticket.venueLogoUrl} />
            <span>{ticket.venueName}</span>
          </p>
          {/* The demo page's own switch already changes both screens. */}
          {!sandbox && <LanguageSwitch />}
        </div>

        <div className="pager-ticket">
          {labelled ? (
            <h1 className="pager-order">
              {ticket.labelKind === 'number' && (
                <>
                  <span className="pager-order-label">{t('pager.ticket.orderLabel')}</span>{' '}
                </>
              )}
              <bdi className="pager-order-number" style={labelSize}>
                {label}
              </bdi>
            </h1>
          ) : (
            <h1 className="pager-code-card">
              <span className="pager-code-caption">{t('pager.ticket.codeLabel')}</span>{' '}
              <MatchCode code={code} className="pager-code" longClassName="pager-code--long" />
            </h1>
          )}

          {status === 'waiting' && !labelled && (
            <p className="pager-status" role="status">
              {t('pager.ticket.showCode')}
            </p>
          )}

          {status === 'waiting' && labelled && (
            <div role="status">
              <p className="pager-status">{t('pager.ticket.waiting')}</p>
              <p className="pager-text">
                {ahead > 0 ? t('pager.ticket.ahead', { count: ahead }) : t('pager.ticket.next')}
              </p>
            </div>
          )}

          {ready && (
            <div role="alert">
              {/* Keyed on the ring, so a call again shakes it once more. */}
              <RenniliMark key={ringMark} size={56} wiggle className="pager-ready-mark" />
              <p className="pager-status">{t('pager.ticket.ready')}</p>
              <p className="pager-text">
                {acknowledged ? t('pager.ticket.acknowledged') : t('pager.ticket.readyHint')}
              </p>
            </div>
          )}

          {status === 'collected' && (
            <p className="pager-status" role="status">
              {t('pager.ticket.collected')}
            </p>
          )}

          {status === 'expired' && (
            <div role="status">
              <p className="pager-status">
                {labelled ? t('pager.ticket.expired') : t('pager.ticket.codeExpired')}
              </p>
              <p className="pager-text">
                {labelled ? t('pager.ticket.expiredHint') : t('pager.ticket.codeExpiredHint')}
              </p>
            </div>
          )}
        </div>

        {ready && !acknowledged && (
          <div className="pager-actions">
            {acknowledgeFailed && (
              <p className="pager-error" role="alert">
                {t('pager.ticket.acknowledgeFailed')}
              </p>
            )}
            <button
              type="button"
              className="pager-button"
              disabled={acknowledging}
              onClick={() => void acknowledge()}
            >
              {t('pager.ticket.gotIt')}
            </button>
            {alarm.ringing && (
              <button
                type="button"
                className="pager-button pager-button--secondary"
                onClick={() => readyAlarm.silence()}
              >
                {t('pager.ticket.stopSound')}
              </button>
            )}
          </div>
        )}

        {status === 'waiting' && iosTab && (
          <div className="pager-warning" role="note">
            <p className="pager-warning-title">
              <Lock size={16} strokeWidth={2.4} aria-hidden="true" />
              {t('pager.ticket.iphoneTitle')}
            </p>
            <p className="pager-warning-text">{t('pager.ticket.iphoneBody')}</p>
          </div>
        )}

        {/* The browser lets a page ring, buzz or ask to notify only after a tap,
            and the scan gave none, so this is the one tap the customer makes. */}
        {isOpenStatus(status) && !acknowledged && !alarm.unlocked && (
          <div className="pager-panel">
            <p className="pager-text">{t('pager.ticket.notifyHint')}</p>
            <button
              type="button"
              className="pager-button"
              onClick={() => {
                readyAlarm.unlock();
                askPushPermission(pushKey);
                setPushAsked((count) => count + 1);
              }}
            >
              {t('pager.ticket.notify')}
            </button>
          </div>
        )}

        {status === 'waiting' && !iosTab && (
          <p className="pager-text">{t('pager.ticket.keepOpen')}</p>
        )}

        {pollFailed && (
          <p className="pager-text" role="status">
            {t('pager.ticket.connectionLost')}
          </p>
        )}

        <footer className="pager-footer">
          <RenniliLogo size={20} />
          <ByRtylr />
        </footer>
      </main>
    </div>
  );
};

export default Ticket;
