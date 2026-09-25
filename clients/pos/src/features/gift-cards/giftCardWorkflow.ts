import { createOneKitClient } from '@voxire/api';
import { GiftCardServiceClient } from '@voxire/voxie/pos/gift_card/v1/client';
import type { GiftCard } from '../../shared/types/pos';

const giftCardClient = createOneKitClient(GiftCardServiceClient, 'pos');

interface IssueGiftCardParams {
  codeInput: string;
  amountInput: string;
  issue: (partial: Omit<GiftCard, 'id' | 'issuedAt' | 'active'>) => GiftCard;
}

type IssueGiftCardResult = { ok: true; card: GiftCard } | { ok: false; error: string };

// Creates the card on the backend first so it exists for every terminal,
// then mirrors it into the local store for instant display.
export async function issueGiftCard({
  codeInput,
  amountInput,
  issue,
}: IssueGiftCardParams): Promise<IssueGiftCardResult> {
  try {
    const code = codeInput.trim();
    if (!code) return { ok: false, error: 'Code and amount required' };
    const amount = Number(amountInput);
    if (!Number.isFinite(amount) || amount <= 0) {
      return { ok: false, error: 'Gift card amount must be greater than zero' };
    }

    await giftCardClient.createGiftCard({
      code,
      paidOrderId: '1',
      initialAmount: amount,
      customerName: 'Walk-in',
      customerEmail: 'noreply@rtylr.com',
      expiresAt: new Date(Date.now() + 365 * 24 * 60 * 60 * 1000).toISOString(),
    });

    const card = issue({ code, balance: amount });
    return { ok: true, card };
  } catch (error) {
    return {
      ok: false,
      error: error instanceof Error ? error.message : 'Gift card issuance failed',
    };
  }
}
