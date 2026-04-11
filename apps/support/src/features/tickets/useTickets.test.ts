import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/supportService', () => ({
  supportService: {
    getTickets: vi.fn(),
    getTicket: vi.fn(),
    getTicketMessages: vi.fn(),
    createTicket: vi.fn(),
    updateTicket: vi.fn(),
    updateTicketStatus: vi.fn(),
    sendTicketMessage: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import {
  useTickets,
  useTicket,
  useTicketMessages,
  useCreateTicket,
  useUpdateTicket,
  useUpdateTicketStatus,
  useSendTicketMessage,
} from './useTickets';
import { supportService } from '../../shared/services/supportService';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useTickets', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches tickets on mount', async () => {
    const data = [{ id: 1, subject: 'Bug report' }];
    (supportService.getTickets as any).mockResolvedValue(data);
    const { result } = renderHook(() => useTickets(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });

  it('passes params to the service', async () => {
    (supportService.getTickets as any).mockResolvedValue([]);
    const params = { status: 'open', search: 'login' };
    const { result } = renderHook(() => useTickets(params), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(supportService.getTickets).toHaveBeenCalledWith(params);
  });

  it('starts in loading state', () => {
    (supportService.getTickets as any).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useTickets(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useTicket', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches a single ticket', async () => {
    const ticket = { id: 5, subject: 'Help' };
    (supportService.getTicket as any).mockResolvedValue(ticket);
    const { result } = renderHook(() => useTicket(5), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(ticket);
  });
});

describe('useTicketMessages', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches messages for a ticket', async () => {
    const msgs = [{ id: 1, body: 'Hello' }];
    (supportService.getTicketMessages as any).mockResolvedValue(msgs);
    const { result } = renderHook(() => useTicketMessages(5), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(msgs);
  });
});

describe('useCreateTicket', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls createTicket service method', async () => {
    (supportService.createTicket as any).mockResolvedValue({ id: 10 });
    const { result } = renderHook(() => useCreateTicket(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync({ subject: 'New issue' }));
    expect(supportService.createTicket).toHaveBeenCalledWith({ subject: 'New issue' });
  });
});

describe('useUpdateTicket', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls updateTicket with id and data', async () => {
    (supportService.updateTicket as any).mockResolvedValue({ id: 1 });
    const { result } = renderHook(() => useUpdateTicket(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync({ id: 1, data: { subject: 'Updated' } }));
    expect(supportService.updateTicket).toHaveBeenCalledWith(1, { subject: 'Updated' });
  });
});

describe('useUpdateTicketStatus', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls updateTicketStatus with id and status', async () => {
    (supportService.updateTicketStatus as any).mockResolvedValue({ ok: true });
    const { result } = renderHook(() => useUpdateTicketStatus(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync({ id: 2, status: 'closed' }));
    expect(supportService.updateTicketStatus).toHaveBeenCalledWith(2, 'closed');
  });
});

describe('useSendTicketMessage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls sendTicketMessage with ticketId and data', async () => {
    (supportService.sendTicketMessage as any).mockResolvedValue({ id: 20 });
    const { result } = renderHook(() => useSendTicketMessage(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync({ ticketId: 5, data: { body: 'Reply here' } }));
    expect(supportService.sendTicketMessage).toHaveBeenCalledWith(5, { body: 'Reply here' });
  });
});
