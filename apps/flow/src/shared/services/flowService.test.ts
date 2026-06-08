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

import { flowService } from './flowService';
import { safe } from '../lib/api';

describe('flowService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('getDashboard calls GET /flow/dashboard', () => {
    flowService.getDashboard();
    expect(safe.get).toHaveBeenCalledWith('/flow/dashboard');
  });

  // Workspaces
  it('getWorkspaces calls GET /flow/workspaces', () => {
    flowService.getWorkspaces();
    expect(safe.get).toHaveBeenCalledWith('/flow/workspaces');
  });

  it('createWorkspace calls POST /flow/workspaces', () => {
    const data = { name: 'WS' };
    flowService.createWorkspace(data);
    expect(safe.post).toHaveBeenCalledWith('/flow/workspaces', data);
  });

  it('updateWorkspace calls PUT /flow/workspaces with id', () => {
    flowService.updateWorkspace(1, { name: 'Updated' });
    expect(safe.put).toHaveBeenCalledWith('/flow/workspaces', { id: 1, name: 'Updated' });
  });

  it('deleteWorkspace calls DELETE /flow/workspaces with id', () => {
    flowService.deleteWorkspace(1);
    expect(safe.delete).toHaveBeenCalledWith('/flow/workspaces', { data: { id: 1 } });
  });

  // Projects
  it('getProjects calls GET /flow/projects with params', () => {
    const params = { workspaceId: 1 };
    flowService.getProjects(params);
    expect(safe.get).toHaveBeenCalledWith('/flow/projects', { params });
  });

  it('createProject calls POST /flow/projects', () => {
    const data = { name: 'Proj' };
    flowService.createProject(data);
    expect(safe.post).toHaveBeenCalledWith('/flow/projects', data);
  });

  it('updateProject calls PUT /flow/projects with id', () => {
    flowService.updateProject(2, { name: 'Updated' });
    expect(safe.put).toHaveBeenCalledWith('/flow/projects', { id: 2, name: 'Updated' });
  });

  it('deleteProject calls DELETE /flow/projects with id', () => {
    flowService.deleteProject(2);
    expect(safe.delete).toHaveBeenCalledWith('/flow/projects', { data: { id: 2 } });
  });

  // Tasks
  it('getTasks calls GET /flow/tasks with params', () => {
    const params = { status: 'open' };
    flowService.getTasks(params);
    expect(safe.get).toHaveBeenCalledWith('/flow/tasks', { params });
  });

  it('createTask calls POST /flow/tasks', () => {
    const data = { title: 'Task' };
    flowService.createTask(data);
    expect(safe.post).toHaveBeenCalledWith('/flow/tasks', data);
  });

  it('updateTask calls PUT /flow/tasks with id', () => {
    flowService.updateTask(3, { title: 'Updated' });
    expect(safe.put).toHaveBeenCalledWith('/flow/tasks', { id: 3, title: 'Updated' });
  });

  it('deleteTask calls DELETE /flow/tasks with id', () => {
    flowService.deleteTask(3);
    expect(safe.delete).toHaveBeenCalledWith('/flow/tasks', { data: { id: 3 } });
  });

  // Labels
  it('getLabels calls GET /flow/labels', () => {
    flowService.getLabels();
    expect(safe.get).toHaveBeenCalledWith('/flow/labels');
  });

  it('createLabel calls POST /flow/labels', () => {
    const data = { name: 'Urgent' };
    flowService.createLabel(data);
    expect(safe.post).toHaveBeenCalledWith('/flow/labels', data);
  });
});
