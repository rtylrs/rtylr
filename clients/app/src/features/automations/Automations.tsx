import { useMemo, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import {
  Activity,
  ArrowDown,
  ArrowRight,
  Check,
  CircleDot,
  Clock3,
  Mail,
  Pause,
  Play,
  Plus,
  Search,
  ShoppingBag,
  Sparkles,
  Trash2,
  Workflow,
  Zap,
} from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import {
  useListAutomations,
  useUpdateAutomation,
  useDeleteAutomation,
  type Automation,
} from '../../modules/crm/shared/hooks/queries/useAutomations';
import { Button, ConfirmModal, SkeletonList } from '../../modules/crm/shared/components/ui';
import './automations.css';

type AutomationFilter = 'all' | 'active' | 'paused' | 'draft';

function runCount(automation: Automation): number {
  return automation.runCount ?? automation.enrolled ?? 0;
}

function actionCount(automation: Automation): number {
  if (automation.actions) return automation.actions.length;
  return Math.max(0, (automation.nodes?.length ?? 1) - 1);
}

function mostRecentRun(automations: Automation[]): string | null {
  return automations.reduce<string | null>((latest, automation) => {
    if (!automation.lastRunAt || Number.isNaN(Date.parse(automation.lastRunAt))) return latest;
    if (!latest || Date.parse(automation.lastRunAt) > Date.parse(latest)) {
      return automation.lastRunAt;
    }
    return latest;
  }, null);
}

export default function Automations() {
  const navigate = useNavigate();
  const { t, i18n } = useTranslation();
  const { data: automations = [], isLoading, isError, refetch } = useListAutomations();
  const update = useUpdateAutomation();
  const remove = useDeleteAutomation();
  const [deleteId, setDeleteId] = useState<number | null>(null);
  const [query, setQuery] = useState('');
  const [filter, setFilter] = useState<AutomationFilter>('all');
  const deleteTarget = automations.find((item) => item.id === deleteId);

  const visible = useMemo(
    () =>
      automations.filter(
        (item) =>
          item.name.toLocaleLowerCase().includes(query.toLocaleLowerCase()) &&
          (filter === 'all' || item.status === filter)
      ),
    [automations, filter, query]
  );
  const activeCount = automations.filter((item) => item.status === 'active').length;
  const totalRuns = automations.reduce((total, item) => total + runCount(item), 0);
  const latestRun = mostRecentRun(automations);
  const latestRunLabel = latestRun
    ? new Intl.DateTimeFormat(i18n.language, { dateStyle: 'medium', timeStyle: 'short' }).format(
        new Date(latestRun)
      )
    : t('automations.noRunsYet');

  const openBuilder = () => navigate({ to: '/automations/new' });

  return (
    <>
      <HelmetContainer title={`${t('common.nav.automations')} · rtylr`} />
      <main className="automation-studio">
        <header className="automation-studio-header">
          <div>
            <div className="automation-eyebrow">
              <span className="automation-eyebrow-mark">
                <Zap size={12} fill="currentColor" />
              </span>
              {t('automations.studioLabel')}
            </div>
            <h1>{t('common.nav.automations')}</h1>
            <p>{t('automations.subtitle')}</p>
          </div>
          <Button onClick={openBuilder}>
            <Plus size={16} />
            {t('automations.newWorkflow')}
          </Button>
        </header>

        <section className="automation-hero" aria-labelledby="automation-hero-title">
          <div className="automation-hero-copy">
            <span className="automation-hero-overline">
              <Sparkles size={14} /> {t('automations.heroEyebrow')}
            </span>
            <h2 id="automation-hero-title">{t('automations.heroTitle')}</h2>
            <p>{t('automations.heroDescription')}</p>
            <Button onClick={openBuilder} className="automation-hero-button">
              {t('automations.buildWorkflow')} <ArrowRight size={16} />
            </Button>
            <div className="automation-hero-footnote">
              <span className="automation-live-dot" />
              {t('automations.heroFootnote')}
            </div>
          </div>

          <div
            className="automation-preview"
            role="img"
            aria-label={t('automations.workflowPreview')}
          >
            <div className="automation-preview-toolbar">
              <span>
                <span className="automation-window-dot" />
                <span className="automation-window-dot" />
                <span className="automation-window-dot" />
              </span>
              <span>{t('automations.workflowPreview')}</span>
              <span className="automation-preview-status">
                <CircleDot size={11} /> {t('automations.previewExample')}
              </span>
            </div>
            <div className="automation-preview-canvas">
              <div className="automation-preview-node automation-preview-trigger">
                <span className="automation-node-icon">
                  <ShoppingBag size={16} />
                </span>
                <span>
                  <small>{t('automations.triggerLabel')}</small>
                  <strong>{t('crm.automationTriggers.order_created')}</strong>
                </span>
                <Check size={14} className="automation-node-check" />
              </div>
              <div className="automation-preview-connector">
                <span />
                <ArrowDown size={14} />
              </div>
              <div className="automation-preview-node automation-preview-action">
                <span className="automation-node-icon">
                  <Mail size={16} />
                </span>
                <span>
                  <small>{t('automations.actionLabel', { number: 1 })}</small>
                  <strong>{t('crm.automationActions.send_email')}</strong>
                </span>
                <span className="automation-node-menu">···</span>
              </div>
              <button type="button" className="automation-preview-add" onClick={openBuilder}>
                <Plus size={15} /> {t('automations.addStep')}
              </button>
            </div>
            <div className="automation-preview-footer">
              <span>
                <Activity size={14} /> {t('automations.workflowShape')}
              </span>
              <span>
                01 <span className="automation-preview-footer-dim">/ 02</span>
              </span>
            </div>
          </div>
          <div className="automation-orbit automation-orbit-one" />
          <div className="automation-orbit automation-orbit-two" />
        </section>

        <section className="automation-stat-grid" aria-label={t('automations.workspaceOverview')}>
          <article className="automation-stat-card">
            <span className="automation-stat-icon automation-stat-icon-cyan">
              <Workflow size={17} />
            </span>
            <span className="automation-stat-label">{t('automations.totalWorkflows')}</span>
            <strong>{automations.length}</strong>
            <small>{t('automations.inWorkspace')}</small>
          </article>
          <article className="automation-stat-card">
            <span className="automation-stat-icon automation-stat-icon-lime">
              <Zap size={17} />
            </span>
            <span className="automation-stat-label">{t('automations.activeWorkflows')}</span>
            <strong>{activeCount}</strong>
            <small>{t('automations.runningNow')}</small>
          </article>
          <article className="automation-stat-card">
            <span className="automation-stat-icon automation-stat-icon-amber">
              <Activity size={17} />
            </span>
            <span className="automation-stat-label">{t('automations.totalRuns')}</span>
            <strong>{totalRuns.toLocaleString(i18n.language)}</strong>
            <small>{t('automations.latestRun', { date: latestRunLabel })}</small>
          </article>
        </section>

        <section className="automation-workflows" aria-labelledby="automation-workflows-title">
          <div className="automation-workflows-heading">
            <div>
              <div className="automation-section-kicker">
                <span /> {t('automations.workspace')}
              </div>
              <h2 id="automation-workflows-title">{t('automations.yourWorkflows')}</h2>
              <p>{t('automations.workflowListDescription')}</p>
            </div>
            <span className="automation-workflow-total">
              {t('automations.workflowCount', { count: automations.length })}
            </span>
          </div>

          <div className="automation-workflow-controls">
            <div
              className="automation-filter-tabs"
              role="group"
              aria-label={t('automations.filterByStatus')}
            >
              {(['all', 'active', 'paused', 'draft'] as const).map((value) => (
                <button
                  type="button"
                  key={value}
                  className={filter === value ? 'is-selected' : ''}
                  aria-pressed={filter === value}
                  onClick={() => setFilter(value)}
                >
                  {t(`automations.filter.${value}`)}
                  {value === 'all' && <span>{automations.length}</span>}
                </button>
              ))}
            </div>
            <label className="automation-search">
              <Search size={16} />
              <span className="sr-only">{t('automations.search')}</span>
              <input
                type="search"
                aria-label={t('automations.search')}
                placeholder={t('automations.search')}
                value={query}
                onChange={(event) => setQuery(event.target.value)}
              />
            </label>
          </div>

          {isLoading ? (
            <SkeletonList count={4} />
          ) : isError ? (
            <div role="alert" className="automation-load-error">
              <p>{t('crmErrors.generic')}</p>
              <Button variant="outline" onClick={() => void refetch()}>
                {t('crm.actions.refresh')}
              </Button>
            </div>
          ) : !automations.length ? (
            <div className="automation-empty-studio">
              <div className="automation-empty-graphic" aria-hidden="true">
                <div className="automation-empty-node">
                  <ShoppingBag size={17} />
                  <span>{t('crm.automationTriggers.order_created')}</span>
                </div>
                <div className="automation-empty-link">
                  <span />
                </div>
                <div className="automation-empty-node">
                  <Mail size={17} />
                  <span>{t('crm.automationActions.send_email')}</span>
                </div>
                <div className="automation-empty-spark">
                  <Sparkles size={19} />
                </div>
              </div>
              <div className="automation-empty-copy">
                <span>{t('automations.readyWhenYouAre')}</span>
                <h3>{t('crm.automations.emptyTitle')}</h3>
                <p>{t('automations.emptyDescription')}</p>
                <Button onClick={openBuilder}>
                  <Plus size={15} /> {t('automations.startBuilding')}
                </Button>
              </div>
            </div>
          ) : !visible.length ? (
            <p className="automation-no-matches">{t('automations.noMatches')}</p>
          ) : (
            <ul className="automation-list">
              {visible.map((item) => (
                <li key={item.id} className="automation-row">
                  <button
                    type="button"
                    className="automation-open"
                    onClick={() => navigate({ to: '/automations/new', search: { id: item.id } })}
                  >
                    <span className="automation-row-icon">
                      <Workflow size={19} />
                    </span>
                    <span className="automation-row-main">
                      <strong>{item.name}</strong>
                      <span className="automation-row-flow">
                        <span>{t(`crm.automationTriggers.${item.triggerType}`, item.trigger)}</span>
                        <ArrowRight size={13} />
                        <span>{t('automations.actionCount', { count: actionCount(item) })}</span>
                      </span>
                    </span>
                    <span className={`automation-status automation-status-${item.status}`}>
                      <span />
                      {t(`automations.${item.status}`)}
                    </span>
                    <span className="automation-run-summary">
                      <Activity size={14} /> {t('automations.runs', { count: runCount(item) })}
                    </span>
                    <span className="automation-row-last-run">
                      <Clock3 size={14} />{' '}
                      {item.lastRunAt
                        ? new Intl.DateTimeFormat(i18n.language, { dateStyle: 'medium' }).format(
                            new Date(item.lastRunAt)
                          )
                        : t('automations.noRunsYet')}
                    </span>
                  </button>
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
          {!isLoading && !isError && automations.length > 0 && (
            <div className="automation-list-footer">
              <span>{t('automations.latestRun', { date: latestRunLabel })}</span>
              <button type="button" onClick={openBuilder}>
                {t('automations.newWorkflow')} <Plus size={14} />
              </button>
            </div>
          )}
        </section>

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
      </main>
    </>
  );
}
