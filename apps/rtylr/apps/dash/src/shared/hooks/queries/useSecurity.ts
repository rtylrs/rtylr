import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { securityClient } from '../../lib/clients';
import { securityStateSchema } from '../../schemas/security';
import { toast } from 'sonner';
import i18next from 'i18next';

export type { SecurityState, Session, TrustedDevice } from '../../schemas/security';

export const useSecurity = () => {
  return useQuery({
    queryKey: ['security'],
    queryFn: async () => {
      const res = await securityClient.getSecurityState({});
      return securityStateSchema.parse(res.state ?? res);
    },
  });
};

export const useToggleTwoFactor = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (enabled: boolean) => {
      await securityClient.update2FA({ enabled });
      return enabled;
    },
    onSuccess: (enabled) => {
      toast.success(
        enabled
          ? i18next.t('common.toasts.twoFactorEnabled', 'Two-factor enabled')
          : i18next.t('common.toasts.twoFactorDisabled', 'Two-factor disabled')
      );
      qc.invalidateQueries({ queryKey: ['security'] });
    },
    onError: () =>
      toast.error(
        i18next.t('common.toasts.failedTwoFactorUpdate', 'Failed to update two-factor settings')
      ),
  });
};

export const useBeginMFAEnrollment = () => {
  return useMutation({
    mutationFn: async () => {
      const response = await securityClient.beginMFAEnrollment({});
      return {
        secret: response.secret ?? '',
        otpauthUri: response.otpauthUri ?? '',
      };
    },
  });
};

export const useCompleteMFAEnrollment = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (code: string) => {
      const response = await securityClient.completeMFAEnrollment({ code });
      return response.recoveryCodes ?? [];
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['security'] });
    },
  });
};

export const useVerifyMFA = () => {
  return useMutation({
    mutationFn: async (code: string) => {
      await securityClient.verifyMFA({ code });
    },
  });
};

export const useMFARecoveryCode = () => {
  return useMutation({
    mutationFn: async (code: string) => {
      await securityClient.useMFARecoveryCode({ code });
    },
  });
};

export const useRevokeSession = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (_id: string) => {
      await securityClient.invalidateSessions({});
    },
    onSuccess: () => {
      toast.success(i18next.t('common.toasts.signedOutSessions', 'Signed out of all sessions'));
      qc.invalidateQueries({ queryKey: ['security'] });
    },
    onError: () =>
      toast.error(i18next.t('common.toasts.failedSignOutSessions', 'Failed to sign out sessions')),
  });
};
