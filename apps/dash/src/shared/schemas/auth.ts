import { z } from 'zod';

export const loginResponseSchema = z.object({
  token: z.string(),
  data: z.any().optional(),
});

export const userInfoSchema = z.object({
  userId: z.string(),
  token: z.string(),
  user: z
    .object({
      email: z.string(),
      role: z.string(),
      fullName: z.string().optional(),
      addons: z.array(z.string()).optional(),
    })
    .passthrough(),
  visits: z.any().optional(),
});

export type LoginResponse = z.infer<typeof loginResponseSchema>;
export type UserInfo = z.infer<typeof userInfoSchema>;
