import React, { useId, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { resolveAppUrl } from '@voxire/config';
import { HelmetContainer } from '@voxire/components';
import { Alert, Switch } from '@voxire/ui';
import type { ListTicketsResponse, Ticket } from '@voxire/voxie/pager/v1/types';
import { Link } from '@tanstack/react-router';
import {
  BellRing,
  Check,
  MonitorSmartphone,
  Printer,
  QrCode,
  RotateCcw,
  Settings2,
  X,
} from 'lucide-react';
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
type QueueTab = 'new' | 'waiting' | 'ready';

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

/** Whole minutes between creation and the board's last answer. */
function minutesSince(createdAt: string | undefined, serverTime: string | undefined): number {
  const since = Date.parse(serverTime ?? '') - Date.parse(createdAt ?? '');
  return Number.isNaN(since) ? 0 : Math.max(0, Math.floor(since / 60_000));
}

interface NewPhonesDockProps {
  phones: Ticket[];
  selectedId: string | undefined;
  onSelect: (id: string) => void;
  labelMode: LabelMode;
  serverTime: string | undefined;
  busyId: string | undefined;
  onLabel: (id: string, label: string, done: () => void) => void;
  onRemove: (id: string) => void;
}

/**
 * High-volume scan dock: scanned phones sit as compact selectable chips above
 * one shared quick-label bar, so 1 or 30 scans stay tidy without giant boxes.
 */
function NewPhonesDock({
  phones,
  selectedId,
  onSelect,
  labelMode,
  serverTime,
  busyId,
  onLabel,
  onRemove,
}: NewPhonesDockProps): React.JSX.Element {
  const { t } = useTranslation('pager');
  const inputId = useId();
  const inputRef = useRef<HTMLInputElement>(null);
  const [value, setValue] = useState('');
  const active = phones.find((p) => p.id === selectedId) ?? phones[0];
  const numeric = labelMode === 'number';
  const activeCode = active?.code ?? '';
  const busy = Boolean(active && busyId === active.id);

  if (phones.length === 0) {
    return (
      <section className="pager-board-dock pager-board-dock--empty">
        <div className="pager-board-lane-head">
          <h2 className="pager-board-lane-title">
            {t('pager.board.newColumn')}
            <span className="pager-board-count">0</span>
          </h2>
        </div>
        <p className="pager-board-empty">{t('pager.board.emptyNew')}</p>
      </section>
    );
  }

  return (
    <section className="pager-board-dock">
      <div className="pager-board-lane-head">
        <h2 className="pager-board-lane-title">
          {t('pager.board.newColumn')}
          <span className="pager-board-count pager-board-count--accent">{phones.length}</span>
        </h2>
      </div>

      <ul className="pager-board-chips" role="list">
        {phones.map((ticket) => {
          const code = ticket.code ?? '';
          const minutes = minutesSince(ticket.createdAt, serverTime);
          const isSelected = active?.id === ticket.id;
          const isBusy = busyId === ticket.id;
          return (
            <li key={ticket.id} className="pager-board-chip-item">
              <button
                type="button"
                aria-pressed={isSelected}
                aria-label={t('pager.board.selectPhone', { code })}
                className={`pager-board-chip ${isSelected ? 'pager-board-chip--selected' : ''}`}
                onClick={() => {
                  if (ticket.id) onSelect(ticket.id);
                  inputRef.current?.focus();
                }}
              >
                <MatchCode
                  code={code}
                  className="pager-board-chip-code"
                  longClassName="pager-board-chip-code--long"
                />
                <span className="pager-board-chip-time">
                  {minutes < 1
                    ? t('pager.board.waitNow')
                    : t('pager.board.waitAgo', { count: minutes })}
                </span>
              </button>
              <button
                type="button"
                className="pager-board-chip-remove"
                aria-label={t('pager.board.removePhone', { code })}
                disabled={isBusy}
                onClick={() => ticket.id && onRemove(ticket.id)}
              >
                <X size={13} strokeWidth={2.5} aria-hidden="true" />
              </button>
            </li>
          );
        })}
      </ul>

      {active && (
        <form
          className="pager-board-assign"
          onSubmit={(event) => {
            event.preventDefault();
            if (value.trim() && !busy && active.id) {
              onLabel(active.id, value.trim(), () => {
                setValue('');
                inputRef.current?.focus();
              });
            }
          }}
        >
          <div className="pager-board-assign-badge">
            <span className="pager-board-assign-caption">{t('pager.board.codeCaption')}</span>
            <MatchCode
              code={activeCode}
              className="pager-board-assign-code"
              longClassName="pager-board-assign-code--long"
            />
          </div>
          <label htmlFor={inputId} className="sr-only">
            {t(`pager.board.labelFor.${labelMode}`, { code: activeCode })}
          </label>
          <input
            ref={inputRef}
            id={inputId}
            dir="auto"
            placeholder={t(`pager.board.labelInput.${labelMode}`)}
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
            className="pager-board-assign-input"
          />
          <Button
            type="submit"
            size="sm"
            className="pager-board-assign-submit"
            disabled={!value.trim() || busy}
          >
            {t('pager.board.saveLabel')}
          </Button>
        </form>
      )}
    </section>
  );
}

const Board = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const queryClient = useQueryClient();
  const enabledId = useId();
  const labelModeId = useId();
  const [boardError, setBoardError] = useState<BoardError | null>(null);
  const [selectedNewId, setSelectedNewId] = useState<string | undefined>(undefined);
  const [activeTab, setActiveTab] = useState<QueueTab>('new');
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [qrOpen, setQrOpen] = useState(false);
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
  const pagerEnabled = settings.data?.enabled ?? false;
  const open = tickets.data?.tickets ?? [];
  const newPhones = open.filter((ticket) => ticket.status === 'waiting' && !ticket.labelKind);
  const waiting = open.filter((ticket) => ticket.status === 'waiting' && ticket.labelKind);
  const ready = open.filter((ticket) => ticket.status === 'ready');
  const busyId =
    (changeTicket.isPending ? changeTicket.variables?.id : undefined) ??
    (labelTicket.isPending ? labelTicket.variables?.id : undefined);

  const renderOrderTile = (ticket: Ticket, kind: 'waiting' | 'ready'): React.JSX.Element => {
    const minutes = minutesSince(ticket.createdAt, tickets.data?.serverTime);
    const isBusy = busyId === ticket.id;
    return (
      <li
        key={ticket.id}
        className={`pager-board-tile ${kind === 'ready' ? 'pager-board-tile--ready' : ''}`}
      >
        <div className="pager-board-tile-top">
          <div className="pager-board-tile-id">
            {ticket.labelKind === 'number' && (
              <span className="pager-board-tile-caption">{t('pager.board.orderCaption')}</span>
            )}
            <bdi
              className={
                ticket.labelKind === 'number'
                  ? 'pager-board-tile-label pager-board-tile-label--num'
                  : 'pager-board-tile-label'
              }
            >
              {ticket.label}
            </bdi>
          </div>
          <span className="pager-board-tile-time">
            {minutes < 1 ? t('pager.board.waitNow') : t('pager.board.waitAgo', { count: minutes })}
          </span>
        </div>

        {ticket.acknowledgedAt && (
          <div className="pager-board-tile-meta">
            <Badge variant="success">{t('pager.board.acknowledged')}</Badge>
          </div>
        )}

        <div className="pager-board-tile-actions">
          {kind === 'waiting' ? (
            <>
              <Button
                size="sm"
                variant="default"
                className="pager-board-ready-btn flex-1"
                aria-label={t('pager.board.actionFor', {
                  action: t('pager.board.markReady'),
                  label: ticket.label ?? '',
                })}
                disabled={isBusy}
                onClick={() => changeTicket.mutate({ id: ticket.id ?? '', action: 'ready' })}
              >
                <BellRing size={14} aria-hidden="true" />
                {t('pager.board.markReady')}
              </Button>
              <Button
                size="sm"
                variant="ghost"
                aria-label={t('pager.board.actionFor', {
                  action: t('pager.board.remove'),
                  label: ticket.label ?? '',
                })}
                disabled={isBusy}
                onClick={() => changeTicket.mutate({ id: ticket.id ?? '', action: 'collect' })}
              >
                <X size={15} aria-hidden="true" />
              </Button>
            </>
          ) : (
            <>
              <Button
                size="sm"
                variant="default"
                className="pager-board-collect-btn flex-1"
                aria-label={t('pager.board.actionFor', {
                  action: t('pager.board.collect'),
                  label: ticket.label ?? '',
                })}
                disabled={isBusy}
                onClick={() => changeTicket.mutate({ id: ticket.id ?? '', action: 'collect' })}
              >
                <Check size={14} aria-hidden="true" />
                {t('pager.board.collect')}
              </Button>
              <Button
                size="sm"
                variant="outline"
                aria-label={t('pager.board.actionFor', {
                  action: t('pager.board.remind'),
                  label: ticket.label ?? '',
                })}
                disabled={isBusy}
                onClick={() => changeTicket.mutate({ id: ticket.id ?? '', action: 'remind' })}
              >
                <BellRing size={14} aria-hidden="true" />
              </Button>
              <Button
                size="sm"
                variant="ghost"
                aria-label={t('pager.board.actionFor', {
                  action: t('pager.board.undo'),
                  label: ticket.label ?? '',
                })}
                disabled={isBusy}
                onClick={() => changeTicket.mutate({ id: ticket.id ?? '', action: 'undo' })}
              >
                <RotateCcw size={14} aria-hidden="true" />
              </Button>
            </>
          )}
        </div>
      </li>
    );
  };

  const renderOrderLane = (
    title: string,
    emptyText: string,
    items: Ticket[],
    kind: 'waiting' | 'ready'
  ): React.JSX.Element => (
    <section className={`pager-board-lane ${kind === 'ready' ? 'pager-board-lane--ready' : ''}`}>
      <div className="pager-board-lane-head">
        <h2 className="pager-board-lane-title">
          {title}
          <span
            className={`pager-board-count ${kind === 'ready' && items.length > 0 ? 'pager-board-count--ready' : ''}`}
          >
            {items.length}
          </span>
        </h2>
      </div>
      {items.length === 0 ? (
        <p className="pager-board-empty">{emptyText}</p>
      ) : (
        <ul className="pager-board-grid" role="list">
          {items.map((ticket) => renderOrderTile(ticket, kind))}
        </ul>
      )}
    </section>
  );

  return (
    <>
      <HelmetContainer title={t('pager.board.pageTitle')} />
      {/* Laid out by its own width, not the window's: the demo shows it beside a phone. */}
      <div className="@container pager-board-root">
        <header className="pager-board-toolbar">
          <div className="pager-board-toolbar-start">
            {!sandbox && (
              <h1 className="pager-board-heading">
                <BellRing size={17} aria-hidden="true" />
                <span>{t('pager.board.heading')}</span>
              </h1>
            )}
            {settings.data && (
              <span
                className={`pager-board-status-pill ${pagerEnabled ? 'pager-board-status-pill--on' : 'pager-board-status-pill--off'}`}
              >
                <span className="pager-board-status-dot" aria-hidden="true" />
                {t(pagerEnabled ? 'pager.board.statusOn' : 'pager.board.statusOff')}
              </span>
            )}
          </div>

          <div className="pager-board-toolbar-actions">
            {!sandbox && customerLink && (
              <Button
                size="sm"
                variant={qrOpen ? 'default' : 'outline'}
                aria-expanded={qrOpen}
                onClick={() => setQrOpen((openState) => !openState)}
              >
                <QrCode size={15} aria-hidden="true" />
                <span>{t('pager.board.qrButton')}</span>
              </Button>
            )}
            {!sandbox && (
              <>
                <Link
                  to={PAGER_DISPLAY_PATH}
                  className="portal-button portal-button--outline portal-button--sm"
                >
                  <MonitorSmartphone size={15} aria-hidden="true" />
                  {t('pager.board.displayLink')}
                </Link>
                <Link
                  to={PAGER_POSTER_PATH}
                  className="portal-button portal-button--outline portal-button--sm"
                >
                  <Printer size={15} aria-hidden="true" />
                  {t('pager.board.posterLink')}
                </Link>
              </>
            )}
            <Button
              size="sm"
              variant={settingsOpen ? 'default' : 'outline'}
              aria-expanded={settingsOpen}
              onClick={() => setSettingsOpen((openState) => !openState)}
            >
              <Settings2 size={15} aria-hidden="true" />
              <span>{t('pager.board.settingsButton')}</span>
            </Button>
          </div>
        </header>

        {boardError && <Alert variant="destructive">{t(`pager.board.${boardError}`)}</Alert>}
        {settings.data && !settings.data.enabled && (
          <Alert variant="warning">{t('pager.board.disabledNotice')}</Alert>
        )}

        {qrOpen && customerLink && (
          <section className="pager-board-drawer pager-board-qr-drawer">
            <div className="rounded-lg bg-white p-2">
              <VenueQr url={customerLink} size={140} />
            </div>
            <div className="min-w-0 flex-1 space-y-1">
              <h2 className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                {t('pager.board.qrHeading')}
              </h2>
              <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
                {t('pager.board.qrHint')}
              </p>
              <a
                href={customerLink}
                target="_blank"
                rel="noreferrer"
                dir="ltr"
                className="inline-block break-all pt-1 text-xs underline"
                style={{ color: 'var(--dash-text)' }}
              >
                {customerLink}
              </a>
            </div>
          </section>
        )}

        {settingsOpen && (
          <section className="pager-board-drawer" aria-label={t('pager.board.settingsHeading')}>
            {settings.isError ? (
              <div className="space-y-2">
                <Alert variant="destructive">{t('pager.board.loadFailed')}</Alert>
                <Button size="sm" variant="outline" onClick={() => void settings.refetch()}>
                  {t('pager.board.retry')}
                </Button>
              </div>
            ) : (
              <div className="pager-board-settings-grid">
                <div>
                  <Switch
                    id={enabledId}
                    label={t('pager.board.enabledLabel')}
                    aria-describedby={`${enabledId}-hint`}
                    checked={pagerEnabled}
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
                  </div>
                )}

                {!canEditSettings && (
                  <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
                    {t('pager.board.settingsAdminOnly')}
                  </p>
                )}
              </div>
            )}
          </section>
        )}

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
          <>
            {/* Segmented tabs on compact screens: keeps 20+ tickets clean without stacked boxes. */}
            <div
              className="pager-board-tabs"
              role="tablist"
              aria-label={t('pager.board.tabsLabel')}
            >
              {(
                [
                  { id: 'new', label: t('pager.board.newColumn'), count: newPhones.length },
                  { id: 'waiting', label: t('pager.board.waitingColumn'), count: waiting.length },
                  { id: 'ready', label: t('pager.board.readyColumn'), count: ready.length },
                ] as const
              ).map((tab) => (
                <button
                  key={tab.id}
                  type="button"
                  role="tab"
                  aria-selected={activeTab === tab.id}
                  className={`pager-board-tab ${activeTab === tab.id ? 'pager-board-tab--active' : ''}`}
                  onClick={() => setActiveTab(tab.id)}
                >
                  <span>{tab.label}</span>
                  <span className="pager-board-tab-count">{tab.count}</span>
                </button>
              ))}
            </div>

            <div className="pager-board-workspace" data-active-tab={activeTab}>
              <div className="pager-board-section pager-board-section--new">
                <NewPhonesDock
                  phones={newPhones}
                  selectedId={selectedNewId}
                  onSelect={setSelectedNewId}
                  labelMode={labelMode}
                  serverTime={tickets.data.serverTime}
                  busyId={busyId}
                  onLabel={(id, label, done) => {
                    labelTicket.mutate(
                      { id, label },
                      {
                        onSuccess: () => {
                          done();
                          setSelectedNewId(undefined);
                        },
                      }
                    );
                  }}
                  onRemove={(id) => changeTicket.mutate({ id, action: 'collect' })}
                />
              </div>

              <div className="pager-board-lanes">
                <div className="pager-board-section pager-board-section--waiting">
                  {renderOrderLane(
                    t('pager.board.waitingColumn'),
                    t('pager.board.emptyWaiting'),
                    waiting,
                    'waiting'
                  )}
                </div>
                <div className="pager-board-section pager-board-section--ready">
                  {renderOrderLane(
                    t('pager.board.readyColumn'),
                    t('pager.board.emptyReady'),
                    ready,
                    'ready'
                  )}
                </div>
              </div>
            </div>
          </>
        )}
      </div>
    </>
  );
};

export default Board;
