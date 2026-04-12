import { z } from 'zod';

export const reportTemplateSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  name: z.string(),
  description: z.string().optional().nullable(),
  category: z.string().optional().nullable(),
  params: z
    .array(
      z.object({
        key: z.string(),
        label: z.string(),
        type: z.string().optional(),
        required: z.boolean().optional(),
      })
    )
    .optional(),
  data: z.array(z.unknown()).optional(),
});

export type ReportTemplate = z.infer<typeof reportTemplateSchema>;
