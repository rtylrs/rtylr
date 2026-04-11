import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { hrApi, CreateDepartmentInput, UpdateDepartmentInput } from '../../shared/services/hrApi';

export function useDepartments() {
  return useQuery({
    queryKey: ['departments'],
    queryFn: () => unwrap(hrApi.listDepartments()),
  });
}

export function useCreateDepartment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateDepartmentInput) => unwrap(hrApi.createDepartment(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['departments'] }),
  });
}

export function useUpdateDepartment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateDepartmentInput) => unwrap(hrApi.updateDepartment(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['departments'] }),
  });
}
