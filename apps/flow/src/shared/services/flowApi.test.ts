import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ResultAsync } from 'neverthrow';

vi.mock('../lib/api', () => ({
  safe: {
    get: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: [] }), () => ({}) as any)),
    post: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
    put: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
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
  it('getDashboard calls GET /flow/dashboard with optional workspaceId', () => {
    flowApi.getDashboard('ws-1');
    expect(safe.get).toHaveBeenCalledWith('/flow/dashboard', {
      params: { workspace_id: 'ws-1' },
    });
  });

  it('getDashboard without workspaceId passes undefined params', () => {
    flowApi.getDashboard();
    expect(safe.get).toHaveBeenCalledWith('/flow/dashboard', { params: undefined });
  });

  // Workspaces
  it('getWorkspaces calls GET /flow/workspaces', () => {
    flowApi.getWorkspaces();
    expect(safe.get).toHaveBeenCalledWith('/flow/workspaces');
  });

  it('createWorkspace calls POST /flow/workspaces', () => {
    const data = { name: 'Dev', color: '#ff0000' };
    flowApi.createWorkspace(data);
    expect(safe.post).toHaveBeenCalledWith('/flow/workspaces', data);
  });

  it('updateWorkspace calls PUT /flow/workspaces with id in body', () => {
    flowApi.updateWorkspace('5', { name: 'Updated' });
    expect(safe.put).toHaveBeenCalledWith('/flow/workspaces', { id: 5, name: 'Updated' });
  });

  it('deleteWorkspace calls DELETE /flow/workspaces with id in body', () => {
    flowApi.deleteWorkspace('5');
    expect(safe.delete).toHaveBeenCalledWith('/flow/workspaces', { data: { id: 5 } });
  });

  it('addWorkspaceMember calls POST /flow/workspaces/member', () => {
    const data = { workspaceId: 'ws-1', userId: 'u-1', role: 'admin' };
    flowApi.addWorkspaceMember(data);
    expect(safe.post).toHaveBeenCalledWith('/flow/workspaces/member', data);
  });

  // Projects
  it('getProjects calls GET /flow/projects with workspaceId', () => {
    flowApi.getProjects('ws-1');
    expect(safe.get).toHaveBeenCalledWith('/flow/projects', {
      params: { workspace_id: 'ws-1' },
    });
  });

  it('createProject calls POST /flow/projects', () => {
    const data = { workspaceId: 'ws-1', name: 'Project A' };
    flowApi.createProject(data);
    expect(safe.post).toHaveBeenCalledWith('/flow/projects', data);
  });

  it('deleteProject calls DELETE /flow/projects with id in body', () => {
    flowApi.deleteProject('7');
    expect(safe.delete).toHaveBeenCalledWith('/flow/projects', { data: { id: 7 } });
  });

  // Tasks
  it('getTasks calls GET /flow/tasks with snake_case params', () => {
    const params = { projectId: 'p-1', status: 'todo' as const };
    flowApi.getTasks(params);
    expect(safe.get).toHaveBeenCalledWith('/flow/tasks', {
      params: { project_id: 'p-1', status: 'todo' },
    });
  });

  it('getTasks with workspaceId sends workspace_id param', () => {
    flowApi.getTasks({ workspaceId: 'ws-1' });
    expect(safe.get).toHaveBeenCalledWith('/flow/tasks', {
      params: { workspace_id: 'ws-1' },
    });
  });

  it('getMyTasks calls GET /flow/tasks/my', () => {
    flowApi.getMyTasks();
    expect(safe.get).toHaveBeenCalledWith('/flow/tasks/my');
  });

  it('createTask calls POST /flow/tasks', () => {
    const data = { projectId: 'p-1', workspaceId: 'ws-1', title: 'Task 1' };
    flowApi.createTask(data);
    expect(safe.post).toHaveBeenCalledWith('/flow/tasks', data);
  });

  it('updateTask calls PUT /flow/tasks with id in body', () => {
    flowApi.updateTask('9', { title: 'Updated' });
    expect(safe.put).toHaveBeenCalledWith('/flow/tasks', { id: 9, title: 'Updated' });
  });

  it('deleteTask calls DELETE /flow/tasks with id in body', () => {
    flowApi.deleteTask('9');
    expect(safe.delete).toHaveBeenCalledWith('/flow/tasks', { data: { id: 9 } });
  });

  it('updateTaskStatus calls PUT /flow/tasks/status', () => {
    flowApi.updateTaskStatus('9', 'done');
    expect(safe.put).toHaveBeenCalledWith('/flow/tasks/status', { id: 9, status: 'done' });
  });

  it('addComment calls POST /flow/tasks/comment', () => {
    const data = { taskId: 't-1', content: 'Looks good' };
    flowApi.addComment(data);
    expect(safe.post).toHaveBeenCalledWith('/flow/tasks/comment', data);
  });

  it('deleteComment calls DELETE /flow/tasks/comment with id in body', () => {
    flowApi.deleteComment('3');
    expect(safe.delete).toHaveBeenCalledWith('/flow/tasks/comment', { data: { id: 3 } });
  });

  // Labels
  it('getLabels calls GET /flow/labels', () => {
    flowApi.getLabels();
    expect(safe.get).toHaveBeenCalledWith('/flow/labels');
  });

  it('createLabel calls POST /flow/labels', () => {
    const data = { name: 'Bug', color: 'red' };
    flowApi.createLabel(data);
    expect(safe.post).toHaveBeenCalledWith('/flow/labels', data);
  });

  // Milestones
  it('getMilestones calls GET /flow/milestones', () => {
    flowApi.getMilestones('p-1');
    expect(safe.get).toHaveBeenCalledWith('/flow/milestones', {
      params: { project_id: 'p-1' },
    });
  });

  it('updateMilestone calls PUT /flow/milestones/:id', () => {
    flowApi.updateMilestone('m-1', { name: 'Updated' });
    expect(safe.put).toHaveBeenCalledWith('/flow/milestones/m-1', { name: 'Updated' });
  });

  // Cycles
  it('getCycles calls GET /flow/cycles', () => {
    flowApi.getCycles();
    expect(safe.get).toHaveBeenCalledWith('/flow/cycles', { params: undefined });
  });

  it('updateCycle calls PUT /flow/cycles/:id', () => {
    flowApi.updateCycle('cy-1', { name: 'Sprint 2' });
    expect(safe.put).toHaveBeenCalledWith('/flow/cycles/cy-1', { name: 'Sprint 2' });
  });

  // Views
  it('getViews calls GET /flow/views', () => {
    flowApi.getViews();
    expect(safe.get).toHaveBeenCalledWith('/flow/views');
  });

  it('deleteView calls DELETE /flow/views/:id', () => {
    flowApi.deleteView('v-1');
    expect(safe.delete).toHaveBeenCalledWith('/flow/views/v-1');
  });

  // Insights, Workload, Roadmap
  it('getInsights calls GET /flow/insights', () => {
    flowApi.getInsights();
    expect(safe.get).toHaveBeenCalledWith('/flow/insights', { params: undefined });
  });

  it('getTeamWorkload calls GET /flow/team/workload', () => {
    flowApi.getTeamWorkload();
    expect(safe.get).toHaveBeenCalledWith('/flow/team/workload', { params: undefined });
  });

  it('getRoadmap calls GET /flow/roadmap', () => {
    flowApi.getRoadmap();
    expect(safe.get).toHaveBeenCalledWith('/flow/roadmap', { params: undefined });
  });
});
