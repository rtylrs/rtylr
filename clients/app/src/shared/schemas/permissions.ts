import { z } from 'zod';

export const permissionDefSchema = z.object({
  key: z.string(),
  appId: z.string(),
  resource: z.string(),
  action: z.string(),
  isSensitive: z.boolean().default(false),
  isHighRisk: z.boolean().default(false),
});

export const permissionCatalogSchema = z.object({
  permissions: z.array(permissionDefSchema).default([]),
});

export const permissionProfileSchema = z.object({
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
export type PermissionCatalog = z.infer<typeof permissionCatalogSchema>;
export type PermissionProfile = z.infer<typeof permissionProfileSchema>;
