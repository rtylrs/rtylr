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

/** Two amounts within one cent are the same money on the wire. */
const MATCH_TOLERANCE = 0.01;

/**
 * Binds each transaction to at most one candidate and never reuses a
 * candidate: once bound (or already listed in usedCandidateIds, e.g. from
 * earlier manual matches), it is excluded for every subsequent transaction.
 */
export function findExactMatches(
  transactions: ReconcileTransaction[],
  candidates: ReconcileCandidate[],
  usedCandidateIds: ReadonlySet<string> = new Set()
): Array<{ transactionId: string; candidate: ReconcileCandidate }> {
  const consumed = new Set(usedCandidateIds);
  return transactions.flatMap((transaction) => {
    if (!transaction.id) return [];
    const candidate = candidates.find(
      (item) =>
        !consumed.has(item.id) &&
        Math.abs(Math.abs(item.amount) - Math.abs(transaction.amount)) <= MATCH_TOLERANCE
    );
    if (!candidate) return [];
    consumed.add(candidate.id);
    return [{ transactionId: transaction.id, candidate }];
  });
}
