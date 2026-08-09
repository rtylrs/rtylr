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

export function descriptionContains(lines: JournalLine[], keyword: string): JournalLine[] {
  return lines.filter((l) => l.description?.toLowerCase().includes(keyword.toLowerCase()));
}

export function totalsByAccount(
  lines: JournalLine[]
): Record<string, { debits: number; credits: number; net: number }> {
  const result: Record<string, { debits: number; credits: number; net: number }> = {};
  for (const l of lines) {
    if (!result[l.accountId]) result[l.accountId] = { debits: 0, credits: 0, net: 0 };
    if (l.type === 'debit') result[l.accountId].debits += l.amount;
    else result[l.accountId].credits += l.amount;
    result[l.accountId].net = result[l.accountId].debits - result[l.accountId].credits;
  }
  return result;
}
