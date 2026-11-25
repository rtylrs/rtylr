import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { flowApi, type ApiTaskTemplate } from '../../shared/services/flowApi';

/** A project's description templates by task type (T-380). */
export function useTaskTemplates(projectId: string | undefined) {
  return useQuery<ApiTaskTemplate[]>({
    queryKey: ['task-templates', projectId],
    queryFn: async () => (await unwrap(flowApi.listTaskTemplates(projectId!))).templates ?? [],
    enabled: projectId !== undefined && projectId !== '',
    staleTime: 5 * 60 * 1000,
  });
}

export function useSetTaskTemplate() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (vars: { projectId: string; taskType: string; body: string }) =>
      unwrap(flowApi.setTaskTemplate(vars.projectId, vars.taskType, vars.body)),
    onSuccess: (_data, vars) => {
      void qc.invalidateQueries({ queryKey: ['task-templates', vars.projectId] });
      toast.success(
        t(vars.body.trim() ? 'flowToasts.template.saved' : 'flowToasts.template.removed')
      );
    },
    onError: () => toast.error(t('flowToasts.template.saveFailed')),
  });
}
