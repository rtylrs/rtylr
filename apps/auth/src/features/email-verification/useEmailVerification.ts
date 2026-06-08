import { useMutation } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import type { ResendVerificationRequest } from '@voxire/contracts/common/auth/v1/resend_verification';
import type { VerifyEmailRequest } from '@voxire/contracts/common/auth/v1/verify_email';
import { toast } from 'sonner';
import { safe } from '../../shared/lib/api';

export const useEmailVerification = () => {
  return useMutation({
    mutationFn: (payload: VerifyEmailRequest) =>
      unwrap(safe.post<{ verified: boolean }>('/auth/verify-email', payload)),
    onError: () => toast.error('Email verification failed. Please try again.'),
  });
};

export const useResendVerificationEmail = () => {
  return useMutation({
    mutationFn: (payload: ResendVerificationRequest) =>
      unwrap(safe.post<{ ok: boolean }>('/auth/resend-verification', payload)),
    onError: () => toast.error('Failed to resend verification email'),
  });
};
