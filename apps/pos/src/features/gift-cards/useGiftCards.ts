import { unwrap } from '@voxire/api';
import { useCallback, useState } from 'react';
import { posService } from '../../shared/services/api/posService';
import { useGiftCardStore } from '../../shared/stores/giftCardStore';
import type { GiftCard } from '../../shared/types/pos';
import { normalizeRemoteGiftCard, normalizeRemoteGiftCardList } from './giftCardRemote';

export function useGiftCards() {
  const store = useGiftCardStore();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await unwrap(posService.getGiftCards());
      const remoteCards = normalizeRemoteGiftCardList(res);
      if (remoteCards.length > 0 || (res as { gift_cards?: unknown[] })?.gift_cards) {
        useGiftCardStore.getState().setCards(remoteCards);
      }
    } catch (e) {
      const err = e as { response?: { data?: { message?: string } }; message?: string };
      setError(err?.response?.data?.message ?? err?.message ?? 'Failed to load gift cards');
    } finally {
      setLoading(false);
    }
  }, []);

  const checkBalance = useCallback(async (code: string): Promise<GiftCard | null> => {
    const normalizedCode = code.trim();
    if (!normalizedCode) return null;

    try {
      const res = await unwrap(posService.getGiftCardBalance(normalizedCode));
      const remoteCard = normalizeRemoteGiftCard(res);
      if (remoteCard) return remoteCard;

      return useGiftCardStore.getState().findByCode(normalizedCode) ?? null;
    } catch {
      const local = useGiftCardStore.getState().findByCode(normalizedCode);
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
