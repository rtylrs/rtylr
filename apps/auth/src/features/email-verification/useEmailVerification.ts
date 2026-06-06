import { useMutation } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { toast } from 'sonner';
import { safe } from '../../shared/lib/api';

export interface VerifyEmailPayload {
  token: string;
}

export const useEmailVerification = () => {
  return useMutation({
    mutationFn: (payload: VerifyEmailPayload) =>
      unwrap(safe.post<{ verified: boolean }>('/auth/verify-email', payload)),
    onError: () => toast.error('Email verification failed. Please try again.'),
  });
};

export const useResendVerificationEmail = () => {
  return useMutation({
    mutationFn: (payload: { email: string }) =>
      unwrap(safe.post<{ ok: boolean }>('/auth/resend-verification', payload)),
    onError: () => toast.error('Failed to resend verification email'),
  });
};
