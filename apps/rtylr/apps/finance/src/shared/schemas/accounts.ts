import { z } from 'zod';

const resourceIdSchema = z.union([z.string(), z.number()]).transform(String);
const nullableResourceIdSchema = z
  .union([z.string(), z.number()])
  .nullable()
  .transform((value) => (value == null ? null : String(value)));
const minorUnitsSchema = z
  .union([z.number(), z.string()])
  .transform((value) => Number(value))
  .refine((value) => Number.isFinite(value), 'amount_minor_units must be a finite number');
const optionalNullableString = z.string().nullable().optional();

export const bankAccountSchema = z
  .object({
    id: resourceIdSchema,
    name: z.string(),
    type: z.string().nullable(),
    balance: z.number().nullable(),
    currency: z.string().nullable(),
    institution: z.string().nullable(),
    mask: z.string().nullable(),
  })
  .strict();

/**
 * Bank-feed rows are the wire contract, not the legacy normalized shape used
 * by the reconciliation UI. Keep the wire fields here and map them below so a
 * valid /finance/bank-feeds response cannot be rejected merely because it has
 * snake_case names (or because the REST adapter already camel-cased them).
 */
export const bankTransactionSchema = z
  .object({
    id: resourceIdSchema,
    // Normalized shape accepted by older callers/tests.
    matched: z.boolean().optional(),
    description: optionalNullableString,
    date: optionalNullableString,
    category: optionalNullableString,
    amount: z.number().nullable().optional(),

    // Camel-cased fields produced by the REST response interceptor.
    ownerId: nullableResourceIdSchema.optional(),
    bankAccountId: nullableResourceIdSchema.optional(),
    amountMinorUnits: minorUnitsSchema.optional(),
    currency: optionalNullableString,
    occurredAt: optionalNullableString,
    externalId: optionalNullableString,
    matchedExpenseId: nullableResourceIdSchema.optional(),
    matchedInvoiceId: nullableResourceIdSchema.optional(),
    createdAt: optionalNullableString,

    // Raw snake_case fields are accepted too: this schema is also used at
    // boundaries that have not gone through the key converter.
    owner_id: nullableResourceIdSchema.optional(),
    bank_account_id: nullableResourceIdSchema.optional(),
    amount_minor_units: minorUnitsSchema.optional(),
    occurred_at: optionalNullableString,
    external_id: optionalNullableString,
    matched_expense_id: nullableResourceIdSchema.optional(),
    matched_invoice_id: nullableResourceIdSchema.optional(),
    created_at: optionalNullableString,
  })
  .strict()
  .refine(
    (value) =>
      value.amount !== undefined ||
      value.amountMinorUnits !== undefined ||
      value.amount_minor_units !== undefined,
    { message: 'a bank transaction must include an amount', path: ['amount'] }
  )
  .transform((value) => {
    const matchedExpenseId = value.matchedExpenseId ?? value.matched_expense_id ?? null;
    const matchedInvoiceId = value.matchedInvoiceId ?? value.matched_invoice_id ?? null;
    const minorUnits = value.amountMinorUnits ?? value.amount_minor_units;

    return {
      id: value.id,
      matched: value.matched ?? (matchedExpenseId !== null || matchedInvoiceId !== null),
      description: value.description ?? null,
      date: value.date ?? value.occurredAt ?? value.occurred_at ?? null,
      category: value.category ?? null,
      amount: minorUnits === undefined ? (value.amount ?? null) : minorUnits / 100,
      currency: value.currency ?? null,
      bankAccountId: value.bankAccountId ?? value.bank_account_id ?? null,
      matchedExpenseId,
      matchedInvoiceId,
    };
  });

export const chartOfAccountSchema = z
  .object({
    id: z.union([z.string(), z.number()]).transform(String).optional(),
    name: z.string().default(''),
    code: z.string().optional().nullable(),
    type: z.string().optional().nullable(),
    balance: z.number().optional().nullable(),
  })
  .passthrough();

export type BankAccountRaw = z.infer<typeof bankAccountSchema>;
export type BankTransactionRaw = z.infer<typeof bankTransactionSchema>;
export type ChartOfAccountRaw = z.infer<typeof chartOfAccountSchema>;
