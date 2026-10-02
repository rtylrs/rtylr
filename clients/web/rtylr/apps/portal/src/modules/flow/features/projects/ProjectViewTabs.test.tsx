import React from 'react';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import ProjectViewTabs from './ProjectViewTabs';
import { initTestI18n } from '../../shared/testing/i18n';
const state = vi.hoisted(() => ({ pathname: '/flow/p/2', canAdminister: true, canCreate: true }));
vi.mock('../../shared/lib/currentUser', () => ({
  useCanAdministerProject: () => state.canAdminister,
}));
vi.mock('../../shared/stores/userStore', () => ({
  useUserStore: (selector: (value: unknown) => unknown) =>
    selector({
      user: {
        effectiveAccess: {
          isOwner: state.canCreate,
          subscribedApps: ['flow'],
          apps: [],
          dashPermissions: [],
          appPermissions: {},
        },
      },
    }),
}));
beforeEach(async () => {
  await initTestI18n();
  state.pathname = '/flow/p/2';
  state.canAdminister = true;
  state.canCreate = true;
});

const mockUseFeatureFlag = vi.fn();

vi.mock('@tanstack/react-router', () => ({
  useLocation: () => ({ pathname: state.pathname }),
  Link: ({
    children,
    to,
    className,
    style,
    ...props
  }: React.PropsWithChildren<{
    to: string;
    className?: string;
    style?: React.CSSProperties;
  }>) => (
    <a href={to} className={className} style={style} {...props}>
      {children}
    </a>
  ),
}));

vi.mock('@voxire/config', () => ({
  useFeatureFlag: (flag: string) => mockUseFeatureFlag(flag),
  FLAG_KEYS: { APP_FLOW_PREVIEW_ROUTES: 'app.flow.previewRoutes' },
}));

function renderTabs(
  project: { name: string; color: string; icon?: string } = {
    name: 'Test Project',
    color: '#64C5B9',
  }
) {
  return render(
    <ProjectViewTabs project={project} projectId="2" taskCount={3} onNewTask={vi.fn()} />
  );
}

describe('ProjectViewTabs preview-route gating', () => {
  it('hides Overview, Timeline and Dashboard tabs when the preview-routes flag is off', () => {
    mockUseFeatureFlag.mockReturnValue(false);
    renderTabs();

    expect(screen.queryByText('Overview')).not.toBeInTheDocument();
    expect(screen.queryByText('Timeline')).not.toBeInTheDocument();
    expect(screen.queryByText('Dashboard')).not.toBeInTheDocument();

    expect(screen.getByText('Board')).toBeInTheDocument();
    expect(screen.getByText('List')).toBeInTheDocument();
    expect(screen.getByText('Calendar')).toBeInTheDocument();
  });

  it('shows all tabs when the preview-routes flag is on', () => {
    mockUseFeatureFlag.mockReturnValue(true);
    renderTabs();

    expect(screen.getByText('Overview')).toBeInTheDocument();
    expect(screen.getByText('Timeline')).toBeInTheDocument();
    expect(screen.getByText('Dashboard')).toBeInTheDocument();
    expect(screen.getByText('Board')).toBeInTheDocument();
    expect(screen.getByText('List')).toBeInTheDocument();
    expect(screen.getByText('Calendar')).toBeInTheDocument();
  });
});

describe('ProjectViewTabs shows the project glyph (#1085)', () => {
  it('draws an svg tinted with the project colour, not a bare dot', () => {
    // The header carried a plain coloured dot, which said nothing the sidebar
    // dot was not already saying. G-C5 wants the colour on the header icon.
    const { container } = renderTabs({ name: 'Web', color: '#112233', icon: 'rocket' });

    const tinted = [...container.querySelectorAll('svg')].filter((el) =>
      (el.getAttribute('style') ?? '').includes('112233')
    );
    expect(tinted, 'no svg carries the project colour').toHaveLength(1);
  });

  it('falls back to a glyph when the project has no icon', () => {
    // A project created before icons existed has none, and the header must
    // still render something rather than an empty box.
    const { container } = renderTabs({ name: 'Web', color: '#112233' });
    expect(container.querySelector('svg')).toBeInTheDocument();
  });

  it('renders the icon at the header size, not the row size', () => {
    // G-I1: 16px in rows, 20px in headers.
    const { container } = renderTabs({ name: 'Web', color: '#112233', icon: 'rocket' });
    const tinted = [...container.querySelectorAll('svg')].find((el) =>
      (el.getAttribute('style') ?? '').includes('112233')
    );
    expect(tinted?.getAttribute('width')).toBe('20');
  });
});

describe('Portal project navigation', () => {
  it.each([
    ['/flow/p/2', 'Board'],
    ['/flow/p/2/calendar', 'Calendar'],
  ])('marks the current view at %s', (pathname, label) => {
    state.pathname = pathname;
    renderTabs();
    expect(screen.getByRole('link', { name: label })).toHaveAttribute('aria-current', 'page');
  });
  it('hides creation and settings for a viewer', () => {
    state.canCreate = false;
    state.canAdminister = false;
    renderTabs();
    expect(screen.queryByRole('button', { name: 'New task' })).not.toBeInTheDocument();
    expect(screen.queryByRole('link', { name: 'Settings' })).not.toBeInTheDocument();
  });
});
