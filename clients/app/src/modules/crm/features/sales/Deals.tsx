import React, { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { Columns3, Handshake, List, Plus, Settings2 } from 'lucide-react';
import { Button } from '../../shared/components/ui';
import {
  amountOf,
  type Deal,
  DEAL_SORTS,
  DEAL_STATUSES,
  type DealSort,
  type DealStatus,
  dueState,
  initials,
  type Stage,
  stageById,
} from './data/model';
import {
  type BoardFilters,
  useDeal,
  useDealBoard,
  useDealList,
  useDealStats,
  useMoveDeal,
  usePipelines,
  useStageDeals,
} from './data/queries';
import {
  DealStatusBadge,
  DueDate,
  EmptyBlock,
  ListFooter,
  LoadFailed,
  Money,
  RepFilterSelect,
  RepName,
  SearchBox,
  StatTile,
  useMoney,
  useNextPageOnScroll,
  useRepNames,
} from './components/common';
import { DealFormModal, LoseDealModal, WinDealModal } from './components/DealModals';

const DealCard: React.FC<{
  deal: Deal;
  names: Map<string, string>;
  onOpen: () => void;
  draggable: boolean;
}> = ({ deal, names, onOpen, draggable }) => {
  const { t } = useTranslation();
  const owner = deal.assignedTo ? names.get(deal.assignedTo) : undefined;
  const closeState = deal.status === 'open' ? dueState(deal.expectedCloseDate) : undefined;
  return (
    <button
      type="button"
      className="sales-deal-card"
      draggable={draggable}
      onDragStart={(event) => {
        event.dataTransfer.setData('text/x-deal', deal.id ?? '');
        event.dataTransfer.effectAllowed = 'move';
      }}
      onClick={onOpen}
    >
      <span className="sales-deal-title">{deal.title}</span>
      {(() => {
        const subtitle = [deal.companyName !== deal.title ? deal.companyName : '', deal.contactName]
          .filter(Boolean)
          .join(' · ');
        return subtitle ? <span className="sales-cell-sub">{subtitle}</span> : null;
      })()}
      <span className="sales-deal-meta">
        <Money value={deal.value} billing={deal.billing} />
        {owner && (
          <span className="sales-avatar" title={owner} aria-label={owner}>
            {initials(owner)}
          </span>
        )}
      </span>
      {(deal.expectedCloseDate || deal.nextFollowUp) && (
        <span className="sales-deal-meta">
          {deal.expectedCloseDate ? (
            <span
              className={closeState === 'overdue' ? 'sales-due' : undefined}
              data-state={closeState}
            >
              {t('crmSales.deals.closes')} <DueDate date={deal.expectedCloseDate} />
            </span>
          ) : (
            <span />
          )}
          {deal.nextFollowUp && dueState(deal.nextFollowUp) !== 'upcoming' && (
            <span
              className="sales-badge"
              data-tone={dueState(deal.nextFollowUp) === 'overdue' ? 'red' : 'amber'}
            >
              {t('crmSales.deals.followUpDue')}
            </span>
          )}
        </span>
      )}
    </button>
  );
};

const StageColumn: React.FC<{
  stage: Stage;
  deals: number;
  amount?: string;
  weighted?: string;
  filters: BoardFilters;
  names: Map<string, string>;
  onOpen: (deal: Deal) => void;
  onDropDeal: (dealId: string, stage: Stage) => void;
}> = ({ stage, deals, amount, weighted, filters, names, onOpen, onDropDeal }) => {
  const { t, i18n } = useTranslation();
  const money = useMoney();
  const list = useStageDeals(stage.id, filters);
  const sentinel = useNextPageOnScroll(list);
  const [over, setOver] = useState(false);
  const rows = list.data?.pages.flatMap((page) => page.deals ?? []) ?? [];
  return (
    <section
      className={`sales-column${over ? ' is-drop-target' : ''}`}
      aria-label={stage.name}
      onDragOver={(event) => {
        if (!event.dataTransfer.types.includes('text/x-deal')) return;
        event.preventDefault();
        setOver(true);
      }}
      onDragLeave={() => setOver(false)}
      onDrop={(event) => {
        event.preventDefault();
        setOver(false);
        const id = event.dataTransfer.getData('text/x-deal');
        if (id) onDropDeal(id, stage);
      }}
    >
      <div className="sales-column-head">
        <div className="sales-column-title">
          <span>{stage.name}</span>
          <span className="sales-tab-count">
            {new Intl.NumberFormat(i18n.language).format(deals)}
          </span>
        </div>
        <div className="sales-column-meta">
          {money(amount)}
          {stage.kind === 'open' && (
            <>
              {' · '}
              {t('crmSales.deals.weightedShort', {
                amount: money(weighted),
                probability: stage.probability ?? 0,
              })}
            </>
          )}
        </div>
      </div>
      {rows.map((deal) => (
        <DealCard
          key={deal.id}
          deal={deal}
          names={names}
          onOpen={() => onOpen(deal)}
          draggable={deal.status === 'open'}
        />
      ))}
      {list.hasNextPage && (
        <Button
          ref={sentinel}
          variant="ghost"
          onClick={() => void list.fetchNextPage()}
          disabled={list.isFetchingNextPage}
        >
          {list.isFetchingNextPage ? t('crmSales.common.loading') : t('crmSales.common.loadMore')}
        </Button>
      )}
      {list.isPending && (
        <span className="sales-muted" style={{ fontSize: 12 }}>
          {t('crmSales.common.loading')}
        </span>
      )}
    </section>
  );
};

const DealTable: React.FC<{ pipelineId?: string; rep: string; search: string }> = ({
  pipelineId,
  rep,
  search,
}) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const names = useRepNames();
  const board = useDealBoard({ pipelineId, rep, search: '' });
  const [statuses, setStatuses] = useState<DealStatus[]>(['open']);
  const [sort, setSort] = useState<DealSort>('close_date');
  const deals = useDealList({ pipelineId, statuses, rep, search, sort });
  const rows = deals.data?.pages.flatMap((page) => page.deals ?? []) ?? [];
  const pipeline = board.data?.pipeline;
  return (
    <div className="sales-card">
      <div className="sales-filters" style={{ padding: 12 }}>
        <div className="sales-chip-row" role="group" aria-label={t('crmSales.fields.status')}>
          {DEAL_STATUSES.map((status) => (
            <button
              key={status}
              type="button"
              className="sales-chip"
              aria-pressed={statuses.includes(status)}
              onClick={() =>
                setStatuses((current) =>
                  current.includes(status)
                    ? current.filter((item) => item !== status)
                    : [...current, status]
                )
              }
            >
              {t(`crmSales.dealStatus.${status}`)}
            </button>
          ))}
        </div>
        <span className="sales-spacer" />
        <select
          className="sales-select"
          aria-label={t('crmSales.common.sortBy')}
          value={sort}
          onChange={(event) => setSort(event.target.value as DealSort)}
        >
          {DEAL_SORTS.map((item) => (
            <option key={item} value={item}>
              {t(`crmSales.deals.sort.${item}`)}
            </option>
          ))}
        </select>
      </div>
      {deals.isError && rows.length === 0 ? (
        <LoadFailed title={t('crmSales.deals.loadFailed')} onRetry={() => void deals.refetch()} />
      ) : rows.length === 0 ? (
        <p className="sales-empty">
          {deals.isPending ? t('crmSales.common.loading') : t('crmSales.deals.noMatches')}
        </p>
      ) : (
        <>
          <div className="sales-table-wrap">
            <table className="sales-table">
              <thead>
                <tr>
                  <th>{t('crmSales.fields.deal')}</th>
                  <th>{t('crmSales.fields.stage')}</th>
                  <th>{t('crmSales.fields.value')}</th>
                  <th>{t('crmSales.fields.owner')}</th>
                  <th>{t('crmSales.fields.expectedClose')}</th>
                  <th>{t('crmSales.fields.followUp')}</th>
                  <th>{t('crmSales.fields.status')}</th>
                </tr>
              </thead>
              <tbody>
                {rows.map((deal) => (
                  <tr key={deal.id} onClick={() => navigate({ to: `/crm/sales/deals/${deal.id}` })}>
                    <td>
                      <div className="sales-cell-main">{deal.title}</div>
                      <div className="sales-cell-sub">
                        {[deal.companyName, deal.contactName].filter(Boolean).join(' · ')}
                      </div>
                    </td>
                    <td>{stageById(pipeline, deal.stageId)?.name ?? '—'}</td>
                    <td>
                      <Money value={deal.value} billing={deal.billing} />
                    </td>
                    <td>
                      <RepName id={deal.assignedTo} names={names} />
                    </td>
                    <td>
                      <DueDate
                        date={deal.status === 'open' ? deal.expectedCloseDate : undefined}
                        emptyLabel={deal.expectedCloseDate || '—'}
                      />
                    </td>
                    <td>
                      <DueDate date={deal.nextFollowUp} />
                    </td>
                    <td>
                      <DealStatusBadge status={deal.status} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <ListFooter shown={rows.length} total={deals.data?.pages[0]?.total} pages={deals} />
        </>
      )}
    </div>
  );
};

const Deals: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const money = useMoney();
  const names = useRepNames();
  const [pipelineId, setPipelineId] = useState<string | undefined>();
  const [rep, setRep] = useState('all');
  const [search, setSearch] = useState('');
  const [view, setView] = useState<'board' | 'list'>('board');
  const [creating, setCreating] = useState(false);
  const [winning, setWinning] = useState<Deal | null>(null);
  const [losing, setLosing] = useState<Deal | null>(null);
  const filters = useMemo<BoardFilters>(
    () => ({ pipelineId, rep, search }),
    [pipelineId, rep, search]
  );
  const board = useDealBoard(filters);
  const stats = useDealStats(board.data?.pipeline?.id ?? pipelineId, rep);
  const move = useMoveDeal();
  const pipeline = board.data?.pipeline;
  const columns = board.data?.columns ?? [];
  const allDeals = columns.reduce((sum, column) => sum + (column.deals ?? 0), 0);

  const openDeal = (deal: Deal) => navigate({ to: `/crm/sales/deals/${deal.id}` });
  const dropDeal = (dealId: string, stage: Stage): void => {
    const placeholder: Deal = { id: dealId, title: '', pipelineId: pipeline?.id };
    if (stage.kind === 'won') setWinning(placeholder);
    else if (stage.kind === 'lost') setLosing(placeholder);
    else if (stage.id) move.mutate({ id: dealId, stageId: stage.id });
  };

  return (
    <div className="sales-page">
      <HelmetContainer title={`${t('crmSales.deals.title')} · rtylr`} />
      <header className="sales-header">
        <div>
          <h1>{t('crmSales.deals.title')}</h1>
          <p>{t('crmSales.deals.subtitle')}</p>
        </div>
        <div className="sales-actions">
          <Button variant="outline" onClick={() => navigate({ to: '/crm/sales/pipelines' })}>
            <Settings2 size={13} />
            {t('crmSales.deals.editPipeline')}
          </Button>
          <Button onClick={() => setCreating(true)}>
            <Plus size={13} />
            {t('crmSales.deals.add')}
          </Button>
        </div>
      </header>

      <div className="sales-stats">
        <StatTile
          label={t('crmSales.deals.stats.open')}
          value={money(stats.data?.openAmount)}
          hint={t('crmSales.deals.stats.openHint', { count: stats.data?.openDeals ?? 0 })}
        />
        <StatTile
          label={t('crmSales.deals.stats.weighted')}
          value={money(stats.data?.weightedAmount)}
          hint={t('crmSales.deals.stats.weightedHint')}
        />
        <StatTile
          label={t('crmSales.deals.stats.wonThisMonth')}
          value={money(stats.data?.wonAmountThisMonth)}
          hint={t('crmSales.deals.stats.wonHint', { count: stats.data?.wonThisMonth ?? 0 })}
        />
        <StatTile
          label={t('crmSales.deals.stats.winRate')}
          value={`${Math.round(stats.data?.winRate ?? 0)}%`}
          hint={t('crmSales.deals.stats.winRateHint', {
            days: Math.round(stats.data?.averageDaysToWin ?? 0),
            amount: money(amountOf(stats.data?.averageWonAmount)),
          })}
        />
      </div>

      <div className="sales-filters">
        <SearchBox
          value={search}
          onChange={setSearch}
          placeholder={t('crmSales.deals.searchPlaceholder')}
        />
        <RepFilterSelect value={rep} onChange={setRep} label={t('crmSales.fields.owner')} />
        {board.data && (board.data.pipeline?.stages?.length ?? 0) > 0 && (
          <PipelinePicker value={pipeline?.id} onChange={setPipelineId} />
        )}
        <span className="sales-spacer" />
        <div className="sales-chip-row" role="group" aria-label={t('crmSales.deals.view')}>
          <button
            type="button"
            className="sales-chip"
            aria-pressed={view === 'board'}
            onClick={() => setView('board')}
          >
            <Columns3 size={13} />
            {t('crmSales.deals.board')}
          </button>
          <button
            type="button"
            className="sales-chip"
            aria-pressed={view === 'list'}
            onClick={() => setView('list')}
          >
            <List size={13} />
            {t('crmSales.deals.list')}
          </button>
        </div>
      </div>

      {board.isError && !board.data ? (
        <LoadFailed title={t('crmSales.deals.loadFailed')} onRetry={() => void board.refetch()} />
      ) : !board.data ? (
        <p className="sales-empty" role="status">
          {t('crmSales.common.loading')}
        </p>
      ) : view === 'list' ? (
        <DealTable pipelineId={pipeline?.id} rep={rep} search={search} />
      ) : allDeals === 0 && search === '' && rep === 'all' ? (
        <div className="sales-card">
          <EmptyBlock
            icon={Handshake}
            title={t('crmSales.deals.emptyTitle')}
            description={t('crmSales.deals.emptyHint')}
          >
            <Button variant="outline" onClick={() => navigate({ to: '/crm/sales/leads' })}>
              {t('crmSales.deals.goToLeads')}
            </Button>
            <Button onClick={() => setCreating(true)}>
              <Plus size={13} />
              {t('crmSales.deals.add')}
            </Button>
          </EmptyBlock>
        </div>
      ) : (
        <div className="sales-board" aria-label={pipeline?.name}>
          {columns.map((column) =>
            column.stage ? (
              <StageColumn
                key={column.stage.id}
                stage={column.stage}
                deals={column.deals ?? 0}
                amount={column.amount}
                weighted={column.weightedAmount}
                filters={filters}
                names={names}
                onOpen={openDeal}
                onDropDeal={dropDeal}
              />
            ) : null
          )}
        </div>
      )}

      <DealFormModal
        open={creating}
        onClose={() => setCreating(false)}
        pipelineId={pipeline?.id}
        onCreated={openDeal}
      />
      {winning && <DealCloser deal={winning} kind="won" onClose={() => setWinning(null)} />}
      {losing && <DealCloser deal={losing} kind="lost" onClose={() => setLosing(null)} />}
    </div>
  );
};

const PipelinePicker: React.FC<{ value?: string; onChange: (id: string) => void }> = ({
  value,
  onChange,
}) => {
  const { t } = useTranslation();
  const { data: pipelines = [] } = usePipelines();
  if (pipelines.length < 2) return null;
  return (
    <select
      className="sales-select"
      aria-label={t('crmSales.fields.pipeline')}
      value={value ?? ''}
      onChange={(event) => onChange(event.target.value)}
    >
      {pipelines.map((pipeline) => (
        <option key={pipeline.id} value={pipeline.id}>
          {pipeline.name}
        </option>
      ))}
    </select>
  );
};

/** Loads the full deal a board drop names before closing it. */
const DealCloser: React.FC<{ deal: Deal; kind: 'won' | 'lost'; onClose: () => void }> = ({
  deal,
  kind,
  onClose,
}) => {
  const full = useDeal(deal.id).data?.deal;
  if (!full) return null;
  return kind === 'won' ? (
    <WinDealModal deal={full} open onClose={onClose} />
  ) : (
    <LoseDealModal deal={full} open onClose={onClose} />
  );
};

export default Deals;
