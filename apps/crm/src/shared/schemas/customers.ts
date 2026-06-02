import { z } from 'zod';

/**
 * Customer wire shape mirrors the backend `crmv1.Customer` proto
 * (onion customers module, customer_list_handler.go → crmv1.Customer).
 * The @voxire/api response interceptor converts snake_case → camelCase, so
 * the schema fields below match the proto's `json=` camelCase names:
 * firstName, lastName, visitCount, lastVisitAt, loyaltyTier, isActive, …
 *
 * The backend does NOT send a combined `name`, `initials`, `visits`,
 * `lastVisit`, or `consent`. Several existing screens still read those
 * legacy fields, so we derive display-only `name`, `initials`, `visits`,
 * and `lastVisit` from the canonical fields via a transform and keep
 * `consent` / `birthday` / `referredBy` optional for backwards
 * compatibility. Derived fields are not sent back to the API.
 */
const customerWireSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  ownerId: z.number().optional(),
  firstName: z.string().default(''),
  lastName: z.string().default(''),
  email: z.string().default(''),
  phone: z.string().default(''),
  dateOfBirth: z.string().nullable().optional(),
  gender: z.string().optional(),
  address: z.string().optional(),
  city: z.string().optional(),
  country: z.string().optional(),
  loyaltyPoints: z.number().default(0),
  totalSpent: z.number().default(0),
  visitCount: z.number().default(0),
  lastVisitAt: z.string().nullable().optional(),
  loyaltyTier: z.string().optional(),
  tags: z.array(z.string()).optional(),
  notes: z.string().optional(),
  isActive: z.boolean().optional(),
  createdAt: z.string().optional(),
  updatedAt: z.string().optional(),
  // Legacy / client-only fields the backend does not currently emit.
  consent: z
    .object({
      email: z.boolean(),
      sms: z.boolean(),
      marketing: z.boolean(),
    })
    .optional(),
  referredBy: z.number().nullable().optional(),
});

function deriveInitials(firstName: string, lastName: string): string {
  const first = firstName.trim().charAt(0);
  const last = lastName.trim().charAt(0);
  return `${first}${last}`.toUpperCase();
}

export const customerSchema = customerWireSchema.transform((c) => {
  const fullName = `${c.firstName} ${c.lastName}`.trim();
  return {
    ...c,
    // Display-only derived fields (not part of the API contract).
    name: fullName,
    initials: deriveInitials(c.firstName, c.lastName),
    visits: c.visitCount,
    lastVisit: c.lastVisitAt ?? '',
    birthday: c.dateOfBirth ?? undefined,
  };
});

export const timelineEntrySchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  type: z.enum(['visit', 'purchase', 'email', 'sms', 'note', 'points', 'reward']).catch('visit'),
  title: z.string(),
  description: z.string().optional(),
  at: z.string(),
  meta: z.record(z.string(), z.unknown()).optional(),
});

export type Customer = z.infer<typeof customerSchema>;
export type TimelineEntry = z.infer<typeof timelineEntrySchema>;
