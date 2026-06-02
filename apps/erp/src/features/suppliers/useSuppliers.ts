import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  erpApi,
  CreateSupplierInput,
  UpdateSupplierInput,
  Supplier,
} from '../../shared/services/erpApi';
import { unwrapList } from '../../shared/lib/unwrapList';

export interface SupplierListResult {
  suppliers: Supplier[];
  hasMore: boolean;
  page: number;
  pageSize: number;
}

export interface UseSupplierParams {
  page?: number;
  pageSize?: number;
}

// Backend (erp.v1.Supplier) shape, camelCased by @voxire/api on the way in.
// The backend exposes `isActive` (boolean), not a `status` string, and does not
// return contactName, productCount, categories, or any performance metrics.
const backendSupplierSchema = z.object({
  id: z.number(),
  name: z.string(),
  email: z.string().default(''),
  phone: z.string().default(''),
  address: z.string().default(''),
  city: z.string().default(''),
  country: z.string().default(''),
  taxId: z.string().default(''),
  paymentTerms: z.string().default(''),
  notes: z.string().default(''),
  isActive: z.boolean().default(true),
  createdAt: z.string().default(''),
  updatedAt: z.string().default(''),
});

type BackendSupplier = z.infer<typeof backendSupplierSchema>;

// Adapt the backend supplier onto the shared Supplier shape the UI expects.
// Fields the backend never returns degrade to safe defaults: contactName -> '',
// productCount -> 0, categories -> [], and the boolean isActive maps to the
// display-cased Active/Inactive status. Performance metrics stay undefined so
// the scorecard renders its empty/"-" state.
function adaptSupplier(s: BackendSupplier): Supplier {
  return {
    id: s.id,
    name: s.name,
    contactName: '',
    email: s.email,
    phone: s.phone,
    productCount: 0,
    categories: [],
    status: s.isActive ? 'Active' : 'Inactive',
  };
}

export function useSuppliers(params: UseSupplierParams = {}) {
  const page = params.page ?? 1;
  const pageSize = params.pageSize ?? 50;
  return useQuery<SupplierListResult>({
    queryKey: ['suppliers', { page, pageSize }],
    queryFn: async () => {
      const res = await unwrap(erpApi.listSuppliers({ page, page_size: pageSize }));
      const suppliers = z
        .array(backendSupplierSchema)
        .parse(unwrapList(res, 'suppliers'))
        .map(adaptSupplier);
      const hasMore = (res as unknown as Record<string, unknown>)?.hasMore === true;
      return { suppliers, hasMore, page, pageSize };
    },
  });
}

export function useCreateSupplier() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateSupplierInput) => unwrap(erpApi.createSupplier(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['suppliers'] }),
    onError: () => toast.error('Failed to create supplier'),
  });
}

export function useUpdateSupplier() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateSupplierInput) => unwrap(erpApi.updateSupplier(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['suppliers'] }),
    onError: () => toast.error('Failed to update supplier'),
  });
}
