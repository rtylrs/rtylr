import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/bookService', () => ({
  bookService: {
    getAppointments: vi.fn(),
    getAppointment: vi.fn(),
    createAppointment: vi.fn(),
    updateAppointment: vi.fn(),
    cancelAppointment: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import {
  useAppointments,
  useAppointment,
  useCreateAppointment,
  useUpdateAppointment,
  useCancelAppointment,
} from './useAppointments';
import { bookService } from '../../shared/services/bookService';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useAppointments', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches appointments on mount', async () => {
    const data = [{ id: 1, date: '2026-04-11' }];
    (bookService.getAppointments as any).mockResolvedValue(data);
    const { result } = renderHook(() => useAppointments(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
    expect(bookService.getAppointments).toHaveBeenCalledWith(undefined);
  });

  it('passes filters to the service', async () => {
    (bookService.getAppointments as any).mockResolvedValue([]);
    const filters = { search: 'john', status: 'confirmed' };
    const { result } = renderHook(() => useAppointments(filters), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(bookService.getAppointments).toHaveBeenCalledWith(filters);
  });

  it('starts in loading state', () => {
    (bookService.getAppointments as any).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useAppointments(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useAppointment', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches a single appointment by id', async () => {
    const appt = { id: 5, date: '2026-04-12' };
    (bookService.getAppointment as any).mockResolvedValue(appt);
    const { result } = renderHook(() => useAppointment(5), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(appt);
  });

  it('does not fetch when id is undefined', () => {
    const { result } = renderHook(() => useAppointment(undefined), {
      wrapper: createWrapper(),
    });
    expect(result.current.fetchStatus).toBe('idle');
    expect(bookService.getAppointment).not.toHaveBeenCalled();
  });
});

describe('useCreateAppointment', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls createAppointment service method', async () => {
    const created = { id: 10 };
    (bookService.createAppointment as any).mockResolvedValue(created);
    const { result } = renderHook(() => useCreateAppointment(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync({ date: '2026-04-11' }));
    expect(bookService.createAppointment).toHaveBeenCalledWith({ date: '2026-04-11' });
  });
});

describe('useUpdateAppointment', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls updateAppointment with id and data', async () => {
    (bookService.updateAppointment as any).mockResolvedValue({ id: 1 });
    const { result } = renderHook(() => useUpdateAppointment(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync({ id: 1, data: { status: 'confirmed' } }));
    expect(bookService.updateAppointment).toHaveBeenCalledWith(1, { status: 'confirmed' });
  });
});

describe('useCancelAppointment', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls cancelAppointment with the id', async () => {
    (bookService.cancelAppointment as any).mockResolvedValue({ ok: true });
    const { result } = renderHook(() => useCancelAppointment(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync(3));
    expect(bookService.cancelAppointment).toHaveBeenCalledWith(3);
  });
});
