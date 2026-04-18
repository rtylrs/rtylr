import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { crmApi, Segment } from '../../shared/services/crmApi';
import { segmentSchema } from '../../shared/schemas/segments';
import { customerSchema } from '../../shared/schemas/customers';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useSegments() {
  return useQuery({
    queryKey: ['segments'],
    queryFn: async () => {
      const res = await unwrap(crmApi.getSegments());
      return z.array(segmentSchema).parse(unwrapList(res, 'segments'));
    },
  });
}

export function useSegmentById(id: number | undefined) {
  return useQuery({
    queryKey: ['segment', id],
    queryFn: async () => {
      const res = await unwrap(crmApi.getSegmentById(id!));
      return segmentSchema.parse(res);
    },
    enabled: !!id,
  });
}

export function useSegmentCustomers(id: number | undefined) {
  return useQuery({
    queryKey: ['segment-customers', id],
    queryFn: async () => {
      const res = await unwrap(crmApi.getSegmentCustomers(id!));
      return z.array(customerSchema).parse(unwrapList(res, 'customers'));
    },
    enabled: !!id,
  });
}

export function useCreateSegment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Segment>) => unwrap(crmApi.createSegment(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['segments'] }),
  });
}

export function useUpdateSegment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<Segment> }) =>
      unwrap(crmApi.updateSegment(id, data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['segments'] }),
  });
}

export function useDeleteSegment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(crmApi.deleteSegment(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['segments'] }),
  });
}
