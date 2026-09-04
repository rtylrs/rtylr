// @vitest-environment happy-dom
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';

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
