// Pure helpers for gift card balance and validation logic.

import type { GiftCard } from '../types/pos';

/** Returns true if the gift card is active and has a positive balance. */
export function isGiftCardUsable(card: GiftCard): boolean {
  return card.active && Number.isFinite(card.balance) && card.balance > 0;
}

/** Calculate the amount to charge to a gift card, not exceeding the balance. */
export function giftCardChargeAmount(card: GiftCard, amountDue: number): number {
  if (!isGiftCardUsable(card)) return 0;
  if (!Number.isFinite(amountDue) || amountDue <= 0) return 0;
  return Math.min(card.balance, amountDue);
}

/** Calculate the remaining balance after a charge. */
export function giftCardBalanceAfterCharge(card: GiftCard, charged: number): number {
  if (!Number.isFinite(charged) || charged < 0) return Math.max(0, card.balance);
  return Math.max(0, Math.round((card.balance - charged) * 100) / 100);
}

/** Return true if a gift card code has the correct format. */
export function isValidGiftCardCode(code: string): boolean {
  return /^[A-Za-z0-9-]{4,32}$/.test(code.trim());
}

/** Format a gift card balance for display. */
export function formatGiftCardBalance(balance: number): string {
  if (!Number.isFinite(balance) || balance < 0) return '$0.00';
  return `$${balance.toFixed(2)}`;
}

/** Return the gift card with an updated balance after a purchase. */
export function applyGiftCardCharge(card: GiftCard, charged: number): GiftCard {
  return {
    ...card,
    balance: giftCardBalanceAfterCharge(card, charged),
  };
}
