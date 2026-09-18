import { useMutation } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
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
  const { t } = useTranslation();
  return useMutation<VerifyEmailResponse, Error, VerifyEmailRequest>({
    mutationFn: (payload) => authClient.verifyEmail(payload),
    onError: () => toast.error(t('auth.verify.defaultFailure')),
  });
};

export const useResendVerificationEmail = () => {
  const { t } = useTranslation();
  return useMutation<ResendVerificationResponse, Error, ResendVerificationRequest>({
    mutationFn: (payload) => authClient.resendVerification(payload),
    onError: () => toast.error(t('auth.verify.resendFailedToast')),
  });
};
