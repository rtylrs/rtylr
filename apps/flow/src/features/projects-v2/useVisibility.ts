import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { projectV2 } from '../../shared/lib/clients';

export type ProjectVisibility = 'private' | 'public';

/**
 * Flips a project between private and public.
 *
 * Uses the dedicated setProjectVisibility RPC rather than updateProject:
 * UpdateProjectBody has no visibility field, so the general update cannot carry
 * one, and the dedicated route is where the server applies the project-admin
 * gate and the "My Tasks is always private" rule.
 */
export function useSetProjectVisibility() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
    mutationFn: async (vars: { id: string; visibility: ProjectVisibility }) => {
      await projectV2.setProjectVisibility({ id: vars.id, visibility: vars.visibility });
    },
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ['project-v2', vars.id] });
      qc.invalidateQueries({ queryKey: ['projects-v2'] });
      // 'projects' (the sidebar) is a separate query key mid-migration and
      // also renders a lock icon from this same visibility field.
      qc.invalidateQueries({ queryKey: ['projects'] });
      toast.success(t('flow.projectSettings.visibilitySaved'));
    },
    onError: () => toast.error(t('toasts.project.visibilityFailed')),
  });
}
