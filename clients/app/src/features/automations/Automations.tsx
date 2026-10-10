import { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Plus, Play, Pause, Trash2 } from 'lucide-react';
import { AutomationsIcon } from '@voxire/ui';
import { EmptyState } from '@voxire/ui';
import {
  useListAutomations,
  useUpdateAutomation,
  useDeleteAutomation,
} from '../../modules/crm/shared/hooks/queries/useAutomations';
import {
  PageHeader,
  Button,
  SkeletonList,
  ConfirmModal,
} from '../../modules/crm/shared/components/ui';
import './automations.css';

export default function Automations() {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const { data: automations = [], isLoading, isError, refetch } = useListAutomations();
  const update = useUpdateAutomation();
  const remove = useDeleteAutomation();
  const [deleteId, setDeleteId] = useState<number | null>(null);
  const [query, setQuery] = useState('');
  const [status, setStatus] = useState('all');
  const visible = automations.filter(
    (item) =>
      item.name.toLocaleLowerCase().includes(query.toLocaleLowerCase()) &&
      (status === 'all' || item.status === status)
  );
  const deleteTarget = automations.find((item) => item.id === deleteId);

  return (
    <>
      <HelmetContainer title={`${t('common.nav.automations')} · rtylr`} />
      <div className="portal-automations">
        <PageHeader
          title={t('common.nav.automations')}
          subtitle={t('automations.subtitle')}
          actions={
            <Button onClick={() => navigate({ to: '/automations/new' })}>
              <Plus size={16} />
              {t('crmNav.newAutomation')}
            </Button>
          }
        />
        <div className="automation-filters">
          <input
            type="search"
            aria-label={t('automations.search')}
            placeholder={t('automations.search')}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
          <select
            aria-label={t('automations.status')}
            value={status}
            onChange={(e) => setStatus(e.target.value)}
          >
            <option value="all">{t('automations.all')}</option>
            {['active', 'paused', 'draft'].map((value) => (
              <option key={value} value={value}>
                {t(`automations.${value}`)}
              </option>
            ))}
          </select>
        </div>
        {isLoading ? (
          <SkeletonList count={4} />
        ) : isError ? (
          <div role="alert" className="automation-empty">
            <p>{t('crmErrors.generic')}</p>
            <Button variant="outline" onClick={() => void refetch()}>
              {t('crm.actions.refresh')}
            </Button>
          </div>
        ) : !automations.length ? (
          <EmptyState
            icon={AutomationsIcon}
            title={t('crm.automations.emptyTitle')}
            description={t('automations.emptyDescription')}
            action={{
              label: t('crm.automations.createOne'),
              onClick: () => navigate({ to: '/automations/new' }),
            }}
          />
        ) : !visible.length ? (
          <p className="automation-empty">{t('automations.noMatches')}</p>
        ) : (
          <ul className="automation-list">
            {visible.map((item) => (
              <li key={item.id} className="automation-row">
                <button
                  type="button"
                  className="automation-open"
                  onClick={() => navigate({ to: '/automations/new', search: { id: item.id } })}
                >
                  <AutomationsIcon size={19} />
                  <span>
                    <strong>{item.name}</strong>
                    <small>
                      {t(`crm.automationTriggers.${item.triggerType}`, item.trigger)} ·{' '}
                      {t('automations.runs', { count: item.runCount ?? item.enrolled })}
                    </small>
                  </span>
                </button>
                <span className="automation-status">{t(`automations.${item.status}`)}</span>
                <div className="automation-row-actions">
                  {item.status !== 'draft' && (
                    <Button
                      variant="ghost"
                      disabled={update.isPending}
                      aria-label={t(
                        item.status === 'active' ? 'automations.pause' : 'automations.resume'
                      )}
                      onClick={() =>
                        update.mutate({
                          id: item.id,
                          data: {
                            isActive: item.status !== 'active',
                            status: item.status === 'active' ? 'paused' : 'active',
                          },
                        })
                      }
                    >
                      {item.status === 'active' ? <Pause size={16} /> : <Play size={16} />}
                    </Button>
                  )}
                  <Button
                    variant="ghost"
                    disabled={remove.isPending}
                    aria-label={t('crm.automations.deleteAutomation')}
                    onClick={() => setDeleteId(item.id)}
                  >
                    <Trash2 size={16} />
                  </Button>
                </div>
              </li>
            ))}
          </ul>
        )}
        <ConfirmModal
          open={deleteId !== null}
          onClose={() => !remove.isPending && setDeleteId(null)}
          onConfirm={() => {
            if (deleteId === null || remove.isPending) return;
            remove.mutate(deleteId, {
              onSuccess: () => setDeleteId(null),
              onError: () => toast.error(t('crm.automations.deleteFailed')),
            });
          }}
          title={t('crm.automations.deleteConfirmTitle')}
          description={t('automations.deleteDescription', { name: deleteTarget?.name ?? '' })}
          confirmLabel={t('crmCommon.delete')}
          variant="danger"
        />
      </div>
    </>
  );
}
