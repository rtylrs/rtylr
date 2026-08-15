import { useQuery } from '@tanstack/react-query';
import { createOneKitClient, unwrap } from '@voxire/api';
import { StaffServiceClient } from '@voxire/voxie/hr/staff/v1/client';
import { DepartmentServiceClient } from '@voxire/voxie/hr/department/v1/client';
import type { Staff } from '@voxire/voxie/hr/staff/v1/types';
import { safe } from '../../shared/lib/api';
import { profileDetailSchema } from '../../shared/schemas/profile';

const staffClient = createOneKitClient(StaffServiceClient, 'flow');
const departmentClient = createOneKitClient(DepartmentServiceClient, 'flow');

export interface MyStaffRecord extends Staff {
  departmentName?: string;
}

/**
 * `hr.staff.v1` has no self-lookup or `user_id` filter, so resolving "my own
 * record" means listing the business's staff and matching client-side on
 * `user_id` - the same data any authenticated business member can already
 * fetch through this same endpoint today. Sensitive fields (salary, address,
 * emergency contact) are gated server-side and simply arrive absent when the
 * caller isn't allowed to see them, including on their own record until the
 * self-view exception ships (rtylr HR PR #262).
 */
export function useMyStaffRecord(authUserId: string | undefined) {
  return useQuery({
    queryKey: ['my-staff-record', authUserId],
    queryFn: async (): Promise<MyStaffRecord | null> => {
      const [staffResponse, departmentsResponse] = await Promise.all([
        staffClient.listStaff({}),
        departmentClient.listDepartments({}),
      ]);
      const staff = (staffResponse.staff ?? []).find((member) => member.userId === authUserId);
      if (!staff) return null;
      const department = (departmentsResponse.departments ?? []).find(
        (item) => item.id === staff.departmentId
      );
      return { ...staff, departmentName: department?.name };
    },
    enabled: !!authUserId,
  });
}

export function useMyProfileDetail(staffId: string | undefined) {
  return useQuery({
    queryKey: ['my-profile-detail', staffId],
    queryFn: async () => {
      const res = await unwrap(safe.get<unknown>('/hr/staff/detail', { params: { id: staffId } }));
      return profileDetailSchema.parse(res);
    },
    enabled: !!staffId,
  });
}
