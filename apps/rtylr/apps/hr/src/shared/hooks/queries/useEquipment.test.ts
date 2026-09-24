import React from 'react';
import { act, renderHook } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  createEquipmentRequest: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => ({ createEquipmentRequest: mocks.createEquipmentRequest }),
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
}));
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));
vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

import { useCreateEquipmentRequest } from './useEquipment';

const wrapper = ({ children }: { children: React.ReactNode }): React.JSX.Element => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return React.createElement(QueryClientProvider, { client }, children);
};

describe('useCreateEquipmentRequest', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.createEquipmentRequest.mockResolvedValue({ request: { id: '1' } });
  });

  it('does not accept a user urgency choice while the live adapter drops it', async () => {
    const { result } = renderHook(() => useCreateEquipmentRequest(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({
        employeeId: '42',
        equipmentType: 'Laptop',
        description: 'Replacement device',
        neededByDate: '2026-10-01',
      });
    });

    expect(mocks.createEquipmentRequest).toHaveBeenCalledWith({
      employeeId: '42',
      equipmentType: 'Laptop',
      description: 'Replacement device',
      neededByDate: '2026-10-01',
      urgency: 'medium',
    });
  });
});
