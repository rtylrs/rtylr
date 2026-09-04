import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import Milestones from './Milestones';

const mutateSpy = vi.fn();
// The mutation resolves with the { project } envelope the contract returns.
// Mocking a bare project here hid a real bug: the component read .id off the
// envelope and got undefined, and only the typecheck caught it.
const createProjectSpy = vi.fn((_vars: unknown, opts?: { onSuccess?: (p: unknown) => void }) =>
  opts?.onSuccess?.({ project: { id: 'p-new', name: 'Ras Beirut refit' } })
);

const projectsState = vi.hoisted(() => ({
  data: [{ id: 'p-1', name: 'Website Redesign' }] as Array<{ id: string; name: string }>,
}));
const createProjectState = vi.hoisted(() => ({ isPending: false }));
const accessState = vi.hoisted(() => ({ canCreate: true }));

// The create form asks which project the milestone belongs to, so the screen
// reads the project list, and can create one inline when none will do.
vi.mock('../projects/useProjects', () => ({
  useProjects: () => ({ data: projectsState.data, isLoading: false }),
  useCreateProject: () => ({ mutate: createProjectSpy, isPending: createProjectState.isPending }),
}));

// The inline create is permission gated exactly as the sidebar's is. The
// permission itself has its own tests; here it is a boolean input.
vi.mock('../../shared/lib/roleUtils', () => ({
  canFlowPermission: () => accessState.canCreate,
}));

vi.mock('@voxire/auth', () => ({
  useUserStore: (selector: (s: unknown) => unknown) =>
    selector({ user: { role: 'member' }, effectiveAccess: {} }),
}));

const milestonesState = vi.hoisted(() => ({ data: [] as Array<Record<string, unknown>> }));

vi.mock('./useMilestones', () => ({
  useMilestones: () => ({ data: milestonesState.data, isLoading: false }),
  useCreateMilestone: () => ({ mutate: mutateSpy, isPending: false }),
}));

vi.mock('../../shared/stores/flowStore', () => ({
  useFlowStore: () => ({ activeWorkspaceId: 'ws1' }),
}));

describe('Milestones create form', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-08-20T12:00:00.000Z'));
    mutateSpy.mockClear();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('submits the exact date picked by the user, not the current date', () => {
    render(<Milestones />);

    fireEvent.click(screen.getByText('flow.milestones.createFirst'));
    fireEvent.change(screen.getByPlaceholderText('flow.milestones.namePlaceholder'), {
      target: { value: 'Launch' },
    });

    fireEvent.click(screen.getByLabelText('flow.milestones.targetDate'));
    fireEvent.click(screen.getByRole('gridcell', { name: /august 31, 2026/i }));

    fireEvent.click(screen.getByText('flow.milestones.create'));

    expect(mutateSpy).toHaveBeenCalledTimes(1);
    expect(mutateSpy.mock.calls[0][0].targetDate).toBe('2026-08-31');
  });

  it('never submits a milestone without an explicit target date', () => {
    render(<Milestones />);

    fireEvent.click(screen.getByText('flow.milestones.createFirst'));
    fireEvent.change(screen.getByPlaceholderText('flow.milestones.namePlaceholder'), {
      target: { value: 'Launch' },
    });

    // No date picked - the Create button must be disabled so a click cannot
    // silently submit today's date in its place.
    const createButton = screen.getByText('flow.milestones.create');
    expect(createButton).toBeDisabled();

    fireEvent.click(createButton);
    expect(mutateSpy).not.toHaveBeenCalled();
  });

  it('gives the target date picker an accessible name that opens a calendar', () => {
    render(<Milestones />);

    fireEvent.click(screen.getByText('flow.milestones.createFirst'));

    fireEvent.click(screen.getByLabelText('flow.milestones.targetDate'));
    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });
});

describe('Milestones progress bar (#1080)', () => {
  function milestone(over: Record<string, unknown> = {}): Record<string, unknown> {
    return {
      id: 'm-1',
      projectId: 'p-1',
      name: 'Launch',
      status: 'active',
      targetDate: '2026-09-30',
      taskCount: 0,
      doneCount: 0,
      ...over,
    };
  }

  beforeEach(() => {
    milestonesState.data = [];
  });

  afterEach(() => {
    milestonesState.data = [];
  });

  function fillWidth(container: HTMLElement): string | undefined {
    // The fill is the only element inside the rounded track.
    const track = container.querySelector('.rounded-full.overflow-hidden');
    return (track?.firstElementChild as HTMLElement | null)?.style.width;
  }

  it('renders an empty track when progress is unknown', () => {
    // computedProgress is optional. `width: undefined%` is invalid, the
    // browser drops the declaration, and the fill took its parent's width, so
    // a milestone with unknown progress looked like a completed one.
    milestonesState.data = [milestone({ computedProgress: undefined })];
    const { container } = render(<Milestones />);
    expect(fillWidth(container)).toBe('0%');
  });

  it('says there are no tasks yet instead of a bare percent sign', () => {
    milestonesState.data = [milestone({ computedProgress: undefined })];
    render(<Milestones />);
    expect(screen.getByText('flow.milestones.noTasksYet')).toBeInTheDocument();
    expect(screen.queryByText('%')).not.toBeInTheDocument();
  });

  it('renders the real percentage when there are tasks', () => {
    milestonesState.data = [milestone({ taskCount: 19, doneCount: 7, computedProgress: 37 })];
    const { container } = render(<Milestones />);
    expect(screen.getByText('37%')).toBeInTheDocument();
    expect(fillWidth(container)).toBe('37%');
  });

  it('clamps a percentage above a hundred', () => {
    milestonesState.data = [milestone({ taskCount: 3, doneCount: 3, computedProgress: 140 })];
    const { container } = render(<Milestones />);
    expect(fillWidth(container)).toBe('100%');
  });

  it('shows a full track only when the work is actually done', () => {
    milestonesState.data = [milestone({ taskCount: 4, doneCount: 4, computedProgress: 100 })];
    const { container } = render(<Milestones />);
    expect(fillWidth(container)).toBe('100%');
    expect(screen.getByText('100%')).toBeInTheDocument();
  });
});

// ─── Creating a project from the milestone form ──────────────────────────────
//
// Every milestone needs a project (project_id is @required on v2), so with no
// projects the form was a dead end: the only option carried value="" and Create
// stayed disabled with no way out. These cover the inline escape hatch.

describe('Milestones inline project creation', () => {
  beforeEach(() => {
    createProjectSpy.mockClear();
    createProjectState.isPending = false;
    projectsState.data = [{ id: 'p-1', name: 'Website Redesign' }];
    accessState.canCreate = true;
    mutateSpy.mockClear();
  });

  const openForm = (): void => {
    fireEvent.click(
      screen.queryByText('flow.milestones.createFirst') ??
        screen.getByText('flow.milestones.newMilestone')
    );
  };

  const chooseNewProject = (): void => {
    fireEvent.change(screen.getByLabelText('flow.milestones.project', { selector: 'select' }), {
      target: { value: '__new__' },
    });
  };

  it('offers the new-project option and does not treat it as a project id', () => {
    render(<Milestones />);
    openForm();

    const select = screen.getByLabelText('flow.milestones.project', { selector: 'select' });
    expect(select).toHaveValue('p-1');
    expect(screen.getByText('flow.milestones.newProjectOption')).toBeInTheDocument();
  });

  it('reveals a name input instead of submitting a milestone', () => {
    render(<Milestones />);
    openForm();
    chooseNewProject();

    expect(
      screen.getByPlaceholderText('flow.milestones.newProjectPlaceholder')
    ).toBeInTheDocument();
    expect(mutateSpy).not.toHaveBeenCalled();
    expect(createProjectSpy).not.toHaveBeenCalled();
  });

  it('creates the project, selects it, and keeps what was already typed', () => {
    render(<Milestones />);
    openForm();
    fireEvent.change(screen.getByPlaceholderText('flow.milestones.namePlaceholder'), {
      target: { value: 'Launch' },
    });
    chooseNewProject();
    fireEvent.change(screen.getByPlaceholderText('flow.milestones.newProjectPlaceholder'), {
      target: { value: 'Ras Beirut refit' },
    });
    fireEvent.click(screen.getByText('flow.milestones.createProject'));

    expect(createProjectSpy).toHaveBeenCalledTimes(1);
    expect(createProjectSpy.mock.calls[0][0].name).toBe('Ras Beirut refit');

    // The new project is now selected and the milestone name survived.
    expect(screen.getByLabelText('flow.milestones.project', { selector: 'select' })).toHaveValue(
      'p-new'
    );
    expect(screen.getByPlaceholderText('flow.milestones.namePlaceholder')).toHaveValue('Launch');
  });

  it('does not send a blank project name', () => {
    render(<Milestones />);
    openForm();
    chooseNewProject();

    const createButton = screen.getByText('flow.milestones.createProject');
    expect(createButton).toBeDisabled();
    fireEvent.click(createButton);
    expect(createProjectSpy).not.toHaveBeenCalled();
  });

  it('restores the previous selection when cancelled', () => {
    render(<Milestones />);
    openForm();
    chooseNewProject();
    fireEvent.click(screen.getByText('flow.milestones.cancelProject'));

    const select = screen.getByLabelText('flow.milestones.project', { selector: 'select' });
    expect(select).toHaveValue('p-1');
    expect(createProjectSpy).not.toHaveBeenCalled();
  });

  it('keeps the typed name on screen when the create fails', () => {
    createProjectSpy.mockImplementation((_vars, opts) => opts?.onError?.(new Error('nope')));
    render(<Milestones />);
    openForm();
    chooseNewProject();
    fireEvent.change(screen.getByPlaceholderText('flow.milestones.newProjectPlaceholder'), {
      target: { value: 'Doomed' },
    });
    fireEvent.click(screen.getByText('flow.milestones.createProject'));

    expect(screen.getByPlaceholderText('flow.milestones.newProjectPlaceholder')).toHaveValue(
      'Doomed'
    );
  });

  it('is reachable with no projects at all, which is the dead end it fixes', () => {
    projectsState.data = [];
    render(<Milestones />);
    openForm();

    // Without this option the form cannot be completed: Create needs a
    // project id and there is no other way to obtain one from here.
    expect(screen.getByText('flow.milestones.newProjectOption')).toBeInTheDocument();
    expect(screen.getByText('flow.milestones.create')).toBeDisabled();
  });

  it('hides the option when the user cannot create projects', () => {
    accessState.canCreate = false;
    render(<Milestones />);
    openForm();

    expect(screen.queryByText('flow.milestones.newProjectOption')).not.toBeInTheDocument();
  });

  it('explains the dead end when there are no projects and no permission', () => {
    // Hiding the option is right, but on its own it leaves the original
    // defect for this user: the only option carries value="", so Create stays
    // disabled with nothing saying why.
    accessState.canCreate = false;
    projectsState.data = [];
    render(<Milestones />);
    openForm();

    expect(screen.getByText('flow.milestones.noProjectsGate')).toBeInTheDocument();
    expect(screen.getByText('flow.milestones.create')).toBeDisabled();
  });

  it('does not explain a dead end that is not there', () => {
    accessState.canCreate = false;
    projectsState.data = [{ id: 'p-1', name: 'Website Redesign' }];
    render(<Milestones />);
    openForm();

    expect(screen.queryByText('flow.milestones.noProjectsGate')).not.toBeInTheDocument();
  });

  it('does not explain a dead end to someone who can create a project', () => {
    accessState.canCreate = true;
    projectsState.data = [];
    render(<Milestones />);
    openForm();

    expect(screen.queryByText('flow.milestones.noProjectsGate')).not.toBeInTheDocument();
    expect(screen.getByText('flow.milestones.newProjectOption')).toBeInTheDocument();
  });

  it('points the select at the reason so a screen reader reads it', () => {
    accessState.canCreate = false;
    projectsState.data = [];
    const { container } = render(<Milestones />);
    openForm();

    const select = container.querySelector('select');
    expect(select?.getAttribute('aria-describedby')).toBe('milestone-project-gate');
  });
});
