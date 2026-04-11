import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { hrExtras } from '../services/hrExtras';
import { hrApi, StaffMember } from '../services/hrApi';

/* ── 31. useStaffById ── */
export function useStaffById(id: number | undefined) {
  return useQuery({
    queryKey: ['staff', id],
    enabled: id !== undefined,
    queryFn: async () => {
      const res = await hrApi.listStaff();
      return (res.data ?? []).find((s: StaffMember) => s.id === id) ?? null;
    },
  });
}

/* ── 32. useTimeOff (extras) ── */
export function useTimeOff() {
  return useQuery({ queryKey: ['time-off-extras'], queryFn: () => unwrap(hrExtras.listTimeOff()) });
}

export function useSubmitTimeOff() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: hrExtras.submitTimeOff,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['time-off'] }),
  });
}

export function useApproveTimeOff() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, approved, note }: { id: number; approved: boolean; note?: string }) =>
      hrExtras.approveTimeOff(id, approved, note),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['time-off'] }),
  });
}

/* ── 33. useShiftSchedule ── */
export function useShiftSchedule(weekStart?: string) {
  return useQuery({
    queryKey: ['schedule', weekStart ?? 'all'],
    queryFn: () => unwrap(hrExtras.listSchedule(weekStart)),
  });
}

export function useUpsertShift() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: hrExtras.upsertScheduleShift,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['schedule'] }),
  });
}

export function useDeleteShift() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrExtras.deleteScheduleShift(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['schedule'] }),
  });
}

export function usePublishWeek() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (weekStart: string) => unwrap(hrExtras.publishWeek(weekStart)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['schedule'] }),
  });
}

/* ── 34. useTimesheets ── */
export function useTimesheets() {
  return useQuery({ queryKey: ['timesheets'], queryFn: () => unwrap(hrExtras.listTimesheets()) });
}

export function useApproveTimesheet() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrExtras.approveTimesheet(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['timesheets'] }),
  });
}

/* ── 35. useReviews ── */
export function useReviews() {
  return useQuery({ queryKey: ['reviews'], queryFn: () => unwrap(hrExtras.listReviews()) });
}

export function useCreateReview() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: hrExtras.createReview,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['reviews'] }),
  });
}

/* ── 36. useOneOnOnes ── */
export function useOneOnOnes() {
  return useQuery({ queryKey: ['one-on-ones'], queryFn: () => unwrap(hrExtras.listOneOnOnes()) });
}

export function useCreateOneOnOne() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: hrExtras.createOneOnOne,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['one-on-ones'] }),
  });
}

/* ── 37. useOnboardingChecklist ── */
export function useOnboardingChecklist() {
  return useQuery({ queryKey: ['onboarding'], queryFn: () => unwrap(hrExtras.listOnboarding()) });
}

export function useToggleOnboardingTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrExtras.toggleOnboardingTask(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['onboarding'] }),
  });
}

export function useOffboardingChecklist() {
  return useQuery({ queryKey: ['offboarding'], queryFn: () => unwrap(hrExtras.listOffboarding()) });
}

export function useToggleOffboardingTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrExtras.toggleOffboardingTask(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['offboarding'] }),
  });
}

/* ── 38. useDocuments ── */
export function useDocuments() {
  return useQuery({ queryKey: ['documents'], queryFn: () => unwrap(hrExtras.listDocuments()) });
}

export function useAddDocument() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: hrExtras.addDocument,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['documents'] }),
  });
}

export function useDeleteDocument() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrExtras.deleteDocument(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['documents'] }),
  });
}

export function useExpiringDocuments(daysAhead = 30) {
  return useQuery({
    queryKey: ['documents', 'expiring', daysAhead],
    queryFn: () => unwrap(hrExtras.expiringDocuments(daysAhead)),
  });
}

/* ── 39. useBenefits ── */
export function useBenefits() {
  return useQuery({
    queryKey: ['benefits', 'plans'],
    queryFn: () => unwrap(hrExtras.listBenefitPlans()),
  });
}

export function useBenefitEnrollments() {
  return useQuery({
    queryKey: ['benefits', 'enrollments'],
    queryFn: () => unwrap(hrExtras.listBenefitEnrollments()),
  });
}

export function useEnrollBenefit() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: hrExtras.enrollBenefit,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['benefits'] }),
  });
}

/* ── 40. useTraining ── */
export function useTraining() {
  return useQuery({
    queryKey: ['training', 'courses'],
    queryFn: () => unwrap(hrExtras.listCourses()),
  });
}

export function useTrainingProgress() {
  return useQuery({
    queryKey: ['training', 'progress'],
    queryFn: () => unwrap(hrExtras.listTrainingProgress()),
  });
}

/* ── extras ── */
export function useBirthdays() {
  return useQuery({ queryKey: ['birthdays'], queryFn: () => unwrap(hrExtras.listBirthdays()) });
}

export function useAnniversaries() {
  return useQuery({
    queryKey: ['anniversaries'],
    queryFn: () => unwrap(hrExtras.listAnniversaries()),
  });
}

export function useAvailability() {
  return useQuery({
    queryKey: ['availability'],
    queryFn: () => unwrap(hrExtras.listAvailability()),
  });
}

export function useSwaps() {
  return useQuery({ queryKey: ['swaps'], queryFn: () => unwrap(hrExtras.listSwaps()) });
}

export function useResolveSwap() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, accepted }: { id: number; accepted: boolean }) =>
      unwrap(hrExtras.resolveSwap(id, accepted)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['swaps'] }),
  });
}

export function useCreateSwap() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: hrExtras.createSwap,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['swaps'] }),
  });
}

export function useBreaks() {
  return useQuery({ queryKey: ['breaks'], queryFn: () => unwrap(hrExtras.listBreaks()) });
}

export function useStartBreak() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      staffId,
      staffName,
      type,
    }: {
      staffId: number;
      staffName: string;
      type: 'Lunch' | 'Short' | 'Other';
    }) => hrExtras.startBreak(staffId, staffName, type),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['breaks'] }),
  });
}

export function useEndBreak() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrExtras.endBreak(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['breaks'] }),
  });
}

export function useOvertimeAlerts() {
  return useQuery({
    queryKey: ['overtime-alerts'],
    queryFn: () => unwrap(hrExtras.listOvertimeAlerts()),
  });
}

export function useHrSettings() {
  return useQuery({ queryKey: ['settings'], queryFn: () => unwrap(hrExtras.getSettings()) });
}

export function useUpdateHrSettings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: hrExtras.updateSettings,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['settings'] }),
  });
}

export function useEmployeeProfile(id: number | undefined) {
  return useQuery({
    queryKey: ['profile', id],
    enabled: id !== undefined,
    queryFn: () => (id !== undefined ? hrExtras.getProfile(id) : Promise.resolve(null)),
  });
}
