import { z } from 'zod';

export const templateSchema = z.object({
  id: z.string(),
  name: z.string(),
  channel: z.enum(['email', 'sms']).catch('email'),
  // NOTE: crm.template.v1.Template has no category/thumbnail fields - these
  // are UI-only concepts today and don't persist to the backend.
  category: z.string().default('general'),
  subject: z.string().nullable().optional(),
  body: z.string(),
  thumbnail: z.string().nullable().optional(),
  // Optional on the wire (crm.template.v1.Template.updatedAt is optional and
  // emitted omitempty) - a required field here would break the whole list.
  updatedAt: z.string().default(''),
});

export type Template = z.infer<typeof templateSchema>;
