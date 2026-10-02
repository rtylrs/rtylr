import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { canEffectiveAccess } from '@voxire/permissions';
import { FolderKanban, ListTodo } from 'lucide-react';
import type { StoredUser } from '../../shared/stores/userStore';
import type { PortalCommand } from '../../shell/command-palette/commands';
import { flowApi } from './shared/services/flowApi';

export function useFlowCommands(user: StoredUser | null, enabled: boolean, query: string) {
  const [search, setSearch] = useState(query);
  useEffect(() => {
    const timer = setTimeout(() => setSearch(query.trim()), 250);
    return () => clearTimeout(timer);
  }, [query]);
  const projects = useQuery({
    queryKey: ['projects'],
    queryFn: async () => (await unwrap(flowApi.getProjects())).projects ?? [],
    enabled: enabled && canEffectiveAccess(user?.effectiveAccess, 'flow:projects:view', 'flow'),
  });
  const tasks = useQuery({
    queryKey: ['tasks', 'command-search', search],
    queryFn: async () =>
      (await unwrap(search.length >= 3 ? flowApi.searchTasks(search) : flowApi.getTasks({})))
        .tasks ?? [],
    enabled: enabled && canEffectiveAccess(user?.effectiveAccess, 'flow:tasks:view', 'flow'),
  });
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
