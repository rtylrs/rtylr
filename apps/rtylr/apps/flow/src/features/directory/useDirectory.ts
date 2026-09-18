import { useQuery } from '@tanstack/react-query';
import { createOneKitClient } from '@voxire/api';
import { StaffServiceClient } from '@voxire/voxie/hr/staff/v1/client';
import { DepartmentServiceClient } from '@voxire/voxie/hr/department/v1/client';
import type { Staff } from '@voxire/voxie/hr/staff/v1/types';

const staffClient = createOneKitClient(StaffServiceClient, 'flow');
const departmentClient = createOneKitClient(DepartmentServiceClient, 'flow');

export interface DirectoryMember extends Staff {
  departmentName?: string;
}

/**
 * The whole business's roster, gated server-side by view_directory (onion
 * PR #267) - a caller without it gets a real 403 from listStaff, not just a
 * hidden UI. Sensitive fields (address, emergency contact) are already
 * field-gated separately and simply arrive absent when not permitted.
 */
export function useDirectory(search?: string) {
  return useQuery({
    queryKey: ['directory', search],
    queryFn: async (): Promise<DirectoryMember[]> => {
      const [staffResponse, departmentsResponse] = await Promise.all([
        staffClient.listStaff({ search: search || undefined }),
        departmentClient.listDepartments({}),
      ]);
      const departments = departmentsResponse.departments ?? [];
      return (staffResponse.staff ?? []).map((member) => ({
        ...member,
        departmentName: departments.find((d) => d.id === member.departmentId)?.name,
      }));
    },
  });
}
