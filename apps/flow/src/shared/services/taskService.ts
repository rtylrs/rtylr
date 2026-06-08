import { safe } from '../lib/api';
import type {
  ApiTask,
  ApiTaskDetail,
  ApiTaskActivity,
  ApiTaskComment,
  ApiSubtask,
  ApiAttachment,
  ApiTimeLog,
} from './flowApi';

export const taskService = {
  getTasks: (params: { projectId?: string; status?: ApiTask['status']; assigneeId?: string }) => {
    const q: Record<string, string> = {};
    if (params.projectId) q.project_id = params.projectId;
    if (params.status) q.status = params.status;
    if (params.assigneeId) q.assignee_id = params.assigneeId;
    return safe.get<ApiTask[]>('/flow/tasks', { params: q });
  },

  getMyTasks: () => safe.get<ApiTask[]>('/flow/tasks/my'),

  getDetail: (id: string) => safe.get<ApiTaskDetail>('/flow/tasks/detail', { params: { id } }),

  create: (data: {
    projectId: string;
    workspaceId: string;
    title: string;
    status?: ApiTask['status'];
    priority?: ApiTask['priority'];
    assigneeId?: string;
    dueDate?: string;
  }) =>
    safe.post<ApiTask>('/flow/tasks', {
      projectId: data.projectId,
      workspaceId: data.workspaceId,
      title: data.title,
      priority: data.priority,
      assigneeId: data.assigneeId,
      dueDate: data.dueDate,
    }),

  update: (id: string, data: Partial<ApiTask>) =>
    safe.put<ApiTask>('/flow/tasks', { id: parseInt(id), ...data }),

  delete: (id: string) => safe.delete('/flow/tasks', { data: { id: parseInt(id) } }),

  updateStatus: (id: string, status: ApiTask['status']) =>
    safe.put('/flow/tasks/status', { id: parseInt(id), status }),

  addComment: (data: { taskId: string; content: string }) =>
    safe.post<ApiTaskComment>('/flow/tasks/comment', data),

  deleteComment: (id: string) => safe.delete('/flow/tasks/comment', { data: { id: parseInt(id) } }),

  getActivity: (id: string) => safe.get<ApiTaskActivity[]>(`/flow/tasks/${id}/activity`),

  getSubtasks: (id: string) => safe.get<ApiSubtask[]>(`/flow/tasks/${id}/subtasks`),

  createSubtask: (taskId: string, data: { title: string; assigneeId?: string }) =>
    safe.post<ApiSubtask>(`/flow/tasks/${taskId}/subtasks`, data),

  getAttachments: (id: string) => safe.get<ApiAttachment[]>(`/flow/tasks/${id}/attachments`),

  getTimeLogs: (id: string) => safe.get<ApiTimeLog[]>(`/flow/tasks/${id}/time`),

  logTime: (id: string, data: { durationMinutes: number; description?: string }) =>
    safe.post<ApiTimeLog>(`/flow/tasks/${id}/time`, data),
};
