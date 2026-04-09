import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { hrApi, CreateStaffInput, UpdateStaffInput } from '../services/hrApi';

export function useStaff() {
  return useQuery({
    queryKey: ['staff'],
    queryFn: () => hrApi.listStaff().then((r) => r.data || []),
  });
}

export function useCreateStaff() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateStaffInput) => hrApi.createStaff(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['staff'] }),
  });
}

export function useUpdateStaff() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateStaffInput) => hrApi.updateStaff(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['staff'] }),
  });
}

export function useDeleteStaff() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => hrApi.deleteStaff(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['staff'] }),
  });
}

export function useClockIn() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (staffId: number) => hrApi.clockIn(staffId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['staff'] });
      qc.invalidateQueries({ queryKey: ['hr-dashboard'] });
    },
  });
}

export function useClockOut() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (staffId: number) => hrApi.clockOut(staffId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['staff'] });
      qc.invalidateQueries({ queryKey: ['hr-dashboard'] });
    },
  });
}
