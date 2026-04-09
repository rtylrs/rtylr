import { useCallback, useState } from 'react';
import { posService } from '../services/posService';

export function useOverride() {
  const [verifying, setVerifying] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const verify = useCallback(async (pin: string): Promise<boolean> => {
    setVerifying(true);
    setError(null);
    try {
      const res = await posService.verifyManagerPin(pin);
      const ok = !!(res?.data?.ok ?? res?.data?.verified ?? res?.data?.success);
      if (!ok) setError('Incorrect manager PIN');
      return ok;
    } catch (e: any) {
      setError(e?.response?.data?.message ?? 'Incorrect manager PIN');
      return false;
    } finally {
      setVerifying(false);
    }
  }, []);

  return { verify, verifying, error, clearError: () => setError(null) };
}
