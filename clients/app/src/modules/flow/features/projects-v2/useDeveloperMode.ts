import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { projectV2 } from '../../shared/lib/clients';
import { useProject } from './useAllProjects';
import { toPointsScale, type PointsScale } from '../../shared/lib/storyPoints';

/**
 * Whether a project has developer features on. Every developer feature (code
 * links, the Development section, dev fields) gates on this one flag, so a
 * project with it off looks exactly as it did before developer projects.
 */
export function useIsDeveloperProject(projectId: string | undefined): boolean {
  const { data: project } = useProject(projectId);
  return project?.developerMode === true;
}

export function useSetProjectDeveloperMode() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
    mutationFn: async (vars: { id: string; enabled: boolean }) => {
      await projectV2.setProjectDeveloperMode({ id: vars.id, enabled: vars.enabled });
    },
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ['project-v2', vars.id] });
      qc.invalidateQueries({ queryKey: ['projects-v2'] });
      toast.success(
        t(
          vars.enabled
            ? 'flow.projectSettings.developerModeOn'
            : 'flow.projectSettings.developerModeOff'
        )
      );
    },
    onError: () => toast.error(t('flowToasts.project.developerModeFailed')),
  });
}

/** The scale the project's story points use (T-381); Fibonacci when unset. */
export function useProjectPointsScale(projectId: string | undefined): PointsScale {
  const { data: project } = useProject(projectId);
  return toPointsScale(project?.pointsScale);
}

export function useSetProjectPointsScale() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
    mutationFn: async (vars: { id: string; scale: PointsScale }) => {
      await projectV2.setProjectPointsScale({ id: vars.id, scale: vars.scale });
    },
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ['project-v2', vars.id] });
      qc.invalidateQueries({ queryKey: ['projects-v2'] });
      toast.success(t('flow.projectSettings.pointsScaleSaved'));
    },
    onError: () => toast.error(t('flowToasts.project.pointsScaleFailed')),
  });
}
