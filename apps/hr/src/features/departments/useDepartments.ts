import { createOneKitClient, unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { DepartmentServiceClient } from '@voxire/voxie/hr/department/v1/client';
import { StaffServiceClient } from '@voxire/voxie/hr/staff/v1/client';
import { departmentSchema, type Department } from '../../shared/schemas/departments';
import { hrApi, CreateDepartmentInput, UpdateDepartmentInput } from '../../shared/services/hrApi';

const departmentClient = createOneKitClient(DepartmentServiceClient, 'hr');
const staffClient = createOneKitClient(StaffServiceClient, 'hr');

export function useDepartments(includeCompensation = false) {
  return useQuery({
    queryKey: ['departments', includeCompensation],
    queryFn: async (): Promise<Department[]> => {
      const [departmentsRes, staffRes] = await Promise.all([
        departmentClient.listDepartments({}),
        staffClient.listStaff({ search: '', departmentId: '', status: '' }),
      ]);
      const staffList = staffRes.staff ?? [];
      const staffById = new Map(staffList.map((s) => [s.id, s]));

      return (departmentsRes.departments ?? []).map((d) => {
        const members = staffList.filter((s) => s.departmentId === d.id);
        const manager = staffById.get(d.managerStaffId);
        const salaries = includeCompensation
          ? members.map((s) => s.salary ?? 0).filter((s) => s > 0)
          : [];
        const avgSalary = salaries.length
          ? salaries.reduce((sum, s) => sum + s, 0) / salaries.length
          : 0;

        return departmentSchema.parse({
          id: d.id,
          name: d.name,
          description: d.description,
          staffCount: members.length,
          head: manager ? `${manager.firstName} ${manager.lastName}`.trim() : '',
          avgSalary: includeCompensation ? avgSalary : undefined,
        });
      });
    },
  });
}

// hr.department.v1 has no create/update consumers wired in the UI yet - these
// remain on the legacy (currently 404ing) hrApi paths until a real screen
// needs them.
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
