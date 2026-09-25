import { createOneKitClient } from '@voxire/api';
import { z } from 'zod';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { ReceiptServiceClient } from '@voxire/voxie/pos/receipt/v1/client';
import { useReceiptsStore } from '../../shared/stores/receiptsStore';
import type { Receipt } from '../../shared/types/pos';
import { receiptSchema } from '../../shared/utils/schemas/receipts';

const receiptClient = createOneKitClient(ReceiptServiceClient, 'pos');

function parseRemoteReceiptId(id: string): number | null {
  const parsed = Number(id);
  return Number.isInteger(parsed) && parsed > 0 ? parsed : null;
}

export function useReceipts() {
  const local = useReceiptsStore((s) => s.receipts);
  const addLocal = useReceiptsStore((s) => s.add);
  const markEmailed = useReceiptsStore((s) => s.markEmailed);
  const [remote, setRemote] = useState<Receipt[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await receiptClient.listReceipts({});
      setRemote(z.array(receiptSchema).parse(res.receipts) as Receipt[]);
    } catch (e) {
      const err = e as { response?: { data?: { message?: string } }; message?: string };
      setError(err?.response?.data?.message ?? err?.message ?? 'Failed to load receipts');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const email = useCallback(
    async (id: string, address: string) => {
      const remoteId = parseRemoteReceiptId(id);
      if (!remoteId) return false;

      try {
        await receiptClient.emailReceipt({ id: String(remoteId), email: address });
        markEmailed(id);
        return true;
      } catch {
        return false;
      }
    },
    [markEmailed]
  );

  // Merge local + remote once per input change instead of an O(n×m) rebuild
  // on every render of every consumer.
  const all = useMemo(() => {
    const localIds = new Set(local.map((l) => l.id));
    return [...local, ...remote.filter((r) => !localIds.has(r.id))];
  }, [local, remote]);
  return { receipts: all, loading, error, reload: load, email, add: addLocal };
}
