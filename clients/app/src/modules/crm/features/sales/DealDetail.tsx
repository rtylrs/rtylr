import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { Link, useNavigate, useParams } from '@tanstack/react-router';
import { FileSignature, Pencil, RotateCcw, Trash2 } from 'lucide-react';
import { Breadcrumbs, Button, ConfirmModal } from '../../shared/components/ui';
import { amountOf, dateOf, formatRelative, isClosedDeal, stageById } from './data/model';
import {
  useDeal,
  useDeleteDeal,
  useMoveDeal,
  usePipelines,
  useReopenDeal,
  useUpdateDeal,
} from './data/queries';
import {
  DealStatusBadge,
  DueDate,
  LoadFailed,
  Money,
  RepPicker,
  useMoney,
} from './components/common';
import { LogActivityPanel, SubjectTasks, Timeline } from './components/Activity';
import { DealFormModal, LoseDealModal, WinDealModal } from './components/DealModals';
import ContractModal from './components/ContractModal';

const DealDetail: React.FC = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const money = useMoney();
  const { id } = useParams({ strict: false }) as { id?: string };
  const query = useDeal(id);
  const { data: pipelines = [] } = usePipelines();
  const move = useMoveDeal();
  const reopen = useReopenDeal();
  const update = useUpdateDeal();
  const remove = useDeleteDeal();
  const [editing, setEditing] = useState(false);
  const [winning, setWinning] = useState(false);
  const [losing, setLosing] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [drafting, setDrafting] = useState(false);

  if (query.isError) {
    return (
      <LoadFailed
        title={t('crmSales.dealDetail.loadFailed')}
        onRetry={() => void query.refetch()}
      />
    );
  }
  const deal = query.data?.deal;
  if (!deal) {
    return (
      <p className="sales-empty" role="status">
        {t('crmSales.common.loading')}
      </p>
    );
  }

  const contract = query.data?.contract;
  const pipeline = pipelines.find((item) => item.id === deal.pipelineId);
  const stage = stageById(pipeline, deal.stageId);
  const closed = isClosedDeal(deal);
  const stages = pipeline?.stages ?? [];
  const currentPosition = stage?.position ?? -1;
  const weighted =
    stage?.kind === 'open' ? (amountOf(deal.amount) * (stage.probability ?? 0)) / 100 : undefined;
  const fields = {
    title: deal.title ?? '',
    contactName: deal.contactName ?? '',
    companyName: deal.companyName ?? '',
    value: deal.value ?? '0',
    billing: deal.billing ?? 'one_time',
    expectedCloseDate: deal.expectedCloseDate ?? '',
    source: deal.source ?? '',
    notes: deal.notes ?? '',
  };

  return (
    <div className="sales-page">
      <HelmetContainer title={`${deal.title} · rtylr`} />
      <Breadcrumbs
        items={[
          { label: t('crmSales.deals.title'), to: '/crm/sales' },
          { label: deal.title ?? '' },
        ]}
      />
      <header className="sales-header">
        <div>
          <h1 style={{ display: 'flex', alignItems: 'center', gap: 10, flexWrap: 'wrap' }}>
            {deal.title}
            <DealStatusBadge status={deal.status} />
          </h1>
          <p>
            <Money value={deal.value} billing={deal.billing} />
            {pipeline && ` · ${pipeline.name}`}
            {stage && ` · ${stage.name}`}
          </p>
        </div>
        <div className="sales-actions">
          {!closed && <Button onClick={() => setWinning(true)}>{t('crmSales.win.action')}</Button>}
          {!closed && (
            <Button variant="outline" onClick={() => setLosing(true)}>
              {t('crmSales.lose.action')}
            </Button>
          )}
          {closed && (
            <Button
              variant="outline"
              onClick={() => reopen.mutate({ id: deal.id ?? '' })}
              disabled={reopen.isPending}
            >
              <RotateCcw size={13} />
              {t('crmSales.dealDetail.reopen')}
            </Button>
          )}
          <Button variant="outline" onClick={() => setEditing(true)}>
            <Pencil size={13} />
            {t('crmSales.common.edit')}
          </Button>
          {!contract && (
            <Button
              variant="ghost"
              onClick={() => setDeleting(true)}
              aria-label={t('crmSales.common.delete')}
            >
              <Trash2 size={13} />
            </Button>
          )}
        </div>
      </header>

      {stages.length > 0 && (
        <div className="sales-stages" role="group" aria-label={t('crmSales.fields.stage')}>
          {stages
            .filter((item) => item.kind === 'open')
            .map((item) => (
              <button
                key={item.id}
                type="button"
                className={`sales-stage-step${item.id === deal.stageId ? ' is-current' : (item.position ?? 0) < currentPosition ? ' is-past' : ''}`}
                aria-current={item.id === deal.stageId ? 'step' : undefined}
                disabled={closed || item.id === deal.stageId || move.isPending}
                onClick={() => item.id && move.mutate({ id: deal.id ?? '', stageId: item.id })}
                title={t('crmSales.dealDetail.probability', { probability: item.probability ?? 0 })}
              >
                {item.name}
              </button>
            ))}
        </div>
      )}

      {deal.status === 'lost' && (
        <div className="sales-banner" role="status">
          {t('crmSales.dealDetail.lostBanner', { reason: deal.lostReason || '—' })}
        </div>
      )}

      <div className="sales-detail">
        <aside style={{ display: 'flex', flexDirection: 'column', gap: 16, minWidth: 0 }}>
          <section className="sales-panel" aria-label={t('crmSales.dealDetail.details')}>
            <h2>{t('crmSales.dealDetail.details')}</h2>
            <dl className="sales-fields">
              <dt>{t('crmSales.fields.owner')}</dt>
              <dd>
                <RepPicker
                  label={t('crmSales.fields.owner')}
                  value={deal.assignedTo}
                  onChange={(assignedTo) =>
                    update.mutate({ id: deal.id ?? '', deal: { ...fields, assignedTo } })
                  }
                />
              </dd>
              <dt>{t('crmSales.fields.value')}</dt>
              <dd>
                <Money value={deal.value} billing={deal.billing} />
              </dd>
              <dt>{t('crmSales.dealDetail.firstYear')}</dt>
              <dd>{money(deal.amount)}</dd>
              {weighted !== undefined && (
                <>
                  <dt>{t('crmSales.dealDetail.weighted')}</dt>
                  <dd>{money(weighted)}</dd>
                </>
              )}
              <dt>{t('crmSales.fields.expectedClose')}</dt>
              <dd>
                {closed ? deal.expectedCloseDate || '—' : <DueDate date={deal.expectedCloseDate} />}
              </dd>
              <dt>{t('crmSales.fields.contact')}</dt>
              <dd>
                {deal.customerId ? (
                  <Link className="sales-link" to={`/crm/customers/${deal.customerId}`}>
                    {deal.contactName || t('crmSales.dealDetail.customer')}
                  </Link>
                ) : (
                  deal.contactName || '—'
                )}
              </dd>
              <dt>{t('crmSales.fields.company')}</dt>
              <dd>{deal.companyName || '—'}</dd>
              <dt>{t('crmSales.fields.source')}</dt>
              <dd>{deal.source || '—'}</dd>
              {deal.leadId && (
                <>
                  <dt>{t('crmSales.dealDetail.fromLead')}</dt>
                  <dd>
                    <Link className="sales-link" to={`/crm/sales/leads/${deal.leadId}`}>
                      {t('crmSales.dealDetail.openLead')}
                    </Link>
                  </dd>
                </>
              )}
              <dt>{t('crmSales.dealDetail.inStage')}</dt>
              <dd>{formatRelative(dateOf(deal.stageChangedAt), i18n.language)}</dd>
              <dt>{t('crmSales.dealDetail.opened')}</dt>
              <dd>{formatRelative(dateOf(deal.createdAt), i18n.language)}</dd>
            </dl>
          </section>

          <section className="sales-panel" aria-label={t('crmSales.dealDetail.contract')}>
            <div className="sales-panel-head">
              <h2>{t('crmSales.dealDetail.contract')}</h2>
              <FileSignature size={14} aria-hidden="true" />
            </div>
            {contract ? (
              <dl className="sales-fields">
                <dt>{t('crmSales.fields.plan')}</dt>
                <dd>{contract.plan}</dd>
                <dt>{t('crmSales.fields.value')}</dt>
                <dd>
                  <Money value={contract.value} billing={contract.billing} />
                </dd>
                <dt>{t('crmSales.fields.status')}</dt>
                <dd>
                  {contract.signedAt
                    ? t('crmSales.contracts.signedOn', { date: contract.signedAt })
                    : t('crmSales.contracts.draft')}
                </dd>
                <dt />
                <dd>
                  <Link className="sales-link" to="/crm/sales/contracts">
                    {t('crmSales.dealDetail.openContracts')}
                  </Link>
                </dd>
              </dl>
            ) : deal.status === 'won' ? (
              <>
                <p className="sales-muted" style={{ margin: 0, fontSize: 13 }}>
                  {t('crmSales.dealDetail.noContract')}
                </p>
                <Button variant="outline" onClick={() => setDrafting(true)}>
                  {t('crmSales.dealDetail.draftContract')}
                </Button>
              </>
            ) : (
              <p className="sales-muted" style={{ margin: 0, fontSize: 13 }}>
                {t('crmSales.dealDetail.contractOnWin')}
              </p>
            )}
          </section>

          {deal.notes && (
            <section className="sales-panel" aria-label={t('crmSales.fields.notes')}>
              <h2>{t('crmSales.fields.notes')}</h2>
              <p
                style={{
                  margin: 0,
                  fontSize: 13,
                  whiteSpace: 'pre-wrap',
                  overflowWrap: 'anywhere',
                }}
              >
                {deal.notes}
              </p>
            </section>
          )}
        </aside>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16, minWidth: 0 }}>
          <LogActivityPanel subject={{ dealId: deal.id }} />
          <SubjectTasks subject={{ dealId: deal.id }} />
          <Timeline subject={{ dealId: deal.id }} />
        </div>
      </div>

      <DealFormModal open={editing} onClose={() => setEditing(false)} deal={deal} />
      <WinDealModal deal={deal} open={winning} onClose={() => setWinning(false)} />
      <LoseDealModal deal={deal} open={losing} onClose={() => setLosing(false)} />
      <ContractModal open={drafting} onClose={() => setDrafting(false)} deal={deal} />
      <ConfirmModal
        open={deleting}
        onClose={() => setDeleting(false)}
        onConfirm={() =>
          remove.mutate(deal.id ?? '', { onSuccess: () => navigate({ to: '/crm/sales' }) })
        }
        title={t('crmSales.dealDetail.deleteTitle')}
        description={t('crmSales.dealDetail.deleteHint')}
        confirmLabel={t('crmSales.common.delete')}
        variant="danger"
      />
    </div>
  );
};

export default DealDetail;
