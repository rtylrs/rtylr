import React, { useId, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { HelmetContainer } from '@voxire/components';
import { Alert, Switch } from '@voxire/ui';
import type { ListTicketsResponse, Ticket } from '@voxire/voxie/pager/v1/types';
import { Link } from '@tanstack/react-router';
import { BellRing, Delete, Printer, Smartphone } from 'lucide-react';
import { Badge, Button } from '../../../../shared/components/ui';
import { useUserStore } from '../../../../shared/stores/userStore';
import { PAGER_POSTER_PATH } from '../../pagerPaths';
import { PAGER_ERROR, errorCode, errorStatus } from '../../shared/lib/apiError';
import { staffPagerClient } from '../../shared/lib/clients';
import { normaliseTicketNumber } from '../../shared/lib/ticketNumber';
import { TICKET_POLL_INTERVAL_MS, isStaleAnswer } from '../../shared/lib/ticketPoller';
import { usePagerVenue } from '../../shared/lib/usePagerVenue';

const TICKETS_KEY = ['pager', 'tickets'] as const;
const SETTINGS_KEY = ['pager', 'settings'] as const;
const KEYPAD_DIGITS = ['1', '2', '3', '4', '5', '6', '7', '8', '9'] as const;

type BoardError =
  'invalidNumber' | 'invalidTransition' | 'tooManyOpen' | 'settingsAdminOnly' | 'actionFailed';

function boardErrorFor(error: unknown): BoardError {
  switch (errorCode(error)) {
    case PAGER_ERROR.invalidNumber:
      return 'invalidNumber';
    case PAGER_ERROR.invalidTransition:
      return 'invalidTransition';
    case PAGER_ERROR.tooManyOpenTickets:
      return 'tooManyOpen';
    default:
      return errorStatus(error) === 403 ? 'settingsAdminOnly' : 'actionFailed';
  }
}

type TicketAction = 'ready' | 'remind' | 'undo' | 'collect';

const TICKET_ACTIONS: Record<TicketAction, (id: string) => Promise<unknown>> = {
  ready: (id) => staffPagerClient.markReady({ id }),
  remind: (id) => staffPagerClient.remindTicket({ id }),
  undo: (id) => staffPagerClient.undoReady({ id }),
  collect: (id) => staffPagerClient.collectTicket({ id }),
};

const Board = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const queryClient = useQueryClient();
  const inputId = useId();
  const enabledId = useId();
  const strictId = useId();
  const [number, setNumber] = useState('');
  const [boardError, setBoardError] = useState<BoardError | null>(null);
  const shown = useRef<ListTicketsResponse | null>(null);

  const tickets = useQuery({
    queryKey: TICKETS_KEY,
    queryFn: async () => {
      const next = await staffPagerClient.listTickets({});
      // A slow answer must not overwrite a newer one already on screen.
      if (shown.current && isStaleAnswer(shown.current.serverTime, next.serverTime)) {
        return shown.current;
      }
      shown.current = next;
      return next;
    },
    refetchInterval: TICKET_POLL_INTERVAL_MS,
    refetchOnWindowFocus: true,
    staleTime: 0,
    retry: false,
  });

  const settings = useQuery({
    queryKey: SETTINGS_KEY,
    queryFn: () => staffPagerClient.getSettings({}),
  });

  // A poll that left before a change would land after it and briefly show the
  // old board, so it is dropped and the list is fetched again afterwards.
  const beforeChange = async (): Promise<void> => {
    setBoardError(null);
    await queryClient.cancelQueries({ queryKey: TICKETS_KEY });
  };
  const afterChange = (): Promise<void> => queryClient.invalidateQueries({ queryKey: TICKETS_KEY });
  const onChangeFailed = (error: unknown): void => setBoardError(boardErrorFor(error));

  const createTicket = useMutation({
    mutationFn: (value: string) => staffPagerClient.createTicket({ number: value }),
    onMutate: beforeChange,
    onSuccess: () => setNumber(''),
    onError: onChangeFailed,
    onSettled: afterChange,
  });

  const changeTicket = useMutation({
    mutationFn: ({ id, action }: { id: string; action: TicketAction }) =>
      TICKET_ACTIONS[action](id),
    onMutate: beforeChange,
    onError: onChangeFailed,
    onSettled: afterChange,
  });

  const updateSettings = useMutation({
    mutationFn: (patch: { enabled?: boolean; strict?: boolean }) =>
      staffPagerClient.updateSettings(patch),
    onMutate: () => setBoardError(null),
    onSuccess: (next) => queryClient.setQueryData(SETTINGS_KEY, next),
    onError: onChangeFailed,
  });

  const customerLink = usePagerVenue().joinUrl;
  // The server refuses a settings change from anyone else; this only saves
  // them the round trip.
  const canEditSettings = useUserStore(
    (state) =>
      Boolean(state.user?.effectiveAccess?.isOwner) || Boolean(state.user?.effectiveAccess?.isAdmin)
  );

  if (errorStatus(tickets.error) === 403 || errorStatus(settings.error) === 403) {
    return (
      <Alert variant="warning">
        <HelmetContainer title={t('pager.board.pageTitle')} />
        {t('pager.board.noAccess')}
      </Alert>
    );
  }

  const open = tickets.data?.tickets ?? [];
  const waiting = open.filter((ticket) => ticket.status === 'waiting');
  const ready = open.filter((ticket) => ticket.status === 'ready');
  const busyId = changeTicket.isPending ? changeTicket.variables?.id : undefined;

  const actionButton = (
    ticket: Ticket,
    action: TicketAction,
    label: string,
    variant: 'default' | 'outline' | 'ghost'
  ): React.JSX.Element => (
    <Button
      size="sm"
      variant={variant}
      aria-label={t('pager.board.actionFor', { action: label, number: ticket.number ?? '' })}
      disabled={busyId === ticket.id}
      onClick={() => changeTicket.mutate({ id: ticket.id ?? '', action })}
    >
      {label}
    </Button>
  );

  const column = (
    title: string,
    emptyText: string,
    items: Ticket[],
    actions: (ticket: Ticket) => React.ReactNode
  ): React.JSX.Element => (
    <section
      className="min-w-0 rounded-xl p-3"
      style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
    >
      <h2 className="mb-3 text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
        {title} ({items.length})
      </h2>
      {items.length === 0 ? (
        <p className="text-sm" style={{ color: 'var(--dash-text-dim)' }}>
          {emptyText}
        </p>
      ) : (
        <ul className="flex flex-col gap-2">
          {items.map((ticket) => {
            const ticketNumber = ticket.number ?? '';
            return (
              <li
                key={ticket.id}
                className="flex flex-wrap items-center justify-between gap-2 rounded-lg p-3"
                style={{ border: '1px solid var(--dash-border)' }}
              >
                <div className="flex min-w-0 flex-wrap items-center gap-2">
                  <bdi
                    className="text-2xl font-bold tabular-nums"
                    style={{ color: 'var(--dash-text)' }}
                  >
                    {ticketNumber}
                  </bdi>
                  {ticket.phoneLinked && (
                    <Smartphone
                      size={18}
                      role="img"
                      aria-label={t('pager.board.phoneLinked')}
                      style={{ color: 'var(--dash-text-muted)' }}
                    />
                  )}
                  {ticket.acknowledgedAt && (
                    <Badge variant="success">{t('pager.board.acknowledged')}</Badge>
                  )}
                </div>
                <div className="flex flex-wrap gap-2">{actions(ticket)}</div>
              </li>
            );
          })}
        </ul>
      )}
    </section>
  );

  const add = (): void => {
    if (number && !createTicket.isPending) createTicket.mutate(number);
  };

  return (
    <>
      <HelmetContainer title={t('pager.board.pageTitle')} />
      <div className="w-full min-w-0 space-y-5">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <BellRing size={17} aria-hidden="true" style={{ color: 'var(--dash-text-muted)' }} />
              <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
                {t('pager.board.heading')}
              </h1>
            </div>
            <p className="mt-1 text-xs" style={{ color: 'var(--dash-text-dim)' }}>
              {t('pager.board.subheading')}
            </p>
          </div>
          {/* The Portal's Button renders a <button> only, so the link wears its classes. */}
          <Link
            to={PAGER_POSTER_PATH}
            className="portal-button portal-button--outline portal-button--sm"
          >
            <Printer size={16} aria-hidden="true" />
            {t('pager.board.posterLink')}
          </Link>
        </div>

        {boardError && <Alert variant="destructive">{t(`pager.board.${boardError}`)}</Alert>}
        {/* The API leaves a false flag out of the answer altogether. */}
        {settings.data && !settings.data.enabled && (
          <Alert variant="warning">{t('pager.board.disabledNotice')}</Alert>
        )}

        {/* On a phone the order is keypad, tickets, settings; on a wide screen the
            tickets take the second column beside the other two. */}
        <div className="grid gap-5 lg:grid-cols-[280px_minmax(0,1fr)] lg:grid-rows-[auto_1fr]">
          <form
            className="space-y-3 lg:col-start-1 lg:row-start-1"
            onSubmit={(event) => {
              event.preventDefault();
              add();
            }}
          >
            <label
              htmlFor={inputId}
              className="block text-xs font-semibold"
              style={{ color: 'var(--dash-text-muted)' }}
            >
              {t('pager.board.numberLabel')}
            </label>
            <input
              id={inputId}
              dir="ltr"
              inputMode="numeric"
              autoComplete="off"
              autoCapitalize="characters"
              spellCheck={false}
              value={number}
              onChange={(event) => setNumber(normaliseTicketNumber(event.target.value))}
              className="h-14 w-full rounded-lg px-3 text-center text-3xl! font-bold tabular-nums outline-none focus:border-[var(--dash-border-focus)]"
              style={{
                background: 'var(--dash-input-bg)',
                border: '1px solid var(--dash-border)',
                color: 'var(--dash-text)',
              }}
            />
            {/* A phone keypad reads left to right in every language. */}
            <div
              dir="ltr"
              role="group"
              aria-label={t('pager.board.keypadLabel')}
              className="grid grid-cols-3 gap-2"
            >
              {KEYPAD_DIGITS.map((digit) => (
                <Button
                  key={digit}
                  variant="outline"
                  className="h-14 text-xl"
                  onClick={() => setNumber((value) => normaliseTicketNumber(value + digit))}
                >
                  {digit}
                </Button>
              ))}
              <Button
                variant="ghost"
                className="h-14"
                aria-label={t('pager.board.backspace')}
                disabled={!number}
                onClick={() => setNumber((value) => value.slice(0, -1))}
              >
                <Delete size={20} aria-hidden="true" />
              </Button>
              <Button
                variant="outline"
                className="h-14 text-xl"
                onClick={() => setNumber((value) => normaliseTicketNumber(`${value}0`))}
              >
                0
              </Button>
              <span />
            </div>
            <Button
              type="submit"
              className="h-12 w-full"
              disabled={!number || createTicket.isPending}
            >
              {t('pager.board.add')}
            </Button>
          </form>

          <div className="min-w-0 space-y-3 lg:col-start-2 lg:row-start-1 lg:row-span-2">
            {tickets.isPending && (
              <p className="text-sm" role="status" style={{ color: 'var(--dash-text-dim)' }}>
                {t('pager.board.loading')}
              </p>
            )}
            {tickets.isError && !tickets.data && (
              <div className="space-y-2">
                <Alert variant="destructive">{t('pager.board.loadFailed')}</Alert>
                <Button size="sm" variant="outline" onClick={() => void tickets.refetch()}>
                  {t('pager.board.retry')}
                </Button>
              </div>
            )}
            {tickets.isError && tickets.data && (
              <p className="text-sm" role="status" style={{ color: 'var(--dash-text-dim)' }}>
                {t('pager.board.connectionLost')}
              </p>
            )}
            {tickets.data && (
              <div className="grid gap-3 md:grid-cols-2">
                {column(
                  t('pager.board.waitingColumn'),
                  t('pager.board.emptyWaiting'),
                  waiting,
                  (ticket) => (
                    <>
                      {actionButton(ticket, 'ready', t('pager.board.markReady'), 'default')}
                      {actionButton(ticket, 'collect', t('pager.board.remove'), 'ghost')}
                    </>
                  )
                )}
                {column(
                  t('pager.board.readyColumn'),
                  t('pager.board.emptyReady'),
                  ready,
                  (ticket) => (
                    <>
                      {actionButton(ticket, 'remind', t('pager.board.remind'), 'outline')}
                      {actionButton(ticket, 'undo', t('pager.board.undo'), 'ghost')}
                      {actionButton(ticket, 'collect', t('pager.board.collect'), 'default')}
                    </>
                  )
                )}
              </div>
            )}
          </div>

          <section className="space-y-3 lg:col-start-1 lg:row-start-2">
            <h2 className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
              {t('pager.board.settingsHeading')}
            </h2>
            {settings.isError ? (
              <div className="space-y-2">
                <Alert variant="destructive">{t('pager.board.loadFailed')}</Alert>
                <Button size="sm" variant="outline" onClick={() => void settings.refetch()}>
                  {t('pager.board.retry')}
                </Button>
              </div>
            ) : (
              <>
                <div>
                  <Switch
                    id={enabledId}
                    label={t('pager.board.enabledLabel')}
                    aria-describedby={`${enabledId}-hint`}
                    checked={settings.data?.enabled ?? false}
                    disabled={!canEditSettings || !settings.data || updateSettings.isPending}
                    onCheckedChange={(enabled) => updateSettings.mutate({ enabled })}
                  />
                  <p
                    id={`${enabledId}-hint`}
                    className="mt-1 text-xs"
                    style={{ color: 'var(--dash-text-dim)' }}
                  >
                    {t('pager.board.enabledHint')}
                  </p>
                </div>
                <div>
                  <Switch
                    id={strictId}
                    label={t('pager.board.strictLabel')}
                    aria-describedby={`${strictId}-hint`}
                    checked={settings.data?.strict ?? false}
                    disabled={!canEditSettings || !settings.data || updateSettings.isPending}
                    onCheckedChange={(strict) => updateSettings.mutate({ strict })}
                  />
                  <p
                    id={`${strictId}-hint`}
                    className="mt-1 text-xs"
                    style={{ color: 'var(--dash-text-dim)' }}
                  >
                    {t('pager.board.strictHint')}
                  </p>
                </div>
                {!canEditSettings && (
                  <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
                    {t('pager.board.settingsAdminOnly')}
                  </p>
                )}
              </>
            )}
            {customerLink && (
              <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
                {t('pager.board.customerLink')}{' '}
                <a
                  href={customerLink}
                  target="_blank"
                  rel="noreferrer"
                  dir="ltr"
                  className="break-all underline"
                  style={{ color: 'var(--dash-text)' }}
                >
                  {customerLink}
                </a>
              </p>
            )}
          </section>
        </div>
      </div>
    </>
  );
};

export default Board;
