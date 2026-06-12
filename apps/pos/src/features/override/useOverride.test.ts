import { renderHook, act } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useOverride } from './useOverride';

const mockPosService = vi.hoisted(() => ({
  verifyManagerPin: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((value) => value),
}));

vi.mock('../../shared/services/api/posService', () => ({
  posService: {
    verifyManagerPin: (...args: unknown[]) => mockPosService.verifyManagerPin(...args),
  },
}));

describe('useOverride', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('returns true for successful manager verification responses', async () => {
    mockPosService.verifyManagerPin.mockResolvedValue({ verified: true });

    const { result } = renderHook(() => useOverride());

    await act(async () => {
      await expect(result.current.verify('1234')).resolves.toBe(true);
    });
    expect(mockPosService.verifyManagerPin).toHaveBeenCalledWith('1234');
    expect(result.current.error).toBeNull();
  });

  it('fails closed for rejected verification responses', async () => {
    mockPosService.verifyManagerPin.mockResolvedValue({ verified: false });

    const { result } = renderHook(() => useOverride());

    await act(async () => {
      await expect(result.current.verify('1111')).resolves.toBe(false);
    });
    expect(result.current.error).toBe('Incorrect manager PIN');
  });

  it('does not call the API for blank manager PINs', async () => {
    const { result } = renderHook(() => useOverride());

    await act(async () => {
      await expect(result.current.verify('   ')).resolves.toBe(false);
    });
    expect(mockPosService.verifyManagerPin).not.toHaveBeenCalled();
    expect(result.current.error).toBe('Enter a manager PIN');
  });

  it('clears errors on request', async () => {
    const { result } = renderHook(() => useOverride());

    await act(async () => {
      await result.current.verify('   ');
      result.current.clearError();
    });

    expect(result.current.error).toBeNull();
  });
});
