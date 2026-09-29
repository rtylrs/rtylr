import { createOneKitClient } from '@voxire/api';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { DepartmentServiceClient } from '@voxire/voxie/hr/department/v1/client';
import type { CreateDepartmentRequest } from '@voxire/voxie/hr/department/v1/types';
import { StaffServiceClient } from '@voxire/voxie/hr/staff/v1/client';
import { departmentSchema, type Department } from '../../shared/schemas/departments';

const departmentClient = createOneKitClient(DepartmentServiceClient, 'hr');
const staffClient = createOneKitClient(StaffServiceClient, 'hr');

export function useDepartments(includeCompensation = false) {
  return useQuery({
    queryKey: ['departments', includeCompensation],
    queryFn: async (): Promise<Department[]> => {
      const departmentsRes = await departmentClient.listDepartments({});
      const departments = departmentsRes.departments ?? [];
      const staffByDepartment = await Promise.all(
        departments.map(async (department) => {
          const response = await staffClient.listStaff({
            search: '',
            departmentId: String(department.id),
            status: '',
          });
          return [department.id, response.staff ?? []] as const;
        })
      );
      const staffByDepartmentId = new Map(staffByDepartment);
      const staffList = staffByDepartment.flatMap(([, staff]) => staff);
      const staffById = new Map(staffList.map((s) => [s.id, s]));

      return departments.map((d) => {
        const members = staffByDepartmentId.get(d.id) ?? [];
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

export function useCreateDepartment() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateDepartmentRequest) => departmentClient.createDepartment(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['departments'] }),
    onError: () => toast.error(t('errors.createDepartment')),
  });
}
