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
    viewClient: mockClient(),
    insightsClient: mockClient(),
    notificationClient: mockClient(),
    settingClient: mockClient(),
    userClient: mockClient(),
    projectV2: mockClient(),
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
  viewClient: mockClients.viewClient,
  insightsClient: mockClients.insightsClient,
  notificationClient: mockClients.notificationClient,
  settingClient: mockClients.settingClient,
  userClient: mockClients.userClient,
  projectV2: mockClients.projectV2,
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
  viewClient,
  insightsClient,
  notificationClient,
  settingClient,
  userClient,
  projectV2,
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
      flowApi.getProjects('ws-1'),
      flowApi.createProject({ name: 'Project A' }),
      flowApi.updateProject('7', { name: 'Renamed' }),
      flowApi.getProjectMembers('7'),
      flowApi.addProjectMember({ project_id: 7, user_id: 2, role: 'member' }),
    ]);

    expect(projectClient.listProjects).toHaveBeenCalledWith({ workspaceId: 'ws-1' });
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
      flowApi.createTask({ workspaceId: '1', projectId: '7', title: 'Task' }),
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
    expect(taskClient.createTask).toHaveBeenCalledWith({
      projectId: '7',
      workspaceId: '1',
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
      flowApi.listLabels('1'),
      flowApi.createLabel({ workspace_id: '1', name: 'Bug', color: '#f00' }),
      flowApi.setTaskLabels({ task_id: 9, label_ids: [2, 3] }),
      flowApi.getMilestones({ projectId: '7' }),
      flowApi.createCycle({ name: 'Sprint', workspaceId: '1' }),
      flowApi.getViews(),
      flowApi.getInsights('1'),
      flowApi.getTeamWorkload('1'),
      flowApi.getRoadmap('1'),
    ]);

    expect(labelClient.listLabels).toHaveBeenCalledWith({ workspaceId: '1' });
    expect(labelClient.createLabel).toHaveBeenCalledWith({
      workspaceId: '1',
      name: 'Bug',
      color: '#f00',
    });
    expect(labelClient.setTaskLabels).toHaveBeenCalledWith({
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
    expect(viewClient.listViews).toHaveBeenCalledWith({});
    expect(insightsClient.getInsights).toHaveBeenCalledWith({ workspaceId: '1' });
    expect(insightsClient.getTeamWorkload).toHaveBeenCalledWith({ workspaceId: '1' });
    expect(insightsClient.getRoadmap).toHaveBeenCalledWith({ workspaceId: '1' });
  });

  it('encodes saved-view filters to the wire JSON string and decodes them back', async () => {
    (viewClient.listViews as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      views: [{ id: '1', name: 'Mine', icon: 'folder', filters: '{"status":"todo"}' }],
    });

    const result = await flowApi.getViews('ws-1');

    expect(viewClient.listViews).toHaveBeenCalledWith({ workspaceId: 'ws-1' });
    expect(result._unsafeUnwrap().views[0].filters).toEqual({ status: 'todo' });

    await flowApi.createView({ workspaceId: 'ws-1', name: 'New', filters: { status: 'todo' } });
    expect(viewClient.createView).toHaveBeenCalledWith(
      expect.objectContaining({ filters: '{"status":"todo"}' })
    );
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
