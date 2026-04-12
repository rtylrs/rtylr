import { z } from 'zod';

export const teamMemberSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  fullName: z.string().optional(),
  name: z.string().optional(),
  email: z.string(),
  role: z.string(),
  apps: z.array(z.string()).default([]),
  isActive: z.boolean().optional(),
  avatarUrl: z.string().optional(),
  lastActive: z.string().optional(),
  createdAt: z.string().optional(),
});

export type TeamMember = z.infer<typeof teamMemberSchema> & { status: string };
