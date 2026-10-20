import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { Link } from '@tanstack/react-router';
import { FileSignature, Plus } from 'lucide-react';
import { Button } from '../../shared/components/ui';
import { type Contract, formatDay } from './data/model';
import { useContracts } from './data/queries';
import { EmptyBlock, LoadFailed, Money } from './components/common';
import ContractModal from './components/ContractModal';

const PAGE = 25;
type Filter = 'all' | 'signed' | 'draft';

const Contracts: React.FC = () => {
  const { t, i18n } = useTranslation();
  const [filter, setFilter] = useState<Filter>('all');
  const [offset, setOffset] = useState(0);
  const [editing, setEditing] = useState<Contract | null>(null);
  const [creating, setCreating] = useState(false);
  const signed = filter === 'all' ? undefined : filter === 'signed';
  const contracts = useContracts(signed, offset, PAGE);
  const rows = contracts.data?.contracts ?? [];
  const total = contracts.data?.total ?? 0;
  const format = (value: number) => new Intl.NumberFormat(i18n.language).format(value);

  return (
    <div className="sales-page">
      <HelmetContainer title={`${t('crmSales.contracts.title')} · rtylr`} />
      <header className="sales-header">
        <div>
          <h1>{t('crmSales.contracts.title')}</h1>
          <p>{t('crmSales.contracts.subtitle')}</p>
        </div>
        <div className="sales-actions">
          <Button onClick={() => setCreating(true)}>
            <Plus size={13} />
            {t('crmSales.contracts.add')}
          </Button>
        </div>
      </header>

      <div className="sales-tabs" role="tablist" aria-label={t('crmSales.fields.status')}>
        {(['all', 'signed', 'draft'] as const).map((item) => (
          <button
            key={item}
            type="button"
            role="tab"
            className="sales-tab"
            aria-selected={filter === item}
            onClick={() => {
              setFilter(item);
              setOffset(0);
            }}
          >
            {t(`crmSales.contracts.tabs.${item}`)}
          </button>
        ))}
      </div>

      <div className="sales-card">
        {contracts.isError && rows.length === 0 ? (
          <LoadFailed
            title={t('crmSales.contracts.loadFailed')}
            onRetry={() => void contracts.refetch()}
          />
        ) : contracts.isPending ? (
          <p className="sales-empty" role="status">
            {t('crmSales.common.loading')}
          </p>
        ) : rows.length === 0 ? (
          <EmptyBlock
            icon={FileSignature}
            title={t('crmSales.contracts.emptyTitle')}
            description={t('crmSales.contracts.emptyHint')}
          />
        ) : (
          <>
            <div className="sales-table-wrap">
              <table className="sales-table">
                <thead>
                  <tr>
                    <th>{t('crmSales.fields.plan')}</th>
                    <th>{t('crmSales.fields.value')}</th>
                    <th>{t('crmSales.contracts.term')}</th>
                    <th>{t('crmSales.fields.status')}</th>
                    <th>{t('crmSales.fields.deal')}</th>
                  </tr>
                </thead>
                <tbody>
                  {rows.map((contract) => (
                    <tr key={contract.id} onClick={() => setEditing(contract)}>
                      <td className="sales-cell-main">{contract.plan}</td>
                      <td>
                        <Money value={contract.value} billing={contract.billing} />
                      </td>
                      <td className="sales-muted">
                        {contract.startDate
                          ? `${formatDay(contract.startDate, i18n.language)} – ${contract.endDate ? formatDay(contract.endDate, i18n.language) : t('crmSales.contracts.openEnded')}`
                          : '—'}
                      </td>
                      <td>
                        {contract.signedAt ? (
                          <span className="sales-badge" data-tone="green">
                            {t('crmSales.contracts.signedOn', {
                              date: formatDay(contract.signedAt, i18n.language),
                            })}
                          </span>
                        ) : (
                          <span className="sales-badge" data-tone="amber">
                            {t('crmSales.contracts.draft')}
                          </span>
                        )}
                      </td>
                      <td onClick={(event) => event.stopPropagation()}>
                        {contract.dealId ? (
                          <Link className="sales-link" to={`/crm/sales/deals/${contract.dealId}`}>
                            {t('crmSales.contracts.openDeal')}
                          </Link>
                        ) : (
                          <span className="sales-muted">—</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <div className="sales-footer">
              <span>
                {t('crmSales.contracts.range', {
                  from: format(offset + 1),
                  to: format(offset + rows.length),
                  total: format(total),
                })}
              </span>
              <span className="sales-actions">
                <Button
                  variant="outline"
                  disabled={offset === 0}
                  onClick={() => setOffset(Math.max(0, offset - PAGE))}
                >
                  {t('crmSales.common.previous')}
                </Button>
                <Button
                  variant="outline"
                  disabled={offset + rows.length >= total}
                  onClick={() => setOffset(offset + PAGE)}
                >
                  {t('crmSales.common.next')}
                </Button>
              </span>
            </div>
          </>
        )}
      </div>

      <ContractModal open={creating} onClose={() => setCreating(false)} />
      <ContractModal
        open={editing !== null}
        onClose={() => setEditing(null)}
        contract={editing ?? undefined}
      />
    </div>
  );
};

export default Contracts;
