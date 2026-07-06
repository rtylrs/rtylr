import { createOneKitClient, getOneKitErrorMessage, unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { TimeOffServiceClient } from '@voxire/voxie/hr/time_off/v1/service_client';
import { StaffServiceClient } from '@voxire/voxie/hr/staff/v1/service_client';
import { timeOffRequestSchema, type TimeOffRequest } from '../../shared/schemas/timeOff';
import { hrApi } from '../../shared/services/hrApi';

const timeOffClient = createOneKitClient(TimeOffServiceClient, 'hr');
const staffClient = createOneKitClient(StaffServiceClient, 'hr');

function normalizeStatus(status: string): 'Pending' | 'Approved' | 'Rejected' {
  const lower = status.toLowerCase();
  if (lower === 'approved') return 'Approved';
  if (lower === 'rejected') return 'Rejected';
  return 'Pending';
}

export function useTimeOff() {
  return useQuery({
    queryKey: ['time-off'],
    queryFn: async (): Promise<TimeOffRequest[]> => {
      const [timeOffRes, staffRes] = await Promise.all([
        timeOffClient.listTimeOff({}),
        staffClient.listStaff({ search: '', departmentId: '', status: '' }),
      ]);
      const staffById = new Map(staffRes.staff.map((s) => [s.id, s]));

      return timeOffRes.requests.map((r) => {
        const staff = staffById.get(r.staffId);
        return timeOffRequestSchema.parse({
          id: r.id,
          staffId: Number(r.staffId),
          staffName: staff ? `${staff.firstName} ${staff.lastName}`.trim() : `Staff #${r.staffId}`,
          startDate: r.startDate,
          endDate: r.endDate,
          reason: r.reason,
          status: normalizeStatus(r.status),
        });
      });
    },
  });
}

export function useSubmitTimeOff() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { staffId: number; startDate: string; endDate: string; reason: string }) =>
      timeOffClient.submitTimeOff({
        staffId: String(data.staffId),
        type: '',
        startDate: data.startDate,
        endDate: data.endDate,
        daysCount: 0,
        reason: data.reason,
        notes: '',
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['time-off'] }),
    onError: (error) =>
      toast.error(getOneKitErrorMessage(error, 'Failed to submit time off request')),
  });
}

// hr.time_off.v1 has no approve RPC - this stays on the legacy (working)
// chi route until a typed contract exists.
export function useApproveTimeOff() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, approved }: { id: number; approved: boolean }) =>
      unwrap(hrApi.approveTimeOff(id, approved)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['time-off'] }),
    onError: () => toast.error('Failed to update time off request'),
  });
}
