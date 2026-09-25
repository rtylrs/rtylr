// @vitest-environment happy-dom
import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mockClient = vi.hoisted(() => ({
  sendInvoice: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => mockClient,
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
}));
vi.mock('@voxire/voxie/finance/invoices/v1/client', () => ({
  InvoicesServiceClient: {},
}));
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));
vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

import { useSendInvoice } from './useInvoices';

const createWrapper = (client: QueryClient) => {
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
};

describe('useSendInvoice', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockClient.sendInvoice.mockResolvedValue({ message: 'sent' });
  });

  it('invalidates both the invoice list and the sent invoice detail', async () => {
    const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries').mockResolvedValue(undefined);
    const { result } = renderHook(() => useSendInvoice(), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      await result.current.mutateAsync({ invoiceId: 'inv-9' });
    });

    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['invoices'] });
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['invoice', 'inv-9'] });
  });
});
