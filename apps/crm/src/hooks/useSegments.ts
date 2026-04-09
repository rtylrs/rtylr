import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { crmApi, Segment } from '../services/crmApi';

export function useSegments() {
  return useQuery({
    queryKey: ['segments'],
    queryFn: () => crmApi.getSegments(),
  });
}

export function useSegmentById(id: number | undefined) {
  return useQuery({
    queryKey: ['segment', id],
    queryFn: () => crmApi.getSegmentById(id!),
    enabled: !!id,
  });
}

export function useSegmentCustomers(id: number | undefined) {
  return useQuery({
    queryKey: ['segment-customers', id],
    queryFn: () => crmApi.getSegmentCustomers(id!),
    enabled: !!id,
  });
}

export function useCreateSegment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Segment>) => crmApi.createSegment(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['segments'] }),
  });
}

export function useUpdateSegment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<Segment> }) => crmApi.updateSegment(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['segments'] }),
  });
}

export function useDeleteSegment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => crmApi.deleteSegment(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['segments'] }),
  });
}
