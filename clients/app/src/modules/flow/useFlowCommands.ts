import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { canEffectiveAccess } from '@voxire/permissions';
import { FolderKanban, ListTodo } from 'lucide-react';
import type { StoredUser } from '../../shared/stores/userStore';
import type { PortalCommand } from '../../shell/command-palette/commands';
import { flowApi } from './shared/services/flowApi';
import { useTaskSearch } from './shared/hooks/queries/useTaskSearch';

export function useFlowCommands(user: StoredUser | null, enabled: boolean, query: string) {
  const projects = useQuery({
    queryKey: ['projects'],
    queryFn: async () => (await unwrap(flowApi.getProjects())).projects ?? [],
    enabled: enabled && canEffectiveAccess(user?.effectiveAccess, 'flow:projects:view', 'flow'),
  });
  const tasks = useTaskSearch(
    query,
    enabled && canEffectiveAccess(user?.effectiveAccess, 'flow:tasks:view', 'flow')
  );
  const commands: PortalCommand[] = enabled
    ? [
        ...(projects.data ?? []).map((project): PortalCommand => ({
          id: `flow-project-${project.id}`,
          label: project.name,
          group: 'projects',
          icon: FolderKanban,
          destination: { path: `/flow/p/${project.id}` },
        })),
        ...(tasks.data ?? []).slice(0, 30).map((task): PortalCommand => ({
          id: `flow-task-${task.id}`,
          label: task.title,
          description: task.identifier,
          group: 'tasks',
          icon: ListTodo,
          destination: { openFlowTask: String(task.id) },
        })),
      ]
    : [];
  return {
    commands,
    loading: enabled && (projects.isLoading || tasks.isLoading),
    failed: enabled && (projects.isError || tasks.isError),
  };
}
