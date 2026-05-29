export type JournalEntryType = 'credit' | 'debit';

export interface JournalLine {
  accountId: string;
  type: JournalEntryType;
  amount: number;
  description?: string;
}

export function sumDebits(lines: JournalLine[]): number {
  return lines.filter((l) => l.type === 'debit').reduce((acc, l) => acc + l.amount, 0);
}

export function sumCredits(lines: JournalLine[]): number {
  return lines.filter((l) => l.type === 'credit').reduce((acc, l) => acc + l.amount, 0);
}

export function isBalanced(lines: JournalLine[], tolerance = 0.01): boolean {
  return Math.abs(sumDebits(lines) - sumCredits(lines)) <= tolerance;
}

export function netAmount(lines: JournalLine[]): number {
  return sumDebits(lines) - sumCredits(lines);
}

export function linesForAccount(lines: JournalLine[], accountId: string): JournalLine[] {
  return lines.filter((l) => l.accountId === accountId);
}

export function accountBalance(lines: JournalLine[], accountId: string): number {
  const filtered = linesForAccount(lines, accountId);
  return sumDebits(filtered) - sumCredits(filtered);
}
