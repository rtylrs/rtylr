import { useMutation } from '@tanstack/react-query';

export interface ForgotPasswordPayload {
  email: string;
}

/**
 * Stub hook for requesting a password reset email. Returns a mutation that
 * simulates a network request, replace `mutationFn` with a real API call
 * when the backend endpoint is ready.
 */
export const useForgotPassword = () => {
  return useMutation({
    mutationFn: async (_payload: ForgotPasswordPayload) => {
      await new Promise((resolve) => setTimeout(resolve, 700));
      return { ok: true } as const;
    },
  });
};
