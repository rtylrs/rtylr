import { z } from 'zod';

// The backend returns { id, ip, userAgent, createdAt, isCurrent }. Accept that
// and the legacy display shape, normalizing to what SecurityRow renders so the
// page doesn't render blank rows.
export const sessionSchema = z
  .object({
    id: z.union([z.string(), z.number()]).transform(String),
    device: z.string().optional(),
    browser: z.string().optional(),
    location: z.string().optional(),
    ip: z.string().default(''),
    userAgent: z.string().optional(),
    current: z.boolean().optional(),
    isCurrent: z.boolean().optional(),
    lastActive: z.string().optional(),
    createdAt: z.string().optional(),
  })
  .transform((s) => ({
    id: s.id,
    device: s.device ?? s.userAgent ?? 'Unknown device',
    browser: s.browser ?? '',
    location: s.location ?? '',
    ip: s.ip,
    current: s.current ?? s.isCurrent ?? false,
    lastActive: s.lastActive ?? s.createdAt ?? '',
  }));

export const trustedDeviceSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  name: z.string(),
  lastSeen: z.string(),
  os: z.string(),
});

// Fields the backend doesn't return (2FA method, sso, trusted devices, password
// timestamp) are optional/defaulted so the page loads. Sessions come back as
// `activeSessions`; normalize to `sessions` which the page reads.
export const securityStateSchema = z
  .object({
    twoFactorEnabled: z.boolean().default(false),
    twoFactorMethod: z.enum(['app', 'sms']).nullable().default(null),
    sessions: z.array(sessionSchema).optional(),
    activeSessions: z.array(sessionSchema).optional(),
    trustedDevices: z.array(trustedDeviceSchema).default([]),
    passwordChangedAt: z.string().optional(),
    ssoEnabled: z.boolean().default(false),
  })
  .transform((s) => ({
    twoFactorEnabled: s.twoFactorEnabled,
    twoFactorMethod: s.twoFactorMethod,
    sessions: s.sessions ?? s.activeSessions ?? [],
    trustedDevices: s.trustedDevices,
    passwordChangedAt: s.passwordChangedAt,
    ssoEnabled: s.ssoEnabled,
  }));

export type SecurityState = z.infer<typeof securityStateSchema>;
export type Session = z.infer<typeof sessionSchema>;
export type TrustedDevice = z.infer<typeof trustedDeviceSchema>;
