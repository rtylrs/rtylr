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
}

type CustomerInput = Partial<Customer>;

interface UpdateCustomerPayload {
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
}

// ── Queries ──────────────────────────────────────────────────────────────────

export const useListCustomers = (params: ListCustomersParams): UseQueryResult<Customer[]> =>
  useQuery<Customer[]>({
    queryKey: ['crm-customers', 'filter', params],
    queryFn: async (): Promise<Customer[]> => {
      const res = await unwrap(safe.get<unknown>('/crm/customers', { params }));
      return z.array(customerSchema).parse(unwrapList(res, 'customers'));
    },
  });

export const useGetCustomer = (id: number | undefined): UseQueryResult<Customer> =>
  useQuery<Customer>({
    queryKey: ['crm-customer', id],
    queryFn: async (): Promise<Customer> => {
      const res = await unwrap(safe.get<unknown>('/crm/customers/detail', { params: { id } }));
      return customerSchema.parse(res);
    },
    enabled: !!id,
  });

export const useGetCustomerFull = (id: number | undefined): UseQueryResult<CustomerFullPayload> =>
  useQuery<CustomerFullPayload>({
    queryKey: ['crm-customer', id, 'full'],
    queryFn: async (): Promise<CustomerFullPayload> => {
      const res = await unwrap(
        safe.get<CustomerFullPayload>('/crm/customers/detail/full', { params: { id } })
      );
      return res as CustomerFullPayload;
    },
    enabled: !!id,
  });

export const useCustomerTimeline = (id: number | undefined): UseQueryResult<TimelineEntry[]> =>
  useQuery<TimelineEntry[]>({
    queryKey: ['crm-customer', id, 'timeline'],
    queryFn: async (): Promise<TimelineEntry[]> => {
      const res = await unwrap(safe.get<unknown>('/crm/customers/timeline', { params: { id } }));
      return z.array(timelineEntrySchema).parse(unwrapList(res, 'entries', 'timeline'));
    },
    enabled: !!id,
  });

// ── Mutations ────────────────────────────────────────────────────────────────

export const useCreateCustomer = (): UseMutationResult<Customer, unknown, CustomerInput> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<Customer, unknown, CustomerInput>({
    mutationFn: async (data: CustomerInput): Promise<Customer> => {
      const res = await unwrap(safe.post<Customer>('/crm/customers', data));
      return res as Customer;
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

export const useUpdateCustomer = (): UseMutationResult<
  Customer,
  unknown,
  UpdateCustomerPayload
> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<Customer, unknown, UpdateCustomerPayload>({
    mutationFn: async (payload: UpdateCustomerPayload): Promise<Customer> => {
      const res = await unwrap(
        safe.put<Customer>('/crm/customers', { id: payload.id, ...payload.data })
      );
      return res as Customer;
    },
    onSuccess: (_data: Customer, variables: UpdateCustomerPayload): void => {
      void qc.invalidateQueries({ queryKey: ['crm-customers'] });
      void qc.invalidateQueries({ queryKey: ['crm-customer', variables.id] });
      toast.success(t('crm.customers.updated'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
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

export const useRecordVisit = (): UseMutationResult<unknown, unknown, number> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, number>({
    mutationFn: (customerId: number): Promise<unknown> =>
      unwrap(safe.post('/crm/customers/record-visit', { customerId })),
    onSuccess: (_data: unknown, customerId: number): void => {
      void qc.invalidateQueries({ queryKey: ['crm-customers'] });
      void qc.invalidateQueries({ queryKey: ['crm-customer', customerId] });
      toast.success(t('crm.customers.visitRecorded'));
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

export const useImportCustomers = (): UseMutationResult<unknown, unknown, File> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, File>({
    mutationFn: async (file: File): Promise<unknown> => {
      const form = new FormData();
      form.append('file', file);
      return unwrap(
        safe.post('/crm/customers/import', form, {
          headers: { 'Content-Type': 'multipart/form-data' },
        })
      );
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-customers'] });
      toast.success(t('crm.customers.imported'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

interface ImportCustomerRowsPayload {
  customers: Record<string, string>[];
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
      unwrap(safe.post('/crm/customers/import', { customers: payload.customers })),
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-customers'] });
      toast.success(t('crm.customers.imported'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useExportCustomers = (): UseMutationResult<unknown, unknown, number> => {
  const { t } = useTranslation();
  return useMutation<unknown, unknown, number>({
    mutationFn: (id: number): Promise<unknown> =>
      unwrap(safe.get('/crm/customers/export', { params: { id } })),
    onSuccess: (): void => {
      toast.success(t('crm.customers.exported'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export type { Customer, TimelineEntry } from '../../schemas/customers';
