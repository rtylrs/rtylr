import { z } from 'zod';

export const permissionDefSchema = z.object({
  key: z.string(),
  label: z.string(),
  group: z.string(),
  description: z.string().optional(),
});

export const permissionMatrixSchema = z.object({
  permissions: z.array(permissionDefSchema),
  roles: z.array(z.string()),
  values: z.record(z.string(), z.record(z.string(), z.boolean())),
});

export const customRoleSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  name: z.string(),
  description: z.string().default(''),
  permissions: z.array(z.string()).default([]),
  appId: z.string().default('dash'),
  roleKey: z.string().default(''),
  level: z.number().default(0),
  parentRoleId: z.union([z.string(), z.number()]).transform(String).nullable().optional(),
  scopeRules: z.array(z.string()).default([]),
  isSystem: z.boolean().default(false),
  isArchived: z.boolean().default(false),
  lifecycle: z.string().default('active'),
  version: z.union([z.string(), z.number()]).transform(Number).default(1),
  memberCount: z.number().default(0),
  createdAt: z.string().default(''),
});

export type PermissionDef = z.infer<typeof permissionDefSchema>;
export type PermissionMatrix = z.infer<typeof permissionMatrixSchema>;
export type CustomRole = z.infer<typeof customRoleSchema>;
