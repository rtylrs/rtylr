import { create } from 'zustand';

// ── Types ────────────────────────────────────────────────────────

export interface Workspace {
  id: string;
  name: string;
  color: string;
}

export interface Member {
  id: string;
  name: string;
  email: string;
  initials: string;
  role: 'owner' | 'admin' | 'member' | 'viewer';
  capacity: number; // hours/week
}

export interface Label {
  id: string;
  name: string;
  color: string;
}

export interface Milestone {
  id: string;
  projectId: string;
  name: string;
  targetDate: string;
  progress: number; // 0-100
  status: 'planning' | 'active' | 'at_risk' | 'completed';
  description?: string;
}

export interface Cycle {
  id: string;
  workspaceId: string;
  name: string;
  startDate: string;
  endDate: string;
  status: 'upcoming' | 'active' | 'completed';
  scope: number; // task count
  completed: number;
}

export interface Project {
  id: string;
  workspaceId: string;
  name: string;
  color: string;
  taskCount: number;
  completedCount: number;
  memberAvatars: string[];
  description?: string;
  health: 'on_track' | 'at_risk' | 'off_track';
  progress: number; // 0-100
  startDate: string;
  targetDate: string;
  lead: string;
  archived?: boolean;
  icon?: string;
}

export interface Comment {
  id: string;
  taskId: string;
  authorId: string;
  author: string;
  authorInitials: string;
  body: string;
  createdAt: string;
  mentions?: string[];
}

export interface ActivityEntry {
  id: string;
  taskId: string;
  type:
    | 'created'
    | 'status_change'
    | 'assignee_change'
    | 'priority_change'
    | 'due_date_change'
    | 'comment'
    | 'label_added'
    | 'attachment';
  actor: string;
  actorInitials: string;
  from?: string;
  to?: string;
  createdAt: string;
  detail?: string;
}

export interface Subtask {
  id: string;
  taskId: string;
  title: string;
  completed: boolean;
}

export interface Attachment {
  id: string;
  taskId: string;
  name: string;
  size: number;
  url: string;
  uploadedBy: string;
  uploadedAt: string;
  mime: string;
}

export interface TimeLog {
  id: string;
  taskId: string;
  userId: string;
  minutes: number;
  loggedAt: string;
  note?: string;
}

export interface Task {
  id: string;
  projectId: string;
  workspaceId: string;
  title: string;
  description?: string;
  status: 'todo' | 'in_progress' | 'in_review' | 'done' | 'cancelled' | 'backlog';
  priority: 'urgent' | 'high' | 'medium' | 'low' | 'no_priority';
  assignee: string;
  assigneeInitials: string;
  assigneeId: string;
  dueDate: string;
  startDate?: string;
  estimate?: number; // hours
  subtaskCount: number;
  completedSubtasks?: number;
  projectName: string;
  labels?: string[];
  cycleId?: string;
  milestoneId?: string;
  parentId?: string;
  dependencies?: string[];
  archived?: boolean;
  createdAt: string;
  updatedAt: string;
  identifier: string; // e.g. FLOW-123
}

export interface SavedView {
  id: string;
  name: string;
  icon: string;
  filters: {
    status?: Task['status'][];
    priority?: Task['priority'][];
    assigneeId?: string;
    projectId?: string;
    labels?: string[];
  };
  groupBy?: 'status' | 'priority' | 'assignee' | 'project' | 'label';
  sortBy?: 'priority' | 'dueDate' | 'updated' | 'created';
  sortDir?: 'asc' | 'desc';
}

export interface TaskTemplate {
  id: string;
  name: string;
  title: string;
  description?: string;
  priority: Task['priority'];
  status: Task['status'];
  labels?: string[];
  subtasks?: string[];
  estimate?: number;
}

export interface Notification {
  id: string;
  type: 'assigned' | 'mention' | 'comment' | 'status_change' | 'due_soon';
  read: boolean;
  title: string;
  description: string;
  project: string;
  projectColor: string;
  actor: string;
  actorInitials: string;
  timestamp: string;
  taskTitle: string;
  taskId: string;
}

// ── Mock Data ────────────────────────────────────────────────────

const MOCK_WORKSPACES: Workspace[] = [
  { id: 'default', name: 'Voxire HQ', color: '#64C5B9' },
  { id: 'clients', name: 'Client Projects', color: '#4583C4' },
  { id: 'internal', name: 'Internal', color: '#f59e0b' },
];

const MOCK_MEMBERS: Member[] = [
  { id: 'u1', name: 'Mohamed H', email: 'me@rtylr.com', initials: 'MH', role: 'owner', capacity: 40 },
  { id: 'u2', name: 'Sara A', email: 'sara@rtylr.com', initials: 'SA', role: 'admin', capacity: 40 },
  { id: 'u3', name: 'Lena K', email: 'lena@rtylr.com', initials: 'LK', role: 'member', capacity: 32 },
  { id: 'u4', name: 'Tarek N', email: 'tarek@rtylr.com', initials: 'TN', role: 'member', capacity: 40 },
  { id: 'u5', name: 'Nora Z', email: 'nora@rtylr.com', initials: 'NZ', role: 'member', capacity: 20 },
];

const MOCK_LABELS: Label[] = [
  { id: 'l1', name: 'Bug', color: '#ef4444' },
  { id: 'l2', name: 'Feature', color: '#3b82f6' },
  { id: 'l3', name: 'Improvement', color: '#22c55e' },
  { id: 'l4', name: 'Design', color: '#a855f7' },
  { id: 'l5', name: 'Docs', color: '#f59e0b' },
  { id: 'l6', name: 'Refactor', color: '#06b6d4' },
];

const MOCK_PROJECTS: Project[] = [
  {
    id: 'rtylr',
    workspaceId: 'default',
    name: 'rtylr Platform',
    color: '#64C5B9',
    taskCount: 24,
    completedCount: 18,
    memberAvatars: ['MH', 'SA', 'LK'],
    description: 'Core platform for rtylr — auth, workspaces, billing.',
    health: 'on_track',
    progress: 75,
    startDate: '2026-02-01',
    targetDate: '2026-05-30',
    lead: 'Mohamed H',
  },
  {
    id: 'flow',
    workspaceId: 'default',
    name: 'Flow App',
    color: '#4583C4',
    taskCount: 28,
    completedCount: 8,
    memberAvatars: ['MH', 'SA'],
    description: 'Project management app inspired by Linear.',
    health: 'on_track',
    progress: 29,
    startDate: '2026-03-15',
    targetDate: '2026-06-30',
    lead: 'Mohamed H',
  },
  {
    id: 'hub',
    workspaceId: 'default',
    name: 'Hub Dashboard',
    color: '#a855f7',
    taskCount: 8,
    completedCount: 7,
    memberAvatars: ['LK'],
    description: 'App launcher and unified dashboard.',
    health: 'at_risk',
    progress: 88,
    startDate: '2026-01-10',
    targetDate: '2026-04-20',
    lead: 'Lena K',
  },
  {
    id: 'branding',
    workspaceId: 'clients',
    name: 'Brand Refresh',
    color: '#f59e0b',
    taskCount: 6,
    completedCount: 2,
    memberAvatars: ['MH', 'TN'],
    description: 'New logo, color system, typography.',
    health: 'on_track',
    progress: 33,
    startDate: '2026-03-20',
    targetDate: '2026-05-10',
    lead: 'Tarek N',
  },
  {
    id: 'mobile',
    workspaceId: 'clients',
    name: 'Mobile App v2',
    color: '#ef4444',
    taskCount: 31,
    completedCount: 10,
    memberAvatars: ['SA', 'LK', 'TN'],
    description: 'Rewrite of the mobile app in React Native.',
    health: 'off_track',
    progress: 32,
    startDate: '2026-02-15',
    targetDate: '2026-07-01',
    lead: 'Sara A',
  },
];

function mkTask(
  id: string,
  projectId: string,
  workspaceId: string,
  title: string,
  status: Task['status'],
  priority: Task['priority'],
  assigneeId: string,
  assignee: string,
  assigneeInitials: string,
  dueDate: string,
  subtaskCount: number,
  projectName: string,
  labels: string[] = [],
  identifier: string,
  description?: string,
  cycleId?: string,
  milestoneId?: string,
  estimate?: number,
  completedSubtasks = 0
): Task {
  return {
    id,
    projectId,
    workspaceId,
    title,
    description,
    status,
    priority,
    assignee,
    assigneeInitials,
    assigneeId,
    dueDate,
    subtaskCount,
    completedSubtasks,
    projectName,
    labels,
    cycleId,
    milestoneId,
    estimate,
    createdAt: '2026-04-01T10:00:00',
    updatedAt: '2026-04-08T10:00:00',
    identifier,
  };
}

const MOCK_TASKS: Task[] = [
  mkTask('t1', 'flow', 'default', 'Design kanban board layout', 'in_progress', 'urgent', 'u1', 'Mohamed H', 'MH', '2026-04-10', 3, 'Flow App', ['l4', 'l2'], 'FLOW-1', 'Define columns, card hover states, empty state.', 'c1', 'm1', 8, 2),
  mkTask('t2', 'flow', 'default', 'Implement drag-and-drop for task cards', 'todo', 'high', 'u1', 'Mohamed H', 'MH', '2026-04-14', 0, 'Flow App', ['l2'], 'FLOW-2', undefined, 'c1', 'm1', 6),
  mkTask('t3', 'rtylr', 'default', 'Fix auth redirect loop on refresh', 'in_review', 'urgent', 'u2', 'Sara A', 'SA', '2026-04-09', 1, 'rtylr Platform', ['l1'], 'RTYL-14', 'Token is being cleared on bootstrap.', 'c1', undefined, 4, 1),
  mkTask('t4', 'rtylr', 'default', 'Set up CI/CD pipeline for flow app', 'todo', 'medium', 'u1', 'Mohamed H', 'MH', '2026-04-18', 2, 'rtylr Platform', ['l3'], 'RTYL-15', undefined, 'c2', undefined, 5),
  mkTask('t5', 'hub', 'default', 'Update dashboard analytics charts', 'done', 'low', 'u3', 'Lena K', 'LK', '2026-04-08', 0, 'Hub Dashboard', ['l3'], 'HUB-22', undefined, 'c1', undefined, 3),
  mkTask('t6', 'flow', 'default', 'Build task detail modal', 'todo', 'high', 'u1', 'Mohamed H', 'MH', '2026-04-16', 4, 'Flow App', ['l2', 'l4'], 'FLOW-3', undefined, 'c2', 'm2', 10),
  mkTask('t7', 'rtylr', 'default', 'Write API docs for task endpoints', 'in_progress', 'medium', 'u4', 'Tarek N', 'TN', '2026-04-12', 0, 'rtylr Platform', ['l5'], 'RTYL-16', undefined, 'c1', undefined, 4),
  mkTask('t8', 'flow', 'default', 'Add keyboard shortcut support', 'todo', 'low', 'u2', 'Sara A', 'SA', '2026-04-20', 0, 'Flow App', ['l3'], 'FLOW-4', undefined, 'c2', 'm2', 6),
  mkTask('t9', 'rtylr', 'default', 'Migrate legacy menu store to Zustand', 'done', 'high', 'u3', 'Lena K', 'LK', '2026-04-07', 2, 'rtylr Platform', ['l6'], 'RTYL-17', undefined, 'c1', undefined, 5, 2),
  mkTask('t10', 'flow', 'default', 'Set up Supabase real-time sync', 'in_review', 'high', 'u1', 'Mohamed H', 'MH', '2026-04-11', 1, 'Flow App', ['l2'], 'FLOW-5', undefined, 'c1', undefined, 8, 0),
  mkTask('t11', 'rtylr', 'default', 'Refactor sidebar navigation', 'done', 'medium', 'u2', 'Sara A', 'SA', '2026-04-06', 0, 'rtylr Platform', ['l6'], 'RTYL-18', undefined, 'c1', undefined, 4),
  mkTask('t12', 'flow', 'default', 'Timeline view implementation', 'todo', 'medium', 'u4', 'Tarek N', 'TN', '2026-04-22', 5, 'Flow App', ['l2'], 'FLOW-6', undefined, 'c2', 'm3', 12, 1),
  mkTask('t13', 'flow', 'default', 'Command bar (cmd+k)', 'todo', 'medium', 'u2', 'Sara A', 'SA', '2026-04-19', 0, 'Flow App', ['l2'], 'FLOW-7', undefined, 'c2', 'm2', 6),
  mkTask('t14', 'flow', 'default', 'Insights and velocity charts', 'backlog', 'low', 'u3', 'Lena K', 'LK', '2026-04-30', 0, 'Flow App', ['l2'], 'FLOW-8', undefined, undefined, 'm3', 8),
  mkTask('t15', 'mobile', 'clients', 'Define onboarding flow screens', 'in_progress', 'high', 'u3', 'Lena K', 'LK', '2026-04-15', 3, 'Mobile App v2', ['l4'], 'MOB-3', undefined, 'c1', undefined, 10),
  mkTask('t16', 'mobile', 'clients', 'Push notifications POC', 'todo', 'medium', 'u4', 'Tarek N', 'TN', '2026-04-28', 2, 'Mobile App v2', ['l2'], 'MOB-4', undefined, 'c2', undefined, 14),
  mkTask('t17', 'branding', 'clients', 'Logo concept exploration', 'in_progress', 'high', 'u4', 'Tarek N', 'TN', '2026-04-11', 0, 'Brand Refresh', ['l4'], 'BR-2', undefined, 'c1', undefined, 6),
  mkTask('t18', 'hub', 'default', 'Dark mode polish pass', 'in_review', 'low', 'u3', 'Lena K', 'LK', '2026-04-09', 0, 'Hub Dashboard', ['l4'], 'HUB-23', undefined, 'c1', undefined, 3),
  mkTask('t19', 'flow', 'default', 'Inline task creation', 'todo', 'medium', 'u1', 'Mohamed H', 'MH', '2026-04-17', 0, 'Flow App', ['l2'], 'FLOW-9', undefined, 'c2', 'm2', 4),
  mkTask('t20', 'flow', 'default', 'Triage inbox for unassigned tasks', 'backlog', 'no_priority', 'u5', 'Nora Z', 'NZ', '2026-05-02', 0, 'Flow App', [], 'FLOW-10', undefined, undefined, 'm3', 5),
];

const MOCK_COMMENTS: Comment[] = [
  { id: 'cm1', taskId: 't1', authorId: 'u2', author: 'Sara A', authorInitials: 'SA', body: 'Loving the direction here. Could we tighten column spacing?', createdAt: '2026-04-06T10:15:00' },
  { id: 'cm2', taskId: 't1', authorId: 'u1', author: 'Mohamed H', authorInitials: 'MH', body: 'Agreed. Pushing new spec shortly.', createdAt: '2026-04-06T11:02:00' },
  { id: 'cm3', taskId: 't3', authorId: 'u3', author: 'Lena K', authorInitials: 'LK', body: 'Confirmed repro on Chrome. Patching the token refresh path.', createdAt: '2026-04-07T08:22:00' },
  { id: 'cm4', taskId: 't6', authorId: 'u2', author: 'Sara A', authorInitials: 'SA', body: '@Mohamed H let me know when you want review.', createdAt: '2026-04-08T14:40:00', mentions: ['u1'] },
];

const MOCK_ACTIVITY: ActivityEntry[] = [
  { id: 'a1', taskId: 't1', type: 'created', actor: 'Mohamed H', actorInitials: 'MH', createdAt: '2026-04-02T09:00:00' },
  { id: 'a2', taskId: 't1', type: 'status_change', actor: 'Mohamed H', actorInitials: 'MH', from: 'todo', to: 'in_progress', createdAt: '2026-04-05T10:30:00' },
  { id: 'a3', taskId: 't1', type: 'assignee_change', actor: 'Sara A', actorInitials: 'SA', to: 'Mohamed H', createdAt: '2026-04-05T11:00:00' },
  { id: 'a4', taskId: 't1', type: 'label_added', actor: 'Mohamed H', actorInitials: 'MH', detail: 'Feature', createdAt: '2026-04-06T09:00:00' },
  { id: 'a5', taskId: 't3', type: 'priority_change', actor: 'Sara A', actorInitials: 'SA', from: 'high', to: 'urgent', createdAt: '2026-04-07T07:45:00' },
];

const MOCK_SUBTASKS: Subtask[] = [
  { id: 's1', taskId: 't1', title: 'Sketch layout', completed: true },
  { id: 's2', taskId: 't1', title: 'Pick column colors', completed: true },
  { id: 's3', taskId: 't1', title: 'Define empty state', completed: false },
  { id: 's4', taskId: 't6', title: 'Collect all fields needed', completed: true },
  { id: 's5', taskId: 't6', title: 'Wire comments panel', completed: false },
  { id: 's6', taskId: 't6', title: 'Wire activity tab', completed: false },
  { id: 's7', taskId: 't6', title: 'Mobile breakpoint', completed: false },
];

const MOCK_ATTACHMENTS: Attachment[] = [
  { id: 'att1', taskId: 't1', name: 'board-wireframe.png', size: 238124, url: '#', uploadedBy: 'Mohamed H', uploadedAt: '2026-04-05T12:00:00', mime: 'image/png' },
  { id: 'att2', taskId: 't6', name: 'task-detail-spec.pdf', size: 512338, url: '#', uploadedBy: 'Sara A', uploadedAt: '2026-04-06T15:00:00', mime: 'application/pdf' },
];

const MOCK_MILESTONES: Milestone[] = [
  { id: 'm1', projectId: 'flow', name: 'MVP Board', targetDate: '2026-04-20', progress: 60, status: 'active', description: 'Kanban board ships with drag-drop.' },
  { id: 'm2', projectId: 'flow', name: 'Task Detail & Shortcuts', targetDate: '2026-05-10', progress: 25, status: 'active' },
  { id: 'm3', projectId: 'flow', name: 'Insights & Reports', targetDate: '2026-06-15', progress: 5, status: 'planning' },
  { id: 'm4', projectId: 'rtylr', name: 'Auth Hardening', targetDate: '2026-04-25', progress: 80, status: 'active' },
  { id: 'm5', projectId: 'mobile', name: 'Beta release', targetDate: '2026-06-01', progress: 40, status: 'at_risk' },
];

const MOCK_CYCLES: Cycle[] = [
  { id: 'c1', workspaceId: 'default', name: 'Cycle 12', startDate: '2026-04-01', endDate: '2026-04-14', status: 'active', scope: 14, completed: 5 },
  { id: 'c2', workspaceId: 'default', name: 'Cycle 13', startDate: '2026-04-15', endDate: '2026-04-28', status: 'upcoming', scope: 9, completed: 0 },
  { id: 'c0', workspaceId: 'default', name: 'Cycle 11', startDate: '2026-03-18', endDate: '2026-03-31', status: 'completed', scope: 12, completed: 12 },
];

const MOCK_TEMPLATES: TaskTemplate[] = [
  { id: 'tpl1', name: 'Bug report', title: 'Bug: ', priority: 'high', status: 'todo', labels: ['l1'], subtasks: ['Reproduce', 'Diagnose', 'Fix', 'Regression test'] },
  { id: 'tpl2', name: 'Feature spec', title: 'Feature: ', priority: 'medium', status: 'backlog', labels: ['l2'], subtasks: ['Write spec', 'Design review', 'Break into tasks'] },
  { id: 'tpl3', name: 'Design review', title: 'Design review: ', priority: 'medium', status: 'todo', labels: ['l4'] },
];

const MOCK_VIEWS: SavedView[] = [
  { id: 'v1', name: 'Urgent & High', icon: 'flame', filters: { priority: ['urgent', 'high'] }, groupBy: 'status', sortBy: 'priority', sortDir: 'asc' },
  { id: 'v2', name: 'Active Sprint', icon: 'timer', filters: {}, groupBy: 'assignee', sortBy: 'dueDate', sortDir: 'asc' },
  { id: 'v3', name: 'Needs Review', icon: 'check-circle', filters: { status: ['in_review'] }, groupBy: 'project', sortBy: 'updated', sortDir: 'desc' },
];

const MOCK_NOTIFICATIONS: Notification[] = [
  {
    id: 'n1',
    type: 'assigned',
    read: false,
    title: 'Assigned to you',
    description: 'Sara assigned you a new task',
    project: 'Flow App',
    projectColor: '#4583C4',
    actor: 'Sara A',
    actorInitials: 'SA',
    timestamp: '2026-04-09T09:30:00',
    taskTitle: 'Build task detail modal',
    taskId: 't6',
  },
  {
    id: 'n2',
    type: 'comment',
    read: false,
    title: 'New comment',
    description: 'Lena commented on your task',
    project: 'rtylr Platform',
    projectColor: '#64C5B9',
    actor: 'Lena K',
    actorInitials: 'LK',
    timestamp: '2026-04-09T08:15:00',
    taskTitle: 'Fix auth redirect loop on refresh',
    taskId: 't3',
  },
  {
    id: 'n3',
    type: 'mention',
    read: false,
    title: 'You were mentioned',
    description: 'Tarek mentioned you in a comment',
    project: 'rtylr Platform',
    projectColor: '#64C5B9',
    actor: 'Tarek N',
    actorInitials: 'TN',
    timestamp: '2026-04-08T17:45:00',
    taskTitle: 'Write API docs for task endpoints',
    taskId: 't7',
  },
  {
    id: 'n4',
    type: 'status_change',
    read: true,
    title: 'Task completed',
    description: 'Lena marked a task as done',
    project: 'Hub Dashboard',
    projectColor: '#a855f7',
    actor: 'Lena K',
    actorInitials: 'LK',
    timestamp: '2026-04-08T16:00:00',
    taskTitle: 'Update dashboard analytics charts',
    taskId: 't5',
  },
  {
    id: 'n5',
    type: 'assigned',
    read: true,
    title: 'Assigned to you',
    description: 'Mohamed assigned you to this task',
    project: 'Flow App',
    projectColor: '#4583C4',
    actor: 'Mohamed H',
    actorInitials: 'MH',
    timestamp: '2026-04-08T14:22:00',
    taskTitle: 'Timeline view implementation',
    taskId: 't12',
  },
];

const MOCK_TIMELOGS: TimeLog[] = [
  { id: 'tl1', taskId: 't1', userId: 'u1', minutes: 90, loggedAt: '2026-04-06T09:00:00', note: 'Initial wireframes' },
  { id: 'tl2', taskId: 't1', userId: 'u1', minutes: 120, loggedAt: '2026-04-07T13:30:00', note: 'Column polish' },
  { id: 'tl3', taskId: 't3', userId: 'u2', minutes: 60, loggedAt: '2026-04-08T10:00:00' },
];

// ── Store ────────────────────────────────────────────────────────

interface FlowState {
  activeWorkspaceId: string;
  workspaces: Workspace[];
  members: Member[];
  projects: Project[];
  tasks: Task[];
  comments: Comment[];
  activity: ActivityEntry[];
  subtasks: Subtask[];
  attachments: Attachment[];
  milestones: Milestone[];
  cycles: Cycle[];
  labels: Label[];
  views: SavedView[];
  templates: TaskTemplate[];
  notifications: Notification[];
  timelogs: TimeLog[];

  // preferences
  commandBarOpen: boolean;
  quickCreateOpen: boolean;
  selectedTaskIds: string[];
  defaultView: 'board' | 'list';
  showArchived: boolean;

  // actions
  setActiveWorkspace: (id: string) => void;
  setCommandBarOpen: (open: boolean) => void;
  setQuickCreateOpen: (open: boolean) => void;
  toggleSelectTask: (id: string) => void;
  clearSelection: () => void;
  selectAll: (ids: string[]) => void;
  updateTask: (id: string, patch: Partial<Task>) => void;
  createTask: (data: Partial<Task> & { title: string; projectId: string; workspaceId: string }) => Task;
  deleteTask: (id: string) => void;
  archiveTask: (id: string) => void;
  bulkUpdateTasks: (ids: string[], patch: Partial<Task>) => void;
  addComment: (taskId: string, body: string) => void;
  toggleSubtask: (id: string) => void;
  addSubtask: (taskId: string, title: string) => void;
  markNotificationRead: (id: string) => void;
  markAllNotificationsRead: () => void;
  logTime: (taskId: string, minutes: number, note?: string) => void;
  setDefaultView: (v: 'board' | 'list') => void;
  setShowArchived: (v: boolean) => void;
  addView: (view: SavedView) => void;
}

export const useFlowStore = create<FlowState>()((set, get) => ({
  activeWorkspaceId: 'default',
  workspaces: MOCK_WORKSPACES,
  members: MOCK_MEMBERS,
  projects: MOCK_PROJECTS,
  tasks: MOCK_TASKS,
  comments: MOCK_COMMENTS,
  activity: MOCK_ACTIVITY,
  subtasks: MOCK_SUBTASKS,
  attachments: MOCK_ATTACHMENTS,
  milestones: MOCK_MILESTONES,
  cycles: MOCK_CYCLES,
  labels: MOCK_LABELS,
  views: MOCK_VIEWS,
  templates: MOCK_TEMPLATES,
  notifications: MOCK_NOTIFICATIONS,
  timelogs: MOCK_TIMELOGS,

  commandBarOpen: false,
  quickCreateOpen: false,
  selectedTaskIds: [],
  defaultView: 'board',
  showArchived: false,

  setActiveWorkspace: (id) => set({ activeWorkspaceId: id }),
  setCommandBarOpen: (open) => set({ commandBarOpen: open }),
  setQuickCreateOpen: (open) => set({ quickCreateOpen: open }),
  toggleSelectTask: (id) =>
    set((s) => ({
      selectedTaskIds: s.selectedTaskIds.includes(id)
        ? s.selectedTaskIds.filter((x) => x !== id)
        : [...s.selectedTaskIds, id],
    })),
  clearSelection: () => set({ selectedTaskIds: [] }),
  selectAll: (ids) => set({ selectedTaskIds: ids }),

  updateTask: (id, patch) =>
    set((s) => ({
      tasks: s.tasks.map((t) =>
        t.id === id ? { ...t, ...patch, updatedAt: new Date().toISOString() } : t
      ),
    })),

  createTask: (data) => {
    const state = get();
    const project = state.projects.find((p) => p.id === data.projectId);
    const id = `t${Date.now()}`;
    const count = state.tasks.filter((t) => t.projectId === data.projectId).length + 1;
    const newTask: Task = {
      id,
      projectId: data.projectId,
      workspaceId: data.workspaceId,
      title: data.title,
      description: data.description,
      status: data.status || 'todo',
      priority: data.priority || 'medium',
      assignee: data.assignee || 'Mohamed H',
      assigneeInitials: data.assigneeInitials || 'MH',
      assigneeId: data.assigneeId || 'u1',
      dueDate: data.dueDate || '',
      subtaskCount: 0,
      completedSubtasks: 0,
      projectName: project?.name || '',
      labels: data.labels || [],
      cycleId: data.cycleId,
      milestoneId: data.milestoneId,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      identifier: `${(project?.name || 'TASK').slice(0, 4).toUpperCase()}-${count}`,
    };
    set((s) => ({ tasks: [newTask, ...s.tasks] }));
    return newTask;
  },

  deleteTask: (id) =>
    set((s) => ({ tasks: s.tasks.filter((t) => t.id !== id) })),

  archiveTask: (id) =>
    set((s) => ({
      tasks: s.tasks.map((t) => (t.id === id ? { ...t, archived: true } : t)),
    })),

  bulkUpdateTasks: (ids, patch) =>
    set((s) => ({
      tasks: s.tasks.map((t) =>
        ids.includes(t.id) ? { ...t, ...patch, updatedAt: new Date().toISOString() } : t
      ),
    })),

  addComment: (taskId, body) =>
    set((s) => ({
      comments: [
        ...s.comments,
        {
          id: `cm${Date.now()}`,
          taskId,
          authorId: 'u1',
          author: 'Mohamed H',
          authorInitials: 'MH',
          body,
          createdAt: new Date().toISOString(),
        },
      ],
    })),

  toggleSubtask: (id) =>
    set((s) => ({
      subtasks: s.subtasks.map((x) =>
        x.id === id ? { ...x, completed: !x.completed } : x
      ),
    })),

  addSubtask: (taskId, title) =>
    set((s) => ({
      subtasks: [
        ...s.subtasks,
        { id: `s${Date.now()}`, taskId, title, completed: false },
      ],
    })),

  markNotificationRead: (id) =>
    set((s) => ({
      notifications: s.notifications.map((n) =>
        n.id === id ? { ...n, read: true } : n
      ),
    })),

  markAllNotificationsRead: () =>
    set((s) => ({
      notifications: s.notifications.map((n) => ({ ...n, read: true })),
    })),

  logTime: (taskId, minutes, note) =>
    set((s) => ({
      timelogs: [
        ...s.timelogs,
        {
          id: `tl${Date.now()}`,
          taskId,
          userId: 'u1',
          minutes,
          note,
          loggedAt: new Date().toISOString(),
        },
      ],
    })),

  setDefaultView: (v) => set({ defaultView: v }),
  setShowArchived: (v) => set({ showArchived: v }),
  addView: (view) => set((s) => ({ views: [...s.views, view] })),
}));

// ── Utility selectors ────────────────────────────────────────────

export const PRIORITY_ORDER: Record<Task['priority'], number> = {
  urgent: 0,
  high: 1,
  medium: 2,
  low: 3,
  no_priority: 4,
};

export const STATUS_LABELS: Record<Task['status'], string> = {
  backlog: 'Backlog',
  todo: 'Todo',
  in_progress: 'In Progress',
  in_review: 'In Review',
  done: 'Done',
  cancelled: 'Cancelled',
};

export const STATUS_COLORS: Record<Task['status'], string> = {
  backlog: '#6b7280',
  todo: '#9ca3af',
  in_progress: '#3b82f6',
  in_review: '#f59e0b',
  done: '#22c55e',
  cancelled: '#ef4444',
};

export const PRIORITY_COLORS: Record<Task['priority'], string> = {
  urgent: '#ef4444',
  high: '#f97316',
  medium: '#3b82f6',
  low: '#6b7280',
  no_priority: '#4b5563',
};

export const PRIORITY_LABELS: Record<Task['priority'], string> = {
  urgent: 'Urgent',
  high: 'High',
  medium: 'Medium',
  low: 'Low',
  no_priority: 'No priority',
};
