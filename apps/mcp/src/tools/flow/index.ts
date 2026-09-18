export { listWorkspaces, createWorkspace } from './workspaces.js';
export { listProjects, getProject, createProject, updateProject, deleteProject } from './projects.js';
export { listTasks, getTask, createTask, updateTask, deleteTask, updateTaskStatus } from './tasks.js';
export { myTasks } from './my-tasks.js';
export { listComments, createComment } from './comments.js';

import { listWorkspaces, createWorkspace } from './workspaces.js';
import { listProjects, getProject, createProject, updateProject, deleteProject } from './projects.js';
import { listTasks, getTask, createTask, updateTask, deleteTask, updateTaskStatus } from './tasks.js';
import { myTasks } from './my-tasks.js';
import { listComments, createComment } from './comments.js';

export const flowTools = [
  listWorkspaces, createWorkspace,
  listProjects, getProject, createProject, updateProject, deleteProject,
  listTasks, getTask, createTask, updateTask, deleteTask, updateTaskStatus,
  myTasks,
  listComments, createComment,
];
