export interface ReconcileCandidate {
  id: string;
  label: string;
  amount: number;
}

export interface ReconcileTransaction {
  id: string;
  description: string;
  amount: number;
  matched?: boolean;
}

export function findExactMatches(
  transactions: ReconcileTransaction[],
  candidates: ReconcileCandidate[]
): Array<{ transactionId: string; candidate: ReconcileCandidate }> {
  return transactions.flatMap((transaction) => {
    const candidate = candidates.find(
      (item) => Math.abs(item.amount) === Math.abs(transaction.amount)
    );
    return transaction.id && candidate ? [{ transactionId: transaction.id, candidate }] : [];
  });
}
