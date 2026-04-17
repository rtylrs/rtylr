import { z } from 'zod';

export const sessionSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  device: z.string(),
  browser: z.string(),
  location: z.string(),
  ip: z.string(),
  current: z.boolean(),
  lastActive: z.string(),
});

export const trustedDeviceSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  name: z.string(),
  lastSeen: z.string(),
  os: z.string(),
});

export const securityStateSchema = z.object({
  twoFactorEnabled: z.boolean(),
  twoFactorMethod: z.enum(['app', 'sms']).nullable(),
  sessions: z.array(sessionSchema).default([]),
  trustedDevices: z.array(trustedDeviceSchema).default([]),
  passwordChangedAt: z.string().optional(),
  ssoEnabled: z.boolean(),
});

export type SecurityState = z.infer<typeof securityStateSchema>;
export type Session = z.infer<typeof sessionSchema>;
export type TrustedDevice = z.infer<typeof trustedDeviceSchema>;
