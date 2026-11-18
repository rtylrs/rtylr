import { useQuery } from '@tanstack/react-query';
import { createOneKitClient } from '@voxire/api';
import { ApiError, StaffServiceClient } from '@voxire/voxie/hr/staff/v1/client';
import { DepartmentServiceClient } from '@voxire/voxie/hr/department/v1/client';
import type { Staff } from '@voxire/voxie/hr/staff/v1/types';

const staffClient = createOneKitClient(StaffServiceClient, 'flow');
const departmentClient = createOneKitClient(DepartmentServiceClient, 'flow');

export interface DirectoryMember extends Staff {
  departmentName?: string;
}

/** Why listStaff refused the caller, from the 403's `code`. */
export type DirectoryDenial = 'disabled' | 'permission';

export function getDirectoryDenial(error: unknown): DirectoryDenial | null {
  if (!(error instanceof ApiError) || error.statusCode !== 403) return null;
  return error.code === 'directory_disabled' ? 'disabled' : 'permission';
}

const MAX_RETRIES = 3;

/**
 * The whole business's roster. listStaff is gated server-side: the business
 * must have the directory on and the caller's role must hold view_directory.
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
    // A 403 is an answer, not a blip: retrying it only delays the message.
    retry: (failureCount, error) => !getDirectoryDenial(error) && failureCount < MAX_RETRIES,
  });
}
