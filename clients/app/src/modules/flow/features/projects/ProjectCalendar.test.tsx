// @vitest-environment jsdom
//
// jsdom rather than the suite default. happy-dom silently discards any
// color-mix() assigned through the style property, which is the path React uses,
// so a tint is simply absent from the DOM there (rtylr#1249). The tint
// assertions below passed under happy-dom only because the same element also
// carried the raw colour as a plain hex, which was the defect fixed in
// rtylr#1246.
// @vitest-environment happy-dom
import React from 'react';
import { beforeEach, describe, expect, it, vi, afterEach } from 'vitest';
import {
  setViewportWidth,
  resetViewport,
  MOBILE_WIDTH,
  DESKTOP_WIDTH,
} from '../../../../../test/viewport';
import { render, screen, fireEvent } from '@testing-library/react';

const { language } = vi.hoisted(() => ({ language: { value: 'en-US' } }));

vi.mock('react-i18next', async (importOriginal) => ({
  ...(await importOriginal<typeof import('react-i18next')>()),
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: language.value } }),
}));

vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ projectId: 'p-9' }),
  useNavigate: () => vi.fn(),
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('../../shared/stores/flowStore', () => ({
  useFlowStore: () => ({ projects: [] }),
  STATUS_COLORS: {
    backlog: '#6b7280',
    todo: '#9ca3af',
    in_progress: 'var(--flow-info-fg)',
    in_review: '#f59e0b',
    done: 'var(--flow-status-approved)',
    cancelled: '#ef4444',
  },
}));
const tasksData = vi.hoisted(() => ({ data: [] as Record<string, unknown>[] }));
vi.mock('../tasks/useTasks', () => ({ useTasks: () => ({ ...tasksData, refetch: vi.fn() }) }));
vi.mock('../../shared/lib/useProjectRealtime', () => ({ useProjectRealtime: () => undefined }));
vi.mock('./ProjectViewTabs', () => ({ default: () => null }));
const currentProjectState = vi.hoisted(() => ({
  project: { id: 'p-9', name: 'Web' } as { id: string; name: string } | undefined,
  isLoading: false,
  loadFailed: false,
  refetch: vi.fn(),
}));
vi.mock('./useCurrentProject', () => ({
  useCurrentProject: () => currentProjectState,
}));

import ProjectCalendar from './ProjectCalendar';

describe('ProjectCalendar localisation (#1052)', () => {
  beforeEach(() => {
    language.value = 'en-US';
  });

  it('renders English weekday headers in English', () => {
    render(<ProjectCalendar />);

    for (const day of ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']) {
      expect(screen.getByText(day)).toBeInTheDocument();
    }
  });

  it('renders Arabic weekday headers when the app language is Arabic', () => {
    language.value = 'ar';

    render(<ProjectCalendar />);

    // The global calendar already renders these; this screen used a hardcoded
    // English array, so an Arabic user saw SUN MON TUE next to a right-to-left
    // layout.
    expect(screen.getByText('الأحد')).toBeInTheDocument();
    expect(screen.getByText('السبت')).toBeInTheDocument();
    expect(screen.queryByText('Sun')).not.toBeInTheDocument();
  });

  it('renders French weekday headers when the app language is French', () => {
    language.value = 'fr';

    render(<ProjectCalendar />);

    expect(screen.getByText('dim.')).toBeInTheDocument();
    expect(screen.queryByText('Sun')).not.toBeInTheDocument();
  });

  it('keeps Sunday as the first column in every language', () => {
    language.value = 'fr';

    const { container } = render(<ProjectCalendar />);
    const headers = [...container.querySelectorAll('div')]
      .map((node) => node.textContent)
      .filter((text): text is string => text === 'dim.');

    // Only the language of the labels changed, not which day the week starts
    // on, so a locale change cannot silently shift every task by a column.
    expect(headers.length).toBeGreaterThan(0);
  });

  it('formats the month label in the app language, not the browser locale', () => {
    language.value = 'fr';

    render(<ProjectCalendar />);

    const month = new Date().toLocaleString('fr', { month: 'long', year: 'numeric' });
    expect(screen.getByText(month)).toBeInTheDocument();
  });
});

describe('ProjectCalendar when the projects request fails', () => {
  afterEach(() => {
    currentProjectState.project = { id: 'p-9', name: 'Web' };
    currentProjectState.loadFailed = false;
  });

  it('offers a retry instead of reporting the project deleted', () => {
    // A failed list leaves `project` undefined exactly as a missing one does.
    currentProjectState.project = undefined;
    currentProjectState.loadFailed = true;
    render(<ProjectCalendar />);

    expect(screen.getByText('flow.loadFailed.title')).toBeInTheDocument();
    expect(screen.queryByText('flow.project.notFound')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'flow.loadFailed.retry' }));
    expect(currentProjectState.refetch).toHaveBeenCalled();
  });

  it('still reports a genuinely missing project as not found', () => {
    currentProjectState.project = undefined;
    render(<ProjectCalendar />);
    expect(screen.getByText('flow.project.notFound')).toBeInTheDocument();
    expect(screen.queryByText('flow.loadFailed.title')).not.toBeInTheDocument();
  });
});

describe('ProjectCalendar pills carry status (#1087)', () => {
  const today = new Date();
  const iso = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`;

  afterEach(() => {
    tasksData.data = [];
  });

  const withTasks = (tasks: Record<string, unknown>[]): void => {
    tasksData.data = tasks;
    render(<ProjectCalendar />);
  };

  it('gives two statuses two different tints', () => {
    // Every task here belongs to the same project, so tinting by project
    // colour made all 34 pills one shade and carried no information.
    withTasks([
      { id: 't1', title: 'Doing', status: 'in_progress', dueDate: iso, priority: 'medium' },
      { id: 't2', title: 'Finished', status: 'done', dueDate: iso, priority: 'medium' },
    ]);

    const doing = screen.getByTitle('Doing');
    const finished = screen.getByTitle('Finished');
    expect(doing.getAttribute('style')).not.toBe(finished.getAttribute('style'));
    expect(doing.style.getPropertyValue('--tone')).toContain('flow-info-fg');
    expect(finished.style.getPropertyValue('--tone')).toContain('flow-status-approved');
    // And that the tone is spent on a tint with a derived foreground. Under
    // happy-dom both declarations were absent, so the assertions above were the
    // only thing this test ever saw.
    expect(doing.style.background).toBe('color-mix(in srgb, var(--tone) 13%, transparent)');
    expect(doing.style.color).toBe('color-mix(in srgb, var(--tone) 50%, var(--flow-text))');
  });

  it('lets a right-to-left title decide its own truncation end', () => {
    // Without this an Arabic task clipped with the ellipsis at the start of
    // the phrase rather than at its end.
    withTasks([
      { id: 't3', title: 'مهمة طويلة جدا', status: 'todo', dueDate: iso, priority: 'medium' },
    ]);
    expect(screen.getByTitle('مهمة طويلة جدا')).toHaveAttribute('dir', 'auto');
  });

  it('falls back to a muted tint for an unknown status', () => {
    withTasks([
      { id: 't4', title: 'Odd', status: 'not-a-status', dueDate: iso, priority: 'medium' },
    ]);
    expect(screen.getByTitle('Odd').getAttribute('style')).toContain('flow-text-muted');
  });
});

describe('ProjectCalendar at the 390px floor (#1215)', () => {
  const today = new Date();
  const iso = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`;

  beforeEach(() => {
    // `language` is a hoisted fixture shared across this file, so a previous
    // block's value leaks in and the weekday labels come out in its language.
    // Pinned here rather than left to test order.
    language.value = 'en-US';
  });

  afterEach(() => {
    resetViewport();
    tasksData.data = [];
  });

  const withTasks = (tasks: Record<string, unknown>[]): void => {
    tasksData.data = tasks;
    render(<ProjectCalendar />);
  };

  it('shows the month as a list, not a seven-column grid', () => {
    setViewportWidth(MOBILE_WIDTH);
    withTasks([
      { id: 't1', title: 'Doing', status: 'in_progress', dueDate: iso, priority: 'medium' },
    ]);

    expect(screen.getAllByRole('heading', { level: 3 }).length).toBeGreaterThan(0);
    // The grid's weekday header row is gone rather than merely hidden.
    expect(screen.queryByText('Sun')).toBeNull();
  });

  it('keeps the grid at desktop width', () => {
    setViewportWidth(DESKTOP_WIDTH);
    withTasks([
      { id: 't1', title: 'Doing', status: 'in_progress', dueDate: iso, priority: 'medium' },
    ]);

    expect(screen.getByText('Sun')).toBeInTheDocument();
    expect(screen.queryAllByRole('heading', { level: 3 })).toHaveLength(0);
  });

  it('keeps the status tint cue in the list', () => {
    // The list has to say what the grid's pills said (rtylr#1214), or the
    // narrow layout is a downgrade rather than a different shape.
    setViewportWidth(MOBILE_WIDTH);
    withTasks([
      { id: 't1', title: 'Doing', status: 'in_progress', dueDate: iso, priority: 'medium' },
    ]);

    const title = screen.getByText('Doing');
    expect(title.style.getPropertyValue('--tone')).toContain('flow-info-fg');
  });
});
