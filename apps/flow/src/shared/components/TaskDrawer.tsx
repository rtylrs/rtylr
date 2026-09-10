import React, { Suspense, useEffect } from 'react';
import { AnimatePresence, motion } from 'motion/react';
import { X } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useRouterState } from '@tanstack/react-router';
import { useMediaQuery } from '@voxire/utils';
/**
 * Lazy, for two reasons.
 *
 * Task detail is a large component with a wide dependency tree, and nothing
 * needs it until a task is actually opened. The router already lazy-loads its
 * route components for the same reason.
 *
 * It also keeps that tree out of the layout's own tests. Importing it eagerly
 * made `FlowLayout.test.tsx` fail at module load on a missing `getCookie`
 * export in its `@voxire/auth` mock, because the layout renders this drawer and
 * the drawer reached all the way through to whatever task detail imports. A
 * layout test should not have to mock task detail's dependencies to render a
 * sidebar.
 */
const TaskDetail = React.lazy(() => import('../../features/tasks/TaskDetail'));

/** Below this the drawer is a full-page overlay instead of a side panel. */
const WIDE = '(min-width: 768px)';

/**
 * Task detail over the list you opened it from, rather than instead of it.
 *
 * Opening a task used to replace the board, so the context you were reading
 * disappeared and coming back re-fetched and re-scrolled it (rtylr#1079, S-12).
 *
 * The open task lives in the URL as `?openTask=<id>`, which means the back
 * button closes the drawer, a refresh keeps it open, and the address bar can be
 * shared. `/task/$id` is untouched and still serves a direct link, so a URL
 * someone already has keeps working and anyone landing cold gets a whole page
 * rather than a panel floating over nothing.
 *
 * Below 768px it is a full-page overlay: a 720px panel on a 390px screen is the
 * same mistake as a seven-column month grid.
 */
export const TaskDrawer: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const isWide = useMediaQuery(WIDE);
  const openTask = useRouterState({
    select: (state) => (state.location.search as { openTask?: string }).openTask,
  });

  const close = React.useCallback(() => {
    // Only `openTask` is dropped. Everything else in the query belongs to the
    // screen underneath, and closing a drawer must not reset its filters.
    void navigate({
      to: '.',
      search: (previous: Record<string, unknown>) => {
        const { openTask: _openTask, ...rest } = previous;
        return rest;
      },
      replace: true,
    });
  }, [navigate]);

  useEffect(() => {
    if (!openTask) return;
    const onKeyDown = (event: KeyboardEvent): void => {
      if (event.key === 'Escape') close();
    };
    document.addEventListener('keydown', onKeyDown);
    return () => document.removeEventListener('keydown', onKeyDown);
  }, [openTask, close]);

  return (
    <AnimatePresence>
      {openTask && (
        <>
          {/* The scrim is a button so it is reachable and announced, rather than
              a div with an onClick that only a mouse can use. */}
          <motion.button
            type="button"
            aria-label={t('common.close')}
            onClick={close}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.15 }}
            className="fixed inset-0 z-40 cursor-default border-0"
            style={{ background: 'rgba(0,0,0,0.45)' }}
          />
          <motion.aside
            role="dialog"
            aria-modal="true"
            aria-label={t('flow.taskDetail.drawerLabel')}
            initial={isWide ? { x: '100%' } : { y: '100%' }}
            animate={isWide ? { x: 0 } : { y: 0 }}
            exit={isWide ? { x: '100%' } : { y: '100%' }}
            transition={{ duration: 0.2, ease: [0.4, 0, 0.2, 1] }}
            className="fixed inset-y-0 end-0 z-50 flex w-full flex-col md:max-w-[720px]"
            style={{
              background: 'var(--flow-bg)',
              borderInlineStart: '1px solid var(--flow-border)',
            }}
          >
            <div
              className="flex shrink-0 items-center justify-end px-3 py-2"
              style={{ borderBottom: '1px solid var(--flow-border)' }}
            >
              <button
                type="button"
                onClick={close}
                aria-label={t('common.close')}
                className="flow-focus flex h-11 w-11 items-center justify-center rounded-md hover:bg-[var(--flow-hover)] md:h-8 md:w-8"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                <X size={16} />
              </button>
            </div>
            <div className="min-h-0 flex-1 overflow-y-auto">
              <Suspense
                fallback={
                  <p className="p-6 text-meta" style={{ color: 'var(--flow-text-muted)' }}>
                    {t('flow.taskDetail.loading')}
                  </p>
                }
              >
                <TaskDetail taskId={openTask} />
              </Suspense>
            </div>
          </motion.aside>
        </>
      )}
    </AnimatePresence>
  );
};

export default TaskDrawer;
