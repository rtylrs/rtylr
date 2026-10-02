import {
  ListTodo,
  Sun,
  FolderKanban,
  Inbox,
  Calendar,
  Timer,
  Milestone,
  Eye,
  Archive,
  Map,
  BarChart3,
  Users,
  Clock,
  Contact,
} from 'lucide-react';
import type { StoredUser } from '../../shared/stores/userStore';
import { canAccessPortalPath } from '../../shared/utils/canAccessPortalPath';

export function getFlowHomePath(
  user: StoredUser | null
): '/flow/my-issues' | '/flow/projects' | '/flow/start' {
  if (canAccessPortalPath(user, '/flow/my-issues')) return '/flow/my-issues';
  if (canAccessPortalPath(user, '/flow/projects')) return '/flow/projects';
  return '/flow/start';
}

export const flowNavItems = [
  { path: '/flow/my-issues', key: 'flow.shell.myTasks', icon: ListTodo },
  { path: '/flow/today', key: 'flowNav.focusToday', icon: Sun },
  { path: '/flow/projects', key: 'flow.allProjects.title', icon: FolderKanban },
  { path: '/flow/inbox', key: 'flowNav.inbox', icon: Inbox },
  { path: '/flow/calendar', key: 'flow.shell.calendar', icon: Calendar },
  { path: '/flow/cycles', key: 'flow.shell.cycles', icon: Timer },
  { path: '/flow/milestones', key: 'flow.shell.milestones', icon: Milestone },
  { path: '/flow/views', key: 'flow.shell.views', icon: Eye },
  { path: '/flow/triage', key: 'flow.shell.triage', icon: Eye },
  { path: '/flow/archive', key: 'flow.shell.archive', icon: Archive },
  { path: '/flow/time', key: 'flow.time.heading', icon: Clock },
  { path: '/flow/directory', key: 'flow.directory.heading', icon: Contact },
  { path: '/flow/timeline', key: 'flow.shell.timeline', icon: Timer, preview: true },
  { path: '/flow/roadmap', key: 'flow.shell.roadmap', icon: Map, preview: true },
  { path: '/flow/insights', key: 'flow.shell.insights', icon: BarChart3, preview: true },
  { path: '/flow/team', key: 'flow.shell.team', icon: Users, preview: true },
] as const;
