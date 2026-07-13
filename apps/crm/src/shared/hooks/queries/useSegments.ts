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

function rulesToFilters(rules: string[] | undefined, existing?: string): string {
  if (!rules) return existing ?? '{}';
  return JSON.stringify({ rules });
}

export const useListSegments = (): UseQueryResult<Segment[]> =>
  useQuery<Segment[]>({
    queryKey: ['crm-segments'],
    queryFn: async (): Promise<Segment[]> => {
      const res = await segmentClient.listSegments({});
      return z.array(segmentSchema).parse(
        (res.segments ?? []).map((s) => {
          const segment = fromApiSegment(s);
          return { ...segment, rules: filtersToRules(segment.filters) };
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
      return { ...segment, rules: filtersToRules(segment.filters) };
    },
    enabled: !!id,
  });

export const useSegmentCustomers = (id: string | undefined): UseQueryResult<Customer[]> =>
  useQuery<Customer[]>({
    queryKey: ['crm-segment', id, 'customers'],
    queryFn: async (): Promise<Customer[]> => {
      const res = await segmentClient.getSegmentCustomers({
        segmentId: id ?? '',
        page: String(1),
        pageSize: String(50),
      });
      return z.array(customerSchema).parse((res.customers ?? []).map(fromApiSegmentCustomer));
    },
    enabled: !!id,
  });

export const useCreateSegment = (): UseMutationResult<Segment, unknown, Partial<Segment>> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<Segment, unknown, Partial<Segment>>({
    mutationFn: async (data: Partial<Segment>): Promise<Segment> => {
      const res = await segmentClient.createSegment({
        name: data.name ?? '',
        description: data.description ?? '',
        filters: rulesToFilters(data.rules, data.filters),
      });
      return segmentSchema.parse({
        id: res.segmentId,
        name: data.name ?? '',
        description: data.description ?? '',
        filters: rulesToFilters(data.rules, data.filters),
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
      await segmentClient.updateSegment({
        id: payload.id,
        body: {
          name: payload.data.name ?? '',
          description: payload.data.description ?? '',
          filters: rulesToFilters(payload.data.rules, payload.data.filters),
        },
      });
      return segmentSchema.parse({
        id: payload.id,
        name: payload.data.name ?? '',
        description: payload.data.description ?? '',
        filters: rulesToFilters(payload.data.rules, payload.data.filters),
        customerCount: payload.data.customerCount ?? 0,
        color: payload.data.color,
        rules: payload.data.rules,
        trend: payload.data.trend,
        trendUp: payload.data.trendUp,
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
