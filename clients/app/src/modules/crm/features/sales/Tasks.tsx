import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { ListTodo } from 'lucide-react';
import { useUserStore } from '../../../../shared/stores/userStore';
import type { FollowUpWindow } from './data/model';
import { useReps, useTaskCounts, useTaskList } from './data/queries';
import {
  EmptyBlock,
  ListFooter,
  LoadFailed,
  RepFilterSelect,
  StatTile,
  useRepNames,
} from './components/common';
import { TaskRow } from './components/Activity';

type Tab = 'due' | 'upcoming' | 'open' | 'completed';
const TABS: Tab[] = ['due', 'upcoming', 'open', 'completed'];

function windowOf(tab: Tab): FollowUpWindow | '' {
  if (tab === 'due') return 'due';
  if (tab === 'upcoming') return 'upcoming';
  return '';
}

const Tasks: React.FC = () => {
  const { t, i18n } = useTranslation();
  const userId = useUserStore((state) => state.user?.id);
  const { data: reps } = useReps();
  const names = useRepNames();
  const [tab, setTab] = useState<Tab>('due');
  // Default to the caller's own work when they can be assigned any.
  const [rep, setRep] = useState<string | null>(null);
  const me = userId ? String(userId) : undefined;
  const effectiveRep = rep ?? (me && reps?.some((item) => item.id === me) ? me : 'all');
  const counts = useTaskCounts(effectiveRep);
  const tasks = useTaskList({
    scope: tab === 'completed' ? 'completed' : 'open',
    rep: effectiveRep,
    due: windowOf(tab),
  });
  const rows = tasks.data?.pages.flatMap((page) => page.tasks ?? []) ?? [];
  const format = (value?: number) => new Intl.NumberFormat(i18n.language).format(value ?? 0);

  return (
    <div className="sales-page">
      <HelmetContainer title={`${t('crmSales.tasks.title')} · rtylr`} />
      <header className="sales-header">
        <div>
          <h1>{t('crmSales.tasks.title')}</h1>
          <p>{t('crmSales.tasks.subtitle')}</p>
        </div>
        <RepFilterSelect
          value={effectiveRep}
          onChange={setRep}
          label={t('crmSales.fields.assignee')}
        />
      </header>

      <div className="sales-stats">
        <StatTile
          label={t('crmSales.tasks.counts.overdue')}
          value={format(counts.data?.overdue)}
          warning={(counts.data?.overdue ?? 0) > 0}
        />
        <StatTile label={t('crmSales.tasks.counts.today')} value={format(counts.data?.dueToday)} />
        <StatTile
          label={t('crmSales.tasks.counts.upcoming')}
          value={format(counts.data?.upcoming)}
        />
      </div>

      <div className="sales-tabs" role="tablist" aria-label={t('crmSales.tasks.title')}>
        {TABS.map((item) => (
          <button
            key={item}
            type="button"
            role="tab"
            className="sales-tab"
            aria-selected={tab === item}
            onClick={() => setTab(item)}
          >
            {t(`crmSales.tasks.tabs.${item}`)}
          </button>
        ))}
      </div>

      <div className="sales-card">
        {tasks.isError && rows.length === 0 ? (
          <LoadFailed title={t('crmSales.tasks.loadFailed')} onRetry={() => void tasks.refetch()} />
        ) : tasks.isPending ? (
          <p className="sales-empty" role="status">
            {t('crmSales.common.loading')}
          </p>
        ) : rows.length === 0 ? (
          <EmptyBlock
            icon={ListTodo}
            title={t(`crmSales.tasks.empty.${tab}`)}
            description={t('crmSales.tasks.emptyHint')}
          />
        ) : (
          <>
            {rows.map((task) => (
              <TaskRow
                key={task.id}
                task={task}
                showSubject
                repName={names.get(task.assignedTo ?? '')}
              />
            ))}
            <ListFooter shown={rows.length} pages={tasks} />
          </>
        )}
      </div>
    </div>
  );
};

export default Tasks;
