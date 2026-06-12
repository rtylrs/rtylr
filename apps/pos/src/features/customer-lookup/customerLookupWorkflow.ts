import { z } from 'zod';
import { customerSchema } from '../../shared/utils/schemas/customers';
import type { Customer } from '../../shared/types/pos';

export const CUSTOMER_ATTACHMENT_ERROR = 'Select a valid customer before attaching it to the order.';

function positiveInteger(value: unknown): number | undefined {
  const numeric = typeof value === 'string' ? Number(value.trim()) : Number(value);
  if (!Number.isInteger(numeric) || numeric <= 0) return undefined;
  return numeric;
}

export function mapCustomers(raw: unknown): Customer[] {
  const parsed = z.array(customerSchema).parse(Array.isArray(raw) ? raw : []);

  return parsed.map((c) => ({
    id: c.id,
    name: c.name,
    phone: c.phone ?? undefined,
    email: c.email ?? undefined,
    loyaltyPoints: c.loyaltyPoints ?? undefined,
    lastVisit: c.lastVisit ?? undefined,
  }));
}

export function buildCustomerAttachment(customer: Pick<Customer, 'id' | 'name'>): {
  customerId: number;
  customerName: string;
} {
  const customerId = positiveInteger(customer.id);
  if (customerId == null) {
    throw new Error(CUSTOMER_ATTACHMENT_ERROR);
  }

  return {
    customerId,
    customerName: customer.name.trim() || 'Customer',
  };
}
