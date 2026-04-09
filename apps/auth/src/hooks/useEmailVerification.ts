import { useMutation } from '@tanstack/react-query';

export interface VerifyEmailPayload {
  token: string;
}

export const useEmailVerification = () => {
  return useMutation({
    mutationFn: async (_payload: VerifyEmailPayload) => {
      await new Promise((resolve) => setTimeout(resolve, 500));
      return { verified: true } as const;
    },
  });
};

export const useResendVerificationEmail = () => {
  return useMutation({
    mutationFn: async (_payload: { email: string }) => {
      await new Promise((resolve) => setTimeout(resolve, 500));
      return { ok: true } as const;
    },
  });
};
