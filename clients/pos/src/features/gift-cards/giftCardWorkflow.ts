import { createOneKitClient } from '@voxire/api';
import { GiftCardServiceClient } from '@voxire/voxie/pos/gift_card/v1/client';
import type { GiftCard } from '../../shared/types/pos';

const giftCardClient = createOneKitClient(GiftCardServiceClient, 'pos');

const CODE_ALPHABET = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789';

export function defaultGiftCardExpiryDate(): string {
  return new Date(Date.now() + 365 * 24 * 60 * 60 * 1000).toISOString().slice(0, 10);
}

export function generateGiftCardCode(): string {
  const bytes = new Uint8Array(8);
  crypto.getRandomValues(bytes);
  const chars = Array.from(bytes, (b) => CODE_ALPHABET[b % CODE_ALPHABET.length]).join('');
  return `GC-${chars.slice(0, 4)}-${chars.slice(4, 8)}`;
}

interface IssueGiftCardParams {
  codeInput: string;
  amountInput: string;
  paidOrderId?: string;
  customerName?: string;
  customerEmail?: string;
  expiresAt?: string;
  issue: (partial: Omit<GiftCard, 'id' | 'issuedAt' | 'active'>) => GiftCard;
}

type IssueGiftCardResult = { ok: true; card: GiftCard } | { ok: false; error: string };

function normalizeExpiryDate(input?: string): string {
  const trimmed = (input ?? '').trim();
  if (/^\d{4}-\d{2}-\d{2}$/.test(trimmed)) return trimmed;
  if (/^\d{4}-\d{2}-\d{2}T/.test(trimmed)) return trimmed.slice(0, 10);
  return defaultGiftCardExpiryDate();
}

function extractApiError(error: unknown, fallback: string): string {
  if (error && typeof error === 'object') {
    const apiMsg = (error as { response?: { data?: { message?: string } } }).response?.data?.message;
    if (apiMsg) return apiMsg;
  }
  return error instanceof Error && error.message ? error.message : fallback;
}

// Creates the card on the backend first so it exists for every terminal,
// then mirrors it into the local store for instant display.
export async function issueGiftCard({
  codeInput,
  amountInput,
  paidOrderId,
  customerName,
  customerEmail,
  expiresAt,
  issue,
}: IssueGiftCardParams): Promise<IssueGiftCardResult> {
  try {
    const code = codeInput.trim();
    if (!code) return { ok: false, error: 'Code and amount required' };
    const amount = Number(amountInput);
    if (!Number.isFinite(amount) || amount <= 0) {
      return { ok: false, error: 'Gift card amount must be greater than zero' };
    }
    const resolvedOrderId = (paidOrderId ?? '1').trim();
    const numericOrderId = Number(resolvedOrderId);
    if (!Number.isInteger(numericOrderId) || numericOrderId <= 0) {
      return { ok: false, error: 'Select a paid order before issuing a gift card' };
    }

    const resolvedName = (customerName ?? '').trim() || 'Walk-in';
    const resolvedEmail = (customerEmail ?? '').trim() || 'noreply@rtylr.com';

    await giftCardClient.createGiftCard({
      code,
      paidOrderId: String(numericOrderId),
      initialAmount: amount,
      customerName: resolvedName,
      customerEmail: resolvedEmail,
      expiresAt: normalizeExpiryDate(expiresAt),
    });

    const card = issue({
      code,
      balance: amount,
      ...(resolvedName !== 'Walk-in' ? { customerName: resolvedName } : {}),
    });
    return { ok: true, card };
  } catch (error) {
    return {
      ok: false,
      error: extractApiError(error, 'Gift card issuance failed'),
    };
  }
}
