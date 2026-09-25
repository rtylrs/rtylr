import { z } from 'zod';

export const permissionDefSchema = z.object({
  key: z.string(),
  label: z.string(),
  group: z.string(),
  description: z.string().optional(),
});

/**
 * The `hub/permission/v1` matrix arrives as `{rows, roles}`, each row holding
 * a `roleValues` map. The screens render a flat permission list plus a
 * role -> key -> enabled lookup, so parse the wire shape and translate it.
 */
export const permissionMatrixSchema = z
  .object({
    rows: z.array(
      z.object({
        key: z.string(),
        label: z.string().default(''),
        category: z.string().default(''),
        roleValues: z.record(z.string(), z.boolean()).default({}),
      })
    ),
    roles: z.array(z.string()),
  })
  .transform(({ rows, roles }) => ({
    permissions: rows.map(({ key, label, category }) => ({ key, label, group: category })),
    roles,
    values: Object.fromEntries(
      roles.map((role) => [
        role,
        Object.fromEntries(rows.map((row) => [row.key, row.roleValues[role] ?? false])),
      ])
    ),
  }));

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
