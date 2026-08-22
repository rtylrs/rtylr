import { unwrap } from '@voxire/api';
import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
  type UseQueryResult,
} from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { z } from 'zod';
import { safe } from '../../lib/api';
import { unwrapList } from '../../lib/unwrapList';
import { customerSchema, timelineEntrySchema } from '../../schemas/customers';
import type { Customer, TimelineEntry } from '../../schemas/customers';

// ── Types ────────────────────────────────────────────────────────────────────

export interface ListCustomersParams {
  search?: string;
  limit: number;
  offset: number;
  tier?: string;
  tag?: string;
}

export type CustomerListFilters = Pick<ListCustomersParams, 'search' | 'tier' | 'tag'>;

type CustomerInput = Partial<Customer>;

export interface UpdateCustomerPayload {
  id: number;
  data: CustomerInput;
}

interface MergeCustomersPayload {
  primaryId: number;
  secondaryId: number;
}

export interface CustomerFullPayload extends Customer {
  loyalty?: Record<string, unknown>;
  visitHistory?: Record<string, unknown>[];
  tier?: string | null;
  loyaltyTransactions?: Record<string, unknown>[];
  recentVisits?: Record<string, unknown>[];
}

interface CustomerMutationResponse {
  customerId?: number;
  message?: string;
}

interface CustomerMutationOptions {
  notify?: boolean;
}

interface CustomerWireInput {
  firstName: string;
  lastName: string;
  email: string;
  phone: string;
  dateOfBirth?: string;
  gender?: string;
  address?: string;
  city?: string;
  country?: string;
  loyaltyTier?: string;
  tags: string[];
  notes: string;
  isActive?: boolean;
  consent?: Customer['consent'];
}

type CustomerCreateWireInput = Omit<CustomerWireInput, 'loyaltyTier' | 'isActive'>;

function splitName(
  input: CustomerInput,
  preferCombinedName = false
): { firstName: string; lastName: string } {
  if (preferCombinedName && String(input.name ?? '').trim()) {
    const parts = String(input.name).trim().split(/\s+/).filter(Boolean);
    return { firstName: parts.shift() ?? '', lastName: parts.join(' ') };
  }
  const firstName = String(input.firstName ?? '').trim();
  const lastName = String(input.lastName ?? '').trim();
  if (firstName) return { firstName, lastName };
  const parts = String(input.name ?? '')
    .trim()
    .split(/\s+/)
    .filter(Boolean);
  return { firstName: parts.shift() ?? '', lastName: lastName || parts.join(' ') };
}

function trimOptionalString(value: unknown): string | undefined {
  const trimmed = String(value ?? '').trim();
  return trimmed || undefined;
}

export function toCustomerWireInput(
  input: CustomerInput,
  preferCombinedName = false
): CustomerWireInput {
  const { firstName, lastName } = splitName(input, preferCombinedName);
  return {
    firstName,
    lastName,
    email: String(input.email ?? '').trim(),
    phone: String(input.phone ?? '').trim(),
    dateOfBirth: String(input.dateOfBirth ?? input.birthday ?? '').trim() || undefined,
    gender: trimOptionalString(input.gender),
    address: trimOptionalString(input.address),
    city: trimOptionalString(input.city),
    country: trimOptionalString(input.country),
    loyaltyTier: trimOptionalString(input.loyaltyTier),
    tags: Array.isArray(input.tags)
      ? input.tags.map((tag) => String(tag).trim()).filter(Boolean)
      : [],
    notes: String(input.notes ?? '').trim(),
    isActive: input.isActive,
    consent: input.consent,
  };
}

export function toCustomerCreateWireInput(input: CustomerInput): CustomerCreateWireInput {
  const wire = toCustomerWireInput(input);
  return {
    firstName: wire.firstName,
    lastName: wire.lastName,
    email: wire.email,
    phone: wire.phone,
    dateOfBirth: wire.dateOfBirth,
    gender: wire.gender,
    address: wire.address,
    city: wire.city,
    country: wire.country,
    tags: wire.tags,
    notes: wire.notes,
    consent: wire.consent,
  };
}

function normalizeCustomerResponse(value: unknown): Customer {
  const envelope = (value ?? {}) as { customer?: unknown };
  return customerSchema.parse(envelope.customer ?? value);
}

export interface CustomerTimelinePage {
  entries: TimelineEntry[];
  nextCursor?: string;
}

// ── Queries ──────────────────────────────────────────────────────────────────

const CUSTOMER_PAGE_SIZE = 100;

export function normalizeCustomerPageSize(value: number): number {
  const parsed = Number(value);
  return Number.isFinite(parsed)
    ? Math.max(1, Math.min(Math.floor(parsed), CUSTOMER_PAGE_SIZE))
    : CUSTOMER_PAGE_SIZE;
}

function normalizeCustomerOffset(value: number): number {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? Math.max(0, Math.floor(parsed)) : 0;
}

async function fetchCustomerPage(
  filters: CustomerListFilters,
  page: number,
  pageSize: number
): Promise<{ customers: Customer[]; total?: number }> {
  const res = await unwrap(
    safe.get<unknown>('/crm/customers', {
      params: {
        page,
        page_size: pageSize,
        search: filters.search,
        tier: filters.tier,
        tag: filters.tag,
      },
    })
  );
  const customers = z.array(customerSchema).parse(unwrapList(res, 'customers'));
  const rawTotal =
    res && typeof res === 'object' && !Array.isArray(res)
      ? (res as Record<string, unknown>).total
      : undefined;
  const total = Number(rawTotal);
  return { customers, total: Number.isFinite(total) ? total : undefined };
}

export const useListCustomers = (params: ListCustomersParams): UseQueryResult<Customer[]> =>
  useQuery<Customer[]>({
    queryKey: ['crm-customers', 'filter', params],
    queryFn: async (): Promise<Customer[]> => {
      const limit = normalizeCustomerPageSize(params.limit);
      const page = Math.floor(normalizeCustomerOffset(params.offset) / limit) + 1;
      const result = await fetchCustomerPage(params, page, limit);
      return result.customers;
    },
  });

/** Loads all customer pages for dashboards, campaign previews, and merge pickers. */
export const useListAllCustomers = (
  filters: CustomerListFilters = {},
  options: { enabled?: boolean } = {}
): UseQueryResult<Customer[]> =>
  useQuery<Customer[]>({
    queryKey: ['crm-customers', 'all', filters],
    enabled: options.enabled ?? true,
    queryFn: async (): Promise<Customer[]> => {
      const customers: Customer[] = [];
      let page = 1;
      const maxPages = 100;

      for (let pageIndex = 0; pageIndex < maxPages; pageIndex += 1) {
        const result = await fetchCustomerPage(filters, page, CUSTOMER_PAGE_SIZE);
        customers.push(...result.customers);
        if (
          result.customers.length < CUSTOMER_PAGE_SIZE ||
          (result.total !== undefined && customers.length >= result.total)
        ) {
          break;
        }
        page += 1;
      }

      return customers;
    },
  });

export const useGetCustomer = (id: number | undefined): UseQueryResult<Customer> =>
  useQuery<Customer>({
    queryKey: ['crm-customer', id],
    queryFn: async (): Promise<Customer> => {
      const res = await unwrap(safe.get<unknown>('/crm/customers/detail', { params: { id } }));
      return normalizeCustomerResponse(res);
    },
    enabled: id != null && Number.isInteger(id) && id > 0,
  });

export const useGetCustomerFull = (id: number | undefined): UseQueryResult<CustomerFullPayload> =>
  useQuery<CustomerFullPayload>({
    queryKey: ['crm-customer', id, 'full'],
    queryFn: async (): Promise<CustomerFullPayload> => {
      const res = await unwrap(safe.get<unknown>('/crm/customers/detail/full', { params: { id } }));
      const envelope = (res ?? {}) as {
        customer?: unknown;
        loyaltyTransactions?: Record<string, unknown>[];
        recentVisits?: Record<string, unknown>[];
        loyalty_transactions?: Record<string, unknown>[];
        recent_visits?: Record<string, unknown>[];
      };
      const customer = customerSchema.parse(envelope.customer ?? res);
      return {
        ...customer,
        loyaltyTransactions: envelope.loyaltyTransactions ?? envelope.loyalty_transactions ?? [],
        recentVisits: envelope.recentVisits ?? envelope.recent_visits ?? [],
        visitHistory: envelope.recentVisits ?? envelope.recent_visits ?? [],
      };
    },
    enabled: id != null && Number.isInteger(id) && id > 0,
  });

export const useCustomerTimeline = (
  id: number | undefined,
  cursor?: string
): UseQueryResult<CustomerTimelinePage> =>
  useQuery<CustomerTimelinePage>({
    queryKey: ['crm-customer', id, 'timeline', cursor],
    queryFn: async (): Promise<CustomerTimelinePage> => {
      const res = (await unwrap(
        safe.get<unknown>('/crm/customers/timeline', {
          params: { customer_id: id, cursor, limit: 50 },
        })
      )) as { entries?: unknown; nextCursor?: string; next_cursor?: string };
      const rawEntries = unwrapList(res, 'entries', 'timeline');
      const entries = rawEntries.map((entry) => {
        const value = (entry ?? {}) as Record<string, unknown>;
        const description = String(value.description ?? '');
        return {
          id: Number(value.id ?? 0),
          type: String(value.type ?? 'visit'),
          title: description || String(value.type ?? 'Activity'),
          description: description || undefined,
          at: String(value.createdAt ?? value.created_at ?? ''),
          meta: (value.metadata ?? value.meta) as Record<string, unknown> | undefined,
        };
      });
      return {
        entries: z.array(timelineEntrySchema).parse(entries),
        nextCursor: res.nextCursor ?? res.next_cursor,
      };
    },
    enabled: id != null && Number.isInteger(id) && id > 0,
  });

// ── Mutations ────────────────────────────────────────────────────────────────

export const useCreateCustomer = (): UseMutationResult<
  CustomerMutationResponse,
  unknown,
  CustomerInput
> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<CustomerMutationResponse, unknown, CustomerInput>({
    mutationFn: async (data: CustomerInput): Promise<CustomerMutationResponse> => {
      const res = await unwrap(
        safe.post<CustomerMutationResponse>('/crm/customers', toCustomerCreateWireInput(data))
      );
      return res as CustomerMutationResponse;
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-customers'] });
      toast.success(t('crm.customers.created'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useUpdateCustomer = ({
  notify = true,
}: CustomerMutationOptions = {}): UseMutationResult<
  CustomerMutationResponse,
  unknown,
  UpdateCustomerPayload
> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<CustomerMutationResponse, unknown, UpdateCustomerPayload>({
    mutationFn: async (payload: UpdateCustomerPayload): Promise<CustomerMutationResponse> => {
      const cachedFromDetails =
        qc.getQueryData<Customer>(['crm-customer', payload.id, 'full']) ??
        qc.getQueryData<Customer>(['crm-customer', payload.id]);
      // List-page edits (for example bulk tag updates) do not have a detail
      // query cached. Reuse the canonical row from any loaded customer page so
      // the backend's required first/last-name fields are preserved.
      const cachedFromLists = qc
        .getQueriesData<Customer[]>({ queryKey: ['crm-customers'] })
        .flatMap(([, rows]) => rows ?? [])
        .find((customer) => customer.id === payload.id);
      const cached = cachedFromDetails ?? cachedFromLists;
      const merged = toCustomerWireInput(
        { ...cached, ...payload.data },
        Object.prototype.hasOwnProperty.call(payload.data, 'name')
      );
      const res = await unwrap(
        safe.put<CustomerMutationResponse>('/crm/customers', { id: payload.id, ...merged })
      );
      return res as CustomerMutationResponse;
    },
    onSuccess: (_data: CustomerMutationResponse, variables: UpdateCustomerPayload): void => {
      void qc.invalidateQueries({ queryKey: ['crm-customers'] });
      void qc.invalidateQueries({ queryKey: ['crm-customer', variables.id] });
      if (notify) toast.success(t('crm.customers.updated'));
    },
    onError: (): void => {
      if (notify) toast.error(t('errors.generic'));
    },
  });
};

export const useDeleteCustomer = (): UseMutationResult<unknown, unknown, number> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, number>({
    mutationFn: (id: number): Promise<unknown> =>
      unwrap(safe.delete('/crm/customers', { data: { id } })),
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-customers'] });
      toast.success(t('crm.customers.deleted'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useMergeCustomers = (): UseMutationResult<unknown, unknown, MergeCustomersPayload> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, MergeCustomersPayload>({
    mutationFn: (payload: MergeCustomersPayload): Promise<unknown> =>
      unwrap(safe.post('/crm/customers/merge', payload)),
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-customers'] });
      toast.success(t('crm.customers.merged'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

interface ImportCustomerRowsPayload {
  customers: Record<string, unknown>[];
  dedupe: boolean;
}

export const useImportCustomerRows = (): UseMutationResult<
  unknown,
  unknown,
  ImportCustomerRowsPayload
> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, ImportCustomerRowsPayload>({
    mutationFn: (payload: ImportCustomerRowsPayload): Promise<unknown> =>
      unwrap(safe.post('/crm/customers/import', payload)),
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-customers'] });
      toast.success(t('crm.customers.imported'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export interface CustomerExportPayload {
  customer?: unknown;
  transactions?: unknown[];
  timeline?: unknown[];
}

export const useExportCustomers = (): UseMutationResult<CustomerExportPayload, unknown, number> => {
  const { t } = useTranslation();
  return useMutation<CustomerExportPayload, unknown, number>({
    mutationFn: (id: number): Promise<CustomerExportPayload> =>
      unwrap(
        safe.get<CustomerExportPayload>('/crm/customers/export', { params: { customer_id: id } })
      ) as Promise<CustomerExportPayload>,
    onSuccess: (): void => {
      toast.success(t('crm.customers.exported'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export type { Customer, TimelineEntry } from '../../schemas/customers';
