import { useMemo } from 'react';
import { useFlowStore } from '../../shared/stores/flowStore';
import { apiProjectToProject } from '../../shared/lib/projectMapper';
import { useProjects } from './useProjects';

export function useCurrentProject(workspaceId: string | undefined, projectId: string | undefined) {
  const storeProject = useFlowStore((s) =>
    projectId ? s.projects.find((p) => p.id === projectId) : undefined
  );
  const { data: apiProjects = [], isLoading } = useProjects();

  const apiProject = useMemo(() => {
    if (!projectId) return undefined;
    return apiProjects.find((p) => String(p.id) === projectId);
  }, [apiProjects, projectId]);

  return {
    project: storeProject ?? (apiProject ? apiProjectToProject(apiProject) : undefined),
    isLoading,
  };
}
