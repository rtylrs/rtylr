import { useMutation } from '@tanstack/react-query';
import { createOneKitClient } from '@voxire/api';
import { AuthServiceClient } from '@voxire/voxie/common/auth/v1/client';
import type {
  ResendVerificationRequest,
  ResendVerificationResponse,
  VerifyEmailRequest,
  VerifyEmailResponse,
} from '@voxire/voxie/common/auth/v1/types';
import { toast } from 'sonner';

const authClient = createOneKitClient(AuthServiceClient, 'auth');

export const useEmailVerification = () => {
  return useMutation<VerifyEmailResponse, Error, VerifyEmailRequest>({
    mutationFn: (payload) => authClient.verifyEmail(payload),
    onError: () => toast.error('Email verification failed. Please try again.'),
  });
};

export const useResendVerificationEmail = () => {
  return useMutation<ResendVerificationResponse, Error, ResendVerificationRequest>({
    mutationFn: (payload) => authClient.resendVerification(payload),
    onError: () => toast.error('Failed to resend verification email'),
  });
};
