import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
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
import { SegmentServiceClient } from '@voxire/voxie/crm/segment/v1/client';
import type { Segment as ApiSegment } from '@voxire/voxie/crm/segment/v1/types';
import type { Customer as ApiSegmentCustomer } from '@voxire/voxie/crm/customer/v1/types';
import { segmentSchema } from '../../schemas/segments';
import type { Segment } from '../../schemas/segments';
import { customerSchema } from '../../schemas/customers';
import type { Customer } from '../../schemas/customers';

const segmentClient = createOneKitClient(SegmentServiceClient, 'crm');

interface UpdateSegmentPayload {
  id: string;
  data: Partial<Segment>;
}

export function mergeSegmentData(
  current: Segment | undefined,
  data: Partial<Segment>
): Partial<Segment> {
  return { ...current, ...data };
}

function fromApiSegment(s: ApiSegment): Segment {
  return segmentSchema.parse({
    id: s.id,
    ownerId: s.ownerId,
    name: s.name,
    description: s.description,
    filters: s.filters,
    customerCount: s.customerCount,
    createdAt: s.createdAt,
    updatedAt: s.updatedAt,
  });
}

function fromApiSegmentCustomer(c: ApiSegmentCustomer): Customer {
  return customerSchema.parse(c);
}

// Segments only carry a single `filters` JSON string on the wire; the UI's
// rule-builder produces a `rules: string[]` array that gets wrapped into
// `{ rules }` before serializing, and unwrapped back out when read.
function filtersToRules(filters: string): string[] {
  try {
    const parsed: unknown = JSON.parse(filters);
    if (
      parsed &&
      typeof parsed === 'object' &&
      Array.isArray((parsed as { rules?: unknown }).rules)
    ) {
      return (parsed as { rules: unknown[] }).rules.filter(
        (r): r is string => typeof r === 'string'
      );
    }
  } catch {
    /* fall through */
  }
  return [];
}

function matchAllFromFilters(filters: string): boolean {
  try {
    const parsed: unknown = JSON.parse(filters);
    if (parsed && typeof parsed === 'object' && 'matchAll' in parsed) {
      return (parsed as { matchAll?: unknown }).matchAll !== false;
    }
  } catch {
    /* fall through */
  }
  return true;
}

function rulesToFilters(rules: string[] | undefined, existing?: string, matchAll = true): string {
  if (!rules) return existing ?? '{}';
  return JSON.stringify({ rules, matchAll });
}

export const useListSegments = (): UseQueryResult<Segment[]> =>
  useQuery<Segment[]>({
    queryKey: ['crm-segments'],
    queryFn: async (): Promise<Segment[]> => {
      const res = await segmentClient.listSegments({});
      return z.array(segmentSchema).parse(
        (res.segments ?? []).map((s) => {
          const segment = fromApiSegment(s);
          return {
            ...segment,
            rules: filtersToRules(segment.filters),
            matchAll: matchAllFromFilters(segment.filters),
          };
        })
      );
    },
  });

export const useGetSegment = (id: string | undefined): UseQueryResult<Segment> =>
  useQuery<Segment>({
    queryKey: ['crm-segment', id],
    queryFn: async (): Promise<Segment> => {
      const res = await segmentClient.getSegmentDetail({ id: id ?? '' });
      if (!res.segment) throw new Error('Segment not found');
      const segment = fromApiSegment(res.segment);
      return {
        ...segment,
        rules: filtersToRules(segment.filters),
        matchAll: matchAllFromFilters(segment.filters),
      };
    },
    enabled: id != null && Number.isInteger(Number(id)) && Number(id) > 0,
  });

export const useSegmentCustomers = (id: string | undefined): UseQueryResult<Customer[]> =>
  useQuery<Customer[]>({
    queryKey: ['crm-segment', id, 'customers'],
    queryFn: async (): Promise<Customer[]> => {
      const customers: Customer[] = [];
      const pageSize = 100;
      const maxPages = 100;

      for (let page = 1; page <= maxPages; page += 1) {
        const res = await segmentClient.getSegmentCustomers({
          segmentId: id ?? '',
          page: String(page),
          pageSize: String(pageSize),
        });
        const pageCustomers = z
          .array(customerSchema)
          .parse((res.customers ?? []).map(fromApiSegmentCustomer));
        customers.push(...pageCustomers);

        const total = Number(res.total);
        if (
          pageCustomers.length === 0 ||
          (Number.isFinite(total) ? customers.length >= total : pageCustomers.length < pageSize)
        ) {
          break;
        }
      }

      return customers;
    },
    enabled: id != null && Number.isInteger(Number(id)) && Number(id) > 0,
  });

export const useCreateSegment = (): UseMutationResult<Segment, unknown, Partial<Segment>> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<Segment, unknown, Partial<Segment>>({
    mutationFn: async (data: Partial<Segment>): Promise<Segment> => {
      const res = await segmentClient.createSegment({
        name: data.name ?? '',
        description: data.description?.trim() || data.name?.trim() || 'Customer segment',
        filters: rulesToFilters(data.rules, data.filters, data.matchAll),
      });
      return segmentSchema.parse({
        id: res.segmentId,
        name: data.name ?? '',
        description: data.description?.trim() || data.name?.trim() || 'Customer segment',
        filters: rulesToFilters(data.rules, data.filters, data.matchAll),
        customerCount: 0,
        color: data.color,
        rules: data.rules,
        trend: data.trend,
        trendUp: data.trendUp,
      });
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-segments'] });
      toast.success(t('crm.segments.created'));
    },
    onError: (error): void => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
};

export const useUpdateSegment = (): UseMutationResult<Segment, unknown, UpdateSegmentPayload> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<Segment, unknown, UpdateSegmentPayload>({
    mutationFn: async (payload: UpdateSegmentPayload): Promise<Segment> => {
      const current = qc
        .getQueryData<Segment[]>(['crm-segments'])
        ?.find((segment) => segment.id === payload.id);
      const merged = mergeSegmentData(current, payload.data);
      await segmentClient.updateSegment({
        id: payload.id,
        body: {
          name: merged.name ?? '',
          description: merged.description?.trim() || merged.name?.trim() || 'Customer segment',
          filters: rulesToFilters(merged.rules, merged.filters, merged.matchAll),
        },
      });
      return segmentSchema.parse({
        id: payload.id,
        name: merged.name ?? '',
        description: merged.description?.trim() || merged.name?.trim() || 'Customer segment',
        filters: rulesToFilters(merged.rules, merged.filters, merged.matchAll),
        customerCount: merged.customerCount ?? 0,
        color: merged.color,
        rules: merged.rules,
        trend: merged.trend,
        trendUp: merged.trendUp,
        matchAll: merged.matchAll,
      });
    },
    onSuccess: (_data: Segment, variables: UpdateSegmentPayload): void => {
      void qc.invalidateQueries({ queryKey: ['crm-segments'] });
      void qc.invalidateQueries({ queryKey: ['crm-segment', variables.id] });
      toast.success(t('crm.segments.updated'));
    },
    onError: (error): void => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
};

export const useDeleteSegment = (): UseMutationResult<unknown, unknown, string> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, string>({
    mutationFn: (id: string) => segmentClient.deleteSegment({ id }),
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-segments'] });
      toast.success(t('crm.segments.deleted'));
    },
    onError: (error): void => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
};

export type { Segment } from '../../schemas/segments';
