import { z } from 'zod';

const groupSummarySchema = z.object({
  id: z.number(),
  name: z.string(),
  color: z.string(),
});

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
  groups: z.array(groupSummarySchema).optional().default([]),
});

export type TeamMember = z.infer<typeof teamMemberSchema> & { status: string };
