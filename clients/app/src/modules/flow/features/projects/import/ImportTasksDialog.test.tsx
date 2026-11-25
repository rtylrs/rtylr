import React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../../shared/testing/i18n';

const mocks = vi.hoisted(() => ({
  mutate: vi.fn(),
}));

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));
vi.mock('../../tasks/useTasks', () => ({
  useImportTasks: () => ({ mutate: mocks.mutate, isPending: false }),
}));
vi.mock('../useProjectMembers', () => ({
  useProjectMembers: () => ({
    data: [{ userId: '7', name: 'Mhmd Issa', email: 'mhmd@example.test' }],
  }),
}));
vi.mock('../../../shared/hooks/queries/useFlowTeam', () => ({ useFlowTeam: () => ({ data: [] }) }));
vi.mock('../useStatuses', () => ({
  useProjectStatuses: () => ({ data: [{ id: '30', name: 'To Do' }] }),
}));

import ImportTasksDialog from './ImportTasksDialog';

function chooseFile(text: string) {
  const input = document.querySelector('input[type="file"]') as HTMLInputElement;
  const file = new File([text], 'tasks.csv', { type: 'text/csv' });
  fireEvent.change(input, { target: { files: [file] } });
}

beforeEach(async () => {
  vi.clearAllMocks();
  await initTestI18n();
});

describe('ImportTasksDialog', () => {
  it('previews the rows and sends them resolved to ids', async () => {
    render(<ImportTasksDialog projectId="12" open onOpenChange={vi.fn()} />);
    chooseFile(
      'Title,Assignee email,Section,Subtasks\nLogin page,mhmd@example.test,To Do,Form; Errors\nLogout,,,'
    );
    expect(await screen.findByText('2 tasks ready to import.')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Import 2 tasks' }));
    expect(mocks.mutate).toHaveBeenCalledWith(
      {
        projectId: '12',
        rows: [
          expect.objectContaining({
            title: 'Login page',
            assigneeId: '7',
            sectionId: '30',
            subtasks: ['Form', 'Errors'],
          }),
          expect.objectContaining({ title: 'Logout', subtasks: undefined }),
        ],
      },
      expect.anything()
    );
  });

  it('blocks the import and names each problem', async () => {
    render(<ImportTasksDialog projectId="12" open onOpenChange={vi.fn()} />);
    chooseFile('Title,Assignee email\nLogin page,nobody@example.test');
    expect(
      await screen.findByText('No team member has the email nobody@example.test')
    ).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Import 1 task' })).toBeDisabled();
  });

  it('shows the rows the server refused', async () => {
    mocks.mutate.mockImplementation((_vars, opts: { onSuccess: (r: unknown) => void }) =>
      opts.onSuccess({ created: 0, errors: [{ row: 1, message: 'You cannot assign tasks' }] })
    );
    render(<ImportTasksDialog projectId="12" open onOpenChange={vi.fn()} />);
    chooseFile('Title\nLogin page');
    fireEvent.click(await screen.findByRole('button', { name: 'Import 1 task' }));
    await waitFor(() =>
      expect(screen.getByText('Row 1: You cannot assign tasks')).toBeInTheDocument()
    );
  });
});
