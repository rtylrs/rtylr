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

function renderTabs() {
  return render(
    <ProjectViewTabs
      project={{ name: 'Test Project', color: '#64C5B9' }}
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
