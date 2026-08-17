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
  return { id: '10', name: 'DEV', description: '', ...over } as WorkType;
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

describe('LogTimeDialog: work type flat list', () => {
  it('renders every work type as both a dropdown option and a quick-pick button', () => {
    render(
      <LogTimeDialog
        {...defaultProps}
        workTypes={[workType({ id: '10', name: 'DEV' }), workType({ id: '20', name: 'ADMIN' })]}
      />
    );
    expect(screen.getByRole('button', { name: 'DEV' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ADMIN' })).toBeInTheDocument();
    const select = screen.getByLabelText('flow.time.workType') as HTMLSelectElement;
    expect(Array.from(select.options).map((o) => o.textContent)).toEqual(
      expect.arrayContaining(['DEV', 'ADMIN'])
    );
  });

  it("auto-fills Task Description read-only from the picked work type's own description", () => {
    render(
      <LogTimeDialog
        {...defaultProps}
        workTypes={[
          workType({ id: '10', name: 'Development', description: 'Build / revamp websites' }),
        ]}
      />
    );
    fireEvent.click(screen.getByRole('button', { name: 'Development' }));

    const taskDescription = screen.getByLabelText('flow.time.taskDescription') as HTMLInputElement;
    expect(taskDescription).toHaveValue('Build / revamp websites');
    expect(taskDescription).toHaveAttribute('readonly');
  });

  it('quick-pick button and dropdown stay in sync', () => {
    render(
      <LogTimeDialog
        {...defaultProps}
        workTypes={[workType({ id: '10', name: 'DEV' }), workType({ id: '20', name: 'ADMIN' })]}
      />
    );
    fireEvent.click(screen.getByRole('button', { name: 'ADMIN' }));
    const select = screen.getByLabelText('flow.time.workType') as HTMLSelectElement;
    expect(select).toHaveValue('20');
  });

  it('sends the picked work type id and the employee-typed description separately on submit', () => {
    render(
      <LogTimeDialog
        {...defaultProps}
        workTypes={[
          workType({ id: '10', name: 'Development', description: 'Build / revamp websites' }),
        ]}
      />
    );
    fireEvent.click(screen.getByRole('button', { name: 'Development' }));
    fireEvent.change(screen.getByLabelText('flow.time.duration'), { target: { value: '1:00' } });
    fireEvent.change(screen.getByLabelText('flow.time.project'), { target: { value: '1' } });
    fireEvent.change(screen.getByLabelText('flow.time.description'), {
      target: { value: 'Rebuilt the header' },
    });

    fireEvent.click(screen.getByRole('button', { name: 'flow.time.logAmount' }));

    expect(createLineMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        workTypeId: '10',
        description: 'Rebuilt the header',
        projectId: '1',
      }),
      expect.anything()
    );
  });
});
