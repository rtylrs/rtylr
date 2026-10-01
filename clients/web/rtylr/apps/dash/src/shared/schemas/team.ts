import { z } from 'zod';

const groupSummarySchema = z.object({
  id: z.number(),
  name: z.string(),
  color: z.string(),
});

export const permissionProfileAssignmentSchema = z.object({
  appId: z.string(),
  permissionProfileIds: z.array(z.string()),
  assignmentSource: z.string(),
  status: z.string(),
  scopes: z.array(z.object({ scopeType: z.string(), scopeId: z.string() })),
  groupId: z.string().optional(),
});

export type PermissionProfileAssignment = z.infer<typeof permissionProfileAssignmentSchema>;

export const teamMemberSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  fullName: z.string().optional(),
  name: z.string().optional(),
  email: z.string(),
  memberType: z.string(),
  licensedAppIds: z.array(z.string()).default([]),
  permissionProfileAssignments: z.array(permissionProfileAssignmentSchema).default([]),
  isActive: z.boolean().optional(),
  avatarUrl: z.string().optional(),
  lastActive: z.string().optional(),
  createdAt: z.string().optional(),
  groups: z.array(groupSummarySchema).optional().default([]),
});

export type TeamMember = z.infer<typeof teamMemberSchema> & { status: string };
