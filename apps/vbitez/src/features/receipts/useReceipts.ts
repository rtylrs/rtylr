import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useCallback, useEffect, useState } from 'react';
import { posService } from '../../shared/services/posService';
import { useReceiptsStore } from '../../shared/stores/receiptsStore';
import type { Receipt } from '../../shared/types/pos';
import { receiptSchema } from '../../shared/schemas/receipts';

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
      const res = await unwrap(posService.getReceipts());
      const raw = res?.receipts ?? res?.data ?? [];
      setRemote(z.array(receiptSchema).parse(Array.isArray(raw) ? raw : []) as Receipt[]);
    } catch (e: any) {
      setError(e?.response?.data?.message ?? e?.message ?? 'Failed to load receipts');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const email = useCallback(
    async (id: string, address: string) => {
      try {
        await unwrap(posService.emailReceipt(Number(id), address));
        markEmailed(id);
        return true;
      } catch {
        return false;
      }
    },
    [markEmailed]
  );

  const all = [...local, ...remote.filter((r) => !local.some((l) => l.id === r.id))];
  return { receipts: all, loading, error, reload: load, email, add: addLocal };
}
