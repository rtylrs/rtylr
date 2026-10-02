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

export interface StaffReportingLine {
  id: number;
  name: string;
  managerStaffId: number;
}

// useStaffReportingLines returns raw staff ids alongside names, which the
// reporting-line editor needs and toStaffMember deliberately drops (it shapes
// rows for the list screen).
export function useStaffReportingLines() {
  return useQuery({
    queryKey: ['hr', 'staff', 'reporting-lines'],
    queryFn: async (): Promise<StaffReportingLine[]> => {
      const res = await staffClient.listStaff({ search: '', departmentId: '', status: '' });
      return (res.staff ?? []).map((s) => ({
        id: Number(s.id),
        name: `${s.firstName} ${s.lastName}`.trim() || s.email || '',
        managerStaffId: Number(s.managerStaffId ?? 0),
      }));
    },
  });
}

// useSetStaffManager goes through the GENERATED client, not hrApi.updateStaff.
// hrApi puts to /hr/staff with no id, while the contract declares
// PUT /hr/staff/{id} with a body - which is why the legacy path 404s and why the
// comment below says these hooks were never wired to a screen.
export function useSetStaffManager() {
  const qc = useQueryClient();
  return useMutation({
    // UpdateStaffBody is a REPLACEMENT, not a patch: every field is @required, so
    // sending only the manager would blank the rest of the record. The screen
    // fetches the staff member first and passes the whole thing back with the one
    // field changed.
    mutationFn: ({ staff, managerStaffId }: { staff: ApiStaff; managerStaffId: number }) =>
      staffClient.updateStaff({
        id: String(staff.id),
        body: {
          departmentId: String(staff.departmentId ?? 0),
          managerStaffId: String(managerStaffId),
          firstName: staff.firstName ?? '',
          lastName: staff.lastName ?? '',
          email: staff.email ?? '',
          phone: staff.phone ?? '',
          role: staff.role ?? '',
          employmentType: staff.employmentType ?? '',
          status: staff.status ?? '',
          salary: staff.salary ?? 0,
          hourlyRate: staff.hourlyRate ?? 0,
          overtimeRate: staff.overtimeRate ?? 0,
          endDate: staff.endDate ?? '',
          address: staff.address ?? '',
          emergencyContactName: staff.emergencyContactName ?? '',
          emergencyContactPhone: staff.emergencyContactPhone ?? '',
          apps: staff.apps ?? '',
          notes: staff.notes ?? '',
          photoUrl: staff.photoUrl ?? '',
          bio: staff.bio ?? '',
        },
      }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['hr', 'staff'] });
    },
    onError: (error) => toast.error(getOneKitErrorMessage(error, 'Failed to set the manager')),
  });
}

// useStaffRecord fetches one staff member in full, which the replacement update
// requires: without it a save would blank every field it did not send.
export function useStaffRecord(id: number | null) {
  return useQuery({
    queryKey: ['hr', 'staff', 'record', id],
    queryFn: async (): Promise<ApiStaff> => {
      const res = await staffClient.getStaff({ id: String(id) });
      return res.staff as ApiStaff;
    },
    enabled: !!id,
  });
}

export function useStaff() {
  return useQuery({
    queryKey: ['hr', 'staff'],
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
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hr', 'staff'] }),
    onError: (error) => toast.error(getOneKitErrorMessage(error, 'Failed to create staff member')),
  });
}

export function useUpdateStaff() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateStaffInput) => unwrap(hrApi.updateStaff(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hr', 'staff'] }),
    onError: (error) => toast.error(getOneKitErrorMessage(error, 'Failed to update staff member')),
  });
}

export function useDeleteStaff() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrApi.deleteStaff(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hr', 'staff'] }),
    onError: (error) => toast.error(getOneKitErrorMessage(error, 'Failed to remove staff member')),
  });
}

export function useClockIn() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (staffId: number) => unwrap(hrApi.clockIn(staffId)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hr', 'staff'] });
      qc.invalidateQueries({ queryKey: ['hr', 'hr-dashboard'] });
    },
    onError: (error) => toast.error(getOneKitErrorMessage(error, 'Failed to clock in')),
  });
}

export function useClockOut() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (staffId: number) => unwrap(hrApi.clockOut(staffId)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hr', 'staff'] });
      qc.invalidateQueries({ queryKey: ['hr', 'hr-dashboard'] });
    },
    onError: (error) => toast.error(getOneKitErrorMessage(error, 'Failed to clock out')),
  });
}
