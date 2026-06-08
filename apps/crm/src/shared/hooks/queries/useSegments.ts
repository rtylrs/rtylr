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
import { segmentSchema } from '../../schemas/segments';
import type { Segment } from '../../schemas/segments';
import { customerSchema } from '../../schemas/customers';
import type { Customer } from '../../schemas/customers';

interface UpdateSegmentPayload {
  id: number;
  data: Partial<Segment>;
}

export const useListSegments = (): UseQueryResult<Segment[]> =>
  useQuery<Segment[]>({
    queryKey: ['crm-segments'],
    queryFn: async (): Promise<Segment[]> => {
      const res = await unwrap(safe.get<unknown>('/crm/segments'));
      return z.array(segmentSchema).parse(unwrapList(res, 'segments'));
    },
  });

export const useGetSegment = (id: number | undefined): UseQueryResult<Segment> =>
  useQuery<Segment>({
    queryKey: ['crm-segment', id],
    queryFn: async (): Promise<Segment> => {
      // The live segment-detail route is GET /crm/segments/detail?id= (the
      // segments module does not register a GET /crm/segments/{id}). It wraps
      // the segment in a `{ segment, customers }` envelope.
      const res = await unwrap(safe.get<unknown>('/crm/segments/detail', { params: { id } }));
      const envelope = (res ?? {}) as { segment?: unknown };
      return segmentSchema.parse(envelope.segment ?? res);
    },
    enabled: !!id,
  });

export const useSegmentCustomers = (id: number | undefined): UseQueryResult<Customer[]> =>
  useQuery<Customer[]>({
    queryKey: ['crm-segment', id, 'customers'],
    queryFn: async (): Promise<Customer[]> => {
      // The segment-customers handler reads the `segment_id` query parameter.
      // Query params are sent verbatim (the api interceptor only converts the
      // request body camelCase → snake_case), so the snake_case key must be
      // passed literally here.
      const res = await unwrap(
        safe.get<unknown>('/crm/segments/customers', { params: { segment_id: id } })
      );
      return z.array(customerSchema).parse(unwrapList(res, 'customers'));
    },
    enabled: !!id,
  });

export const useCreateSegment = (): UseMutationResult<Segment, unknown, Partial<Segment>> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<Segment, unknown, Partial<Segment>>({
    mutationFn: async (data: Partial<Segment>): Promise<Segment> => {
      const rules = data.rules ?? [];
      const body = {
        name: data.name,
        description: data.description,
        filters: data.filters ?? { rules },
      };
      const res = await unwrap(safe.post<Segment>('/crm/segments', body));
      return res as Segment;
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-segments'] });
      toast.success(t('crm.segments.created'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useUpdateSegment = (): UseMutationResult<Segment, unknown, UpdateSegmentPayload> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<Segment, unknown, UpdateSegmentPayload>({
    mutationFn: async (payload: UpdateSegmentPayload): Promise<Segment> => {
      const res = await unwrap(
        safe.put<Segment>('/crm/segments', { id: payload.id, ...payload.data })
      );
      return res as Segment;
    },
    onSuccess: (_data: Segment, variables: UpdateSegmentPayload): void => {
      void qc.invalidateQueries({ queryKey: ['crm-segments'] });
      void qc.invalidateQueries({ queryKey: ['crm-segment', variables.id] });
      toast.success(t('crm.segments.updated'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useDeleteSegment = (): UseMutationResult<unknown, unknown, number> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, number>({
    mutationFn: (id: number): Promise<unknown> =>
      unwrap(safe.delete('/crm/segments', { data: { id } })),
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-segments'] });
      toast.success(t('crm.segments.deleted'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export type { Segment } from '../../schemas/segments';
