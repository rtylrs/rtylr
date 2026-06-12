import { unwrap } from '@voxire/api';
import { useCallback, useState } from 'react';
import { posService } from '../../shared/services/api/posService';

export function useOverride() {
  const [verifying, setVerifying] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const verify = useCallback(async (pin: string): Promise<boolean> => {
    const normalizedPin = pin.trim();
    if (!normalizedPin) {
      setError('Enter a manager PIN');
      return false;
    }

    setVerifying(true);
    setError(null);
    try {
      const res = await unwrap(posService.verifyManagerPin(normalizedPin));
      const ok = !!(res?.ok ?? res?.verified ?? res?.success);
      if (!ok) setError('Incorrect manager PIN');
      return ok;
    } catch (e) {
      const err = e as { response?: { data?: { message?: string } } };
      setError(err?.response?.data?.message ?? 'Incorrect manager PIN');
      return false;
    } finally {
      setVerifying(false);
    }
  }, []);

  return { verify, verifying, error, clearError: () => setError(null) };
}
