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
    labelClient: mockClient(),
    milestoneV2: mockClient(),
    cycleV2: mockClient(),
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
  labelClient: mockClients.labelClient,
  milestoneV2: mockClients.milestoneV2,
  cycleV2: mockClients.cycleV2,
  viewV2: mockClients.viewV2,
  insightsV2: mockClients.insightsV2,
  notificationClient: mockClients.notificationClient,
  settingClient: mockClients.settingClient,
  userClient: mockClients.userClient,
  projectV2: mockClients.projectV2,
  taskV2: mockClients.taskV2,
  labelV2: mockClients.labelV2,
}));

import { DEFAULT_PROJECT_ICON, flowApi } from './flowApi';
import {
  businessClient,
  milestoneV2,
  cycleV2,
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

  it('uses the generated business client', async () => {
    await flowApi.getMyBusinesses();

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
      icon: DEFAULT_PROJECT_ICON,
      color: '#6366f1',
      visibility: 'private',
      startDate: '',
      dueDate: '',
    });
    // The project sub-resources moved to v2 as well. They carry no workspace,
    // so unlike listProjects they were not broken on v1; they were simply the
    // last v1 flow calls in the write path.
    expect(projectV2.updateProject).toHaveBeenCalledWith({
      id: '7',
      body: {
        name: 'Renamed',
        description: '-',
        icon: DEFAULT_PROJECT_ICON,
        color: '#6366f1',
        status: 'active',
        startDate: '',
        dueDate: '',
      },
    });
    expect(projectV2.listProjectMembers).toHaveBeenCalledWith({ projectId: '7' });
    expect(projectV2.addProjectMember).toHaveBeenCalledWith({
      projectId: '7',
      userId: '2',
      role: 'member',
    });

    // No v1 assertions here any more: projectClient no longer exists in this app,
    // which is a stronger guarantee than asserting it went uncalled.
  });

  // deleteProject is gone: v2 has no such call, so project removal now archives
  // to the recycle bin. The assertion moved rather than being deleted, because
  // "which client does removal go through" is exactly what regressed before.
  it('routes project removal through v2 archive', async () => {
    await flowApi.archiveProject('7');

    expect(projectV2.archiveProject).toHaveBeenCalledWith({ id: '7' });
  });

  it('keeps task boundary conversions at the generated client edge', async () => {
    await Promise.all([
      flowApi.getTasks({ projectId: '7', status: 'todo' }),
      flowApi.createTask({ projectId: '7', title: 'Task' }),
      flowApi.updateTask('9', { archived: false, status: 'done' }),
      flowApi.updateTaskStatus('9', 'done'),
      flowApi.deleteTask('9'),
    ]);

    expect(taskV2.listTasks).toHaveBeenCalledWith({
      projectId: '7',
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
    expect(taskV2.updateTaskStatus).toHaveBeenCalledWith({ id: '9', body: { status: 'done' } });
    expect(taskV2.deleteTask).toHaveBeenCalledWith({ id: '9' });
  });

  it('uses generated label, planning, view, and insight clients', async () => {
    await Promise.all([
      flowApi.listLabels(),
      flowApi.createLabel({ name: 'Bug', color: '#f00' }),
      flowApi.setTaskLabels({ task_id: 9, label_ids: [2, 3] }),
      flowApi.getMilestones('7'),
      flowApi.createCycle({ name: 'Sprint', projectId: '7' }),
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
    // v2 lists business-wide; project_id is only an optional narrowing filter.
    expect(milestoneV2.listMilestones).toHaveBeenCalledWith({
      projectId: '7',
      page: 1,
      limit: 200,
      sortBy: '',
      sortDir: '',
    });
    // Cycles are project-scoped on v2: project_id is @required @gt(0), so there
    // is no business-wide cycle list and no workspace on the wire.
    expect(cycleV2.createCycle).toHaveBeenCalledWith({
      projectId: '7',
      name: 'Sprint',
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

  // This is the seam that was broken and untested. The S2 backend chain landed on
  // flow/view/v1 while this client is v2, so `visibility` was dropped at the wire
  // and the toggle did nothing. The test one layer up passed the whole time,
  // because it mocks flowApi itself and never checks what flowApi sends onward.
  it('sends visibility through to the v2 view client', async () => {
    await flowApi.updateView('7', { name: 'Mine', visibility: 'private' });

    expect(viewV2.updateView).toHaveBeenCalledWith(
      expect.objectContaining({ id: '7', name: 'Mine', visibility: 'private' })
    );
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

describe('flowApi.updateTask sends a genuine partial update', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // What onion receives is what survives JSON.stringify, and that drops
  // undefined. Asserting the serialised key set rather than the object is the
  // point of these tests: toHaveBeenCalledWith treats `{title: undefined}` and
  // `{}` as equal, which is why the previous `?? '-'` defaults were never
  // caught despite this file already calling updateTask.
  const sentKeys = (call = 0): string[] =>
    Object.keys(JSON.parse(JSON.stringify(vi.mocked(taskV2.updateTask).mock.calls[call][0].body)));

  const sentBody = (call = 0) => vi.mocked(taskV2.updateTask).mock.calls[call][0].body;

  it('omits every field the caller did not supply', async () => {
    await flowApi.updateTask('9', { status: 'done' });

    expect(sentKeys()).toEqual(['status']);
    expect(sentBody().status).toBe('done');
    expect(sentBody().title).toBeUndefined();
    expect(sentBody().description).toBeUndefined();
    expect(sentBody().dueDate).toBeUndefined();
    expect(sentBody().startDate).toBeUndefined();
  });

  it('sends an explicitly cleared title rather than dropping it', async () => {
    await flowApi.updateTask('9', { title: '' });

    expect(sentKeys()).toEqual(['title']);
    expect(sentBody().title).toBe('');
  });

  it('maps the no_priority UI value to the none the column accepts', async () => {
    await flowApi.updateTask('9', { priority: 'no_priority' });

    expect(sentKeys()).toEqual(['priority']);
    expect(sentBody().priority).toBe('none');
  });

  it('passes a real priority through untouched', async () => {
    await flowApi.updateTask('9', { priority: 'high' });

    expect(sentBody().priority).toBe('high');
  });

  it('keeps archived load bearing in both directions', async () => {
    await flowApi.updateTask('9', { archived: false });
    expect(sentKeys()).toEqual(['archived']);
    expect(sentBody().archived).toBe(false);

    vi.clearAllMocks();
    await flowApi.updateTask('9', { title: 'Renamed' });
    expect(sentKeys()).toEqual(['title']);
    expect(sentBody().archived).toBeUndefined();
  });

  it('maps no_priority on the create path too', async () => {
    await flowApi.createTask({ projectId: '7', title: 'Task', priority: 'no_priority' });

    expect(taskV2.createTask).toHaveBeenCalledWith(
      expect.objectContaining({ priority: 'none', title: 'Task' })
    );
  });

  it('forwards the per-user pin and effort fields on a task update', async () => {
    await flowApi.updateTask('9', { focusPin: true, effortMinutes: 30 });

    expect(sentKeys()).toEqual(['effortMinutes', 'focusPin']);
    expect(sentBody().focusPin).toBe(true);
    expect(sentBody().effortMinutes).toBe(30);
  });
});

describe('flowApi project icon writes (#1045)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // `flow_projects.icon` is read as a glyph by every reader in Flow and by the
  // backend's own seeds. Writing a lucide component name into it made the
  // project render as that word. In the dev database this had corrupted ten
  // rows: folder, rocket, globe, flask, building-2.
  const LUCIDE_NAMES = ['folder', 'rocket', 'globe', 'flask', 'building-2'];

  it('creates a project with a glyph, not a lucide component name', async () => {
    await flowApi.createProject({ name: 'Project A' });

    const sent = vi.mocked(projectV2.createProject).mock.calls[0][0] as { icon: string };
    expect(sent.icon).toBe(DEFAULT_PROJECT_ICON);
    expect(LUCIDE_NAMES).not.toContain(sent.icon);
  });

  it('updates a project with a glyph when the caller supplies no icon', async () => {
    await flowApi.updateProject('7', { name: 'Renamed' });

    const sent = vi.mocked(projectV2.updateProject).mock.calls[0][0] as {
      body: { icon: string };
    };
    expect(sent.body.icon).toBe(DEFAULT_PROJECT_ICON);
    expect(LUCIDE_NAMES).not.toContain(sent.body.icon);
  });

  it('passes an existing icon through untouched, whatever it is', async () => {
    // No data migration here: a row already holding a lucide name keeps it
    // until #1085 settles the convention.
    await flowApi.updateProject('7', { name: 'Renamed', icon: 'building-2' });

    const sent = vi.mocked(projectV2.updateProject).mock.calls[0][0] as {
      body: { icon: string };
    };
    expect(sent.body.icon).toBe('building-2');
  });

  it('uses the same glyph every reader falls back to', () => {
    // AllProjects.tsx and Profile.tsx both render `project.icon || '<glyph>'`.
    expect(DEFAULT_PROJECT_ICON).toBe('\u{1F4CB}');
  });
});

describe('flowApi.getTimeLogs coerces the int64 wire string (#1079)', () => {
  it('adds the durations instead of concatenating them', async () => {
    // OneKit puts an int64 on the wire as a STRING, while ApiTimeLog declares
    // `durationMinutes: number`. TypeScript believed the declaration, so
    // `logs.reduce((sum, l) => sum + l.durationMinutes, 0)` concatenated:
    // 2:00, 1:00, 2:15 and 1:30 rendered as "20100226:30 of 40:00" with the
    // bar full and an "Over estimate" badge. The correct total is 6:45.
    (taskV2.getTimeLogs as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      timeLogs: [
        { id: '1', taskId: '42', userId: '7', durationMinutes: '120' },
        { id: '2', taskId: '42', userId: '7', durationMinutes: '60' },
        { id: '3', taskId: '42', userId: '7', durationMinutes: '135' },
        { id: '4', taskId: '42', userId: '7', durationMinutes: '90' },
      ],
    });

    const logs = (await flowApi.getTimeLogs('42'))._unsafeUnwrap().timeLogs;

    for (const log of logs) {
      expect(typeof log.durationMinutes, 'each duration is a number').toBe('number');
    }
    expect(logs.reduce((sum, l) => sum + l.durationMinutes, 0)).toBe(405);
  });

  it('leaves an already numeric duration alone', async () => {
    // The generated client may be fixed later, or a mock may send numbers.
    (taskV2.getTimeLogs as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      timeLogs: [{ id: '1', taskId: '42', userId: '7', durationMinutes: 195 }],
    });

    const logs = (await flowApi.getTimeLogs('42'))._unsafeUnwrap().timeLogs;
    expect(logs[0].durationMinutes).toBe(195);
  });

  it('survives a response with no time logs', async () => {
    (taskV2.getTimeLogs as ReturnType<typeof vi.fn>).mockResolvedValueOnce({});
    expect((await flowApi.getTimeLogs('42'))._unsafeUnwrap().timeLogs).toEqual([]);
  });
});
