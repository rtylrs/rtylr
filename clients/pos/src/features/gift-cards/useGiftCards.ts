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

export interface RedeemResult {
  ok: boolean;
  remaining?: number;
  error?: string;
}

export function useGiftCards() {
  // Slice selectors: the hook must not re-render on unrelated store writes.
  const cards = useGiftCardStore((s) => s.cards);
  const issue = useGiftCardStore((s) => s.issue);
  const addBalance = useGiftCardStore((s) => s.addBalance);
  const deactivate = useGiftCardStore((s) => s.deactivate);
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

  // Server-authoritative redemption: stored value must never be mutated
  // locally only, or balances resurrect on refresh and other terminals never
  // see the redemption. Fails closed - no local mutation on any error.
  const redeemCard = useCallback(async (code: string, amountInput: string): Promise<RedeemResult> => {
    const normalizedCode = code.trim();
    const amount = Number(amountInput);
    if (!normalizedCode) return { ok: false, error: 'Enter a gift card code' };
    if (!Number.isFinite(amount) || amount <= 0) {
      return { ok: false, error: 'Redeem amount must be greater than zero' };
    }

    try {
      const res = await giftCardClient.redeemGiftCard({
        code: normalizedCode,
        amount,
        orderId: '1',
      });
      if (res.giftCard) {
        const redeemed = giftCardFromServer(res.giftCard);
        useGiftCardStore.getState().setCards(
          useGiftCardStore
            .getState()
            .cards.map((c) => (c.code === redeemed.code ? { ...redeemed, id: c.id } : c))
        );
        return { ok: true, remaining: redeemed.balance };
      }
      return { ok: false, error: res.message ?? 'Redemption failed' };
    } catch (error) {
      return {
        ok: false,
        error:
          error instanceof Error && error.message.includes('status')
            ? `Server refused redemption (${error.message.split(':').pop()?.trim() || 'error'})`
            : 'Redemption failed. Check your connection and try again.',
      };
    }
  }, []);

  return {
    cards,
    loading,
    error,
    refresh,
    issue,
    redeem: redeemCard,
    checkBalance,
    addBalance,
    deactivate,
  };
}
