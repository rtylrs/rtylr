import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../../shared/testing/i18n';
import type { ApiStatus, ApiTask } from '../../../shared/services/flowApi';
import DeleteStatusDialog from './DeleteStatusDialog';

const remote = vi.hoisted(() => ({ tasks: [] as Partial<ApiTask>[], remove: vi.fn() }));
vi.mock('../../tasks/useTasks', () => ({ useTasks: () => ({ data: remote.tasks }) }));
vi.mock('../useStatuses', () => ({
  useDeleteStatus: () => ({ mutate: remote.remove, isPending: false }),
}));

const status = (
  id: string,
  name: string,
  bucket: ApiStatus['bucket'],
  sortOrder: number,
  isDefault = false
): ApiStatus => ({
  id,
  projectId: '2',
  name,
  bucket,
  legacyStatus: 'todo',
  color: '#8b5cf6',
  sortOrder,
  isDefault,
  createdAt: '',
});

const TODO = status('1', 'To Do', 'backlog', 10, true);
const REVIEW = status('2', 'Review', 'active', 20);
const DONE = status('3', 'Done', 'done', 900);
const STATUSES = [TODO, REVIEW, DONE];

const renderDialog = (target: ApiStatus, onClose = vi.fn()) =>
  render(
    <DeleteStatusDialog projectId="2" status={target} statuses={STATUSES} onClose={onClose} />
  );

beforeEach(async () => {
  await initTestI18n();
  vi.clearAllMocks();
  remote.tasks = [];
});

describe('DeleteStatusDialog', () => {
  it('counts the tasks on the status and requires a target before deleting', () => {
    remote.tasks = [
      { id: 'a', statusId: '2' },
      { id: 'b', sectionId: '2' },
      { id: 'c', statusId: '1' },
    ];
    renderDialog(REVIEW);

    expect(
      screen.getByText('2 tasks are in this status. Choose the status they move to.')
    ).toBeInTheDocument();
    const confirm = screen.getByRole('button', { name: 'Delete status' });
    expect(confirm).toBeDisabled();

    const target = screen.getByRole('combobox', { name: 'Move tasks to' });
    expect(screen.queryByRole('option', { name: 'Review' })).not.toBeInTheDocument();
    fireEvent.change(target, { target: { value: '3' } });
    fireEvent.click(confirm);

    expect(remote.remove).toHaveBeenCalledWith(
      { id: '2', moveToStatusId: '3' },
      expect.any(Object)
    );
  });

  it('deletes an empty status with a plain confirm', () => {
    renderDialog(REVIEW);

    expect(screen.queryByRole('combobox')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Delete status' }));

    // A safety target still goes along, for tasks this client did not load.
    expect(remote.remove).toHaveBeenCalledWith(
      { id: '2', moveToStatusId: '1' },
      expect.any(Object)
    );
  });

  it('refuses to delete the project only done status and says why', () => {
    renderDialog(DONE);

    expect(
      screen.getByText('A project needs at least one done status, so this one cannot be deleted.')
    ).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Delete status' })).not.toBeInTheDocument();
  });
});
