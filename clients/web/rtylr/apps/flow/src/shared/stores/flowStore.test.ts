import { describe, it, expect, beforeEach } from 'vitest';
import { useFlowStore } from './flowStore';
import type { Task, Project, Workspace, Notification, Subtask, SavedView } from './flowStore';

const { getState, setState } = useFlowStore;

const makeTask = (overrides: Partial<Task> = {}): Task => ({
  id: 't1',
  projectId: 'p1',
  workspaceId: 'w1',
  title: 'Test task',
  status: 'todo',
  priority: 'medium',
  assignee: 'Alice',
  assigneeInitials: 'AL',
  assigneeId: 'u1',
  dueDate: '2026-05-01',
  subtaskCount: 0,
  completedSubtasks: 0,
  projectName: 'Demo',
  labels: [],
  createdAt: '2026-01-01T00:00:00.000Z',
  updatedAt: '2026-01-01T00:00:00.000Z',
  identifier: 'DEMO-1',
  ...overrides,
});

const makeProject = (overrides: Partial<Project> = {}): Project => ({
  id: 'p1',
  workspaceId: 'w1',
  name: 'Demo',
  color: '#3b82f6',
  taskCount: 0,
  completedCount: 0,
  memberAvatars: [],
  health: 'on_track',
  progress: 0,
  startDate: '2026-01-01',
  targetDate: '2026-06-01',
  lead: 'u1',
  ...overrides,
});

beforeEach(() => {
  // Reset entire store to initial state
  setState({
    activeWorkspaceId: '',
    workspaces: [],
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
    commandBarOpen: false,
    quickCreateOpen: false,
    selectedTaskIds: [],
    defaultView: 'board',
    showArchived: false,
  });
});

describe('flowStore', () => {
  // -- Initial state -------------------------------------------------------
  it('has correct initial state', () => {
    const s = getState();
    expect(s.activeWorkspaceId).toBe('');
    expect(s.tasks).toEqual([]);
    expect(s.projects).toEqual([]);
    expect(s.commandBarOpen).toBe(false);
    expect(s.quickCreateOpen).toBe(false);
    expect(s.selectedTaskIds).toEqual([]);
    expect(s.defaultView).toBe('board');
    expect(s.showArchived).toBe(false);
  });

  // -- Workspace -----------------------------------------------------------
  it('setWorkspaces picks personal workspace as default', () => {
    const ws: Workspace[] = [
      { id: 'team', name: 'Team', color: '#f00' },
      { id: 'personal', name: 'Personal', color: '#0f0' },
    ];
    getState().setWorkspaces(ws);
    expect(getState().workspaces).toHaveLength(2);
    expect(getState().activeWorkspaceId).toBe('personal');
  });

  it('setWorkspaces keeps current active workspace if still valid', () => {
    setState({ activeWorkspaceId: 'team' });
    const ws: Workspace[] = [
      { id: 'team', name: 'Team', color: '#f00' },
      { id: 'personal', name: 'Personal', color: '#0f0' },
    ];
    getState().setWorkspaces(ws);
    expect(getState().activeWorkspaceId).toBe('team');
  });

  it('setWorkspaces falls back to first workspace when no personal', () => {
    const ws: Workspace[] = [{ id: 'alpha', name: 'Alpha', color: '#f00' }];
    getState().setWorkspaces(ws);
    expect(getState().activeWorkspaceId).toBe('alpha');
  });

  it('setActiveWorkspace updates active workspace id', () => {
    getState().setActiveWorkspace('w99');
    expect(getState().activeWorkspaceId).toBe('w99');
  });

  // -- UI toggles ----------------------------------------------------------
  it('setCommandBarOpen toggles command bar', () => {
    getState().setCommandBarOpen(true);
    expect(getState().commandBarOpen).toBe(true);
    getState().setCommandBarOpen(false);
    expect(getState().commandBarOpen).toBe(false);
  });

  it('setQuickCreateOpen toggles quick create', () => {
    getState().setQuickCreateOpen(true);
    expect(getState().quickCreateOpen).toBe(true);
  });

  it('setDefaultView switches view', () => {
    getState().setDefaultView('list');
    expect(getState().defaultView).toBe('list');
  });

  it('setShowArchived toggles archived visibility', () => {
    getState().setShowArchived(true);
    expect(getState().showArchived).toBe(true);
  });

  // -- Task selection ------------------------------------------------------
  it('toggleSelectTask adds and removes task from selection', () => {
    getState().toggleSelectTask('t1');
    expect(getState().selectedTaskIds).toEqual(['t1']);
    getState().toggleSelectTask('t1');
    expect(getState().selectedTaskIds).toEqual([]);
  });

  it('selectAll replaces selection with given ids', () => {
    getState().selectAll(['t1', 't2', 't3']);
    expect(getState().selectedTaskIds).toEqual(['t1', 't2', 't3']);
  });

  it('clearSelection empties selected task ids', () => {
    setState({ selectedTaskIds: ['t1', 't2'] });
    getState().clearSelection();
    expect(getState().selectedTaskIds).toEqual([]);
  });

  // -- Task CRUD -----------------------------------------------------------
  it('createTask adds a task with generated id and identifier', () => {
    setState({ projects: [makeProject()] });
    const created = getState().createTask({
      title: 'New task',
      projectId: 'p1',
      workspaceId: 'w1',
    });
    expect(created.title).toBe('New task');
    expect(created.status).toBe('todo');
    expect(created.priority).toBe('medium');
    expect(created.identifier).toMatch(/^DEMO-\d+$/);
    expect(getState().tasks).toHaveLength(1);
    expect(getState().tasks[0].id).toBe(created.id);
  });

  it('createTask uses provided status and priority', () => {
    setState({ projects: [makeProject()] });
    const created = getState().createTask({
      title: 'Urgent fix',
      projectId: 'p1',
      workspaceId: 'w1',
      status: 'in_progress',
      priority: 'urgent',
    });
    expect(created.status).toBe('in_progress');
    expect(created.priority).toBe('urgent');
  });

  it('updateTask patches a single task and updates timestamp', () => {
    const task = makeTask();
    setState({ tasks: [task] });
    getState().updateTask('t1', { status: 'done', priority: 'high' });
    const updated = getState().tasks[0];
    expect(updated.status).toBe('done');
    expect(updated.priority).toBe('high');
    expect(updated.updatedAt).not.toBe(task.updatedAt);
  });

  it('updateTask does not modify other tasks', () => {
    setState({ tasks: [makeTask({ id: 't1' }), makeTask({ id: 't2', title: 'Other' })] });
    getState().updateTask('t1', { title: 'Changed' });
    expect(getState().tasks.find((t) => t.id === 't2')?.title).toBe('Other');
  });

  it('deleteTask removes the task', () => {
    setState({ tasks: [makeTask({ id: 't1' }), makeTask({ id: 't2' })] });
    getState().deleteTask('t1');
    expect(getState().tasks).toHaveLength(1);
    expect(getState().tasks[0].id).toBe('t2');
  });

  it('archiveTask sets archived flag', () => {
    setState({ tasks: [makeTask()] });
    getState().archiveTask('t1');
    expect(getState().tasks[0].archived).toBe(true);
  });

  it('bulkUpdateTasks patches multiple tasks', () => {
    setState({
      tasks: [makeTask({ id: 't1' }), makeTask({ id: 't2' }), makeTask({ id: 't3' })],
    });
    getState().bulkUpdateTasks(['t1', 't3'], { status: 'done' });
    expect(getState().tasks.find((t) => t.id === 't1')?.status).toBe('done');
    expect(getState().tasks.find((t) => t.id === 't2')?.status).toBe('todo');
    expect(getState().tasks.find((t) => t.id === 't3')?.status).toBe('done');
  });

  // -- Comments ------------------------------------------------------------
  it('addComment appends a comment to the list', () => {
    getState().addComment('t1', 'Hello world');
    const comments = getState().comments;
    expect(comments).toHaveLength(1);
    expect(comments[0].body).toBe('Hello world');
    expect(comments[0].taskId).toBe('t1');
    expect(comments[0].author).toBe('You');
  });

  it('addComment uses provided author', () => {
    getState().addComment('t1', 'Review done', { name: 'Bob', initials: 'BO', id: 'u2' });
    expect(getState().comments[0].author).toBe('Bob');
    expect(getState().comments[0].authorInitials).toBe('BO');
  });

  // -- Subtasks ------------------------------------------------------------
  it('addSubtask creates a new uncompleted subtask', () => {
    getState().addSubtask('t1', 'Write tests');
    const subs = getState().subtasks;
    expect(subs).toHaveLength(1);
    expect(subs[0].title).toBe('Write tests');
    expect(subs[0].completed).toBe(false);
  });

  it('toggleSubtask flips completed state', () => {
    const sub: Subtask = { id: 's1', taskId: 't1', title: 'Sub', completed: false };
    setState({ subtasks: [sub] });
    getState().toggleSubtask('s1');
    expect(getState().subtasks[0].completed).toBe(true);
    getState().toggleSubtask('s1');
    expect(getState().subtasks[0].completed).toBe(false);
  });

  // -- Notifications -------------------------------------------------------
  it('markNotificationRead marks a single notification as read', () => {
    const n: Notification = {
      id: 'n1',
      type: 'assigned',
      read: false,
      title: 'Assigned',
      description: 'You were assigned',
      project: 'Demo',
      projectColor: '#f00',
      actor: 'Bob',
      actorInitials: 'BO',
      timestamp: '2026-01-01',
      taskTitle: 'Task',
      taskId: 't1',
    };
    setState({ notifications: [n, { ...n, id: 'n2' }] });
    getState().markNotificationRead('n1');
    expect(getState().notifications.find((x) => x.id === 'n1')?.read).toBe(true);
    expect(getState().notifications.find((x) => x.id === 'n2')?.read).toBe(false);
  });

  it('markAllNotificationsRead marks all as read', () => {
    const base: Notification = {
      id: 'n1',
      type: 'mention',
      read: false,
      title: 'T',
      description: 'D',
      project: 'P',
      projectColor: '#f00',
      actor: 'A',
      actorInitials: 'AA',
      timestamp: '',
      taskTitle: 'T',
      taskId: 't1',
    };
    setState({ notifications: [base, { ...base, id: 'n2' }, { ...base, id: 'n3' }] });
    getState().markAllNotificationsRead();
    expect(getState().notifications.every((n) => n.read)).toBe(true);
  });

  // -- Time logs -----------------------------------------------------------
  it('logTime appends a time log entry', () => {
    getState().logTime('t1', 60, 'Coding session');
    const logs = getState().timelogs;
    expect(logs).toHaveLength(1);
    expect(logs[0].minutes).toBe(60);
    expect(logs[0].note).toBe('Coding session');
    expect(logs[0].taskId).toBe('t1');
  });

  // -- Views ---------------------------------------------------------------
  it('addView appends a saved view', () => {
    const view: SavedView = {
      id: 'v1',
      name: 'My view',
      icon: 'filter',
      filters: { status: ['todo'] },
    };
    getState().addView(view);
    expect(getState().views).toHaveLength(1);
    expect(getState().views[0].name).toBe('My view');
  });

  // -- Collection setters --------------------------------------------------
  it('setMembers replaces members array', () => {
    getState().setMembers([
      { id: 'm1', name: 'Alice', email: 'a@b.com', initials: 'AL', capacity: 40 },
    ]);
    expect(getState().members).toHaveLength(1);
  });

  it('setLabels replaces labels array', () => {
    getState().setLabels([{ id: 'l1', name: 'Bug', color: '#f00' }]);
    expect(getState().labels).toHaveLength(1);
    expect(getState().labels[0].name).toBe('Bug');
  });
});
