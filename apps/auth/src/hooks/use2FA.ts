import { useMutation, useQuery } from '@tanstack/react-query';

export interface TwoFactorSetupResponse {
  secret: string;
  qrCodeUrl: string;
  recoveryCodes: string[];
}

const randomSecret = () =>
  Array.from({ length: 16 })
    .map(() => 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567'[Math.floor(Math.random() * 32)])
    .join('');

const randomCode = () =>
  Math.random().toString(36).slice(2, 6).toUpperCase() +
  '-' +
  Math.random().toString(36).slice(2, 6).toUpperCase();

/**
 * Stub 2FA hook. Fetches a fake TOTP secret + QR code URL + recovery codes.
 * Replace mutationFn with real API when backend is ready.
 */
export const use2FASetup = () => {
  return useQuery<TwoFactorSetupResponse>({
    queryKey: ['2fa-setup'] as const,
    queryFn: async () => {
      await new Promise((resolve) => setTimeout(resolve, 300));
      const secret = randomSecret();
      const label = encodeURIComponent('rtylr');
      const account = encodeURIComponent('user@rtylr.com');
      const otpauth = `otpauth://totp/${label}:${account}?secret=${secret}&issuer=${label}`;
      const qrCodeUrl = `https://api.qrserver.com/v1/create-qr-code/?size=220x220&data=${encodeURIComponent(
        otpauth
      )}`;
      return {
        secret,
        qrCodeUrl,
        recoveryCodes: Array.from({ length: 8 }).map(() => randomCode()),
      };
    },
  });
};

export const use2FAVerify = () => {
  return useMutation({
    mutationFn: async (_payload: { code: string }) => {
      await new Promise((resolve) => setTimeout(resolve, 500));
      return { ok: true } as const;
    },
  });
};

export const use2FADisable = () => {
  return useMutation({
    mutationFn: async (_payload: { password: string }) => {
      await new Promise((resolve) => setTimeout(resolve, 500));
      return { ok: true } as const;
    },
  });
};
