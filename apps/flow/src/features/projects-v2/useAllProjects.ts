import { useQuery } from '@tanstack/react-query';
import type { Project } from '@voxire/voxie/flow/project/v2/types';
import { projectV2 } from '../../shared/lib/clients';

interface ListProjectsParams {
  status?: string;
  visibility?: string;
  q?: string;
  page?: number;
  limit?: number;
}

export function useAllProjects(params: ListProjectsParams = {}) {
  return useQuery<{ projects: Project[]; total: number }>({
    queryKey: ['projects-v2', params],
    queryFn: async () => {
      const res = await projectV2.listProjects({
        status: params.status ?? '',
        visibility: params.visibility ?? '',
        q: params.q ?? '',
        page: params.page ?? 1,
        limit: params.limit ?? 50,
      });
      return {
        projects: (res as { projects?: Project[] }).projects ?? [],
        total: (res as { pagination?: { total?: number } }).pagination?.total ?? 0,
      };
    },
  });
}

export function useProject(projectId: string | undefined) {
  return useQuery({
    queryKey: ['project-v2', projectId],
    queryFn: async () => {
      const res = await projectV2.getProject({ id: projectId! });
      return (res as { project?: Project }).project ?? undefined;
    },
    enabled: !!projectId,
  });
}

export interface ArchivedProject {
  project: Project;
}

export function useArchivedProjects(page = 1, limit = 20) {
  return useQuery<{ projects: Project[]; total: number }>({
    queryKey: ['archived-projects', page, limit],
    queryFn: async () => {
      const res = await projectV2.listArchivedProjects({ page, limit });
      return {
        projects: (res as { projects?: Project[] }).projects ?? [],
        total: (res as { pagination?: { total?: number } }).pagination?.total ?? 0,
      };
    },
  });
}
