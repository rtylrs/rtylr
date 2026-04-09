import { useMutation } from '@tanstack/react-query';

export interface ResetPasswordPayload {
  token: string;
  password: string;
}

/**
 * Stub hook for resetting a password using a one-time token from email.
 */
export const useResetPassword = () => {
  return useMutation({
    mutationFn: async (_payload: ResetPasswordPayload) => {
      await new Promise((resolve) => setTimeout(resolve, 800));
      return { ok: true } as const;
    },
  });
};
