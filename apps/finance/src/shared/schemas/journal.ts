import { z } from 'zod';

export const journalEntrySchema = z
  .object({
    id: z.union([z.string(), z.number()]).transform(String).optional(),
    date: z.string().optional().nullable(),
    description: z.string().optional().nullable(),
    debit: z.number().optional().nullable(),
    credit: z.number().optional().nullable(),
    account: z.string().optional().nullable(),
  })
  .passthrough();

export type JournalEntryRaw = z.infer<typeof journalEntrySchema>;
