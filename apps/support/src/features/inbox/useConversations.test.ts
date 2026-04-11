import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/supportService', () => ({
  supportService: {
    getConversations: vi.fn(),
    getConversation: vi.fn(),
    assignConversation: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import { useConversations, useConversation, useAssignConversation } from './useConversations';
import { supportService } from '../../shared/services/supportService';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useConversations', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches conversations on mount', async () => {
    const data = [{ id: 1, channel: 'email' }];
    (supportService.getConversations as any).mockResolvedValue(data);
    const { result } = renderHook(() => useConversations(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });

  it('passes params to the service', async () => {
    (supportService.getConversations as any).mockResolvedValue([]);
    const params = { status: 'open', channel: 'chat' };
    const { result } = renderHook(() => useConversations(params), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(supportService.getConversations).toHaveBeenCalledWith(params);
  });

  it('starts in loading state', () => {
    (supportService.getConversations as any).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useConversations(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useConversation', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches a single conversation by id', async () => {
    const convo = { id: 7, channel: 'phone' };
    (supportService.getConversation as any).mockResolvedValue(convo);
    const { result } = renderHook(() => useConversation(7), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(convo);
  });

  it('does not fetch when id is falsy', () => {
    const { result } = renderHook(() => useConversation(0), {
      wrapper: createWrapper(),
    });
    expect(result.current.fetchStatus).toBe('idle');
  });
});

describe('useAssignConversation', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls assignConversation with id and assigneeId', async () => {
    (supportService.assignConversation as any).mockResolvedValue({ ok: true });
    const { result } = renderHook(() => useAssignConversation(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync({ id: 1, assigneeId: 42 }));
    expect(supportService.assignConversation).toHaveBeenCalledWith(1, 42);
  });
});
