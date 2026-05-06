import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { GiftCard } from '../types/pos';

interface GiftCardState {
  cards: GiftCard[];
}

interface GiftCardActions {
  issue: (partial: Omit<GiftCard, 'id' | 'issuedAt' | 'active'>) => GiftCard;
  redeem: (code: string, amount: number) => { ok: boolean; remaining?: number; error?: string };
  addBalance: (code: string, amount: number) => void;
  deactivate: (code: string) => void;
  findByCode: (code: string) => GiftCard | undefined;
}

export const useGiftCardStore = create<GiftCardState & GiftCardActions>()(
  persist(
    (set, get) => ({
      cards: [],
      issue: (partial) => {
        const card: GiftCard = {
          ...partial,
          id: 'gc_' + Math.random().toString(36).slice(2, 10),
          issuedAt: new Date().toISOString(),
          active: true,
        };
        set((s) => ({ cards: [card, ...s.cards] }));
        return card;
      },
      redeem: (code, amount) => {
        const card = get().cards.find((c) => c.code === code);
        if (!card) return { ok: false, error: 'Card not found' };
        if (!card.active) return { ok: false, error: 'Card is inactive' };
        if (card.balance < amount) return { ok: false, error: 'Insufficient balance' };
        set((s) => ({
          cards: s.cards.map((c) => (c.code === code ? { ...c, balance: c.balance - amount } : c)),
        }));
        return { ok: true, remaining: card.balance - amount };
      },
      addBalance: (code, amount) =>
        set((s) => ({
          cards: s.cards.map((c) => (c.code === code ? { ...c, balance: c.balance + amount } : c)),
        })),
      deactivate: (code) =>
        set((s) => ({
          cards: s.cards.map((c) => (c.code === code ? { ...c, active: false } : c)),
        })),
      findByCode: (code) => get().cards.find((c) => c.code === code),
    }),
    { name: 'pos-gift-cards' }
  )
);
