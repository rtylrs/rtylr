import { useFlowStore } from '../stores/flowStore';
import { isFlowLocal } from './flowLocalMode';
import {
  SEED_WORKSPACES,
  SEED_MEMBERS,
  SEED_PROJECTS,
  SEED_TASKS,
  SEED_SECTIONS,
  SEED_CUSTOM_FIELDS,
  SEED_LABELS,
  SEED_VIEWS,
  SEED_MILESTONES,
  SEED_CYCLES,
  SEED_COMMENTS,
  SEED_SUBTASKS,
  SEED_ACTIVITY,
  SEED_ATTACHMENTS,
  SEED_NOTIFICATIONS,
  SEED_TIMELOGS,
} from '../data/seedFlowStore';

export function initFlowLocalData(): void {
  if (!isFlowLocal()) return;
  const s = useFlowStore.getState();
  if (s.workspaces.length > 0) return;

  useFlowStore.setState({
    workspaces: SEED_WORKSPACES,
    members: SEED_MEMBERS,
    projects: SEED_PROJECTS,
    tasks: SEED_TASKS,
    sections: SEED_SECTIONS,
    customFieldDefinitions: SEED_CUSTOM_FIELDS,
    labels: SEED_LABELS,
    views: SEED_VIEWS,
    milestones: SEED_MILESTONES,
    cycles: SEED_CYCLES,
    comments: SEED_COMMENTS,
    subtasks: SEED_SUBTASKS,
    activity: SEED_ACTIVITY,
    attachments: SEED_ATTACHMENTS,
    notifications: SEED_NOTIFICATIONS,
    timelogs: SEED_TIMELOGS,
    templates: [
      {
        id: 'tpl-bug',
        name: 'Bug report',
        title: '[Bug] ',
        description: 'Steps to reproduce:\n1.\n2.\n\nExpected:\nActual:',
        priority: 'high',
        status: 'todo',
        labels: ['lbl-tech'],
        subtasks: ['Reproduce', 'Fix', 'Verify'],
        estimate: 2,
      },
      {
        id: 'tpl-feature',
        name: 'Feature',
        title: '',
        description: 'User story:\n\nAcceptance criteria:',
        priority: 'medium',
        status: 'todo',
        estimate: 4,
      },
    ],
    activeWorkspaceId: 'personal',
  });
}
