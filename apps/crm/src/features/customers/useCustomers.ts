import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { crmApi, Customer, CustomersParams } from '../../shared/services/crmApi';
import { customerSchema, timelineEntrySchema } from '../../shared/schemas/customers';

import { unwrapList } from '../../shared/lib/unwrapList';

export function useCustomers(params?: CustomersParams) {
  return useQuery({
    queryKey: ['customers', params],
    queryFn: async () => {
      const res = await unwrap(crmApi.getCustomers(params));
      return z.array(customerSchema).parse(unwrapList(res, 'customers'));
    },
  });
}

export function useCustomer(id: number | undefined) {
  return useQuery({
    queryKey: ['customer', id],
    queryFn: async () => {
      const res = await unwrap(crmApi.getCustomerDetail(id!));
      return customerSchema.parse(res);
    },
    enabled: !!id,
  });
}

export function useCustomerDetail(id: number | undefined) {
  return useCustomer(id);
}

export function useCustomerTimeline(id: number | undefined) {
  return useQuery({
    queryKey: ['customer-timeline', id],
    queryFn: async () => {
      const res = await unwrap(crmApi.getCustomerTimeline(id!));
      return z.array(timelineEntrySchema).parse(unwrapList(res, 'entries', 'timeline'));
    },
    enabled: !!id,
  });
}

export function useCreateCustomer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Customer>) => unwrap(crmApi.createCustomer(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['customers'] }),
  });
}

export function useUpdateCustomer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<Customer> }) =>
      unwrap(crmApi.updateCustomer(id, data)),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ['customers'] });
      qc.invalidateQueries({ queryKey: ['customer', vars.id] });
    },
  });
}

export function useDeleteCustomer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(crmApi.deleteCustomer(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['customers'] }),
  });
}

export function useRecordVisit() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (customerId: number) => unwrap(crmApi.recordVisit(customerId)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['customers'] }),
  });
}

export function useImportCustomers() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (rows: Record<string, unknown>[]) => unwrap(crmApi.importCustomers(rows)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['customers'] }),
  });
}

export function useMergeCustomers() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ primaryId, secondaryId }: { primaryId: number; secondaryId: number }) =>
      unwrap(crmApi.mergeCustomers(primaryId, secondaryId)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['customers'] }),
  });
}

export function useExportCustomerData() {
  return useMutation({
    mutationFn: (id: number) => unwrap(crmApi.exportCustomerData(id)),
  });
}
