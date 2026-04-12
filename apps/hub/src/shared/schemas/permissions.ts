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
  color: z.string(),
  permissions: z.array(z.string()),
  members: z.number().default(0),
});

export type PermissionDef = z.infer<typeof permissionDefSchema>;
export type PermissionMatrix = z.infer<typeof permissionMatrixSchema>;
export type CustomRole = z.infer<typeof customRoleSchema>;
