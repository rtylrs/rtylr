// @vitest-environment happy-dom
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';

const { flagState, storeState } = vi.hoisted(() => ({
  flagState: { showPreviewRoutes: false },
  storeState: {
    commandBarOpen: true,
    setCommandBarOpen: vi.fn(),
    setQuickCreateOpen: vi.fn(),
  },
}));

vi.mock('@voxire/config', () => ({
  FLAG_KEYS: { APP_FLOW_PREVIEW_ROUTES: 'app.flow.preview_routes' },
  useFeatureFlag: () => flagState.showPreviewRoutes,
}));

vi.mock('react-i18next', () => ({
  // Returning the key makes the assertions read as the destinations they are.
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@tanstack/react-router', () => ({ useNavigate: () => vi.fn() }));

vi.mock('../stores/flowStore', () => ({ useFlowStore: () => storeState }));

vi.mock('../../features/tasks/useTasks', () => ({ useTasks: () => ({ data: [] }) }));
vi.mock('../../features/projects/useProjects', () => ({ useProjects: () => ({ data: [] }) }));
vi.mock('../lib/taskMapper', () => ({ apiTaskToTask: (task: unknown) => task }));
vi.mock('../services/flowApi', () => ({ flowApi: { searchTasks: vi.fn() } }));
vi.mock('@voxire/api', () => ({ unwrap: async (value: unknown) => await value }));
vi.mock('@tanstack/react-query', () => ({ useQuery: () => ({ data: [] }) }));

vi.mock('motion/react', () => ({
  AnimatePresence: ({ children }: { children: React.ReactNode }) => children,
  motion: {
    div: ({ children, ...rest }: { children: React.ReactNode }) =>
      React.createElement('div', rest, children),
  },
}));

import { CommandBar } from './CommandBar';

// The four destinations router.tsx wraps in AppFlagGuard(APP_FLOW_PREVIEW_ROUTES).
const PREVIEW_LABELS = [
  'flow.shell.timeline',
  'flow.shell.roadmap',
  'flow.shell.insights',
  'flow.shell.team',
];
// Reachable regardless of the flag, so the filter is not over-reaching.
const ALWAYS_LABELS = [
  'flow.commandBar.goToInbox',
  'flow.shell.calendar',
  'flow.shell.cycles',
  'flow.shell.milestones',
];

describe('CommandBar preview destinations (#1048)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    storeState.commandBarOpen = true;
  });

  it('offers no preview destination while the flag is off', () => {
    flagState.showPreviewRoutes = false;

    render(<CommandBar />);

    for (const label of PREVIEW_LABELS) {
      expect(screen.queryByText(label), `${label} should be hidden`).not.toBeInTheDocument();
    }
  });

  it('still offers the destinations that do not need the flag', () => {
    flagState.showPreviewRoutes = false;

    render(<CommandBar />);

    for (const label of ALWAYS_LABELS) {
      expect(screen.getByText(label), `${label} should be offered`).toBeInTheDocument();
    }
  });

  it('offers the preview destinations once the flag is on', () => {
    flagState.showPreviewRoutes = true;

    render(<CommandBar />);

    for (const label of PREVIEW_LABELS) {
      expect(screen.getByText(label), `${label} should be offered`).toBeInTheDocument();
    }
  });
});

describe('CommandBar selection uses the accent, not the hover grey (#1097 D-061)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    storeState.commandBarOpen = true;
  });

  const rowFor = (label: string): HTMLElement =>
    screen.getByText(label).closest('button') as HTMLElement;

  it('marks the row Enter will run with the accent tint', () => {
    render(<CommandBar />);
    const row = rowFor(ALWAYS_LABELS[1]);
    // Selection follows the pointer as well as the keyboard, so this is how a
    // row becomes the one Enter will run.
    fireEvent.mouseEnter(row);

    // --flow-hover also means "the pointer is here", so it could not say which
    // row Enter would run.
    expect(row.getAttribute('style')).toContain('accent-muted');
    expect(row.getAttribute('style')).not.toContain('flow-hover');
  });

  it('pairs the tint with its own foreground token', () => {
    // Reusing --accent-fg on this tint put 40 of 56 light-mode combinations
    // under 4.5:1, which is why --accent-on-muted exists (#970).
    render(<CommandBar />);
    const row = rowFor(ALWAYS_LABELS[1]);
    fireEvent.mouseEnter(row);
    expect(row.getAttribute('style')).toContain('accent-on-muted');
  });

  it('leaves the rows that are not selected transparent', () => {
    render(<CommandBar />);
    fireEvent.mouseEnter(rowFor(ALWAYS_LABELS[1]));

    const other = rowFor(ALWAYS_LABELS[2]);
    expect(other.getAttribute('style')).toContain('transparent');
    expect(other.getAttribute('style')).not.toContain('accent-muted');
  });
});
