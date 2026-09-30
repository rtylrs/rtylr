/**
 * The "close with $0.00 counted cash?" dialog is a second confirm path in this
 * app, and it had the same double-fire: a disabled button is not the only way
 * to submit twice, so the handler latches and re-arms after a failure.
 */
import React from 'react';
import { act, cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import CloseSessionScreen from './index';

const closeSession = vi.hoisted(() => vi.fn());
const closeDrawer = vi.hoisted(() => vi.fn());
const enqueue = vi.hoisted(() => vi.fn());
const navigate = vi.hoisted(() => vi.fn());
const toastMock = vi.hoisted(() => ({ success: vi.fn(), error: vi.fn(), warning: vi.fn() }));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => ({ closeSession }),
}));

vi.mock('react-router-dom', () => ({ useNavigate: () => navigate }));

vi.mock('sonner', () => ({ toast: toastMock }));

vi.mock('../../layouts/ScreenLayout', () => ({
  default: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

vi.mock('../day-summary/useDaySummary', () => ({
  useDaySummary: () => ({ summary: null }),
}));

vi.mock('../../shared/stores/cashDrawerStore', () => ({
  useCashDrawerStore: (selector: (s: Record<string, unknown>) => unknown) =>
    selector({
      movements: [],
      openingFloat: 100,
      expectedCash: () => 100,
      closeDrawer,
    }),
}));

vi.mock('../../shared/stores/sessionCloseQueueStore', () => ({
  useSessionCloseQueueStore: (selector: (s: Record<string, unknown>) => unknown) =>
    selector({ enqueue }),
}));

vi.mock('./closeSessionWorkflow', async () => {
  const actual = await vi.importActual<typeof import('./closeSessionWorkflow')>(
    './closeSessionWorkflow'
  );
  return {
    ...actual,
    closeSessionWithFallback: vi.fn(
      async ({
        closeSession: doClose,
        payload,
      }: {
        closeSession: (p: unknown) => Promise<void>;
        payload: unknown;
      }) => {
        await doClose(payload);
        return { status: 'online' as const };
      }
    ),
    resolveOpenSessionId: vi.fn(async () => 5),
  };
});

async function openZeroDialog() {
  render(<CloseSessionScreen />);
  fireEvent.click(screen.getByText('Close Session'));
  return screen.findByRole('button', { name: /Close anyway|Closing/ });
}

// A touchscreen double-tap delivers both clicks before React re-renders, so the
// `disabled` attribute has not landed yet. This is the burst the guard must stop.
function doubleTap(element: HTMLElement) {
  act(() => {
    element.click();
    element.click();
  });
}

describe('CloseSessionScreen $0 confirm dialog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    closeSession.mockResolvedValue(undefined);
  });

  afterEach(() => {
    cleanup();
  });

  it('closes the session once when "Close anyway" is double-clicked', async () => {
    let release: (() => void) | undefined;
    closeSession.mockImplementation(
      () =>
        new Promise<void>((resolve) => {
          release = resolve;
        })
    );

    const confirm = await openZeroDialog();
    doubleTap(confirm);

    expect(closeSession).toHaveBeenCalledTimes(1);

    release?.();
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/day-summary'));
  });

  it('closes the session once when the main button is double-tapped', async () => {
    let release: (() => void) | undefined;
    closeSession.mockImplementation(
      () =>
        new Promise<void>((resolve) => {
          release = resolve;
        })
    );

    render(<CloseSessionScreen />);
    fireEvent.change(document.querySelector('input')!, { target: { value: '100' } });
    doubleTap(screen.getByText('Close Session'));

    expect(closeSession).toHaveBeenCalledTimes(1);

    release?.();
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/day-summary'));
  });

  it('disables the dialog buttons while the close is in flight', async () => {
    let release: (() => void) | undefined;
    closeSession.mockImplementation(
      () =>
        new Promise<void>((resolve) => {
          release = resolve;
        })
    );

    const confirm = await openZeroDialog();
    fireEvent.click(confirm);

    await waitFor(() => expect((confirm as HTMLButtonElement).disabled).toBe(true));
    expect((screen.getByRole('button', { name: 'Cancel' }) as HTMLButtonElement).disabled).toBe(true);

    release?.();
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/day-summary'));
  });

  it('re-arms the confirm button after a failed close so it can be retried', async () => {
    closeSession.mockRejectedValueOnce(new Error('server rejected'));

    const confirm = await openZeroDialog();
    fireEvent.click(confirm);

    await waitFor(() => expect(toastMock.error).toHaveBeenCalledWith('server rejected'));
    expect(closeSession).toHaveBeenCalledTimes(1);

    // Dialog survives the failure and the latch is released, so a retry works.
    const retry = screen.getByRole('button', { name: 'Close anyway' });
    expect((retry as HTMLButtonElement).disabled).toBe(false);
    fireEvent.click(retry);

    await waitFor(() => expect(closeSession).toHaveBeenCalledTimes(2));
    expect(navigate).toHaveBeenCalledWith('/day-summary');
  });
});
