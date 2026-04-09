import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { hrApi, CreateDepartmentInput, UpdateDepartmentInput } from '../services/hrApi';

export function useDepartments() {
  return useQuery({
    queryKey: ['departments'],
    queryFn: () => hrApi.listDepartments().then((r) => r.data || []),
  });
}

export function useCreateDepartment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateDepartmentInput) => hrApi.createDepartment(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['departments'] }),
  });
}

export function useUpdateDepartment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateDepartmentInput) => hrApi.updateDepartment(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['departments'] }),
  });
}
