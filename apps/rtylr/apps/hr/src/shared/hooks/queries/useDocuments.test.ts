import React from 'react';
import { act, renderHook } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  post: vi.fn(),
}));

vi.mock('../../lib/clients', () => ({
  hrTransport: { post: mocks.post },
}));
vi.mock('@voxire/api', () => ({
  unwrap: (value: unknown) => value,
}));
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));
vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

import { useUploadDocument } from './useDocuments';

const wrapper = ({ children }: { children: React.ReactNode }): React.JSX.Element => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return React.createElement(QueryClientProvider, { client }, children);
};

describe('useUploadDocument', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.post.mockResolvedValue({ document: { id: '1' } });
  });

  it('sends the collected expiry date with the upload', async () => {
    const { result } = renderHook(() => useUploadDocument(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({
        employeeId: 42,
        documentType: 'visa',
        title: 'Visa',
        storageMode: 'stored',
        fileUrl: 'https://files.example/visa.pdf',
        expiryDate: '2026-12-31',
      });
    });

    expect(mocks.post).toHaveBeenCalledWith('/hr/documents', {
      employeeId: 42,
      documentType: 'visa',
      title: 'Visa',
      storageMode: 'stored',
      fileUrl: 'https://files.example/visa.pdf',
      expiryDate: '2026-12-31',
    });
  });
});
