import { useCallback, useState } from 'react';
import { posService } from '../services/posService';
import { useGiftCardStore } from '../stores/giftCardStore';
import type { GiftCard } from '../types/pos';

export function useGiftCards() {
  const store = useGiftCardStore();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      await posService.getGiftCards();
    } catch (e: any) {
      setError(e?.response?.data?.message ?? e?.message ?? 'Failed to load gift cards');
    } finally {
      setLoading(false);
    }
  }, []);

  const checkBalance = useCallback(async (code: string): Promise<GiftCard | null> => {
    try {
      const res = await posService.getGiftCardBalance(code);
      return (res?.data ?? null) as GiftCard | null;
    } catch {
      const local = useGiftCardStore.getState().findByCode(code);
      return local ?? null;
    }
  }, []);

  return {
    cards: store.cards,
    loading,
    error,
    refresh,
    issue: store.issue,
    redeem: store.redeem,
    checkBalance,
    addBalance: store.addBalance,
    deactivate: store.deactivate,
  };
}
