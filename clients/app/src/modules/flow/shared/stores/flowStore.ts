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
  commentCount?: number;
  attachmentCount?: number;
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
  quickCreateOpen: boolean;
  quickCreateColumnId: string | undefined;
  projectCreateOpen: boolean;
  selectedTaskIds: string[];
  setMembers: (members: Member[]) => void;
  setProjects: (projects: Project[]) => void;
  setQuickCreateOpen: (open: boolean, columnId?: string) => void;
  setProjectCreateOpen: (open: boolean) => void;
  toggleSelectTask: (id: string) => void;
  clearSelection: () => void;
  selectAll: (ids: string[]) => void;
}

export const useFlowStore = create<FlowState>()((set) => ({
  members: [],
  projects: [],
  quickCreateOpen: false,
  quickCreateColumnId: undefined,
  projectCreateOpen: false,
  selectedTaskIds: [],
  setMembers: (members) => set({ members }),
  setProjects: (projects) => set({ projects }),
  setQuickCreateOpen: (open, columnId) =>
    set({ quickCreateOpen: open, quickCreateColumnId: open ? columnId : undefined }),
  setProjectCreateOpen: (open) => set({ projectCreateOpen: open }),
  toggleSelectTask: (id) =>
    set((state) => ({
      selectedTaskIds: state.selectedTaskIds.includes(id)
        ? state.selectedTaskIds.filter((selected) => selected !== id)
        : [...state.selectedTaskIds, id],
    })),
  clearSelection: () => set({ selectedTaskIds: [] }),
  selectAll: (ids) => set({ selectedTaskIds: ids }),
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
