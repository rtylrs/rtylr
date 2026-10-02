import React, { Suspense, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useRouterState } from '@tanstack/react-router';
import { Dialog, DialogContent, DialogTitle } from '@voxire/ui';
import { useUserStore } from '../stores/userStore';
import { canFlowPermission } from '../lib/roleUtils';

const TaskDetail = React.lazy(() => import('../../features/tasks/TaskDetail'));

export const TaskDrawer: React.FC = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const access = useUserStore((state) => state.user?.effectiveAccess);
  const openTask = useRouterState({
    select: (state) => (state.location.search as { openTask?: string }).openTask,
  });
  const close = useCallback(() => {
    void navigate({
      to: '.',
      search: (previous: Record<string, unknown>) => {
        const { openTask: _openTask, ...rest } = previous;
        return rest;
      },
      replace: true,
    });
  }, [navigate]);
  return (
    <Dialog
      open={Boolean(openTask && canFlowPermission('flow:tasks:view', access))}
      onOpenChange={(open) => {
        if (!open) close();
      }}
    >
      <DialogContent
        className="portal-flow-dialog portal-flow-task-drawer"
        style={{ translate: 'none' }}
        dir={i18n.dir()}
        closeLabel={t('flowCommon.close')}
        aria-describedby={undefined}
      >
        <DialogTitle className="flow-task-drawer-title px-6 py-4 text-flow-body font-medium">
          {t('flow.taskDetail.drawerLabel')}
        </DialogTitle>
        <div className="min-h-0 flex-1 overflow-y-auto">
          <Suspense
            fallback={
              <p className="p-6 text-flow-meta text-[var(--flow-text-muted)]">
                {t('flow.taskDetail.loading')}
              </p>
            }
          >
            {openTask && <TaskDetail taskId={openTask} />}
          </Suspense>
        </div>
      </DialogContent>
    </Dialog>
  );
};

export default TaskDrawer;
