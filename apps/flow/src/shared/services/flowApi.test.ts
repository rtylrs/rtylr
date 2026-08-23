import { beforeEach, describe, expect, it, vi } from 'vitest';

const mockClients = vi.hoisted(() => {
  const mockClient = (): Record<string, ReturnType<typeof vi.fn>> => {
    const target: Record<string, ReturnType<typeof vi.fn>> = {};
    return new Proxy(target, {
      get: (currentTarget, property: string) =>
        (currentTarget[property] ??= vi.fn(() => Promise.resolve({}))),
    });
  };

  return {
    businessClient: mockClient(),
    dashboardClient: mockClient(),
    workspaceClient: mockClient(),
    projectClient: mockClient(),
    taskClient: mockClient(),
    labelClient: mockClient(),
    milestoneClient: mockClient(),
    cycleClient: mockClient(),
    viewV2: mockClient(),
    insightsV2: mockClient(),
    notificationClient: mockClient(),
    settingClient: mockClient(),
    userClient: mockClient(),
    projectV2: mockClient(),
    taskV2: mockClient(),
    labelV2: mockClient(),
  };
});

vi.mock('../lib/clients', () => ({
  businessClient: mockClients.businessClient,
  dashboardClient: mockClients.dashboardClient,
  workspaceClient: mockClients.workspaceClient,
  projectClient: mockClients.projectClient,
  taskClient: mockClients.taskClient,
  labelClient: mockClients.labelClient,
  milestoneClient: mockClients.milestoneClient,
  cycleClient: mockClients.cycleClient,
  viewV2: mockClients.viewV2,
  insightsV2: mockClients.insightsV2,
  notificationClient: mockClients.notificationClient,
  settingClient: mockClients.settingClient,
  userClient: mockClients.userClient,
  projectV2: mockClients.projectV2,
  taskV2: mockClients.taskV2,
  labelV2: mockClients.labelV2,
}));

import { flowApi } from './flowApi';
import {
  businessClient,
  dashboardClient,
  projectClient,
  taskClient,
  labelClient,
  milestoneClient,
  cycleClient,
  insightsV2,
  notificationClient,
  settingClient,
  userClient,
  projectV2,
  taskV2,
  labelV2,
  viewV2,
} from '../lib/clients';

describe('flowApi OneKit adapters', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('uses the generated dashboard and business clients', async () => {
    await Promise.all([flowApi.getDashboard('ws-1'), flowApi.getMyBusinesses()]);

    expect(dashboardClient.getFlowDashboard).toHaveBeenCalledWith({ workspaceId: 'ws-1' });
    expect(businessClient.listMyBusinesses).toHaveBeenCalledWith({});
  });

  it('uses generated project and member methods', async () => {
    await Promise.all([
      flowApi.getProjects(),
      flowApi.createProject({ name: 'Project A' }),
      flowApi.updateProject('7', { name: 'Renamed' }),
      flowApi.getProjectMembers('7'),
      flowApi.addProjectMember({ project_id: 7, user_id: 2, role: 'member' }),
    ]);

    // The sidebar list is business-scoped and must be v2: a v2-created project
    // is stored with workspace_id 0, and v1's listProjects filters on
    // workspace_id, so a v1 list would never show a newly created project.
    expect(projectV2.listProjects).toHaveBeenCalledWith({
      status: '',
      visibility: '',
      q: '',
      page: 1,
      limit: 100,
    });
    // v2: no workspace id, and description keeps the '-' sentinel because
    // CreateProjectRequest declares it @required and the generated validator
    // rejects an empty string.
    expect(projectV2.createProject).toHaveBeenCalledWith({
      name: 'Project A',
      description: '-',
      icon: 'folder',
      color: '#6366f1',
      visibility: 'private',
      startDate: '',
      dueDate: '',
    });
    expect(projectClient.updateProject).toHaveBeenCalledWith({
      id: '7',
      body: {
        name: 'Renamed',
        description: '-',
        icon: 'folder',
        color: '#6366f1',
        status: 'active',
        startDate: '',
        dueDate: '',
      },
    });
    expect(projectClient.listProjectMembers).toHaveBeenCalledWith({ projectId: '7' });
    expect(projectClient.addProjectMember).toHaveBeenCalledWith({
      projectId: '7',
      userId: '2',
      role: 'member',
    });
  });

  it('keeps task boundary conversions at the generated client edge', async () => {
    await Promise.all([
      flowApi.getTasks({ projectId: '7', workspaceId: '1', status: 'todo' }),
      flowApi.createTask({ projectId: '7', title: 'Task' }),
      flowApi.updateTask('9', { archived: false, status: 'done' }),
      flowApi.updateTaskStatus('9', 'done'),
      flowApi.deleteTask('9'),
    ]);

    expect(taskClient.listTasks).toHaveBeenCalledWith({
      projectId: '7',
      workspaceId: '1',
      status: 'todo',
      assigneeId: undefined,
    });
    // v2 create carries no workspace_id.
    expect(taskV2.createTask).toHaveBeenCalledWith({
      projectId: '7',
      title: 'Task',
      description: '-',
      priority: 'medium',
      assigneeId: undefined,
      dueDate: '-',
      startDate: '-',
      estimatedHours: undefined,
      labels: undefined,
    });
    expect(taskClient.updateTaskStatus).toHaveBeenCalledWith({ id: '9', body: { status: 'done' } });
    expect(taskClient.deleteTask).toHaveBeenCalledWith({ id: '9' });
  });

  it('uses generated label, planning, view, and insight clients', async () => {
    await Promise.all([
      flowApi.listLabels(),
      flowApi.createLabel({ name: 'Bug', color: '#f00' }),
      flowApi.setTaskLabels({ task_id: 9, label_ids: [2, 3] }),
      flowApi.getMilestones({ projectId: '7' }),
      flowApi.createCycle({ name: 'Sprint', workspaceId: '1' }),
      flowApi.getViews(),
      flowApi.getInsights(),
      flowApi.getTeamWorkload(),
      flowApi.getRoadmap(),
    ]);

    // Labels are one business-wide pool on v2, so no workspace is sent. v1's
    // createLabel declared workspace_id @required @gt(0), which is why creating
    // a label 400d once the workspace layer went away.
    expect(labelV2.listLabels).toHaveBeenCalledWith({
      page: 1,
      limit: 200,
      sortBy: '',
      sortDir: '',
    });
    expect(labelV2.createLabel).toHaveBeenCalledWith({ name: 'Bug', color: '#f00' });
    expect(labelV2.setTaskLabels).toHaveBeenCalledWith({
      taskId: '9',
      labelIds: ['2', '3'],
    });
    expect(milestoneClient.listMilestones).toHaveBeenCalledWith({
      projectId: '7',
      workspaceId: undefined,
    });
    expect(cycleClient.createCycle).toHaveBeenCalledWith({
      name: 'Sprint',
      workspaceId: '1',
      startDate: undefined,
      endDate: undefined,
    });
    expect(viewV2.listViews).toHaveBeenCalledWith({ page: 1, limit: 200 });
    // Insights, team workload and the roadmap are business-wide on v2 and take
    // no parameters at all.
    expect(insightsV2.getInsights).toHaveBeenCalledWith({});
    expect(insightsV2.getTeamWorkload).toHaveBeenCalledWith({});
    expect(insightsV2.getRoadmap).toHaveBeenCalledWith({});
  });

  it('encodes saved-view filters to the wire JSON string and decodes them back', async () => {
    (viewV2.listViews as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      views: [{ id: '1', name: 'Mine', icon: 'folder', filters: '{"status":"todo"}' }],
    });

    const result = await flowApi.getViews();

    expect(result._unsafeUnwrap().views[0].filters).toEqual({ status: 'todo' });

    await flowApi.createView({ name: 'New', filters: { status: 'todo' } });
    // The wire carries filters as a JSON string; the object shape is app-side
    // only. Also asserts no workspace rides along on v2.
    expect(viewV2.createView).toHaveBeenCalledWith({
      name: 'New',
      icon: undefined,
      filters: '{"status":"todo"}',
      groupBy: undefined,
      sortBy: undefined,
      sortDir: undefined,
    });
  });

  it('uses generated notification, settings, and profile clients', async () => {
    await Promise.all([
      flowApi.getNotifications(),
      flowApi.markNotificationRead('n-1'),
      flowApi.markAllNotificationsRead(),
      flowApi.getUserSettings(),
      flowApi.updateProfile({ name: 'Ada', email: 'ada@example.com' }),
      flowApi.changePassword({ currentPassword: 'old', newPassword: 'new' }),
    ]);

    expect(notificationClient.listNotifications).toHaveBeenCalledWith({});
    expect(notificationClient.markNotificationRead).toHaveBeenCalledWith({ id: 'n-1' });
    expect(notificationClient.markAllNotificationsRead).toHaveBeenCalledWith({});
    expect(settingClient.getUserSettings).toHaveBeenCalledWith({});
    expect(userClient.updateProfile).toHaveBeenCalledWith({
      name: 'Ada',
      email: 'ada@example.com',
    });
    expect(userClient.changePassword).toHaveBeenCalledWith({
      currentPassword: 'old',
      newPassword: 'new',
    });
  });
});
