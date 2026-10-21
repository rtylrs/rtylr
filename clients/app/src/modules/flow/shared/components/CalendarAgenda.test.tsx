import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, afterEach } from 'vitest';
import { CalendarAgenda } from './CalendarAgenda';
import { monthCells, tasksByLocalDay, agendaDays } from '../lib/monthGrid';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, vars?: Record<string, unknown>) =>
      vars && 'count' in vars ? `${key}:${String(vars.count)}` : key,
    i18n: { language: 'en' },
  }),
}));

interface Task {
  id: string;
  title: string;
  dueDate?: string | null;
  color?: string;
}

const build = (tasks: Task[], month = new Date(2026, 8, 15)) =>
  agendaDays(monthCells(month), tasksByLocalDay(tasks));

const renderAgenda = (tasks: Task[], onOpen = vi.fn()) => {
  const entries = build(tasks);
  render(
    <CalendarAgenda
      entries={entries}
      todayIso="2026-09-08"
      taskId={(t) => t.id}
      taskTitle={(t) => t.title}
      taskTone={(t) => t.color}
      onOpen={onOpen}
    />
  );
  return onOpen;
};

afterEach(() => {
  vi.clearAllMocks();
});

describe('the calendar agenda', () => {
  it('groups tasks under a heading per day', () => {
    renderAgenda([
      { id: '1', title: 'Grease trap', dueDate: '2026-09-08' },
      { id: '2', title: 'Invoice 4471', dueDate: '2026-09-08' },
      { id: '3', title: 'Rota', dueDate: '2026-09-20' },
    ]);

    expect(screen.getAllByRole('heading', { level: 3 })).toHaveLength(2);
    expect(screen.getByText('Grease trap')).toBeInTheDocument();
    expect(screen.getByText('Rota')).toBeInTheDocument();
  });

  it('says which heading is today', () => {
    renderAgenda([{ id: '1', title: 'Grease trap', dueDate: '2026-09-08' }]);
    // The heading for 8 September, which is the todayIso passed in.
    expect(screen.getByText('flow.calendar.today')).toBeInTheDocument();
  });

  it('shows only days that have tasks', () => {
    // Not 30 headings, most of them empty: that is the grid's problem in a
    // taller shape.
    renderAgenda([{ id: '1', title: 'Only one', dueDate: '2026-09-08' }]);
    expect(screen.getAllByRole('heading', { level: 3 })).toHaveLength(1);
  });

  it('says so when the month is empty rather than rendering nothing', () => {
    renderAgenda([]);
    expect(screen.getByText('flow.calendar.agendaEmpty')).toBeInTheDocument();
  });

  it('opens the task that was activated', () => {
    const onOpen = renderAgenda([
      { id: '1', title: 'Grease trap', dueDate: '2026-09-08' },
      { id: '2', title: 'Rota', dueDate: '2026-09-08' },
    ]);

    screen.getByText('Rota').closest('button')!.click();
    expect(onOpen).toHaveBeenCalledTimes(1);
    expect(onOpen.mock.calls[0][0]).toMatchObject({ id: '2' });
  });

  it('gives every row a real button, not a clickable div', () => {
    renderAgenda([{ id: '1', title: 'Grease trap', dueDate: '2026-09-08' }]);
    const row = screen.getByText('Grease trap').closest('button');
    expect(row).not.toBeNull();
    expect(row).toHaveAttribute('type', 'button');
  });

  it('meets the 44px touch minimum rather than the mock 40px', () => {
    // S-04 specifies a 40px row, but there the row is not the tap target. Here
    // the whole row is the button, so it takes the 44px minimum instead.
    renderAgenda([{ id: '1', title: 'Grease trap', dueDate: '2026-09-08' }]);
    expect(screen.getByText('Grease trap').closest('button')!.className).toContain('h-11');
  });

  it('lets a right-to-left title choose its own truncation end', () => {
    renderAgenda([{ id: '1', title: 'مهمة طويلة جدا', dueDate: '2026-09-08' }]);
    expect(screen.getByText('مهمة طويلة جدا')).toHaveAttribute('dir', 'auto');
  });

  it('carries the tint cue the grid pills had', () => {
    renderAgenda([{ id: '1', title: 'Grease trap', dueDate: '2026-09-08', color: '#112233' }]);
    const title = screen.getByText('Grease trap');
    expect(title.style.getPropertyValue('--tone')).toBe('#112233');
  });

  it('leaves a task with no colour untinted rather than painting undefined', () => {
    renderAgenda([{ id: '1', title: 'Grease trap', dueDate: '2026-09-08' }]);
    const title = screen.getByText('Grease trap');
    expect(title.getAttribute('style') ?? '').not.toContain('undefined');
  });
});
