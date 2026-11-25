import { describe, expect, it } from 'vitest';
import {
  buildBars,
  dependencyArrows,
  flattenRows,
  groupBars,
  taskSpan,
  timelineWindow,
  type TimelineTask,
} from './timelineModel';

const task = (over: Partial<TimelineTask>): TimelineTask => ({
  id: '1',
  title: 'Task',
  done: false,
  ...over,
});

describe('taskSpan', () => {
  it('runs from start to due, or covers the one date there is', () => {
    expect(taskSpan(task({ startDate: '2026-11-20', dueDate: '2026-11-25' }))).toEqual({
      start: '2026-11-20',
      end: '2026-11-25',
    });
    expect(taskSpan(task({ dueDate: '2026-11-25T00:00:00Z' }))).toEqual({
      start: '2026-11-25',
      end: '2026-11-25',
    });
    expect(taskSpan(task({ startDate: '2026-11-26', dueDate: '2026-11-25' }))).toEqual({
      start: '2026-11-25',
      end: '2026-11-26',
    });
    expect(taskSpan(task({}))).toBeNull();
  });
});

describe('timelineWindow', () => {
  it('covers every bar with some room, and at least four weeks', () => {
    const days = timelineWindow(
      [task({ startDate: '2026-11-01', dueDate: '2026-12-31' })],
      '2026-11-25'
    );
    expect(days[0]).toBe('2026-10-30');
    expect(days[days.length - 1]).toBe('2027-01-07');
    expect(timelineWindow([], '2026-11-25')).toHaveLength(28);
  });
});

describe('rows and arrows', () => {
  const tasks = [
    task({
      id: 'a',
      title: 'Design',
      startDate: '2026-11-20',
      dueDate: '2026-11-22',
      assigneeId: '7',
    }),
    task({
      id: 'b',
      title: 'Build',
      startDate: '2026-11-23',
      dueDate: '2026-11-28',
      assigneeId: '8',
      dependencies: ['a', 'zz'],
    }),
    task({ id: 'c', title: 'Docs', dueDate: '2026-11-24' }),
  ];

  it('places bars from the window start and groups unassigned work last', () => {
    const bars = buildBars(tasks, '2026-11-18');
    expect(bars.map((b) => [b.task.id, b.offset, b.span])).toEqual([
      ['a', 2, 3],
      ['b', 5, 6],
      ['c', 6, 1],
    ]);
    const groups = groupBars(
      bars,
      'assignee',
      new Map([
        ['7', 'Zed'],
        ['8', 'Amal'],
      ]),
      'Unassigned'
    );
    expect(groups.map((g) => [g.label, g.bars.length])).toEqual([
      ['Amal', 1],
      ['Zed', 1],
      ['Unassigned', 1],
    ]);
  });

  it('draws an arrow only between bars on screen', () => {
    const rows = flattenRows(
      groupBars(buildBars(tasks, '2026-11-18'), 'none', new Map(), ''),
      false
    );
    expect(dependencyArrows(rows)).toEqual([
      { fromTaskId: 'a', toTaskId: 'b', fromRow: 0, toRow: 1, fromDay: 5, toDay: 5 },
    ]);
  });
});
