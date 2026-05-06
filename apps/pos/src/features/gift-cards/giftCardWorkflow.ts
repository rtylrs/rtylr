import type { GiftCard } from '../../shared/types/pos';

interface IssueGiftCardParams {
  codeInput: string;
  amountInput: string;
  issue: (partial: Omit<GiftCard, 'id' | 'issuedAt' | 'active'>) => GiftCard;
}

type IssueGiftCardResult = { ok: true; card: GiftCard } | { ok: false; error: string };

export function issueGiftCard({
  codeInput,
  amountInput,
  issue,
}: IssueGiftCardParams): IssueGiftCardResult {
  try {
    const code = codeInput.trim();
    if (!code) return { ok: false, error: 'Code and amount required' };
    const card = issue({ code, balance: Number(amountInput) });
    return { ok: true, card };
  } catch (error) {
    return {
      ok: false,
      error: error instanceof Error ? error.message : 'Gift card issuance failed',
    };
  }
}
