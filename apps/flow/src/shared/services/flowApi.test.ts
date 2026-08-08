import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ResultAsync } from 'neverthrow';

vi.mock('../lib/api', () => ({
  safe: {
    get: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: [] }), () => ({}) as any)),
    post: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
    put: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
    patch: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
    delete: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
  },
}));

import { flowApi } from './flowApi';
import { safe } from '../lib/api';

describe('flowApi', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // Dashboard
  it('getDashboard calls GET /flow/dashboard/v1/dashboard with optional workspaceId', () => {
    flowApi.getDashboard('ws-1');
    expect(safe.get).toHaveBeenCalledWith('/flow/dashboard/v1/dashboard', {
      params: { workspace_id: 'ws-1' },
    });
  });

  it('getDashboard without workspaceId passes undefined params', () => {
    flowApi.getDashboard();
    expect(safe.get).toHaveBeenCalledWith('/flow/dashboard/v1/dashboard', {
      params: undefined,
    });
  });

  // Workspaces
  it('getWorkspaces calls GET /flow/workspace/v1/workspaces', () => {
    flowApi.getWorkspaces();
    expect(safe.get).toHaveBeenCalledWith('/flow/workspace/v1/workspaces');
  });

  it('createWorkspace calls POST /flow/workspace/v1/workspaces', () => {
    const data = { name: 'Dev', color: '#ff0000' };
    flowApi.createWorkspace(data);
    expect(safe.post).toHaveBeenCalledWith('/flow/workspace/v1/workspaces', {
      name: 'Dev',
      description: '-',
      icon: 'folder',
      color: '#ff0000',
    });
  });

  it('updateWorkspace calls PUT /flow/workspace/v1/workspaces/{id}', () => {
    flowApi.updateWorkspace('5', { name: 'Updated' });
    expect(safe.put).toHaveBeenCalledWith('/flow/workspace/v1/workspaces/5', {
      body: {
        name: 'Updated',
        description: '-',
        icon: 'folder',
        color: '#6366f1',
      },
    });
  });

  it('deleteWorkspace calls DELETE /flow/workspace/v1/workspaces/{id}', () => {
    flowApi.deleteWorkspace('5');
    expect(safe.delete).toHaveBeenCalledWith('/flow/workspace/v1/workspaces/5');
  });

  it('addWorkspaceMember calls POST /flow/workspace/v1/workspaces/members', () => {
    const data = { workspace_id: 1, user_id: 2, role: 'admin' };
    flowApi.addWorkspaceMember(data);
    expect(safe.post).toHaveBeenCalledWith('/flow/workspace/v1/workspaces/members', data);
  });

  // Projects
  it('getProjects calls GET /flow/project/v1/projects with workspaceId', () => {
    flowApi.getProjects('ws-1');
    expect(safe.get).toHaveBeenCalledWith('/flow/project/v1/projects', {
      params: { workspace_id: 'ws-1' },
    });
  });

  it('createProject calls POST /flow/project/v1/projects', () => {
    const data = { workspaceId: '1', name: 'Project A' };
    flowApi.createProject(data);
    expect(safe.post).toHaveBeenCalledWith('/flow/project/v1/projects', {
      workspace_id: 1,
      name: 'Project A',
      description: '-',
      icon: 'folder',
      color: '#6366f1',
      start_date: '1970-01-01',
      due_date: '1970-01-01',
    });
  });

  it('deleteProject calls DELETE /flow/project/v1/projects/{id}', () => {
    flowApi.deleteProject('7');
    expect(safe.delete).toHaveBeenCalledWith('/flow/project/v1/projects/7');
  });

  // Tasks
  it('getTasks calls GET /flow/task/v1/tasks with snake_case params', () => {
    const params = { projectId: 'p-1', status: 'todo' as const };
    flowApi.getTasks(params);
    expect(safe.get).toHaveBeenCalledWith('/flow/task/v1/tasks', {
      params: { project_id: 'p-1', status: 'todo' },
    });
  });

  it('getTasks with workspaceId sends workspace_id param', () => {
    flowApi.getTasks({ workspaceId: 'ws-1' });
    expect(safe.get).toHaveBeenCalledWith('/flow/task/v1/tasks', {
      params: { workspace_id: 'ws-1' },
    });
  });

  it('getMyTasks calls GET /flow/task/v1/tasks/my', () => {
    flowApi.getMyTasks();
    expect(safe.get).toHaveBeenCalledWith('/flow/task/v1/tasks/my');
  });

  it('createTask calls POST /flow/task/v1/tasks', () => {
    const data = { projectId: '1', workspaceId: '1', title: 'Task 1' };
    flowApi.createTask(data);
    expect(safe.post).toHaveBeenCalledWith('/flow/task/v1/tasks', {
      project_id: 1,
      workspace_id: 1,
      title: 'Task 1',
      description: '-',
      priority: 'medium',
      assignee_id: undefined,
      due_date: '-',
      start_date: '-',
      estimated_hours: undefined,
    });
  });

  it('updateTask calls PUT /flow/task/v1/tasks/{id}', () => {
    flowApi.updateTask('9', { title: 'Updated' });
    expect(safe.put).toHaveBeenCalledWith('/flow/task/v1/tasks/9', {
      body: {
        title: 'Updated',
        description: '-',
        status: 'todo',
        priority: 'medium',
        assignee_id: undefined,
        due_date: '-',
        start_date: '-',
        labels: [0],
        archived: undefined,
      },
    });
  });

  it('updateTask forwards archived when set', () => {
    // Regression guard: TaskDetail's Archive button and Archive.tsx's
    // Restore button both call updateTask({ archived: ... }) - if this
    // function's body allowlist doesn't forward the field, the request
    // silently carries no archive intent at all, independent of whether
    // the backend contract can accept it yet.
    flowApi.updateTask('9', { archived: true });
    expect(safe.put).toHaveBeenCalledWith(
      '/flow/task/v1/tasks/9',
      expect.objectContaining({ body: expect.objectContaining({ archived: true }) })
    );
  });

  it('updateTask forwards archived: false rather than omitting it', () => {
    // The restore path (Archive.tsx) un-archives by sending `false`. Omitting
    // the key is a no-op server-side (COALESCE(?, archived) keeps the current
    // value), so a truthiness guard here would turn Restore into a button that
    // reports success and changes nothing.
    flowApi.updateTask('9', { archived: false });
    expect(safe.put).toHaveBeenCalledWith(
      '/flow/task/v1/tasks/9',
      expect.objectContaining({ body: expect.objectContaining({ archived: false }) })
    );
  });

  it('deleteTask calls DELETE /flow/task/v1/tasks/{id}', () => {
    flowApi.deleteTask('9');
    expect(safe.delete).toHaveBeenCalledWith('/flow/task/v1/tasks/9');
  });

  it('updateTaskStatus calls PATCH /flow/task/v1/tasks/{id}/status', () => {
    flowApi.updateTaskStatus('9', 'done');
    expect(safe.patch).toHaveBeenCalledWith('/flow/task/v1/tasks/9/status', {
      body: { status: 'done' },
    });
  });

  it('addComment calls POST /flow/task/v1/tasks/{taskId}/comments', () => {
    const data = { taskId: 't-1', content: 'Looks good' };
    flowApi.addComment(data);
    expect(safe.post).toHaveBeenCalledWith('/flow/task/v1/tasks/t-1/comments', {
      content: 'Looks good',
    });
  });

  it('deleteComment calls DELETE /flow/task/v1/tasks/comments/{id}', () => {
    flowApi.deleteComment('3');
    expect(safe.delete).toHaveBeenCalledWith('/flow/task/v1/tasks/comments/3');
  });

  // Labels
  it('getLabels calls GET /flow/label/v1/labels', () => {
    flowApi.getLabels();
    expect(safe.get).toHaveBeenCalledWith('/flow/label/v1/labels');
  });

  it('createLabel calls POST /flow/label/v1/labels', () => {
    const data = { name: 'Bug', color: 'red' };
    flowApi.createLabel(data);
    expect(safe.post).toHaveBeenCalledWith('/flow/label/v1/labels', data);
  });

  // Milestones
  it('getMilestones calls GET /flow/milestone/v1/milestones', () => {
    flowApi.getMilestones('p-1');
    expect(safe.get).toHaveBeenCalledWith('/flow/milestone/v1/milestones', {
      params: { project_id: 'p-1' },
    });
  });

  it('updateMilestone calls PUT /flow/milestone/v1/milestones/:id', () => {
    flowApi.updateMilestone('m-1', { name: 'Updated' });
    expect(safe.put).toHaveBeenCalledWith('/flow/milestone/v1/milestones/m-1', {
      name: 'Updated',
    });
  });

  // Cycles
  it('getCycles calls GET /flow/cycle/v1/cycles', () => {
    flowApi.getCycles();
    expect(safe.get).toHaveBeenCalledWith('/flow/cycle/v1/cycles', { params: undefined });
  });

  it('updateCycle calls PUT /flow/cycle/v1/cycles/:id', () => {
    flowApi.updateCycle('cy-1', { name: 'Sprint 2' });
    expect(safe.put).toHaveBeenCalledWith('/flow/cycle/v1/cycles/cy-1', { name: 'Sprint 2' });
  });

  // Views
  it('getViews calls GET /flow/view/v1/views', () => {
    flowApi.getViews();
    expect(safe.get).toHaveBeenCalledWith('/flow/view/v1/views');
  });

  it('deleteView calls DELETE /flow/view/v1/views/:id', () => {
    flowApi.deleteView('v-1');
    expect(safe.delete).toHaveBeenCalledWith('/flow/view/v1/views/v-1');
  });

  // Insights, Workload, Roadmap
  it('getInsights calls GET /flow/insights/v1/overview', () => {
    flowApi.getInsights();
    expect(safe.get).toHaveBeenCalledWith('/flow/insights/v1/overview', { params: undefined });
  });

  it('getTeamWorkload calls GET /flow/insights/v1/team/workload', () => {
    flowApi.getTeamWorkload();
    expect(safe.get).toHaveBeenCalledWith('/flow/insights/v1/team/workload', {
      params: undefined,
    });
  });

  it('getRoadmap calls GET /flow/insights/v1/roadmap', () => {
    flowApi.getRoadmap();
    expect(safe.get).toHaveBeenCalledWith('/flow/insights/v1/roadmap', { params: undefined });
  });
});
