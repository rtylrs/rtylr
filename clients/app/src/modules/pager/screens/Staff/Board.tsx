import React, { useId, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { resolveAppUrl } from '@voxire/config';
import { HelmetContainer } from '@voxire/components';
import { Alert, Switch } from '@voxire/ui';
import type { ListTicketsResponse, Ticket } from '@voxire/voxie/pager/v1/types';
import { Link } from '@tanstack/react-router';
import { BellRing, MonitorSmartphone, Printer } from 'lucide-react';
import { Badge, Button } from '../../../../shared/components/ui';
import { useUserStore } from '../../../../shared/stores/userStore';
import { PAGER_DISPLAY_PATH, PAGER_POSTER_PATH } from '../../pagerPaths';
import { PAGER_ERROR, errorCode, errorStatus } from '../../shared/lib/apiError';
import MatchCode from '../../shared/components/MatchCode';
import {
  type StaffPagerClient,
  usePagerSandbox,
  useStaffPagerClient,
} from '../../shared/lib/pagerClients';
import { normaliseTicketNumber } from '../../shared/lib/ticketNumber';
import { TICKET_POLL_INTERVAL_MS, isStaleAnswer } from '../../shared/lib/ticketPoller';
import { usePagerVenue } from '../../shared/lib/usePagerVenue';
import { VenueLogo } from '../../shared/components/VenueLogo';
import { Dropdown } from '../../shared/components/Dropdown';
import { VenueQr } from './VenueQr';

const TICKETS_KEY = ['pager', 'tickets'] as const;
const SETTINGS_KEY = ['pager', 'settings'] as const;
const LABEL_MODES = ['either', 'number', 'name'] as const;
type LabelMode = (typeof LABEL_MODES)[number];

type BoardError =
  'invalidLabel' | 'labelTaken' | 'invalidTransition' | 'settingsAdminOnly' | 'actionFailed';

function boardErrorFor(error: unknown): BoardError {
  switch (errorCode(error)) {
    case PAGER_ERROR.invalidLabel:
      return 'invalidLabel';
    case PAGER_ERROR.labelTaken:
      return 'labelTaken';
    case PAGER_ERROR.invalidTransition:
      return 'invalidTransition';
    default:
      return errorStatus(error) === 403 ? 'settingsAdminOnly' : 'actionFailed';
  }
}

type TicketAction = 'ready' | 'remind' | 'undo' | 'collect';

function runTicketAction(
  client: StaffPagerClient,
  action: TicketAction,
  id: string
): Promise<unknown> {
  switch (action) {
    case 'ready':
      return client.markReady({ id });
    case 'remind':
      return client.remindTicket({ id });
    case 'undo':
      return client.undoReady({ id });
    case 'collect':
      return client.collectTicket({ id });
  }
}

function asLabelMode(value: string | undefined): LabelMode {
  return LABEL_MODES.find((mode) => mode === value) ?? 'either';
}

/** Whole minutes between the scan and the board's last answer. */
function minutesSince(createdAt: string | undefined, serverTime: string | undefined): number {
  const since = Date.parse(serverTime ?? '') - Date.parse(createdAt ?? '');
  return Number.isNaN(since) ? 0 : Math.max(0, Math.floor(since / 60_000));
}

interface NewPhoneCardProps {
  ticket: Ticket;
  labelMode: LabelMode;
  serverTime: string | undefined;
  busy: boolean;
  onLabel: (label: string, done: () => void) => void;
  onRemove: () => void;
}

/** A phone that scanned the QR and waits for the cashier to type its label. */
function NewPhoneCard({
  ticket,
  labelMode,
  serverTime,
  busy,
  onLabel,
  onRemove,
}: NewPhoneCardProps): React.JSX.Element {
  const { t } = useTranslation('pager');
  const inputId = useId();
  const [value, setValue] = useState('');
  const code = ticket.code ?? '';
  const minutes = minutesSince(ticket.createdAt, serverTime);
  const numeric = labelMode === 'number';

  return (
    <li className="space-y-2 rounded-lg p-3" style={{ border: '1px solid var(--dash-border)' }}>
      <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
        {/* Captioned, so the cashier never takes the code for the order number. */}
        <div className="flex flex-col leading-none">
          <span
            className="text-[11px] font-semibold uppercase tracking-wider"
            style={{ color: 'var(--dash-text-dim)' }}
          >
            {t('pager.board.codeCaption')}
          </span>
          <MatchCode
            code={code}
            className="mt-1 text-3xl font-bold tracking-[0.06em] [font-variant-numeric:lining-nums_tabular-nums]"
            longClassName="text-lg!"
          />
        </div>
        <span className="min-w-[5rem] flex-1 text-xs" style={{ color: 'var(--dash-text-dim)' }}>
          {minutes < 1
            ? t('pager.board.scannedNow')
            : t('pager.board.scannedAgo', { count: minutes })}
        </span>
        <Button
          size="sm"
          variant="ghost"
          className="ms-auto"
          aria-label={t('pager.board.removePhone', { code })}
          disabled={busy}
          onClick={onRemove}
        >
          {t('pager.board.remove')}
        </Button>
      </div>
      <form
        className="space-y-1"
        onSubmit={(event) => {
          event.preventDefault();
          if (value.trim() && !busy) onLabel(value.trim(), () => setValue(''));
        }}
      >
        {/* What to type, in sight; the input's name also says which code it is for. */}
        <span
          aria-hidden="true"
          className="block text-xs font-semibold"
          style={{ color: 'var(--dash-text-muted)' }}
        >
          {t(`pager.board.labelInput.${labelMode}`)}
        </span>
        <label htmlFor={inputId} className="sr-only">
          {t(`pager.board.labelFor.${labelMode}`, { code })}
        </label>
        <div className="flex gap-2">
          <input
            id={inputId}
            dir="auto"
            inputMode={numeric ? 'numeric' : 'text'}
            enterKeyHint="done"
            autoComplete="off"
            autoCapitalize={numeric ? 'characters' : 'words'}
            spellCheck={false}
            maxLength={40}
            value={value}
            onChange={(event) =>
              setValue(numeric ? normaliseTicketNumber(event.target.value) : event.target.value)
            }
            className="h-11 min-w-0 flex-1 rounded-lg px-3 text-base outline-none focus:border-[var(--dash-border-focus)]"
            style={{
              background: 'var(--dash-input-bg)',
              border: '1px solid var(--dash-border)',
              color: 'var(--dash-text)',
            }}
          />
          <Button type="submit" size="sm" className="h-11" disabled={!value.trim() || busy}>
            {t('pager.board.saveLabel')}
          </Button>
        </div>
      </form>
    </li>
  );
}

const Board = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const queryClient = useQueryClient();
  const enabledId = useId();
  const labelModeId = useId();
  const [boardError, setBoardError] = useState<BoardError | null>(null);
  const shown = useRef<ListTicketsResponse | null>(null);
  const staffPagerClient = useStaffPagerClient();
  const sandbox = usePagerSandbox();

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

  const labelTicket = useMutation({
    mutationFn: ({ id, label }: { id: string; label: string }) =>
      staffPagerClient.labelTicket({ id, label }),
    onMutate: beforeChange,
    onError: onChangeFailed,
    onSettled: afterChange,
  });

  const changeTicket = useMutation({
    mutationFn: ({ id, action }: { id: string; action: TicketAction }) =>
      runTicketAction(staffPagerClient, action, id),
    onMutate: beforeChange,
    onError: onChangeFailed,
    onSettled: afterChange,
  });

  const updateSettings = useMutation({
    mutationFn: (patch: { enabled?: boolean; labelMode?: string }) =>
      staffPagerClient.updateSettings(patch),
    onMutate: () => setBoardError(null),
    onSuccess: (next) => queryClient.setQueryData(SETTINGS_KEY, next),
    onError: onChangeFailed,
  });

  const venue = usePagerVenue();
  const customerLink = venue.joinUrl;
  // The server refuses a settings change from anyone else; this only saves
  // them the round trip. In the demo the visitor plays the owner.
  const isOwnerOrAdmin = useUserStore(
    (state) =>
      Boolean(state.user?.effectiveAccess?.isOwner) || Boolean(state.user?.effectiveAccess?.isAdmin)
  );
  const canEditSettings = sandbox || isOwnerOrAdmin;

  if (errorStatus(tickets.error) === 403 || errorStatus(settings.error) === 403) {
    return (
      <Alert variant="warning">
        <HelmetContainer title={t('pager.board.pageTitle')} />
        {t('pager.board.noAccess')}
      </Alert>
    );
  }

  const labelMode = asLabelMode(settings.data?.labelMode);
  const open = tickets.data?.tickets ?? [];
  const newPhones = open.filter((ticket) => ticket.status === 'waiting' && !ticket.labelKind);
  const waiting = open.filter((ticket) => ticket.status === 'waiting' && ticket.labelKind);
  const ready = open.filter((ticket) => ticket.status === 'ready');
  const busyId =
    (changeTicket.isPending ? changeTicket.variables?.id : undefined) ??
    (labelTicket.isPending ? labelTicket.variables?.id : undefined);

  const actionButton = (
    ticket: Ticket,
    action: TicketAction,
    label: string,
    variant: 'default' | 'outline' | 'ghost'
  ): React.JSX.Element => (
    <Button
      size="sm"
      variant={variant}
      aria-label={t('pager.board.actionFor', { action: label, label: ticket.label ?? '' })}
      disabled={busyId === ticket.id}
      onClick={() => changeTicket.mutate({ id: ticket.id ?? '', action })}
    >
      {label}
    </Button>
  );

  const columnShell = (
    title: string,
    count: number,
    emptyText: string,
    list: React.ReactNode
  ): React.JSX.Element => (
    <section
      className="min-w-0 rounded-xl p-3"
      style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
    >
      <h2 className="mb-3 text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
        {title} ({count})
      </h2>
      {count === 0 ? (
        <p className="text-sm" style={{ color: 'var(--dash-text-dim)' }}>
          {emptyText}
        </p>
      ) : (
        <ul className="flex flex-col gap-2">{list}</ul>
      )}
    </section>
  );

  const labelledColumn = (
    title: string,
    emptyText: string,
    items: Ticket[],
    actions: (ticket: Ticket) => React.ReactNode
  ): React.JSX.Element =>
    columnShell(
      title,
      items.length,
      emptyText,
      items.map((ticket) => (
        <li
          key={ticket.id}
          className="flex flex-wrap items-center justify-between gap-2 rounded-lg p-3"
          style={{ border: '1px solid var(--dash-border)' }}
        >
          <div className="flex min-w-0 flex-wrap items-center gap-2">
            {ticket.labelKind === 'number' && (
              <span
                className="text-[11px] font-semibold uppercase tracking-wider"
                style={{ color: 'var(--dash-text-dim)' }}
              >
                {t('pager.board.orderCaption')}
              </span>
            )}
            <bdi
              className={
                ticket.labelKind === 'number'
                  ? 'text-2xl font-bold tabular-nums [overflow-wrap:anywhere]'
                  : 'text-xl font-bold [overflow-wrap:anywhere]'
              }
              style={{ color: 'var(--dash-text)' }}
            >
              {ticket.label}
            </bdi>
            {ticket.acknowledgedAt && (
              <Badge variant="success">{t('pager.board.acknowledged')}</Badge>
            )}
          </div>
          <div className="flex flex-wrap gap-2">{actions(ticket)}</div>
        </li>
      ))
    );

  return (
    <>
      <HelmetContainer title={t('pager.board.pageTitle')} />
      {/* Laid out by its own width, not the window's: the demo shows it beside a phone. */}
      <div className="@container w-full min-w-0 space-y-5">
        {/* The demo page names the board itself and has no venue to show a QR
            for, and the display and poster pages need a sign in. */}
        {!sandbox && (
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div className="min-w-0">
              <div className="flex items-center gap-2">
                <BellRing
                  size={17}
                  aria-hidden="true"
                  style={{ color: 'var(--dash-text-muted)' }}
                />
                <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
                  {t('pager.board.heading')}
                </h1>
              </div>
              <p className="mt-1 text-xs" style={{ color: 'var(--dash-text-dim)' }}>
                {t('pager.board.subheading')}
              </p>
            </div>
            <div className="flex flex-wrap gap-2">
              {/* The Portal's Button renders a <button> only, so the links wear its classes. */}
              <Link
                to={PAGER_DISPLAY_PATH}
                className="portal-button portal-button--outline portal-button--sm"
              >
                <MonitorSmartphone size={16} aria-hidden="true" />
                {t('pager.board.displayLink')}
              </Link>
              <Link
                to={PAGER_POSTER_PATH}
                className="portal-button portal-button--outline portal-button--sm"
              >
                <Printer size={16} aria-hidden="true" />
                {t('pager.board.posterLink')}
              </Link>
            </div>
          </div>
        )}

        {boardError && <Alert variant="destructive">{t(`pager.board.${boardError}`)}</Alert>}
        {/* The API leaves a false flag out of the answer altogether. */}
        {settings.data && !settings.data.enabled && (
          <Alert variant="warning">{t('pager.board.disabledNotice')}</Alert>
        )}

        {/* On a phone the order is QR, tickets, settings; on a wide screen the
            tickets take the second column beside the other two. */}
        <div className="grid gap-5 @4xl:grid-cols-[260px_minmax(0,1fr)] @4xl:grid-rows-[auto_1fr]">
          {customerLink ? (
            <section
              className="flex flex-col items-center gap-2 rounded-xl p-4 text-center @xl:grid @xl:grid-cols-[auto_minmax(0,22rem)] @xl:grid-rows-[1fr_auto_auto_auto_1fr] @xl:items-start @xl:justify-center @xl:gap-x-6 @xl:gap-y-1 @xl:text-start @4xl:col-start-1 @4xl:row-start-1 @4xl:flex @4xl:items-center @4xl:gap-2 @4xl:text-center"
              style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
            >
              {/* Beside the QR on a tablet in portrait, so the tickets start higher. */}
              <h2
                className="text-sm font-semibold @xl:col-start-2 @xl:row-start-2"
                style={{ color: 'var(--dash-text)' }}
              >
                {t('pager.board.qrHeading')}
              </h2>
              <div className="rounded-lg bg-white p-2 @xl:col-start-1 @xl:row-span-5 @xl:row-start-1 @xl:self-center">
                <VenueQr url={customerLink} size={180} />
              </div>
              <p
                className="text-xs @xl:col-start-2 @xl:row-start-3"
                style={{ color: 'var(--dash-text-dim)' }}
              >
                {t('pager.board.qrHint')}
              </p>
              <a
                href={customerLink}
                target="_blank"
                rel="noreferrer"
                dir="ltr"
                className="inline-block break-all py-3.5 text-xs underline @xl:col-start-2 @xl:row-start-4"
                style={{ color: 'var(--dash-text)' }}
              >
                {customerLink}
              </a>
            </section>
          ) : (
            <span className="hidden @4xl:block" />
          )}

          <div className="@container min-w-0 space-y-3 @4xl:col-start-2 @4xl:row-span-2 @4xl:row-start-1">
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
              // New phones gets the most room: its cards hold an input.
              <div className="grid gap-3 @md:grid-cols-2 @2xl:grid-cols-[minmax(0,1.25fr)_minmax(0,1fr)_minmax(0,1fr)]">
                {columnShell(
                  t('pager.board.newColumn'),
                  newPhones.length,
                  t('pager.board.emptyNew'),
                  newPhones.map((ticket) => (
                    <NewPhoneCard
                      key={ticket.id}
                      ticket={ticket}
                      labelMode={labelMode}
                      serverTime={tickets.data.serverTime}
                      busy={busyId === ticket.id}
                      onLabel={(label, done) =>
                        labelTicket.mutate({ id: ticket.id ?? '', label }, { onSuccess: done })
                      }
                      onRemove={() =>
                        changeTicket.mutate({ id: ticket.id ?? '', action: 'collect' })
                      }
                    />
                  ))
                )}
                {labelledColumn(
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
                {labelledColumn(
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

          <section className="space-y-3 @4xl:col-start-1 @4xl:row-start-2">
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
                <div className="space-y-1">
                  <label
                    id={`${labelModeId}-label`}
                    htmlFor={labelModeId}
                    className="block text-xs font-semibold"
                    style={{ color: 'var(--dash-text-muted)' }}
                  >
                    {t('pager.board.labelModeLabel')}
                  </label>
                  <Dropdown
                    id={labelModeId}
                    labelId={`${labelModeId}-label`}
                    aria-describedby={`${labelModeId}-hint`}
                    value={labelMode}
                    options={LABEL_MODES.map((mode) => ({
                      value: mode,
                      label: t(`pager.board.labelMode.${mode}`),
                    }))}
                    disabled={!canEditSettings || !settings.data || updateSettings.isPending}
                    onChange={(mode) => updateSettings.mutate({ labelMode: mode })}
                  />
                  <p
                    id={`${labelModeId}-hint`}
                    className="text-xs"
                    style={{ color: 'var(--dash-text-dim)' }}
                  >
                    {t('pager.board.labelModeHint')}
                  </p>
                </div>
                {/* The logo is the business's own, set where the Portal sets it. */}
                {!sandbox && (
                  <div className="space-y-1">
                    <p
                      className="text-xs font-semibold"
                      style={{ color: 'var(--dash-text-muted)' }}
                    >
                      {t('pager.board.logoLabel')}
                    </p>
                    <div className="flex items-center gap-3">
                      <VenueLogo src={venue.logoUrl} className="pager-board-logo" />
                      <a
                        href={new URL('/settings/businesses', resolveAppUrl('dash')).href}
                        className="portal-button portal-button--outline portal-button--sm"
                      >
                        {venue.logoUrl ? t('pager.board.logoChange') : t('pager.board.logoAdd')}
                      </a>
                    </div>
                    <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
                      {t('pager.board.logoHint')}
                    </p>
                  </div>
                )}
                {!canEditSettings && (
                  <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
                    {t('pager.board.settingsAdminOnly')}
                  </p>
                )}
              </>
            )}
          </section>
        </div>
      </div>
    </>
  );
};

export default Board;
