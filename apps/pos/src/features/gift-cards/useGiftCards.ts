import { unwrap } from '@voxire/api';
import { useCallback, useState } from 'react';
import { posService } from '../../shared/services/posService';
import { useGiftCardStore } from '../../shared/stores/giftCardStore';
import type { GiftCard } from '../../shared/types/pos';

export function useGiftCards() {
  const store = useGiftCardStore();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      await unwrap(posService.getGiftCards());
    } catch (e) {
      const err = e as { response?: { data?: { message?: string } }; message?: string };
      setError(err?.response?.data?.message ?? err?.message ?? 'Failed to load gift cards');
    } finally {
      setLoading(false);
    }
  }, []);

  const checkBalance = useCallback(async (code: string): Promise<GiftCard | null> => {
    try {
      const res = await unwrap(posService.getGiftCardBalance(code));
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
