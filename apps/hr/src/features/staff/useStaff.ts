import { createOneKitClient, getOneKitErrorMessage, unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { StaffServiceClient } from '@voxire/voxie/hr/staff/v1/client';
import { DepartmentServiceClient } from '@voxire/voxie/hr/department/v1/client';
import type { Staff as ApiStaff } from '@voxire/voxie/hr/staff/v1/types';
import { staffMemberSchema, type StaffMember } from '../../shared/schemas/staff';
import { hrApi, CreateStaffInput, UpdateStaffInput } from '../../shared/services/hrApi';

const staffClient = createOneKitClient(StaffServiceClient, 'hr');
const departmentClient = createOneKitClient(DepartmentServiceClient, 'hr');

function toStaffMember(s: ApiStaff, departmentName: string): StaffMember {
  return staffMemberSchema.parse({
    id: s.id,
    name: `${s.firstName} ${s.lastName}`.trim() || s.email,
    role: s.role,
    department: departmentName,
    email: s.email,
    employmentType: s.employmentType,
    status: s.status,
  });
}

export function useStaff() {
  return useQuery({
    queryKey: ['staff'],
    queryFn: async (): Promise<StaffMember[]> => {
      const [staffRes, departmentsRes] = await Promise.all([
        staffClient.listStaff({ search: '', departmentId: '', status: '' }),
        departmentClient.listDepartments({}),
      ]);
      const namesById = new Map((departmentsRes.departments ?? []).map((d) => [d.id, d.name]));
      return (staffRes.staff ?? []).map((s) =>
        toStaffMember(s, namesById.get(s.departmentId) ?? '')
      );
    },
  });
}

// hr.staff.v1 has no create/update/delete/clock-in/out consumers wired in the
// UI yet - these remain on the legacy (currently 404ing) hrApi paths until a
// real screen needs them.
export function useCreateStaff() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateStaffInput) => unwrap(hrApi.createStaff(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['staff'] }),
    onError: (error) => toast.error(getOneKitErrorMessage(error, 'Failed to create staff member')),
  });
}

export function useUpdateStaff() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateStaffInput) => unwrap(hrApi.updateStaff(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['staff'] }),
    onError: (error) => toast.error(getOneKitErrorMessage(error, 'Failed to update staff member')),
  });
}

export function useDeleteStaff() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrApi.deleteStaff(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['staff'] }),
    onError: (error) => toast.error(getOneKitErrorMessage(error, 'Failed to remove staff member')),
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
    onError: (error) => toast.error(getOneKitErrorMessage(error, 'Failed to clock in')),
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
    onError: (error) => toast.error(getOneKitErrorMessage(error, 'Failed to clock out')),
  });
}
