import { z } from 'zod';

const finiteNumber = z.preprocess((value) => {
  if (value === null || value === undefined || value === '') return value;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : value;
}, z.number());

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
  id: finiteNumber,
  ownerId: finiteNumber.optional(),
  firstName: z.string().default(''),
  lastName: z.string().default(''),
  email: z.string().default(''),
  phone: z.string().default(''),
  dateOfBirth: z.string().nullable().optional(),
  gender: z.string().optional(),
  address: z.string().optional(),
  city: z.string().optional(),
  country: z.string().optional(),
  loyaltyPoints: finiteNumber.catch(0),
  totalSpent: finiteNumber.catch(0),
  visitCount: finiteNumber.catch(0),
  lastVisitAt: z.string().nullable().optional(),
  loyaltyTier: z.string().optional(),
  tags: z.array(z.string()).optional(),
  notes: z.string().optional(),
  isActive: z.boolean().optional(),
  createdAt: z.string().optional(),
  updatedAt: z.string().optional(),
  // Legacy / client-only fields the backend does not currently emit. The
  // backend stores consent as an opaque JSON record, so accept any partial
  // map instead of demanding all three keys (a row with only {email:true}
  // would otherwise fail the parse and kill the whole list).
  consent: z.record(z.string(), z.boolean()).optional(),
  referredBy: finiteNumber.nullable().optional(),
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
  id: finiteNumber,
  type: z
    .enum([
      'visit',
      'purchase',
      'email',
      'sms',
      'note',
      'points',
      'reward',
      'adjust',
      'expire',
      'award',
      'upgrade',
    ])
    .catch('visit'),
  title: z.string(),
  description: z.string().optional(),
  at: z.string(),
  meta: z.record(z.string(), z.unknown()).optional(),
});

export type Customer = z.infer<typeof customerSchema>;
export type TimelineEntry = z.infer<typeof timelineEntrySchema>;
