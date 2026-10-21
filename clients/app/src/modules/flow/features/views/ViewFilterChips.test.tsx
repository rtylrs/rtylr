import fs from 'node:fs';
import path from 'node:path';
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

import ViewFilterChips, {
  activeFilterChips,
  hasActiveFilters,
  EMPTY_FILTERS,
} from './ViewFilterChips';

const t = (key: string): string => key;

describe('activeFilterChips', () => {
  it('gives no chips for a view that narrows nothing', () => {
    expect(activeFilterChips({}, t)).toEqual([]);
    expect(hasActiveFilters({})).toBe(false);
    expect(hasActiveFilters(undefined)).toBe(false);
  });

  it('gives one chip per status and per priority value', () => {
    const chips = activeFilterChips({ status: ['todo', 'done'], priority: ['urgent'] }, t);
    expect(chips.map((c) => c.id)).toEqual(['status:todo', 'status:done', 'priority:urgent']);
  });

  it('names statuses the locale files were missing', () => {
    // flowStatus.json shipped four of six: backlog and in_review had no entry.
    const chips = activeFilterChips({ status: ['backlog', 'in_review'] }, t);
    expect(chips.map((c) => c.label)).toEqual(['flowStatus.backlog', 'flowStatus.inReview']);
  });

  it('uses the camelCase key for a snake_case status', () => {
    const chips = activeFilterChips({ status: ['in_progress'] }, t);
    expect(chips[0].label).toBe('flowStatus.inProgress');
  });

  it('removing a chip leaves the other values of the same filter alone', () => {
    const chips = activeFilterChips({ status: ['todo', 'done', 'backlog'] }, t);
    const removeDone = chips.find((c) => c.id === 'status:done');
    expect(removeDone?.next.status).toEqual(['todo', 'backlog']);
  });

  it('removing a chip leaves the other filters alone', () => {
    const chips = activeFilterChips({ status: ['todo'], priority: ['high'], projectId: 'p1' }, t);
    const removeStatus = chips.find((c) => c.id === 'status:todo');
    expect(removeStatus?.next.priority).toEqual(['high']);
    expect(removeStatus?.next.projectId).toBe('p1');
  });

  it('gives one chip for each id filter rather than showing a raw id', () => {
    const chips = activeFilterChips({ assigneeId: 'u9', projectId: 'p3', labels: ['l1'] }, t);
    expect(chips.map((c) => c.label)).toEqual([
      'flow.views.filters.assignee',
      'flow.views.filters.project',
      'flow.views.filters.labels',
    ]);
    // The label is what the user reads, and it must not be a raw id. `next`
    // deliberately still carries the other filters' ids: it is the state the
    // view would have once this one chip is gone.
    for (const chip of chips) {
      expect(chip.label).not.toContain('u9');
      expect(chip.label).not.toContain('p3');
      expect(chip.label).not.toContain('l1');
    }
    expect(chips.find((c) => c.id === 'projectId')?.next.assigneeId).toBe('u9');
  });

  it('treats an empty array as no filter', () => {
    expect(hasActiveFilters({ status: [], priority: [], labels: [] })).toBe(false);
  });
});

describe('ViewFilterChips', () => {
  it('renders nothing when there is nothing to show', () => {
    const { container } = render(<ViewFilterChips filters={{}} onChange={vi.fn()} />);
    expect(container).toBeEmptyDOMElement();
  });

  it('reports the remaining filters when a chip is removed', () => {
    const onChange = vi.fn();
    render(<ViewFilterChips filters={{ status: ['todo', 'done'] }} onChange={onChange} />);

    fireEvent.click(screen.getByText('flowStatus.done').closest('button') as HTMLElement);
    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ status: ['todo'] }));
  });

  it('offers Clear filters only when more than one filter is active', () => {
    const { unmount } = render(
      <ViewFilterChips filters={{ status: ['todo'] }} onChange={vi.fn()} />
    );
    expect(screen.queryByText('flow.views.filters.clearAll')).not.toBeInTheDocument();
    unmount();

    render(
      <ViewFilterChips filters={{ status: ['todo'], priority: ['high'] }} onChange={vi.fn()} />
    );
    expect(screen.getByText('flow.views.filters.clearAll')).toBeInTheDocument();
  });

  it('clears every filter at once', () => {
    const onChange = vi.fn();
    render(<ViewFilterChips filters={{ status: ['todo'], projectId: 'p1' }} onChange={onChange} />);

    fireEvent.click(screen.getByText('flow.views.filters.clearAll'));
    expect(onChange).toHaveBeenCalledWith(EMPTY_FILTERS);
  });

  it('names each chip for assistive technology', () => {
    render(<ViewFilterChips filters={{ status: ['todo'] }} onChange={vi.fn()} />);
    expect(
      screen.getByRole('button', { name: 'flow.views.filters.remove: flowStatus.todo' })
    ).toBeInTheDocument();
  });

  it('disables the controls while a save is in flight', () => {
    render(
      <ViewFilterChips
        filters={{ status: ['todo'], priority: ['high'] }}
        disabled
        onChange={vi.fn()}
      />
    );
    for (const button of screen.getAllByRole('button')) expect(button).toBeDisabled();
  });
});

describe('the Views screen wires the toolbar in', () => {
  // The wiring lesson from the 404 probe: a correct, fully tested component
  // that no screen renders leaves the reported defect exactly where it was.
  const views = fs.readFileSync(path.resolve(__dirname, 'Views.tsx'), 'utf8');

  it('renders the toolbar for the selected view', () => {
    expect(views).toContain('<ViewFilterChips');
    expect(views).toContain('filters={displayView.filters}');
  });

  it('saves the remaining filters through the view update', () => {
    expect(views).toMatch(/onChange=\{\(next\)\s*=>\s*updateView\.mutate\(/);
    expect(views).toContain('data: { filters: next }');
  });

  it('offers Clear filters from the filtered-empty result, not My tasks', () => {
    // G-E2: a filtered-empty result gets a plain sentence and a clear
    // affordance, no illustration. The screen only had "My tasks" before,
    // because there was no filter control for it to point at.
    expect(views).toContain('filtersActive');
    expect(views).toContain('data: { filters: EMPTY_FILTERS }');
  });

  it('keeps the illustration for a genuinely empty view', () => {
    // An unfiltered view with no tasks is first-run, not a dead filter.
    expect(views).toContain('tasks.length === 0 && displayView && !filtersActive');
  });
});
