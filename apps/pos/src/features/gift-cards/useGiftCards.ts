import { createOneKitClient } from '@voxire/api';
import { useCallback, useState } from 'react';
import { GiftCardServiceClient } from '@voxire/voxie/pos/gift_card/v1/client';
import type { GiftCard as ServerGiftCard } from '@voxire/voxie/pos/gift_card/v1/types';
import { useGiftCardStore } from '../../shared/stores/giftCardStore';
import type { GiftCard } from '../../shared/types/pos';

const giftCardClient = createOneKitClient(GiftCardServiceClient, 'pos');

function giftCardFromServer(card: ServerGiftCard): GiftCard {
  return {
    id: card.id ?? '',
    code: card.code ?? '',
    balance: card.currentBalance ?? 0,
    issuedAt: card.createdAt || new Date().toISOString(),
    active: card.isActive ?? false,
    ...(card.customerName ? { customerName: card.customerName } : {}),
  };
}

export function useGiftCards() {
  const store = useGiftCardStore();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await giftCardClient.listGiftCards({});
      useGiftCardStore.getState().setCards((res.giftCards ?? []).map(giftCardFromServer));
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
      const res = await giftCardClient.getGiftCardByCode({ code: normalizedCode });
      if (res.giftCard) return giftCardFromServer(res.giftCard);

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
