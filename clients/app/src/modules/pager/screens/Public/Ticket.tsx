import React, { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useParams } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import type { GetPublicTicketResponse } from '@voxire/voxie/pager/v1/types';
import { SearchX } from 'lucide-react';
import { readyAlarm } from '../../../../shared/alarm/readyAlarm';
import { useAlarm } from '../../../../shared/alarm/useAlarm';
import LanguageSwitch from '../../shared/components/LanguageSwitch';
import { publicPagerClient } from '../../shared/lib/clients';
import { ticketAlarmState } from '../../shared/lib/ticketAlarm';
import { createTicketPoller, isOpenStatus, type TicketPoller } from '../../shared/lib/ticketPoller';

// Shorter than the client's default so one stuck request on a weak signal
// cannot hold up the polls behind it for half a minute.
const POLL_TIMEOUT_MS = 10_000;
const LONG_NUMBER_LENGTH = 5;

const Ticket = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const { token: rawToken } = useParams({ strict: false });
  const token = rawToken ?? '';
  const alarm = useAlarm(readyAlarm);

  const [ticket, setTicket] = useState<GetPublicTicketResponse | null>(null);
  const [notFound, setNotFound] = useState(token === '');
  const [pollFailed, setPollFailed] = useState(false);
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
    return () => created.stop();
  }, [token]);

  const number = ticket?.number ?? '';
  const readyTitle = t('pager.ticket.readyTitle', { number });
  useEffect(() => {
    readyAlarm.sync(ticketAlarmState(notFound ? null : ticket, readyTitle));
  }, [ticket, notFound, readyTitle]);
  useEffect(() => () => readyAlarm.release(), []);

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
        {pollFailed ? (
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
            {t('pager.ticket.loading')}
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
      <HelmetContainer title={t('pager.ticket.pageTitle', { number })} />
      <main className="pager-page">
        <div className="pager-topbar">
          <p className="pager-venue">{ticket.venueName}</p>
          <LanguageSwitch />
        </div>

        <div className="pager-ticket">
          <h1 className="pager-order">
            <span className="pager-order-label">{t('pager.ticket.orderLabel')}</span>{' '}
            <bdi
              className={
                number.length > LONG_NUMBER_LENGTH
                  ? 'pager-order-number pager-order-number--long'
                  : 'pager-order-number'
              }
            >
              {number}
            </bdi>
          </h1>

          {status === 'waiting' && (
            <div role="status">
              <p className="pager-status">{t('pager.ticket.waiting')}</p>
              <p className="pager-text">
                {ahead > 0 ? t('pager.ticket.ahead', { count: ahead }) : t('pager.ticket.next')}
              </p>
            </div>
          )}

          {ready && (
            <div role="alert">
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
              <p className="pager-status">{t('pager.ticket.expired')}</p>
              <p className="pager-text">{t('pager.ticket.expiredHint')}</p>
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

        {/* A reload or a shared link arrives without the tap that unlocks sound. */}
        {isOpenStatus(status) && !acknowledged && !alarm.unlocked && (
          <div className="pager-panel">
            <p className="pager-text">{t('pager.ticket.alertMeHint')}</p>
            <button type="button" className="pager-button" onClick={() => readyAlarm.unlock()}>
              {t('pager.ticket.alertMe')}
            </button>
          </div>
        )}

        {status === 'waiting' && <p className="pager-text">{t('pager.ticket.keepOpen')}</p>}

        {pollFailed && (
          <p className="pager-text" role="status">
            {t('pager.ticket.connectionLost')}
          </p>
        )}
      </main>
    </div>
  );
};

export default Ticket;
