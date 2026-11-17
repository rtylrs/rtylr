import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { ProjectServiceClient as ProjectServiceV2 } from '@voxire/voxie/flow/project/v2/client';

// Work (Flow) is a foreign domain: read-only listing of projects so an invoice
// can be tagged with the project it bills. Client app id stays 'finance'.
const projectClient = createOneKitClient(ProjectServiceV2, 'finance');

export interface WorkProject {
  id: string;
  name: string;
  color?: string;
}

export const WORK_PROJECTS_QUERY_KEY = ['flow', 'projects', 'active', 'invoice-picker'] as const;

/**
 * Active, shared Work projects the caller can see, for the invoice Project
 * picker. Personal My Tasks buckets are left out: nothing is billed to them.
 *
 * Degrade: a caller without Work access (or a failed read) yields null and
 * the picker shows no projects; it never blocks invoicing.
 */
export function useWorkProjects() {
  return useQuery<WorkProject[] | null>({
    queryKey: WORK_PROJECTS_QUERY_KEY,
    queryFn: async (): Promise<WorkProject[] | null> => {
      try {
        const res = await projectClient.listProjects({
          status: 'active',
          visibility: '',
          q: '',
          page: 1,
          limit: 100,
        });
        return (res?.projects ?? [])
          .filter((project) => !project.isPersonal && project.id)
          .map((project) => ({
            id: project.id ?? '',
            name: project.name ?? '',
            color: project.color,
          }))
          .sort((a, b) => a.name.localeCompare(b.name));
      } catch {
        return null;
      }
    },
    staleTime: 5 * 60 * 1000,
    retry: false,
  });
}
