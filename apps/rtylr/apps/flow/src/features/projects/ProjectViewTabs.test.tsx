import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import ProjectViewTabs from './ProjectViewTabs';

const mockUseFeatureFlag = vi.fn();

vi.mock('@tanstack/react-router', () => ({
  useLocation: () => ({ pathname: '/w/1/p/2' }),
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
    <ProjectViewTabs
      project={project}
      workspaceId="1"
      projectId="2"
      taskCount={3}
      onNewTask={vi.fn()}
    />
  );
}

describe('ProjectViewTabs preview-route gating', () => {
  it('hides Overview, Timeline and Dashboard tabs when the preview-routes flag is off', () => {
    mockUseFeatureFlag.mockReturnValue(false);
    renderTabs();

    expect(screen.queryByText('flow.projectViewTabs.overview')).not.toBeInTheDocument();
    expect(screen.queryByText('flow.shell.timeline')).not.toBeInTheDocument();
    expect(screen.queryByText('nav.dashboard')).not.toBeInTheDocument();

    expect(screen.getByText('flow.projectViewTabs.board')).toBeInTheDocument();
    expect(screen.getByText('flow.projectViewTabs.list')).toBeInTheDocument();
    expect(screen.getByText('flow.shell.calendar')).toBeInTheDocument();
  });

  it('shows all tabs when the preview-routes flag is on', () => {
    mockUseFeatureFlag.mockReturnValue(true);
    renderTabs();

    expect(screen.getByText('flow.projectViewTabs.overview')).toBeInTheDocument();
    expect(screen.getByText('flow.shell.timeline')).toBeInTheDocument();
    expect(screen.getByText('nav.dashboard')).toBeInTheDocument();
    expect(screen.getByText('flow.projectViewTabs.board')).toBeInTheDocument();
    expect(screen.getByText('flow.projectViewTabs.list')).toBeInTheDocument();
    expect(screen.getByText('flow.shell.calendar')).toBeInTheDocument();
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
