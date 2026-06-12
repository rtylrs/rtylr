import { act, renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useGiftCardStore } from '../../shared/stores/giftCardStore';
import { useGiftCards } from './useGiftCards';

const mockPosService = vi.hoisted(() => ({
  getGiftCards: vi.fn(),
  getGiftCardBalance: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((value) => value),
}));

vi.mock('../../shared/services/api/posService', () => ({
  posService: {
    getGiftCards: (...args: unknown[]) => mockPosService.getGiftCards(...args),
    getGiftCardBalance: (...args: unknown[]) => mockPosService.getGiftCardBalance(...args),
  },
}));

describe('useGiftCards', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useGiftCardStore.setState({ cards: [] });
  });

  it('returns a normalized remote balance and trims the card code', async () => {
    mockPosService.getGiftCardBalance.mockResolvedValue({
      card: {
        id: 7,
        code: ' GC-7 ',
        balance: '25.50',
        issued_at: '2026-01-01T00:00:00.000Z',
        active: true,
      },
    });

    const { result } = renderHook(() => useGiftCards());

    let card;
    await act(async () => {
      card = await result.current.checkBalance(' GC-7 ');
    });

    expect(mockPosService.getGiftCardBalance).toHaveBeenCalledWith('GC-7');
    expect(card).toEqual({
      id: '7',
      code: 'GC-7',
      balance: 25.5,
      issuedAt: '2026-01-01T00:00:00.000Z',
      active: true,
    });
  });

  it('does not call the API for blank card codes', async () => {
    const { result } = renderHook(() => useGiftCards());

    let card;
    await act(async () => {
      card = await result.current.checkBalance('   ');
    });

    expect(card).toBeNull();
    expect(mockPosService.getGiftCardBalance).not.toHaveBeenCalled();
  });

  it('falls back to the local store when the remote lookup fails or returns unusable data', async () => {
    const localCard = {
      id: 'local-gc',
      code: 'GC-LOCAL',
      balance: 15,
      issuedAt: '2026-01-03T00:00:00.000Z',
      active: true,
    };
    useGiftCardStore.setState({ cards: [localCard] });

    mockPosService.getGiftCardBalance.mockRejectedValueOnce(new Error('Offline'));
    const { result } = renderHook(() => useGiftCards());

    let failedRemoteCard;
    await act(async () => {
      failedRemoteCard = await result.current.checkBalance(' GC-LOCAL ');
    });

    expect(failedRemoteCard).toBe(localCard);
    expect(mockPosService.getGiftCardBalance).toHaveBeenLastCalledWith('GC-LOCAL');

    mockPosService.getGiftCardBalance.mockResolvedValueOnce({ data: { code: 'GC-LOCAL', balance: 'bad' } });

    let invalidRemoteCard;
    await act(async () => {
      invalidRemoteCard = await result.current.checkBalance('GC-LOCAL');
    });

    expect(invalidRemoteCard).toBe(localCard);
  });

  it('surfaces clean refresh errors', async () => {
    mockPosService.getGiftCards.mockRejectedValue(new Error('Gift card service unavailable'));

    const { result } = renderHook(() => useGiftCards());

    await act(async () => {
      await result.current.refresh();
    });

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toBe('Gift card service unavailable');
  });
});
