import { useMutation } from '@tanstack/react-query';

export const useMagicLinkRequest = () =>
  useMutation({
    mutationFn: async (_payload: { email: string }) => {
      await new Promise((resolve) => setTimeout(resolve, 600));
      return { ok: true } as const;
    },
  });

export const useMagicLinkConfirm = () =>
  useMutation({
    mutationFn: async (_payload: { token: string }) => {
      await new Promise((resolve) => setTimeout(resolve, 600));
      return { ok: true } as const;
    },
  });
