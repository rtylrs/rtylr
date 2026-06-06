import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { toast } from 'sonner';
import { hrApi, CreateDepartmentInput, UpdateDepartmentInput } from '../../shared/services/hrApi';
import { departmentSchema } from '../../shared/schemas/departments';

export function useDepartments() {
  return useQuery({
    queryKey: ['departments'],
    queryFn: async () => {
      const res = await unwrap(hrApi.listDepartments());
      const raw = Array.isArray(res) ? res : ((res as Record<string, unknown>)?.departments ?? []);
      return z.array(departmentSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useCreateDepartment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateDepartmentInput) => unwrap(hrApi.createDepartment(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['departments'] }),
    onError: () => toast.error('Failed to create department'),
  });
}

export function useUpdateDepartment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateDepartmentInput) => unwrap(hrApi.updateDepartment(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['departments'] }),
    onError: () => toast.error('Failed to update department'),
  });
}
