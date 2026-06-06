import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { toast } from 'sonner';
import { hrApi, CreateStaffInput, UpdateStaffInput } from '../../shared/services/hrApi';
import { staffMemberSchema } from '../../shared/schemas/staff';

export function useStaff() {
  return useQuery({
    queryKey: ['staff'],
    queryFn: async () => {
      const res = await unwrap(hrApi.listStaff());
      const raw = res ?? [];
      return z.array(staffMemberSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useCreateStaff() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateStaffInput) => unwrap(hrApi.createStaff(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['staff'] }),
    onError: () => toast.error('Failed to create staff member'),
  });
}

export function useUpdateStaff() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateStaffInput) => unwrap(hrApi.updateStaff(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['staff'] }),
    onError: () => toast.error('Failed to update staff member'),
  });
}

export function useDeleteStaff() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrApi.deleteStaff(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['staff'] }),
    onError: () => toast.error('Failed to remove staff member'),
  });
}

export function useClockIn() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (staffId: number) => unwrap(hrApi.clockIn(staffId)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['staff'] });
      qc.invalidateQueries({ queryKey: ['hr-dashboard'] });
    },
    onError: () => toast.error('Failed to clock in'),
  });
}

export function useClockOut() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (staffId: number) => unwrap(hrApi.clockOut(staffId)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['staff'] });
      qc.invalidateQueries({ queryKey: ['hr-dashboard'] });
    },
    onError: () => toast.error('Failed to clock out'),
  });
}
