import { useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import type { ApiStatus } from '../../shared/services/flowApi';
import type { Task } from '../../shared/stores/flowStore';
import type { TaskFilters } from '../../shared/filters/taskFilters';
import { taskFiltersToViewFilters } from '../../shared/filters/viewFilters';
import { useCreateView } from './useViews';

/**
 * A project's filters as filters a saved view can hold.
 *
 * Views list tasks from every project, so the project itself becomes a filter.
 * A project's statuses are its own columns, which a view cannot name; each
 * chosen column is saved as the built-in status it stands for.
 */
export function projectFiltersForView(
  filters: TaskFilters,
  projectId: string,
  statuses: readonly ApiStatus[]
): TaskFilters {
  const legacy = new Set<string>();
  for (const id of filters.statuses) {
    const status = statuses.find((s) => s.id === id);
    if (status?.legacyStatus) legacy.add(status.legacyStatus as Task['status']);
  }
  return { ...filters, projects: [projectId], statuses: [...legacy], query: '' };
}

/** Saves filters as a new view and offers to open it. */
export function useSaveFiltersAsView(): {
  save: (name: string, filters: TaskFilters) => void;
  pending: boolean;
} {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const createView = useCreateView();
  return {
    pending: createView.isPending,
    save: (name, filters) =>
      createView.mutate(
        { name, icon: 'eye', filters: taskFiltersToViewFilters(filters) },
        {
          onSuccess: (result) => {
            const id = (result as { view?: { id?: string } } | undefined)?.view?.id;
            toast.success(t('flow.filters.savedAsView', { name }), {
              action: id
                ? {
                    label: t('flow.filters.openView'),
                    onClick: () => void navigate({ to: '/flow/views', search: { view: id } }),
                  }
                : undefined,
            });
          },
        }
      ),
  };
}
