export type TaskStatus = 'todo' | 'in_progress' | 'in_review' | 'done';
export type TaskPriority = 'low' | 'medium' | 'high' | 'urgent';

export interface Task {
  id: string;
  title: string;
  description?: string;
  status: TaskStatus;
  priority: TaskPriority;
  assigneeId?: string;
  projectId: string;
  parentTaskId?: string;
  subtasks?: Task[];
  dueDate?: string;
  tags?: string[];
  createdAt: string;
}

export interface Project {
  id: string;
  name: string;
  description?: string;
  workspaceId: string;
  color?: string;
  icon?: string;
  members?: string[];
  taskCount?: number;
  createdAt: string;
}

export interface Workspace {
  id: string;
  name: string;
  description?: string;
  ownerId: string;
  members: { userId: string; role: 'admin' | 'member' | 'viewer' }[];
  projects?: Project[];
  createdAt: string;
}
