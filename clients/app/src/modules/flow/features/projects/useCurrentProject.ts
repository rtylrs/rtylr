import { useMemo } from 'react';
import { useFlowStore } from '../../shared/stores/flowStore';
import { apiProjectToProject } from '../../shared/lib/projectMapper';
import type { ApiProject } from '../../shared/services/flowApi';
import { useProjects } from './useProjects';
import { useProject } from '../projects-v2/useAllProjects';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';

/**
 * A refusal or a missing project from the direct lookup both mean the caller
 * cannot open it, so either reads as "not found". Anything else is a failed
 * load the user can retry.
 */
function isNotOpenable(error: unknown): boolean {
  if (!error || typeof error !== 'object') return false;
  const candidate = error as {
    statusCode?: unknown;
    status?: unknown;
    response?: { status?: unknown };
  };
  const status = candidate.statusCode ?? candidate.status ?? candidate.response?.status;
  return status === 403 || status === 404;
}

export function useCurrentProject(projectId: string | undefined) {
  const storeProject = useFlowStore((s) =>
    projectId ? s.projects.find((p) => p.id === projectId) : undefined
  );
  const projectsQuery = useProjects();
  const { data: apiProjects = [] } = projectsQuery;
  const listFailed = hasQueryFailed(projectsQuery);

  const apiProject = useMemo(() => {
    if (!projectId) return undefined;
    return apiProjects.find((p) => String(p.id) === projectId);
  }, [apiProjects, projectId]);

  // The sidebar list is curated (ruling P8), so a project someone may open
  // without being listed, such as a business admin following a link to a
  // private project, is not in it. Only once the list has arrived without it
  // is the project asked for directly.
  const listed = storeProject !== undefined || apiProject !== undefined;
  const needsDirect = !!projectId && projectsQuery.isSuccess && !listed;
  const directQuery = useProject(needsDirect ? projectId : undefined);
  const directFailed = hasQueryFailed(directQuery) && !isNotOpenable(directQuery.error);

  const directProject = needsDirect ? directQuery.data : undefined;

  return {
    project:
      storeProject ??
      (apiProject ? apiProjectToProject(apiProject) : undefined) ??
      (directProject ? apiProjectToProject(directProject as unknown as ApiProject) : undefined),
    isLoading:
      (projectsQuery.isPending && !listFailed) ||
      (needsDirect && directQuery.isPending && !hasQueryFailed(directQuery)),
    // Callers need to tell "no such project" from "the list never arrived".
    // Without this a failed request left `project` undefined once isLoading
    // cleared, and every consumer said "Project not found." for a 500.
    loadFailed: listFailed || (needsDirect && directFailed),
    refetch: needsDirect ? directQuery.refetch : projectsQuery.refetch,
  };
}
