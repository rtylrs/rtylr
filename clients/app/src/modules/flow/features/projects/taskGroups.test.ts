import { describe, expect, it } from 'vitest';
import type { Task } from '../../shared/stores/flowStore';
import { NO_EPIC } from '../../shared/lib/taskHierarchy';
import { NO_VALUE_LANE, epicLanes, laneKeyFor, laneKeysFor, valueLanes } from './taskGroups';

const task = (extra: Partial<Task> = {}): Task => ({
  id: '1',
  projectId: '2',
  title: 'Task',
  status: 'todo',
  priority: 'medium',
  assignee: '',
  assigneeInitials: '',
  assigneeId: '',
  dueDate: '',
  subtaskCount: 0,
  projectName: '',
  createdAt: '',
  updatedAt: '',
  identifier: 'T-1',
  ...extra,
});

const LABELS = [
  { id: 7, name: 'Design' },
  { id: 8, name: 'Backend' },
];

describe('laneKeysFor', () => {
  it('puts a card in the lane of every label it carries', () => {
    expect(laneKeysFor(task({ labelIds: ['8', '7'] }), 'label')).toEqual(['8', '7']);
  });

  it('puts a card with no label, cycle or milestone in the no-value lane', () => {
    expect(laneKeysFor(task({}), 'label')).toEqual([NO_VALUE_LANE]);
    expect(laneKeysFor(task({}), 'cycle')).toEqual([NO_VALUE_LANE]);
    expect(laneKeysFor(task({}), 'milestone')).toEqual([NO_VALUE_LANE]);
  });

  it('treats an id the project no longer has as no value', () => {
    const known = new Set(['7']);
    expect(laneKeysFor(task({ labelIds: ['99'] }), 'label', known)).toEqual([NO_VALUE_LANE]);
    expect(laneKeysFor(task({ labelIds: ['99', '7'] }), 'label', known)).toEqual(['7']);
  });

  it('keys cycles and milestones by id', () => {
    expect(laneKeysFor(task({ cycleId: '4' }), 'cycle')).toEqual(['4']);
    expect(laneKeysFor(task({ milestoneId: '5' }), 'milestone')).toEqual(['5']);
  });

  it('leaves the older groupings as they were', () => {
    expect(laneKeysFor(task({ assignee: 'Abed' }), 'assignee')).toEqual(['Abed']);
    expect(laneKeysFor(task({ priority: 'high' }), 'priority')).toEqual(['high']);
  });
});

describe('valueLanes', () => {
  it('lists the labels cards use, in label order, then No label', () => {
    const lanes = valueLanes(
      [task({ labelIds: ['8'] }), task({ labelIds: ['7'] }), task({})],
      'label',
      LABELS
    );
    expect(lanes).toEqual([
      { key: '7', label: 'Design' },
      { key: '8', label: 'Backend' },
      { key: NO_VALUE_LANE, labelKey: 'flow.board.noLabel' },
    ]);
  });

  it('leaves out lanes nobody uses', () => {
    expect(valueLanes([task({ cycleId: '4' })], 'cycle', [{ id: '4', name: 'Sprint 4' }])).toEqual([
      { key: '4', label: 'Sprint 4' },
    ]);
  });
});

describe('epic swimlanes', () => {
  const epic = task({ id: '10', taskType: 'epic', title: 'Payments', epicColor: '#22c55e' });
  const child = task({ id: '11', epicId: '10' });
  const loose = task({ id: '12' });

  it('gives each epic a lane, in board order, then No epic', () => {
    expect(epicLanes([loose, epic, child])).toEqual([
      { key: '10', label: 'Payments', color: '#22c55e' },
      { key: NO_EPIC, labelKey: 'flow.epics.noEpic' },
    ]);
  });

  it('puts an epic and its tasks in the epic lane and the rest in No epic', () => {
    expect(laneKeyFor(epic, 'epic')).toBe('10');
    expect(laneKeyFor(child, 'epic')).toBe('10');
    expect(laneKeyFor(loose, 'epic')).toBe(NO_EPIC);
  });

  it('still has a No epic lane on a board without epics', () => {
    expect(epicLanes([loose])).toEqual([{ key: NO_EPIC, labelKey: 'flow.epics.noEpic' }]);
  });
});
