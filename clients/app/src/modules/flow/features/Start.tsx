import { Plus, Shield } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Button } from '../../../shared/components/ui/button';
import { EmptyState } from '../../../shared/components/dash/EmptyState';
import { useUserStore } from '../../../shared/stores/userStore';
import { useFlowStore } from '../shared/stores/flowStore';
import { canFlowPermission } from '../shared/lib/roleUtils';

export default function Start() {
  const { t } = useTranslation();
  const access = useUserStore((state) => state.user?.effectiveAccess);
  const canCreateTask = canFlowPermission('flow:tasks:create', access);
  const canCreateProject = canFlowPermission('flow:projects:create', access);
  if (!canCreateTask && !canCreateProject)
    return (
      <EmptyState
        icon={Shield}
        title={t('flow.start.noActions')}
        description={t('flow.start.noActionsDescription')}
      />
    );
  return (
    <section className="px-6 py-6">
      <h1 className="text-flow-screen font-semibold">{t('flow.start.title')}</h1>
      <p className="mt-2 text-flow-body text-[var(--flow-text-muted)]">
        {t('flow.start.description')}
      </p>
      <div className="mt-6 flex flex-wrap gap-2">
        {canCreateTask && (
          <Button onClick={() => useFlowStore.getState().setQuickCreateOpen(true)}>
            <Plus size={16} />
            {t('flow.shell.newTask')}
          </Button>
        )}
        {canCreateProject && (
          <Button onClick={() => useFlowStore.getState().setProjectCreateOpen(true)}>
            <Plus size={16} />
            {t('flow.shell.newProject')}
          </Button>
        )}
      </div>
    </section>
  );
}
