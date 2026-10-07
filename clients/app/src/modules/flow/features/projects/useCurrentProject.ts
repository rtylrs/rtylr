import { useMemo } from 'react';
import { useFlowStore } from '../../shared/stores/flowStore';
import { apiProjectToProject } from '../../shared/lib/projectMapper';
import { useProjects } from './useProjects';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';

export function useCurrentProject(projectId: string | undefined) {
  const storeProject = useFlowStore((s) =>
    projectId ? s.projects.find((p) => p.id === projectId) : undefined
  );
  const projectsQuery = useProjects();
  const { data: apiProjects = [] } = projectsQuery;
  const loadFailed = hasQueryFailed(projectsQuery);

  const apiProject = useMemo(() => {
    if (!projectId) return undefined;
    return apiProjects.find((p) => String(p.id) === projectId);
  }, [apiProjects, projectId]);

  return {
    project: storeProject ?? (apiProject ? apiProjectToProject(apiProject) : undefined),
    isLoading: projectsQuery.isPending && !loadFailed,
    // Callers need to tell "no such project" from "the list never arrived".
    // Without this a failed request left `project` undefined once isLoading
    // cleared, and every consumer said "Project not found." for a 500.
    loadFailed,
    refetch: projectsQuery.refetch,
  };
}
