import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../shared/testUtils/i18nTestSetup';
import ProjectDashboardPage from './ProjectDashboardPage';

initTestI18n();

const mocks = vi.hoisted(() => ({
  params: {} as Record<string, string>,
  navigate: vi.fn(),
  useDashboardProjects: vi.fn(),
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

vi.mock('@tanstack/react-router', () => ({
  useParams: () => mocks.params,
  useNavigate: () => mocks.navigate,
  Navigate: ({ to, params }: { to: string; params: Record<string, string> }) => (
    <p data-testid="redirect">{to.replace('$projectId', params.projectId)}</p>
  ),
}));

vi.mock('../../shared/hooks/queries/useInsightsDashboard', async (importOriginal) => {
  const actual =
    await importOriginal<typeof import('../../shared/hooks/queries/useInsightsDashboard')>();
  return {
    requestErrorStatus: actual.requestErrorStatus,
    useDashboardProjects: mocks.useDashboardProjects,
  };
});

vi.mock('../../shared/components/ProjectDashboardPanel', () => ({
  default: ({ projectId }: { projectId: string }) => <p>Panel for {projectId}</p>,
}));

const projects = [
  { id: '8', name: 'Website', status: 'active', visibility: 'public' },
  { id: '9', name: 'Payroll audit', status: 'active', visibility: 'private' },
];

function listed(data: typeof projects) {
  mocks.useDashboardProjects.mockReturnValue({
    data,
    isLoading: false,
    isError: false,
    error: null,
    refetch: vi.fn(),
  });
}

beforeEach(() => {
  vi.clearAllMocks();
  mocks.params = {};
});

describe('ProjectDashboardPage', () => {
  it('shows the chosen project with a picker of the projects the viewer may open', () => {
    mocks.params = { projectId: '8' };
    listed(projects);
    render(<ProjectDashboardPage />);

    expect(screen.getByRole('heading', { name: 'Project dashboard' })).toBeInTheDocument();
    expect(screen.getByText('Panel for 8')).toBeInTheDocument();
    const picker = screen.getByRole('combobox', { name: 'Project' });
    expect(picker).toHaveValue('8');
    expect(screen.getByRole('option', { name: 'Payroll audit (private)' })).toBeInTheDocument();

    fireEvent.change(picker, { target: { value: '9' } });
    expect(mocks.navigate).toHaveBeenCalledWith({
      to: '/insights/projects/$projectId',
      params: { projectId: '9' },
    });
  });

  it('opens the first project from /insights/projects', () => {
    listed(projects);
    render(<ProjectDashboardPage />);
    expect(screen.getByTestId('redirect')).toHaveTextContent('/insights/projects/8');
  });

  it('explains an empty list', () => {
    listed([]);
    render(<ProjectDashboardPage />);
    expect(screen.getByText('No projects to show')).toBeInTheDocument();
    expect(screen.queryByRole('combobox')).not.toBeInTheDocument();
  });

  // A typed URL to a project that is not in the list still asks the server,
  // which answers 404 for a private project the viewer is not a member of.
  it('still opens a project that is not in the picker', () => {
    mocks.params = { projectId: '77' };
    listed(projects);
    render(<ProjectDashboardPage />);
    expect(screen.getByText('Panel for 77')).toBeInTheDocument();
    expect(screen.getByRole('combobox', { name: 'Project' })).toHaveValue('');
  });

  it('explains a missing Business Overview permission', () => {
    mocks.useDashboardProjects.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: Object.assign(new Error('denied'), { statusCode: 403 }),
      refetch: vi.fn(),
    });
    render(<ProjectDashboardPage />);
    expect(
      screen.getByText('You need access to the Business Overview to see project dashboards.')
    ).toBeInTheDocument();
  });
});
