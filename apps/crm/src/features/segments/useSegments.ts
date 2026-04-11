import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { crmApi, Segment } from '../../shared/services/crmApi';

export function useSegments() {
  return useQuery({
    queryKey: ['segments'],
    queryFn: () => unwrap(crmApi.getSegments()),
  });
}

export function useSegmentById(id: number | undefined) {
  return useQuery({
    queryKey: ['segment', id],
    queryFn: () => unwrap(crmApi.getSegmentById(id!)),
    enabled: !!id,
  });
}

export function useSegmentCustomers(id: number | undefined) {
  return useQuery({
    queryKey: ['segment-customers', id],
    queryFn: () => unwrap(crmApi.getSegmentCustomers(id!)),
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
    mutationFn: ({ id, data }: { id: number; data: Partial<Segment> }) => unwrap(crmApi.updateSegment(id, data)),
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
