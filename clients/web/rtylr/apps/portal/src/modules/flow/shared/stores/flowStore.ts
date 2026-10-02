import { create } from 'zustand';

// -- Types ---------------------------------------------------------------

export interface Member {
  id: string;
  name: string;
  email: string;
  initials: string;
  /**
   * Contracted hours per week, from HR's work schedule (#868). Absent when the
   * member has no work schedule assigned, or the schedule records no hours.
   * Optional on purpose: this used to be hardcoded to 40 for everyone, which
   * read a part-timer as under-loaded. Consumers must render absent as unknown
   * rather than substituting a number.
   */
  capacity?: number; // hours/week
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
  computedProgress: number;
  doneCount: number;
  taskCount: number;
  status: 'planning' | 'active' | 'at_risk' | 'completed';
  description?: string;
  project?: { name: string; color: string };
}

export interface Cycle {
  id: string;
  name: string;
  startDate: string;
  endDate: string;
  status: 'upcoming' | 'active' | 'completed';
  scope: number; // task count
  completed: number;
}

export interface Project {
  id: string;
  name: string;
  color: string;
  taskCount: number;
  completedCount: number;
  memberAvatars: string[];
  description?: string;
  status?: string;
  health: 'on_track' | 'at_risk' | 'off_track';
  progress: number; // 0-100
  startDate: string;
  targetDate: string;
  lead: string;
  archived?: boolean;
  icon?: string;
  /**
   * True when the caller sees this project only because they manage someone
   * involved in it (the dept-manager proxy). A per-caller projection from the
   * server, not a property of the project: the same project is false for its own
   * members and true for their manager.
   */
  visibleViaReport?: boolean;
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

/** Sections within a project (list/board grouping). */
export interface ProjectSection {
  id: string;
  projectId: string;
  name: string;
  order: number;
}

/** Custom field definitions (per project). */
export interface CustomFieldDefinition {
  id: string;
  projectId: string;
  name: string;
  type: 'text' | 'number' | 'dropdown';
  options?: string[];
  order: number;
}

export interface Task {
  id: string;
  projectId: string;
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
  /** Quick effort (minutes) for capacity and “small win” picking, distinct from estimate hours. */
  effortMinutes?: number;
  /** Waiting on someone/something, clears mental “open loop” when named. */
  waitingOn?: string;
  /** One-line decision log to reduce second-guessing later. */
  decisionNote?: string;
  /** Pinned to Today / Focus view alongside due-today tasks. */
  focusPin?: boolean;
  subtaskCount: number;
  completedSubtasks?: number;
  projectName: string;
  labelIds?: string[];
  cycleId?: string;
  milestoneId?: string;
  parentId?: string;
  dependencies?: string[];
  archived?: boolean;
  createdAt: string;
  updatedAt: string;
  identifier: string; // e.g. FLOW-123
  sectionId?: string;
  /** Values keyed by CustomFieldDefinition.id */
  customFieldValues?: Record<string, string | number>;
  /** F-A: the custom status (flow.status.v2) applied to this task, if any.
   * Undefined means the task is on its built-in string status alone. */
  statusId?: string;
  statusName?: string;
  statusColor?: string;
}

export interface SavedView {
  id: string;
  name: string;
  icon: string;
  /**
   * 'shared' means every colleague in the business can use this view; 'private'
   * means only its creator. Optional because the server omits it on older
   * responses, and an absent value means shared - which is how every view
   * behaved before the field existed.
   */
  visibility?: 'shared' | 'private';
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

// -- Store ---------------------------------------------------------------

interface FlowState {
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
  sections: ProjectSection[];
  customFieldDefinitions: CustomFieldDefinition[];

  // preferences
  quickCreateOpen: boolean;
  projectCreateOpen: boolean;
  selectedTaskIds: string[];
  defaultView: 'board' | 'list';
  showArchived: boolean;

  // setters for API data
  setMembers: (m: Member[]) => void;
  setProjects: (p: Project[]) => void;
  setTasks: (t: Task[]) => void;
  setComments: (c: Comment[]) => void;
  setActivity: (a: ActivityEntry[]) => void;
  setSubtasks: (s: Subtask[]) => void;
  setAttachments: (a: Attachment[]) => void;
  setMilestones: (m: Milestone[]) => void;
  setCycles: (c: Cycle[]) => void;
  setLabels: (l: Label[]) => void;
  setViews: (v: SavedView[]) => void;
  setTemplates: (t: TaskTemplate[]) => void;
  setNotifications: (n: Notification[]) => void;
  setTimelogs: (t: TimeLog[]) => void;
  setSections: (s: ProjectSection[]) => void;
  setCustomFieldDefinitions: (c: CustomFieldDefinition[]) => void;

  // actions
  setQuickCreateOpen: (open: boolean) => void;
  setProjectCreateOpen: (open: boolean) => void;
  toggleSelectTask: (id: string) => void;
  clearSelection: () => void;
  selectAll: (ids: string[]) => void;
  updateTask: (id: string, patch: Partial<Task>) => void;
  confirmTask: (tempId: string, realId: string) => void;
  deleteTask: (id: string) => void;
  archiveTask: (id: string) => void;
  bulkUpdateTasks: (ids: string[], patch: Partial<Task>) => void;
  addComment: (
    taskId: string,
    body: string,
    author?: { name: string; initials: string; id: string }
  ) => void;
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
  members: [],
  projects: [],
  tasks: [],
  comments: [],
  activity: [],
  subtasks: [],
  attachments: [],
  milestones: [],
  cycles: [],
  labels: [],
  views: [],
  templates: [],
  notifications: [],
  timelogs: [],
  sections: [],
  customFieldDefinitions: [],

  quickCreateOpen: false,
  projectCreateOpen: false,
  selectedTaskIds: [],
  defaultView: 'board',
  showArchived: false,

  // setters for API data
  setMembers: (m) => set({ members: m }),
  setProjects: (p) => set({ projects: p }),
  setTasks: (t) => set({ tasks: t }),
  setComments: (c) => set({ comments: c }),
  setActivity: (a) => set({ activity: a }),
  setSubtasks: (s) => set({ subtasks: s }),
  setAttachments: (a) => set({ attachments: a }),
  setMilestones: (m) => set({ milestones: m }),
  setCycles: (c) => set({ cycles: c }),
  setLabels: (l) => set({ labels: l }),
  setViews: (v) => set({ views: v }),
  setTemplates: (t) => set({ templates: t }),
  setNotifications: (n) => set({ notifications: n }),
  setTimelogs: (t) => set({ timelogs: t }),
  setSections: (s) => set({ sections: s }),
  setCustomFieldDefinitions: (c) => set({ customFieldDefinitions: c }),

  setQuickCreateOpen: (open) => set({ quickCreateOpen: open }),
  setProjectCreateOpen: (open) => set({ projectCreateOpen: open }),
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

  confirmTask: (tempId, realId) =>
    set((s) => ({
      tasks: s.tasks.map((t) => (t.id === tempId ? { ...t, id: realId } : t)),
    })),

  deleteTask: (id) => set((s) => ({ tasks: s.tasks.filter((t) => t.id !== id) })),

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

  addComment: (taskId, body, author) =>
    set((s) => ({
      comments: [
        ...s.comments,
        {
          id: `cm${Date.now()}`,
          taskId,
          authorId: author?.id ?? '',
          author: author?.name ?? 'You',
          authorInitials: author?.initials ?? 'YO',
          body,
          createdAt: new Date().toISOString(),
        },
      ],
    })),

  toggleSubtask: (id) =>
    set((s) => ({
      subtasks: s.subtasks.map((x) => (x.id === id ? { ...x, completed: !x.completed } : x)),
    })),

  addSubtask: (taskId, title) =>
    set((s) => ({
      subtasks: [...s.subtasks, { id: `s${Date.now()}`, taskId, title, completed: false }],
    })),

  markNotificationRead: (id) =>
    set((s) => ({
      notifications: s.notifications.map((n) => (n.id === id ? { ...n, read: true } : n)),
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
          userId: '',
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

// -- Utility selectors ---------------------------------------------------

export const PRIORITY_ORDER: Record<Task['priority'], number> = {
  urgent: 0,
  high: 1,
  medium: 2,
  low: 3,
  no_priority: 4,
};

export const STATUS_COLORS: Record<Task['status'], string> = {
  backlog: 'var(--flow-neutral-fg)',
  todo: 'var(--flow-neutral-fg)',
  in_progress: 'var(--flow-info-fg)',
  in_review: 'var(--flow-warn-fg)',
  done: 'var(--flow-status-approved)',
  cancelled: 'var(--flow-danger-fg)',
};

export const PRIORITY_COLORS: Record<Task['priority'], string> = {
  urgent: 'var(--flow-danger-fg)',
  high: 'var(--flow-priority-high-fg)',
  medium: 'var(--flow-info-fg)',
  low: 'var(--flow-neutral-fg)',
  no_priority: 'var(--flow-neutral-fg)',
};

// Hardcoded pending per-tenant workflow config (priorities/labels differ by business type);
// tracked as separate follow-up work, not in scope here.
export const PRIORITY_LABELS: Record<Task['priority'], string> = {
  urgent: 'Urgent',
  high: 'High',
  medium: 'Medium',
  low: 'Low',
  no_priority: 'No priority',
};
