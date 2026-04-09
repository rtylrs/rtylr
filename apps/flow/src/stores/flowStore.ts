import { create } from 'zustand';

export interface Workspace {
  id: string;
  name: string;
  color: string;
}

export interface Project {
  id: string;
  workspaceId: string;
  name: string;
  color: string;
  taskCount: number;
  completedCount: number;
  memberAvatars: string[];
}

export interface Task {
  id: string;
  projectId: string;
  workspaceId: string;
  title: string;
  status: 'todo' | 'in_progress' | 'in_review' | 'done';
  priority: 'urgent' | 'high' | 'medium' | 'low';
  assignee: string;
  assigneeInitials: string;
  dueDate: string;
  subtaskCount: number;
  projectName: string;
}

interface FlowState {
  activeWorkspaceId: string;
  workspaces: Workspace[];
  projects: Project[];
  tasks: Task[];
  setActiveWorkspace: (id: string) => void;
}

const MOCK_WORKSPACES: Workspace[] = [
  { id: 'default',  name: 'Voxire HQ',      color: '#64C5B9' },
  { id: 'clients',  name: 'Client Projects', color: '#4583C4' },
  { id: 'internal', name: 'Internal',        color: '#f59e0b' },
];

const MOCK_PROJECTS: Project[] = [
  { id: 'rtylr',    workspaceId: 'default',  name: 'rtylr Platform',   color: '#64C5B9', taskCount: 24, completedCount: 18, memberAvatars: ['MH', 'SA', 'LK'] },
  { id: 'flow',     workspaceId: 'default',  name: 'Flow App',         color: '#4583C4', taskCount: 12, completedCount: 4,  memberAvatars: ['MH', 'SA'] },
  { id: 'hub',      workspaceId: 'default',  name: 'Hub Dashboard',    color: '#a855f7', taskCount: 8,  completedCount: 7,  memberAvatars: ['LK'] },
  { id: 'branding', workspaceId: 'clients',  name: 'Brand Refresh',    color: '#f59e0b', taskCount: 6,  completedCount: 2,  memberAvatars: ['MH', 'TN'] },
  { id: 'mobile',   workspaceId: 'clients',  name: 'Mobile App v2',    color: '#ef4444', taskCount: 31, completedCount: 10, memberAvatars: ['SA', 'LK', 'TN'] },
];

const MOCK_TASKS: Task[] = [
  { id: 't1',  projectId: 'flow',  workspaceId: 'default',  title: 'Design kanban board layout',            status: 'in_progress', priority: 'urgent', assignee: 'Mohamed H',   assigneeInitials: 'MH', dueDate: '2026-04-10', subtaskCount: 3, projectName: 'Flow App' },
  { id: 't2',  projectId: 'flow',  workspaceId: 'default',  title: 'Implement drag-and-drop for task cards', status: 'todo',        priority: 'high',   assignee: 'Mohamed H',   assigneeInitials: 'MH', dueDate: '2026-04-14', subtaskCount: 0, projectName: 'Flow App' },
  { id: 't3',  projectId: 'rtylr', workspaceId: 'default',  title: 'Fix auth redirect loop on refresh',     status: 'in_review',   priority: 'urgent', assignee: 'Sara A',      assigneeInitials: 'SA', dueDate: '2026-04-09', subtaskCount: 1, projectName: 'rtylr Platform' },
  { id: 't4',  projectId: 'rtylr', workspaceId: 'default',  title: 'Set up CI/CD pipeline for flow app',    status: 'todo',        priority: 'medium', assignee: 'Mohamed H',   assigneeInitials: 'MH', dueDate: '2026-04-18', subtaskCount: 2, projectName: 'rtylr Platform' },
  { id: 't5',  projectId: 'hub',   workspaceId: 'default',  title: 'Update dashboard analytics charts',     status: 'done',        priority: 'low',    assignee: 'Lena K',      assigneeInitials: 'LK', dueDate: '2026-04-08', subtaskCount: 0, projectName: 'Hub Dashboard' },
  { id: 't6',  projectId: 'flow',  workspaceId: 'default',  title: 'Build task detail modal',               status: 'todo',        priority: 'high',   assignee: 'Mohamed H',   assigneeInitials: 'MH', dueDate: '2026-04-16', subtaskCount: 4, projectName: 'Flow App' },
  { id: 't7',  projectId: 'rtylr', workspaceId: 'default',  title: 'Write API docs for task endpoints',     status: 'in_progress', priority: 'medium', assignee: 'Tarek N',     assigneeInitials: 'TN', dueDate: '2026-04-12', subtaskCount: 0, projectName: 'rtylr Platform' },
  { id: 't8',  projectId: 'flow',  workspaceId: 'default',  title: 'Add keyboard shortcut support',         status: 'todo',        priority: 'low',    assignee: 'Sara A',      assigneeInitials: 'SA', dueDate: '2026-04-20', subtaskCount: 0, projectName: 'Flow App' },
  { id: 't9',  projectId: 'rtylr', workspaceId: 'default',  title: 'Migrate legacy menu store to Zustand',  status: 'done',        priority: 'high',   assignee: 'Lena K',      assigneeInitials: 'LK', dueDate: '2026-04-07', subtaskCount: 2, projectName: 'rtylr Platform' },
  { id: 't10', projectId: 'flow',  workspaceId: 'default',  title: 'Set up Supabase real-time sync',        status: 'in_review',   priority: 'high',   assignee: 'Mohamed H',   assigneeInitials: 'MH', dueDate: '2026-04-11', subtaskCount: 1, projectName: 'Flow App' },
  { id: 't11', projectId: 'rtylr', workspaceId: 'default',  title: 'Refactor sidebar navigation',           status: 'done',        priority: 'medium', assignee: 'Sara A',      assigneeInitials: 'SA', dueDate: '2026-04-06', subtaskCount: 0, projectName: 'rtylr Platform' },
  { id: 't12', projectId: 'flow',  workspaceId: 'default',  title: 'Timeline view implementation',          status: 'todo',        priority: 'medium', assignee: 'Tarek N',     assigneeInitials: 'TN', dueDate: '2026-04-22', subtaskCount: 5, projectName: 'Flow App' },
];

export const useFlowStore = create<FlowState>()((set) => ({
  activeWorkspaceId: 'default',
  workspaces: MOCK_WORKSPACES,
  projects: MOCK_PROJECTS,
  tasks: MOCK_TASKS,
  setActiveWorkspace: (id) => set({ activeWorkspaceId: id }),
}));
