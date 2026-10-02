import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import i18next from 'i18next';
import {
  useMutation,
  useInfiniteQuery,
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

export interface SegmentCustomersPage {
  customers: Customer[];
  total: number;
  page: number;
  pageSize: number;
}

const SEGMENT_CUSTOMERS_PAGE_SIZE = 50;

function positiveNumber(value: string | undefined, fallback: number): number {
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback;
}

export function parseSegmentCustomersPage(
  response: {
    customers?: ApiSegmentCustomer[];
    total?: string;
    page?: string;
    pageSize?: string;
  },
  requestedPage: number
): SegmentCustomersPage {
  const customers = (response.customers ?? []).map(fromApiSegmentCustomer);
  const page = positiveNumber(response.page, requestedPage);
  const pageSize = positiveNumber(response.pageSize, SEGMENT_CUSTOMERS_PAGE_SIZE);
  const total = positiveNumber(response.total, customers.length);
  return { customers, total, page, pageSize };
}

export function getNextSegmentCustomerPage(page: SegmentCustomersPage): number | undefined {
  if (page.customers.length === 0 || page.page * page.pageSize >= page.total) {
    return undefined;
  }
  return page.page + 1;
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
      if (!res.segment) throw new Error(i18next.t('crm.segmentDetail.notFoundTitle'));
      const segment = fromApiSegment(res.segment);
      return { ...segment, rules: filtersToRules(segment.filters) };
    },
    enabled: !!id,
  });

export const useSegmentCustomers = (id: string | undefined) =>
  useInfiniteQuery({
    queryKey: ['crm-segment', id, 'customers'],
    initialPageParam: 1,
    queryFn: async ({ pageParam }): Promise<SegmentCustomersPage> => {
      const res = await segmentClient.getSegmentCustomers({
        segmentId: id ?? '',
        page: String(pageParam),
        pageSize: String(SEGMENT_CUSTOMERS_PAGE_SIZE),
      });
      return parseSegmentCustomersPage(res, pageParam);
    },
    getNextPageParam: getNextSegmentCustomerPage,
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
      toast.error(getOneKitErrorMessage(error, t('crmErrors.generic')));
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
      toast.error(getOneKitErrorMessage(error, t('crmErrors.generic')));
    },
  });
};

export type { Segment } from '../../schemas/segments';
