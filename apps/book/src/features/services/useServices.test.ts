import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/bookService', () => ({
  bookService: {
    getServices: vi.fn(),
    getService: vi.fn(),
    createService: vi.fn(),
    updateService: vi.fn(),
    deleteService: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import {
  useServices,
  useService,
  useCreateService,
  useUpdateService,
  useDeleteService,
} from './useServices';
import { bookService } from '../../shared/services/bookService';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useServices', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches services on mount', async () => {
    const data = [{ id: 1, name: 'Haircut' }];
    (bookService.getServices as any).mockResolvedValue(data);
    const { result } = renderHook(() => useServices(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });

  it('passes filters to the service', async () => {
    (bookService.getServices as any).mockResolvedValue([]);
    const filters = { search: 'cut' };
    const { result } = renderHook(() => useServices(filters), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(bookService.getServices).toHaveBeenCalledWith(filters);
  });

  it('starts in loading state', () => {
    (bookService.getServices as any).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useServices(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches a single service by id', async () => {
    const svc = { id: 2, name: 'Massage' };
    (bookService.getService as any).mockResolvedValue(svc);
    const { result } = renderHook(() => useService(2), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(svc);
  });

  it('does not fetch when id is undefined', () => {
    const { result } = renderHook(() => useService(undefined), {
      wrapper: createWrapper(),
    });
    expect(result.current.fetchStatus).toBe('idle');
  });
});

describe('useCreateService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls createService on mutate', async () => {
    (bookService.createService as any).mockResolvedValue({ id: 3 });
    const { result } = renderHook(() => useCreateService(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync({ name: 'Facial' }));
    expect(bookService.createService).toHaveBeenCalledWith({ name: 'Facial' });
  });
});

describe('useUpdateService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls updateService with id and data', async () => {
    (bookService.updateService as any).mockResolvedValue({ id: 1 });
    const { result } = renderHook(() => useUpdateService(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync({ id: 1, data: { name: 'Updated' } }));
    expect(bookService.updateService).toHaveBeenCalledWith(1, { name: 'Updated' });
  });
});

describe('useDeleteService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls deleteService with id', async () => {
    (bookService.deleteService as any).mockResolvedValue({ ok: true });
    const { result } = renderHook(() => useDeleteService(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync(5));
    expect(bookService.deleteService).toHaveBeenCalledWith(5);
  });
});
