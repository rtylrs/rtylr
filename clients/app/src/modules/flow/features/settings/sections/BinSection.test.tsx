// @vitest-environment happy-dom
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';

const archivedState = vi.hoisted(() => ({
  data: { projects: [] as Record<string, unknown>[], total: 0 },
  isLoading: false,
  isError: false,
  isPaused: false,
  failureCount: 0,
  refetch: vi.fn(),
}));
const purgeMutate = vi.hoisted(() => vi.fn());
const restoreMutate = vi.hoisted(() => vi.fn());

vi.mock('../../projects-v2/useAllProjects', () => ({
  useArchivedProjects: () => archivedState,
}));
vi.mock('../../projects-v2/useBin', () => ({
  useRestoreProject: () => ({ mutate: restoreMutate, isPending: false }),
  usePurgeProject: () => ({ mutateAsync: purgeMutate, isPending: false }),
}));
vi.mock('react-i18next', async (importOriginal) => ({
  ...(await importOriginal<typeof import('react-i18next')>()),
  useTranslation: () => ({ t: (key: string) => key }),
}));
// The real Dialog is a Radix portal; rendering the content inline keeps these
// tests about BinSection's own state handling.
vi.mock('@voxire/ui', () => ({
  EmptyState: () => null,
  Dialog: ({ open, children }: { open: boolean; children: React.ReactNode }) =>
    open ? <div>{children}</div> : null,
  DialogContent: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  DialogHeader: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  DialogTitle: ({ children }: { children: React.ReactNode }) => <h2>{children}</h2>,
  DialogDescription: ({ children }: { children: React.ReactNode }) => <p>{children}</p>,
}));

import BinSection from './BinSection';

const archivedProjects = [
  { id: 'p1', name: 'Alpha' },
  { id: 'p2', name: 'Beta' },
  { id: 'p3', name: 'Gamma' },
];

function startPurge(): void {
  fireEvent.click(screen.getByRole('button', { name: 'flow.settings.bin.purgeSelected' }));
  fireEvent.change(screen.getByLabelText('flow.confirm.typeToConfirm'), {
    target: { value: 'flow.settings.bin.purgeKeyword' },
  });
  fireEvent.click(screen.getByRole('button', { name: 'flow.bin.purgeCta' }));
}

describe('BinSection bulk purge (#1211)', () => {
  beforeEach(() => {
    archivedState.data = { projects: archivedProjects, total: archivedProjects.length };
    purgeMutate.mockReset();
    restoreMutate.mockReset();
    purgeMutate.mockResolvedValue(undefined);
  });

  it('stops the batch on the first failure and keeps the rest selected for a retry', async () => {
    let attempts = 0;
    purgeMutate.mockImplementation(async (id: string) => {
      if (id === 'p2' && attempts++ === 0) throw new Error('transient 500');
    });

    render(<BinSection />);
    fireEvent.click(screen.getAllByRole('checkbox')[0]);
    startPurge();

    await waitFor(() => expect(purgeMutate).toHaveBeenCalledTimes(2));
    // The failed purge stopped the loop before p3, rather than racing it.
    expect(purgeMutate.mock.calls.map(([id]) => id)).toEqual(['p1', 'p2']);

    // The dialog stays open and the failed and unprocessed projects stay
    // selected, so the operator can retry without re-finding them.
    expect(screen.getByRole('button', { name: 'flow.bin.purgeCta' })).toBeInTheDocument();
    const [, alpha, beta, gamma] = screen.getAllByRole('checkbox');
    expect(alpha).not.toBeChecked();
    expect(beta).toBeChecked();
    expect(gamma).toBeChecked();

    fireEvent.click(screen.getByRole('button', { name: 'flow.bin.purgeCta' }));

    await waitFor(() =>
      expect(screen.queryByRole('button', { name: 'flow.bin.purgeCta' })).toBeNull()
    );
    expect(purgeMutate.mock.calls.map(([id]) => id)).toEqual(['p1', 'p2', 'p2', 'p3']);
    for (const box of screen.getAllByRole('checkbox')) expect(box).not.toBeChecked();
  });

  it('clears the selection and closes the dialog when every purge succeeds', async () => {
    render(<BinSection />);
    fireEvent.click(screen.getAllByRole('checkbox')[0]);
    startPurge();

    await waitFor(() =>
      expect(screen.queryByRole('button', { name: 'flow.bin.purgeCta' })).toBeNull()
    );
    expect(purgeMutate.mock.calls.map(([id]) => id)).toEqual(['p1', 'p2', 'p3']);
    for (const box of screen.getAllByRole('checkbox')) expect(box).not.toBeChecked();
  });
});
