import type {
  Workspace,
  Member,
  Project,
  Task,
  Comment,
  ActivityEntry,
  Subtask,
  Attachment,
  Milestone,
  Cycle,
  Label,
  SavedView,
  Notification,
  TimeLog,
  ProjectSection,
  CustomFieldDefinition,
} from '../stores/flowStore';
import { DEMO_USER_ID } from '../lib/currentUser';

function iso(d: Date): string {
  return d.toISOString().slice(0, 10);
}

function isoFull(d: Date): string {
  return d.toISOString();
}

const now = new Date();
const today = new Date(now);
today.setHours(0, 0, 0, 0);
const addDays = (base: Date, n: number) => {
  const d = new Date(base);
  d.setDate(d.getDate() + n);
  return d;
};

const D0 = iso(today);
const D1 = iso(addDays(today, 1));
const D3 = iso(addDays(today, 3));
const D7 = iso(addDays(today, 7));
const D14 = iso(addDays(today, 14));
const DM3 = iso(addDays(today, -3));

export const SEED_WORKSPACES: Workspace[] = [
  { id: 'personal', name: 'Personal', color: '#a78bfa' },
  { id: 'default', name: 'Voxire corp', color: '#64C5B9' },
];

export const SEED_MEMBERS: Member[] = [
  {
    id: DEMO_USER_ID,
    name: 'Alex Chen',
    email: 'alex@acme.test',
    initials: 'AC',
    role: 'admin',
    capacity: 40,
  },
  {
    id: 'u2',
    name: 'Sam Rivera',
    email: 'sam@acme.test',
    initials: 'SR',
    role: 'member',
    capacity: 40,
  },
  {
    id: 'u3',
    name: 'Jordan Lee',
    email: 'jordan@acme.test',
    initials: 'JL',
    role: 'member',
    capacity: 32,
  },
  {
    id: 'u4',
    name: 'Riley Park',
    email: 'riley@acme.test',
    initials: 'RP',
    role: 'viewer',
    capacity: 20,
  },
];

export const SEED_PROJECTS: Project[] = [
  {
    id: 'p-web',
    workspaceId: 'default',
    name: 'Website redesign',
    color: '#3b82f6',
    taskCount: 8,
    completedCount: 2,
    memberAvatars: ['AC', 'SR', 'JL'],
    description: 'Marketing site refresh and CMS migration.',
    health: 'on_track',
    progress: 42,
    startDate: iso(addDays(today, -30)),
    targetDate: D14,
    lead: 'Alex Chen',
  },
  {
    id: 'p-mob',
    workspaceId: 'default',
    name: 'Mobile app',
    color: '#f97316',
    taskCount: 5,
    completedCount: 1,
    memberAvatars: ['AC', 'JL'],
    description: 'iOS and Android feature parity.',
    health: 'at_risk',
    progress: 28,
    startDate: iso(addDays(today, -14)),
    targetDate: D7,
    lead: 'Sam Rivera',
  },
  {
    id: 'p-side',
    workspaceId: 'personal',
    name: 'Ideas backlog',
    color: '#22c55e',
    taskCount: 2,
    completedCount: 0,
    memberAvatars: ['AC'],
    description: 'Someday / maybe.',
    health: 'on_track',
    progress: 5,
    startDate: D0,
    targetDate: D14,
    lead: 'Alex Chen',
  },
];

export const SEED_SECTIONS: ProjectSection[] = [
  { id: 'sec-web-a', projectId: 'p-web', name: 'Backlog', order: 0 },
  { id: 'sec-web-b', projectId: 'p-web', name: 'Current sprint', order: 1 },
  { id: 'sec-web-c', projectId: 'p-web', name: 'Waiting', order: 2 },
  { id: 'sec-mob-a', projectId: 'p-mob', name: 'To do', order: 0 },
  { id: 'sec-mob-b', projectId: 'p-mob', name: 'In flight', order: 1 },
];

export const SEED_CUSTOM_FIELDS: CustomFieldDefinition[] = [
  {
    id: 'cf-web-team',
    projectId: 'p-web',
    name: 'Team',
    type: 'dropdown',
    options: ['Engineering', 'Design', 'Product'],
    order: 0,
  },
  {
    id: 'cf-web-pts',
    projectId: 'p-web',
    name: 'Points',
    type: 'number',
    order: 1,
  },
  {
    id: 'cf-mob-area',
    projectId: 'p-mob',
    name: 'Platform',
    type: 'dropdown',
    options: ['iOS', 'Android', 'Shared'],
    order: 0,
  },
];

const T = (partial: Omit<Task, 'createdAt' | 'updatedAt'>): Task => ({
  ...partial,
  createdAt: isoFull(addDays(today, -8)),
  updatedAt: isoFull(now),
});

export const SEED_TASKS: Task[] = [
  T({
    id: 'task-1',
    projectId: 'p-web',
    workspaceId: 'default',
    title: 'Hero section illustrations',
    description: 'Coordinate with brand; export SVG and WebP.',
    status: 'in_progress',
    priority: 'high',
    assignee: 'Alex Chen',
    assigneeInitials: 'AC',
    assigneeId: DEMO_USER_ID,
    dueDate: D0,
    subtaskCount: 2,
    completedSubtasks: 1,
    projectName: 'Website redesign',
    labels: ['lbl-brand'],
    sectionId: 'sec-web-b',
    identifier: 'WEB-101',
    customFieldValues: { 'cf-web-team': 'Design', 'cf-web-pts': 5 },
    dependencies: ['task-2'],
    effortMinutes: 60,
    focusPin: true,
  }),
  T({
    id: 'task-2',
    projectId: 'p-web',
    workspaceId: 'default',
    title: 'Design system tokens',
    description: 'Define spacing and color tokens in repo.',
    status: 'done',
    priority: 'high',
    assignee: 'Sam Rivera',
    assigneeInitials: 'SR',
    assigneeId: 'u2',
    dueDate: DM3,
    subtaskCount: 0,
    projectName: 'Website redesign',
    labels: ['lbl-tech'],
    sectionId: 'sec-web-b',
    identifier: 'WEB-102',
    customFieldValues: { 'cf-web-team': 'Engineering', 'cf-web-pts': 3 },
  }),
  T({
    id: 'task-3',
    projectId: 'p-web',
    workspaceId: 'default',
    title: 'Analytics instrumentation',
    description: 'Segment events for funnel.',
    status: 'todo',
    priority: 'medium',
    assignee: 'Alex Chen',
    assigneeInitials: 'AC',
    assigneeId: DEMO_USER_ID,
    dueDate: D3,
    subtaskCount: 0,
    projectName: 'Website redesign',
    sectionId: 'sec-web-a',
    identifier: 'WEB-103',
    customFieldValues: { 'cf-web-team': 'Product', 'cf-web-pts': 2 },
  }),
  T({
    id: 'task-4',
    projectId: 'p-web',
    workspaceId: 'default',
    title: 'Legal pages copy review',
    status: 'in_review',
    priority: 'low',
    assignee: 'Jordan Lee',
    assigneeInitials: 'JL',
    assigneeId: 'u3',
    dueDate: D7,
    subtaskCount: 0,
    projectName: 'Website redesign',
    sectionId: 'sec-web-c',
    identifier: 'WEB-104',
    customFieldValues: { 'cf-web-team': 'Product', 'cf-web-pts': 1 },
    waitingOn: 'Legal counsel',
  }),
  T({
    id: 'task-5',
    projectId: 'p-web',
    workspaceId: 'default',
    title: 'Performance budget checklist',
    status: 'backlog',
    priority: 'medium',
    assignee: 'Alex Chen',
    assigneeInitials: 'AC',
    assigneeId: DEMO_USER_ID,
    dueDate: '',
    subtaskCount: 0,
    projectName: 'Website redesign',
    sectionId: 'sec-web-a',
    identifier: 'WEB-105',
  }),
  T({
    id: 'task-6',
    projectId: 'p-mob',
    workspaceId: 'default',
    title: 'Push notification permissions',
    status: 'in_progress',
    priority: 'urgent',
    assignee: 'Alex Chen',
    assigneeInitials: 'AC',
    assigneeId: DEMO_USER_ID,
    dueDate: D1,
    subtaskCount: 1,
    projectName: 'Mobile app',
    labels: ['lbl-ios'],
    sectionId: 'sec-mob-b',
    identifier: 'MOB-201',
    customFieldValues: { 'cf-mob-area': 'iOS' },
  }),
  T({
    id: 'task-7',
    projectId: 'p-mob',
    workspaceId: 'default',
    title: 'Offline cache strategy',
    status: 'todo',
    priority: 'high',
    assignee: 'Jordan Lee',
    assigneeInitials: 'JL',
    assigneeId: 'u3',
    dueDate: D14,
    subtaskCount: 0,
    projectName: 'Mobile app',
    sectionId: 'sec-mob-a',
    identifier: 'MOB-202',
    customFieldValues: { 'cf-mob-area': 'Shared' },
  }),
  T({
    id: 'task-8',
    projectId: 'p-mob',
    workspaceId: 'default',
    title: 'Dark mode QA',
    status: 'done',
    priority: 'medium',
    assignee: 'Sam Rivera',
    assigneeInitials: 'SR',
    assigneeId: 'u2',
    dueDate: DM3,
    subtaskCount: 0,
    projectName: 'Mobile app',
    sectionId: 'sec-mob-b',
    identifier: 'MOB-203',
    customFieldValues: { 'cf-mob-area': 'Android' },
  }),
  T({
    id: 'task-9',
    projectId: 'p-side',
    workspaceId: 'personal',
    title: 'Research AI assistants',
    status: 'todo',
    priority: 'low',
    assignee: 'Alex Chen',
    assigneeInitials: 'AC',
    assigneeId: DEMO_USER_ID,
    dueDate: D14,
    subtaskCount: 0,
    projectName: 'Ideas backlog',
    identifier: 'IDE-01',
  }),
];

export const SEED_LABELS: Label[] = [
  { id: 'lbl-brand', name: 'Brand', color: '#ec4899' },
  { id: 'lbl-tech', name: 'Tech', color: '#3b82f6' },
  { id: 'lbl-ios', name: 'iOS', color: '#64748b' },
];

export const SEED_VIEWS: SavedView[] = [
  {
    id: 'view-1',
    name: 'My priorities',
    icon: 'flame',
    filters: { priority: ['urgent', 'high'] },
    sortBy: 'dueDate',
    sortDir: 'asc',
  },
  {
    id: 'view-2',
    name: 'Due this week',
    icon: 'timer',
    filters: {},
    sortBy: 'dueDate',
    sortDir: 'asc',
  },
];

export const SEED_MILESTONES: Milestone[] = [
  {
    id: 'mil-1',
    projectId: 'p-web',
    name: 'Beta launch',
    targetDate: D7,
    progress: 55,
    status: 'active',
    description: 'Public beta of new site.',
  },
  {
    id: 'mil-2',
    projectId: 'p-mob',
    name: 'Store submission',
    targetDate: D14,
    progress: 20,
    status: 'planning',
  },
];

export const SEED_CYCLES: Cycle[] = [
  {
    id: 'cyc-1',
    workspaceId: 'default',
    name: 'Cycle 12',
    startDate: iso(addDays(today, -7)),
    endDate: D7,
    status: 'active',
    scope: 12,
    completed: 4,
  },
];

export const SEED_COMMENTS: Comment[] = [
  {
    id: 'cm-1',
    taskId: 'task-1',
    authorId: 'u2',
    author: 'Sam Rivera',
    authorInitials: 'SR',
    body: 'Figma link is in the project wiki — ping me if you need export help.',
    createdAt: isoFull(addDays(today, -1)),
  },
];

export const SEED_SUBTASKS: Subtask[] = [
  { id: 'st-1', taskId: 'task-1', title: 'Export hero assets', completed: true },
  { id: 'st-2', taskId: 'task-1', title: 'Hand off to dev', completed: false },
  { id: 'st-3', taskId: 'task-6', title: 'Copy for permission modal', completed: false },
];

export const SEED_ACTIVITY: ActivityEntry[] = [
  {
    id: 'act-1',
    taskId: 'task-1',
    type: 'status_change',
    actor: 'Alex Chen',
    actorInitials: 'AC',
    from: 'todo',
    to: 'in_progress',
    createdAt: isoFull(addDays(today, -1)),
  },
];

export const SEED_ATTACHMENTS: Attachment[] = [
  {
    id: 'att-1',
    taskId: 'task-1',
    name: 'hero-mock.png',
    size: 240000,
    url: '#',
    uploadedBy: 'Sam Rivera',
    uploadedAt: isoFull(addDays(today, -2)),
    mime: 'image/png',
  },
];

export const SEED_NOTIFICATIONS: Notification[] = [
  {
    id: 'n1',
    type: 'assigned',
    read: false,
    title: 'New task assigned',
    description: 'You were assigned to Push notification permissions',
    project: 'Mobile app',
    projectColor: '#f97316',
    actor: 'Jordan Lee',
    actorInitials: 'JL',
    timestamp: isoFull(addDays(today, 0)),
    taskTitle: 'Push notification permissions',
    taskId: 'task-6',
  },
  {
    id: 'n2',
    type: 'comment',
    read: false,
    title: 'New comment',
    description: 'Sam mentioned you on Hero section illustrations',
    project: 'Website redesign',
    projectColor: '#3b82f6',
    actor: 'Sam Rivera',
    actorInitials: 'SR',
    timestamp: isoFull(addDays(today, -1)),
    taskTitle: 'Hero section illustrations',
    taskId: 'task-1',
  },
];

export const SEED_TIMELOGS: TimeLog[] = [
  {
    id: 'tl-1',
    taskId: 'task-1',
    userId: DEMO_USER_ID,
    minutes: 90,
    loggedAt: isoFull(addDays(today, -1)),
    note: 'Illustration review',
  },
];
