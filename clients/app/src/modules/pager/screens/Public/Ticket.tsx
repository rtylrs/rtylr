import React, { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useParams } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import type { GetPublicTicketResponse } from '@voxire/voxie/pager/v1/types';
import { BellRing, Check, Lock, SearchX, Star } from 'lucide-react';
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
  askPushPermissionNow,
  followTicketByPush,
  iosTabWithoutPush,
  notificationPermission,
  usePushKey,
} from '../../shared/lib/pushAlerts';
import { forgetScanByToken, rememberedSlugForToken } from '../../shared/lib/scanSession';
import { ticketAlarmState } from '../../shared/lib/ticketAlarm';
import { createTicketPoller, isOpenStatus, type TicketPoller } from '../../shared/lib/ticketPoller';
import {
  RINGTONES,
  type RingtoneId,
  asRingtone,
  previewRingtone,
  rememberRingtone,
  rememberedRingtone,
  ringtoneSrc,
} from '../../shared/lib/ringtones';

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
  // The customer's own tone, kept on this phone; the demo phone keeps none.
  const [pickedTone, setPickedTone] = useState<RingtoneId | null>(() =>
    sandbox ? null : rememberedRingtone()
  );
  const alarm = useAlarm(readyAlarm);
  const pushKey = usePushKey();
  // Bumped by the "alert me" tap so a permission granted there subscribes.
  const [pushAsked, setPushAsked] = useState(0);
  const [notifPerm, setNotifPerm] = useState<NotificationPermission | 'unsupported'>(() =>
    typeof notificationPermission === 'function' ? notificationPermission() : 'unsupported'
  );
  const [venueSlug] = useState<string | null>(() =>
    !sandbox ? rememberedSlugForToken(token) : null
  );

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
  const ringtone = pickedTone ?? asRingtone(ticket?.ringtone);
  useEffect(() => {
    readyAlarm.sync(ticketAlarmState(notFound ? null : ticket, readyTitle, ringtoneSrc(ringtone)));
  }, [ticket, notFound, readyTitle, ringtone]);
  useEffect(() => () => readyAlarm.release(), []);

  const open = isOpenStatus(ticket?.status);
  const venueName = ticket?.venueName ?? '';
  const ticketFinished =
    ticket?.status === 'collected' ||
    ticket?.status === 'expired' ||
    (ticket?.status === 'ready' && Boolean(ticket.acknowledgedAt));

  useEffect(() => {
    if (!sandbox && ticketFinished && token) {
      forgetScanByToken(token);
    }
  }, [sandbox, ticketFinished, token]);
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
  // Whole minutes to the venue's estimate, by the server's clock; null without one.
  const etaLeft = Date.parse(ticket.expectedReadyAt ?? '') - Date.parse(ticket.serverTime ?? '');
  const etaMinutes = Number.isNaN(etaLeft) ? null : Math.max(0, Math.ceil(etaLeft / 60_000));
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
      if (!sandbox && token) forgetScanByToken(token);
      poller.current?.pollNow();
    } catch {
      setAcknowledgeFailed(true);
    } finally {
      setAcknowledging(false);
    }
  };

  const needsPushPrompt = !sandbox && Boolean(pushKey) && notifPerm === 'default';
  const showNotifyPanel =
    isOpenStatus(status) && !acknowledged && (!alarm.unlocked || needsPushPrompt);
  const showNewOrderCta =
    !sandbox &&
    Boolean(venueSlug) &&
    (status === 'collected' || status === 'expired' || (ready && acknowledged));
  const showNewOrderLink = !sandbox && Boolean(venueSlug) && isOpenStatus(status) && !acknowledged;

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
          {status === 'waiting' && alarm.unlocked && !needsPushPrompt && (
            <p className="pager-live-pill">
              <Check size={14} strokeWidth={2.8} aria-hidden="true" />
              <span>{t('pager.ticket.alertsActive')}</span>
            </p>
          )}

          {labelled ? (
            <div className={ready ? 'pager-order-hero' : 'pager-code-card pager-code-card--order'}>
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
              {status === 'waiting' && (
                <div className="pager-card-status" role="status">
                  <p className="pager-status">{t('pager.ticket.waiting')}</p>
                  <p className="pager-text">
                    {ahead > 0 ? t('pager.ticket.ahead', { count: ahead }) : t('pager.ticket.next')}
                  </p>
                  {etaMinutes !== null && (
                    <p className="pager-eta">
                      {etaMinutes > 0
                        ? t('pager.ticket.etaMinutes', { count: etaMinutes })
                        : t('pager.ticket.etaSoon')}
                    </p>
                  )}
                </div>
              )}
            </div>
          ) : (
            <div className="pager-code-card">
              <h1 className="pager-code-heading">
                <span className="pager-code-caption">{t('pager.ticket.codeLabel')}</span>{' '}
                <MatchCode code={code} className="pager-code" longClassName="pager-code--long" />
              </h1>
              {status === 'waiting' && (
                <p className="pager-card-instruction" role="status">
                  {t('pager.ticket.showCode')}
                </p>
              )}
            </div>
          )}

          {ready && (
            <div className="pager-ready-block" role="alert">
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

        {ticket.reviewUrl && ((ready && acknowledged) || status === 'collected') && (
          <div className="pager-review">
            <p className="pager-text">{t('pager.ticket.reviewPrompt', { venue: venueName })}</p>
            <a
              className="pager-review-link"
              href={ticket.reviewUrl}
              target="_blank"
              rel="noopener noreferrer"
            >
              <Star size={17} strokeWidth={2.4} aria-hidden="true" />
              {t('pager.ticket.reviewCta')}
            </a>
          </div>
        )}

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

        {/* Shown when sound is still locked OR browser notification permission hasn't been asked yet. */}
        {showNotifyPanel && (
          <div className="pager-panel">
            <p className="pager-text">{t('pager.ticket.notifyHint')}</p>
            <button
              type="button"
              className="pager-button"
              onClick={() => {
                readyAlarm.unlock();
                readyAlarm.blip();
                if (pushKey) askPushPermission(pushKey);
                const pending =
                  !sandbox && typeof askPushPermissionNow === 'function'
                    ? askPushPermissionNow()
                    : null;
                if (pending) {
                  void pending.then((next) => {
                    setNotifPerm(next);
                    setPushAsked((count) => count + 1);
                  });
                } else {
                  setPushAsked((count) => count + 1);
                }
              }}
            >
              <BellRing size={18} strokeWidth={2.3} aria-hidden="true" />
              <span>{t('pager.ticket.notify')}</span>
            </button>
          </div>
        )}

        {status === 'waiting' && (
          <div className="pager-ringtones">
            <p className="pager-ringtones-label" id="pager-ringtones-label">
              {t('pager.ticket.ringtoneLabel')}
            </p>
            <div
              className="pager-ringtones-list"
              role="group"
              aria-labelledby="pager-ringtones-label"
            >
              {RINGTONES.map((tone) => (
                <button
                  key={tone.id}
                  type="button"
                  className="pager-ringtone"
                  aria-pressed={ringtone === tone.id}
                  onClick={() => {
                    setPickedTone(tone.id);
                    if (!sandbox) rememberRingtone(tone.id);
                    previewRingtone(tone.id);
                  }}
                >
                  {t(`pager.ringtones.${tone.id}`)}
                </button>
              ))}
            </div>
          </div>
        )}

        {!sandbox && isOpenStatus(status) && !acknowledged && notifPerm === 'denied' && (
          <p className="pager-subtle-note" role="note">
            {t('pager.ticket.notificationsBlocked')}
          </p>
        )}

        {status === 'waiting' && !iosTab && alarm.unlocked && !needsPushPrompt && (
          <p className="pager-subtle-note">{t('pager.ticket.keepOpen')}</p>
        )}

        {showNewOrderCta && venueSlug && (
          <div className="pager-actions">
            <a href={`/${venueSlug}?new=1`} className="pager-button">
              {t('pager.ticket.newOrder')}
            </a>
          </div>
        )}

        {showNewOrderLink && venueSlug && (
          <a href={`/${venueSlug}?new=1`} className="pager-text-action">
            {t('pager.ticket.newOrderLink')}
          </a>
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
