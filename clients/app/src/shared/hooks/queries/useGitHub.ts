import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { githubClient } from '../../lib/clients';

export type GitHubHealth = 'active' | 'suspended' | 'removed' | 'unknown';

export interface GitHubConnection {
  configured: boolean;
  connected: boolean;
  accountLogin: string;
  accountType: string;
  status: GitHubHealth;
  canManage: boolean;
}

export interface GitHubRepository {
  id: string;
  fullName: string;
  private: boolean;
}

export const GITHUB_CONNECTION_KEY = ['github', 'connection'];
const projectRepositoriesKey = (projectId: string) => ['github', 'project-repositories', projectId];

const HEALTH: readonly GitHubHealth[] = ['active', 'suspended', 'removed', 'unknown'];

// The server omits false and empty fields, so absent means false or empty.
export const useGitHubConnection = () =>
  useQuery({
    queryKey: GITHUB_CONNECTION_KEY,
    queryFn: async (): Promise<GitHubConnection> => {
      const c = (await githubClient.getGitHubConnection({})).connection;
      const status = HEALTH.find((value) => value === c?.status) ?? 'unknown';
      return {
        configured: !!c?.configured,
        connected: !!c?.connected,
        accountLogin: c?.accountLogin ?? '',
        accountType: c?.accountType ?? '',
        status,
        canManage: !!c?.canManage,
      };
    },
  });

export const useStartGitHubInstall = () =>
  useMutation({
    mutationFn: async () => (await githubClient.startGitHubInstall({})).installUrl ?? '',
  });

export const useCompleteGitHubInstall = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: { installationId: string; state: string }) =>
      githubClient.completeGitHubInstall(input),
    onSettled: () => qc.invalidateQueries({ queryKey: GITHUB_CONNECTION_KEY }),
  });
};

export const useDisconnectGitHub = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async () => !!(await githubClient.disconnectGitHub({})).uninstalled,
    onSettled: () => qc.invalidateQueries({ queryKey: ['github'] }),
  });
};

export const useInstallationRepositories = (projectId: string, enabled: boolean) =>
  useQuery({
    queryKey: ['github', 'repositories', projectId],
    queryFn: async (): Promise<GitHubRepository[]> =>
      ((await githubClient.listGitHubRepositories({ projectId })).repositories ?? []).map((r) => ({
        id: r.id ?? '',
        fullName: r.fullName ?? '',
        private: !!r.private,
      })),
    enabled: enabled && !!projectId,
  });

export const useProjectRepositories = (projectId: string, enabled: boolean) =>
  useQuery({
    queryKey: projectRepositoriesKey(projectId),
    queryFn: async (): Promise<GitHubRepository[]> =>
      ((await githubClient.listProjectRepositories({ projectId })).repositories ?? []).map((r) => ({
        id: r.repositoryId ?? '',
        fullName: r.fullName ?? '',
        private: !!r.private,
      })),
    enabled: enabled && !!projectId,
  });

export const useSetProjectRepositories = (projectId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (repositoryIds: string[]) =>
      githubClient.setProjectRepositories({ projectId, repositoryIds }),
    onSuccess: () => qc.invalidateQueries({ queryKey: projectRepositoriesKey(projectId) }),
  });
};
