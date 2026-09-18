import { z } from 'zod';

const journalLineSchema = z.object({
  account: z.string().optional().nullable(),
  debit: z.number().optional().nullable(),
  credit: z.number().optional().nullable(),
  description: z.string().optional().nullable(),
});

export const journalEntrySchema = z
  .object({
    id: z.union([z.string(), z.number()]).transform(String).optional(),
    date: z.string().optional().nullable(),
    description: z.string().optional().nullable(),
    memo: z.string().optional().nullable(),
    debit: z.number().optional().nullable(),
    credit: z.number().optional().nullable(),
    account: z.string().optional().nullable(),
    lines: z.array(journalLineSchema).optional(),
  })
  .passthrough();

export type JournalLine = z.infer<typeof journalLineSchema>;

export type JournalEntryRaw = z.infer<typeof journalEntrySchema>;
