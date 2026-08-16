import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ApiProject } from '../../shared/services/flowApi';
import type { WorkType } from './useWorkDay';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en-US' } }),
}));

const createLineMutate = vi.hoisted(() => vi.fn());

vi.mock('./useWorkDay', () => ({
  newClientKey: () => 'client-key-1',
  useCreateWorkDayLine: () => ({ mutate: createLineMutate, isPending: false }),
}));

const projectsState = vi.hoisted(() => ({ value: { data: [] as ApiProject[] } }));
vi.mock('../projects/useProjects', () => ({
  useProjects: () => projectsState.value,
}));

import LogTimeDialog from './LogTimeDialog';

function project(over: Partial<ApiProject> = {}): ApiProject {
  return {
    id: '1',
    workspaceId: 'ws-1',
    name: 'Site fitout',
    color: '#03aed2',
    taskCount: 0,
    completedCount: 0,
    memberAvatars: [],
    ...over,
  };
}

function workType(over: Partial<WorkType> = {}): WorkType {
  return { id: '10', name: 'DEV', parentId: undefined, ...over } as WorkType;
}

const defaultProps = {
  open: true,
  onOpenChange: vi.fn(),
  date: '2026-08-16',
  workDayId: '9',
  workspaceId: 'ws-1',
  workTypes: [] as WorkType[],
};

beforeEach(() => {
  vi.clearAllMocks();
  projectsState.value = { data: [project()] };
});

describe('LogTimeDialog: work type quick-pick', () => {
  it('renders top-level work types as buttons', () => {
    render(
      <LogTimeDialog
        {...defaultProps}
        workTypes={[workType({ id: '10', name: 'DEV' }), workType({ id: '20', name: 'ADMIN' })]}
      />
    );
    expect(screen.getByRole('button', { name: 'DEV' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ADMIN' })).toBeInTheDocument();
  });

  it('reveals child buttons once a parent with children is picked, and selecting a child sets the breadcrumb', () => {
    render(
      <LogTimeDialog
        {...defaultProps}
        workTypes={[
          workType({ id: '10', name: 'DEV', parentId: undefined }),
          workType({ id: '11', name: 'Web Development', parentId: '10' }),
          workType({ id: '12', name: 'Web Maintenance', parentId: '10' }),
        ]}
      />
    );
    fireEvent.click(screen.getByRole('button', { name: 'DEV' }));
    fireEvent.click(screen.getByRole('button', { name: 'Web Development' }));

    expect(screen.getByLabelText('flow.time.taskDescription')).toHaveValue('DEV / Web Development');
  });

  it('sets the breadcrumb immediately for a top-level type with no children', () => {
    render(
      <LogTimeDialog {...defaultProps} workTypes={[workType({ id: '10', name: 'Meetings' })]} />
    );
    fireEvent.click(screen.getByRole('button', { name: 'Meetings' }));
    expect(screen.getByLabelText('flow.time.taskDescription')).toHaveValue('Meetings');
  });

  it('sends the final leaf work type id and both descriptions on submit', () => {
    render(
      <LogTimeDialog
        {...defaultProps}
        workTypes={[
          workType({ id: '10', name: 'DEV', parentId: undefined }),
          workType({ id: '11', name: 'Web Development', parentId: '10' }),
        ]}
      />
    );
    fireEvent.click(screen.getByRole('button', { name: 'DEV' }));
    fireEvent.click(screen.getByRole('button', { name: 'Web Development' }));
    fireEvent.change(screen.getByLabelText('flow.time.duration'), { target: { value: '1:00' } });
    fireEvent.change(screen.getByLabelText('flow.time.project'), { target: { value: '1' } });
    fireEvent.change(screen.getByLabelText('flow.time.description'), {
      target: { value: 'Rebuilt the header' },
    });

    fireEvent.click(screen.getByRole('button', { name: 'flow.time.logAmount' }));

    expect(createLineMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        workTypeId: '11',
        description: 'Rebuilt the header',
        projectId: '1',
      }),
      expect.anything()
    );
  });
});
