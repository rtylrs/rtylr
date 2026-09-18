import { z } from 'zod';

const permissionsMapSchema = z
  .string()
  .transform((raw, ctx) => {
    try {
      return JSON.parse(raw) as unknown;
    } catch {
      ctx.addIssue({ code: z.ZodIssueCode.custom, message: 'permissions is not valid JSON' });
      return z.NEVER;
    }
  })
  .pipe(z.record(z.string(), z.boolean()));

export const roleSchema = z.object({
  id: z.number(),
  ownerId: z.number(),
  name: z.string(),
  description: z.string(),
  permissions: permissionsMapSchema,
  isSystem: z.boolean(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export const permissionCatalogEntrySchema = z.object({
  key: z.string(),
  label: z.string(),
  description: z.string(),
  category: z.string(),
});

export type Role = z.infer<typeof roleSchema>;
export type PermissionCatalogEntry = z.infer<typeof permissionCatalogEntrySchema>;
