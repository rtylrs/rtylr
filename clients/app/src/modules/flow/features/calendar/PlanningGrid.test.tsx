// @vitest-environment jsdom
import React from 'react';
import { render, screen, fireEvent, within } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { weekCells } from '../../shared/lib/calendarSpans';
import { monthCells } from '../../shared/lib/monthGrid';
import { PlanningGrid } from './PlanningGrid';
import { UnscheduledPanel } from './UnscheduledPanel';
import { useCalendarScheduling } from './useCalendarScheduling';
import type { LeaveDay, PlanningCycle, PlanningMilestone } from './useCalendarPlanning';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) =>
      opts ? `${key} ${JSON.stringify(opts)}` : key,
    i18n: { language: 'en-US' },
  }),
}));

const updateMutate = vi.hoisted(() => vi.fn());
vi.mock('../tasks/useTasks', () => ({
  useUpdateTask: () => ({ mutate: updateMutate }),
}));

interface Task {
  id: string;
  title: string;
  projectId: string;
  dueDate?: string;
  startDate?: string;
  status?: string;
}

// The week of Thursday 19 November 2026: Sunday 15 to Saturday 21.
const WEEK = weekCells(new Date(2026, 10, 19));

function Harness({
  tasks,
  editable = () => true,
  cycles = [],
  milestones = [],
  leave = new Map(),
  cells = WEEK,
  view = 'week',
}: {
  tasks: Task[];
  editable?: (task: Task) => boolean;
  cycles?: PlanningCycle[];
  milestones?: PlanningMilestone[];
  leave?: Map<string, LeaveDay>;
  cells?: typeof WEEK;
  view?: 'week' | 'month';
}) {
  const scheduling = useCalendarScheduling(tasks);
  return (
    <>
      <PlanningGrid
        cells={cells}
        view={view}
        tasks={scheduling.tasks}
        cycles={cycles}
        milestones={milestones}
        leave={leave}
        today="2026-11-19"
        tone={() => '#112233'}
        canEdit={editable}
        onOpen={vi.fn()}
        scheduling={scheduling}
      />
      <UnscheduledPanel
        tasks={scheduling.tasks}
        tone={() => undefined}
        canEdit={editable}
        onOpen={vi.fn()}
        onClose={vi.fn()}
        scheduling={scheduling}
        layout="side"
      />
    </>
  );
}

const day = (container: HTMLElement, iso: string): HTMLElement =>
  container.querySelector(`[data-day="${iso}"]`) as HTMLElement;

const barOf = (title: string): HTMLElement => screen.getByTitle(title);

beforeEach(() => {
  updateMutate.mockReset();
});

describe('PlanningGrid bars', () => {
  it('runs a task from its start date to its due date', () => {
    render(
      <Harness
        tasks={[
          {
            id: 't1',
            title: 'Launch',
            projectId: 'p',
            startDate: '2026-11-16',
            dueDate: '2026-11-18',
          },
        ]}
      />
    );
    // Monday is the second column; three days wide.
    expect(barOf('Launch').parentElement?.style.gridColumn).toBe('2 / span 3');
  });

  it('carries a bar that started last week without its start edge', () => {
    render(
      <Harness
        tasks={[
          {
            id: 't1',
            title: 'Carried over',
            projectId: 'p',
            startDate: '2026-11-10',
            dueDate: '2026-11-16',
          },
        ]}
      />
    );
    const wrapper = barOf('Carried over').parentElement!;
    expect(wrapper.style.gridColumn).toBe('1 / span 2');
    expect(wrapper.querySelector('[data-edge="start"]')).toBeNull();
    expect(wrapper.querySelector('[data-edge="end"]')).not.toBeNull();
  });

  it('draws cycles as bars and milestones on their day', () => {
    const { container } = render(
      <Harness
        tasks={[]}
        cycles={[
          {
            id: 'c1',
            projectId: 'p',
            name: 'Sprint 4',
            range: { start: '2026-11-09', end: '2026-11-20' },
            status: 'active',
          },
        ]}
        milestones={[{ id: 'm1', name: 'Beta', day: '2026-11-19', status: 'at_risk' }]}
      />
    );
    const cycle = container.querySelector('[data-cycle="c1"]') as HTMLElement;
    expect(cycle).toHaveAttribute('title', 'flow.calendar.cycleTitle {"name":"Sprint 4"}');
    expect(cycle.parentElement?.style.gridColumn).toBe('1 / span 6');
    const milestone = container.querySelector('[data-milestone="m1"]') as HTMLElement;
    expect(milestone.parentElement?.style.gridColumn).toBe('5 / span 1');
    expect(milestone.style.getPropertyValue('--tone')).toContain('flow-warn-fg');
  });

  it('shades a day with approved leave and says how many are away', () => {
    const { container } = render(
      <Harness tasks={[]} leave={new Map([['2026-11-17', { people: 2 }]])} />
    );
    const tuesday = day(container, '2026-11-17');
    expect(tuesday).toHaveAttribute('data-leave', '2');
    expect(within(tuesday).getByText('flow.calendar.leaveCount {"count":2}')).toBeInTheDocument();
    expect(day(container, '2026-11-18')).not.toHaveAttribute('data-leave');
  });

  it('folds bars past three lanes into a count on a month', () => {
    render(
      <Harness
        view="month"
        cells={monthCells(new Date(2026, 10, 1))}
        tasks={Array.from({ length: 5 }, (_, index) => ({
          id: `t${index}`,
          title: `Task ${index}`,
          projectId: 'p',
          dueDate: '2026-11-19',
        }))}
      />
    );
    expect(screen.getByText('flow.calendar.moreCount {"count":2}')).toBeInTheDocument();
  });
});

describe('PlanningGrid drag to reschedule', () => {
  it('moves a due-only task to the day it is dropped on', () => {
    const { container } = render(
      <Harness tasks={[{ id: 't1', title: 'Move me', projectId: 'p', dueDate: '2026-11-16' }]} />
    );
    fireEvent.dragStart(barOf('Move me'));
    fireEvent.dragOver(day(container, '2026-11-19'));
    fireEvent.drop(day(container, '2026-11-19'));

    expect(updateMutate).toHaveBeenCalledTimes(1);
    expect(updateMutate.mock.calls[0][0]).toEqual({ id: 't1', data: { dueDate: '2026-11-19' } });
    // Drawn on the new day before the server answers.
    expect(barOf('Move me').parentElement?.style.gridColumn).toBe('5 / span 1');
  });

  it('drags the end edge to stretch a task into a range', () => {
    const { container } = render(
      <Harness tasks={[{ id: 't1', title: 'Stretch', projectId: 'p', dueDate: '2026-11-16' }]} />
    );
    const edge = barOf('Stretch').parentElement!.querySelector('[data-edge="end"]')!;
    fireEvent.dragStart(edge);
    fireEvent.drop(day(container, '2026-11-18'));

    expect(updateMutate.mock.calls[0][0]).toEqual({
      id: 't1',
      data: { startDate: '2026-11-16', dueDate: '2026-11-18' },
    });
  });

  it('drags the start edge of a range to change only its start', () => {
    const { container } = render(
      <Harness
        tasks={[
          {
            id: 't1',
            title: 'Range',
            projectId: 'p',
            startDate: '2026-11-17',
            dueDate: '2026-11-19',
          },
        ]}
      />
    );
    const edge = barOf('Range').parentElement!.querySelector('[data-edge="start"]')!;
    fireEvent.dragStart(edge);
    fireEvent.drop(day(container, '2026-11-15'));

    expect(updateMutate.mock.calls[0][0]).toEqual({
      id: 't1',
      data: { startDate: '2026-11-15' },
    });
  });

  it('sets the due date of an unscheduled task dropped on a day', () => {
    const { container } = render(
      <Harness tasks={[{ id: 'u1', title: 'Someday', projectId: 'p' }]} />
    );
    const item = screen.getByTitle('Someday');
    expect(item).toHaveAttribute('draggable', 'true');
    fireEvent.dragStart(item);
    fireEvent.drop(day(container, '2026-11-20'));

    expect(updateMutate.mock.calls[0][0]).toEqual({ id: 'u1', data: { dueDate: '2026-11-20' } });
    // Scheduled now, so it leaves the panel for the grid.
    expect(screen.getByTitle('Someday').closest('aside')).toBeNull();
  });

  it('puts the dates back if the update fails', () => {
    updateMutate.mockImplementation((_vars, options: { onError: () => void }) => options.onError());
    const { container } = render(
      <Harness tasks={[{ id: 't1', title: 'Refused', projectId: 'p', dueDate: '2026-11-16' }]} />
    );
    fireEvent.dragStart(barOf('Refused'));
    fireEvent.drop(day(container, '2026-11-19'));
    expect(barOf('Refused').parentElement?.style.gridColumn).toBe('2 / span 1');
  });

  it('ignores a drop with no drag from the calendar', () => {
    const { container } = render(
      <Harness tasks={[{ id: 't1', title: 'Still', projectId: 'p', dueDate: '2026-11-16' }]} />
    );
    fireEvent.drop(day(container, '2026-11-19'));
    expect(updateMutate).not.toHaveBeenCalled();
  });
});

describe('PlanningGrid for someone who cannot edit', () => {
  it('lets a viewer open tasks but not drag them or their edges', () => {
    render(
      <Harness
        editable={() => false}
        tasks={[
          { id: 't1', title: 'Look only', projectId: 'p', dueDate: '2026-11-16' },
          { id: 'u1', title: 'Unscheduled look', projectId: 'p' },
        ]}
      />
    );
    const bar = barOf('Look only');
    expect(bar).toHaveAttribute('draggable', 'false');
    expect(bar.parentElement!.querySelector('[data-edge]')).toBeNull();
    expect(screen.getByTitle('Unscheduled look')).toHaveAttribute('draggable', 'false');
    expect(screen.getByText('flow.calendar.unscheduled.readOnly')).toBeInTheDocument();
  });
});

describe('UnscheduledPanel', () => {
  it('lists only open tasks with no dates, and searches them by title', () => {
    render(
      <Harness
        tasks={[
          { id: 'u1', title: 'Write docs', projectId: 'p' },
          { id: 'u2', title: 'Fix login', projectId: 'p' },
          { id: 'u3', title: 'Done already', projectId: 'p', status: 'done' },
          { id: 't1', title: 'Dated', projectId: 'p', dueDate: '2026-11-16' },
        ]}
      />
    );
    const panel = screen.getByRole('complementary', { name: 'flow.calendar.unscheduled.title' });
    expect(within(panel).getByTitle('Write docs')).toBeInTheDocument();
    expect(within(panel).getByTitle('Fix login')).toBeInTheDocument();
    expect(within(panel).queryByTitle('Done already')).toBeNull();
    expect(within(panel).queryByTitle('Dated')).toBeNull();

    fireEvent.change(within(panel).getByRole('searchbox'), { target: { value: 'LOGIN' } });
    expect(within(panel).queryByTitle('Write docs')).toBeNull();
    expect(within(panel).getByTitle('Fix login')).toBeInTheDocument();

    fireEvent.change(within(panel).getByRole('searchbox'), { target: { value: 'nothing' } });
    expect(within(panel).getByText('flow.calendar.unscheduled.noMatches')).toBeInTheDocument();
  });
});
