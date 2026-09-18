import type { Customer as RemoteCustomer } from '@voxire/voxie/pos/customer/v1/types';
import type { Customer } from '../../shared/types/pos';

export const CUSTOMER_ATTACHMENT_ERROR = 'Select a valid customer before attaching it to the order.';

function positiveInteger(value: unknown): number | undefined {
  const numeric = typeof value === 'string' ? Number(value.trim()) : Number(value);
  if (!Number.isInteger(numeric) || numeric <= 0) return undefined;
  return numeric;
}

export function mapCustomers(customers: RemoteCustomer[]): Customer[] {
  return customers.map((c) => ({
    id: c.id == null ? '' : String(c.id),
    name:
      `${c.firstName ?? ''} ${c.lastName ?? ''}`.trim() || c.email || c.phone || 'Customer',
    phone: c.phone || undefined,
    email: c.email || undefined,
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
