import { useMutation } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { safe } from '../../shared/lib/api';

export interface VerifyEmailPayload {
  token: string;
}

export const useEmailVerification = () => {
  return useMutation({
    mutationFn: (payload: VerifyEmailPayload) =>
      unwrap(safe.post<{ verified: boolean }>('/auth/verify-email', payload)),
  });
};

export const useResendVerificationEmail = () => {
  return useMutation({
    mutationFn: (payload: { email: string }) =>
      unwrap(safe.post<{ ok: boolean }>('/auth/resend-verification', payload)),
  });
};
