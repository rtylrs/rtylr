import { unwrap } from '@voxire/api';
import { posService } from '../../shared/services/api/posService';
import type { GiftCard } from '../../shared/types/pos';

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

    await unwrap(posService.createGiftCard({ code, initial_amount: amount, amount }));

    const card = issue({ code, balance: amount });
    return { ok: true, card };
  } catch (error) {
    return {
      ok: false,
      error: error instanceof Error ? error.message : 'Gift card issuance failed',
    };
  }
}
