import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { toast } from 'sonner';
import { hrExtras } from '../services/hrExtras';
import { hrApi } from '../services/hrApi';
import { staffMemberSchema } from '../schemas/staff';
import { timeOffEntrySchema } from '../schemas/timeOff';
import {
  scheduleShiftSchema,
  availabilitySchema,
  swapRequestSchema,
  breakLogSchema,
  overtimeAlertSchema,
} from '../schemas/shifts';
import { timesheetRowSchema } from '../schemas/timesheets';
import { reviewSchema, oneOnOneSchema } from '../schemas/reviews';
import { onboardingTaskSchema, offboardingTaskSchema } from '../schemas/onboarding';
import { documentItemSchema } from '../schemas/documents';
import { benefitPlanSchema, benefitEnrollmentSchema } from '../schemas/benefits';
import { trainingCourseSchema, trainingProgressSchema } from '../schemas/training';
import { birthdaySchema, anniversarySchema, employeeProfileSchema } from '../schemas/staff';
import { hrSettingsSchema } from '../schemas/settings';

/* ── 31. useStaffById ── */
export function useStaffById(id: number | undefined) {
  return useQuery({
    queryKey: ['staff', id],
    enabled: id !== undefined,
    queryFn: async () => {
      const list = await unwrap(hrApi.listStaff());
      const raw = list ?? [];
      const parsed = z.array(staffMemberSchema).parse(Array.isArray(raw) ? raw : []);
      return parsed.find((s) => s.id === id) ?? null;
    },
  });
}

/* ── 32. useTimeOff (extras) ── */
export function useTimeOff() {
  return useQuery({
    queryKey: ['time-off-extras'],
    queryFn: async () => {
      const res = await unwrap(hrExtras.listTimeOff());
      const raw = res ?? [];
      return z.array(timeOffEntrySchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useSubmitTimeOff() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Parameters<typeof hrExtras.submitTimeOff>[0]) =>
      unwrap(hrExtras.submitTimeOff(data)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['time-off-extras'] });
      qc.invalidateQueries({ queryKey: ['time-off'] });
    },
    onError: () => toast.error('Failed to submit time off request'),
  });
}

export function useApproveTimeOff() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, approved, note }: { id: number; approved: boolean; note?: string }) =>
      unwrap(hrExtras.approveTimeOff(id, approved, note)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['time-off-extras'] });
      qc.invalidateQueries({ queryKey: ['time-off'] });
    },
    onError: () => toast.error('Failed to update time off request'),
  });
}

/* ── 33. useShiftSchedule ── */
export function useShiftSchedule(weekStart?: string) {
  return useQuery({
    queryKey: ['schedule', weekStart ?? 'all'],
    queryFn: async () => {
      const res = await unwrap(hrExtras.listSchedule(weekStart));
      const raw = res ?? [];
      return z.array(scheduleShiftSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useUpsertShift() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (shift: Parameters<typeof hrExtras.upsertScheduleShift>[0]) =>
      unwrap(hrExtras.upsertScheduleShift(shift)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['schedule'] }),
    onError: () => toast.error('Failed to save shift'),
  });
}

export function useDeleteShift() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrExtras.deleteScheduleShift(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['schedule'] }),
    onError: () => toast.error('Failed to delete shift'),
  });
}

export function usePublishWeek() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (weekStart: string) => unwrap(hrExtras.publishWeek(weekStart)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['schedule'] }),
    onError: () => toast.error('Failed to publish schedule'),
  });
}

/* ── 34. useTimesheets ── */
export function useTimesheets() {
  return useQuery({
    queryKey: ['timesheets'],
    queryFn: async () => {
      const res = await unwrap(hrExtras.listTimesheets());
      const raw = res ?? [];
      return z.array(timesheetRowSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useApproveTimesheet() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrExtras.approveTimesheet(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['timesheets'] }),
    onError: () => toast.error('Failed to approve timesheet'),
  });
}

/* ── 35. useReviews ── */
export function useReviews() {
  return useQuery({
    queryKey: ['reviews'],
    queryFn: async () => {
      const res = await unwrap(hrExtras.listReviews());
      const raw = res ?? [];
      return z.array(reviewSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useCreateReview() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Parameters<typeof hrExtras.createReview>[0]) =>
      unwrap(hrExtras.createReview(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['reviews'] }),
    onError: () => toast.error('Failed to create review'),
  });
}

/* ── 36. useOneOnOnes ── */
export function useOneOnOnes() {
  return useQuery({
    queryKey: ['one-on-ones'],
    queryFn: async () => {
      const res = await unwrap(hrExtras.listOneOnOnes());
      const raw = res ?? [];
      return z.array(oneOnOneSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useCreateOneOnOne() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Parameters<typeof hrExtras.createOneOnOne>[0]) =>
      unwrap(hrExtras.createOneOnOne(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['one-on-ones'] }),
    onError: () => toast.error('Failed to schedule 1:1'),
  });
}

/* ── 37. useOnboardingChecklist ── */
export function useOnboardingChecklist() {
  return useQuery({
    queryKey: ['onboarding'],
    queryFn: async () => {
      const res = await unwrap(hrExtras.listOnboarding());
      const raw = res ?? [];
      return z.array(onboardingTaskSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useToggleOnboardingTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrExtras.toggleOnboardingTask(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['onboarding'] }),
    onError: () => toast.error('Failed to update onboarding task'),
  });
}

export function useOffboardingChecklist() {
  return useQuery({
    queryKey: ['offboarding'],
    queryFn: async () => {
      const res = await unwrap(hrExtras.listOffboarding());
      const raw = res ?? [];
      return z.array(offboardingTaskSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useToggleOffboardingTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrExtras.toggleOffboardingTask(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['offboarding'] }),
    onError: () => toast.error('Failed to update offboarding task'),
  });
}

/* ── 38. useDocuments ── */
export function useDocuments() {
  return useQuery({
    queryKey: ['documents'],
    queryFn: async () => {
      const res = await unwrap(hrExtras.listDocuments());
      const raw = res ?? [];
      return z.array(documentItemSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useAddDocument() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (doc: Parameters<typeof hrExtras.addDocument>[0]) =>
      unwrap(hrExtras.addDocument(doc)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['documents'] }),
    onError: () => toast.error('Failed to upload document'),
  });
}

export function useDeleteDocument() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrExtras.deleteDocument(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['documents'] }),
    onError: () => toast.error('Failed to delete document'),
  });
}

export function useExpiringDocuments(daysAhead = 30) {
  return useQuery({
    queryKey: ['documents', 'expiring', daysAhead],
    queryFn: async () => {
      const res = await unwrap(hrExtras.expiringDocuments(daysAhead));
      const raw = res ?? [];
      return z.array(documentItemSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

/* ── 39. useBenefits ── */
export function useBenefits() {
  return useQuery({
    queryKey: ['benefits', 'plans'],
    queryFn: async () => {
      const res = await unwrap(hrExtras.listBenefitPlans());
      const raw = res ?? [];
      return z.array(benefitPlanSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useBenefitEnrollments() {
  return useQuery({
    queryKey: ['benefits', 'enrollments'],
    queryFn: async () => {
      const res = await unwrap(hrExtras.listBenefitEnrollments());
      const raw = res ?? [];
      return z.array(benefitEnrollmentSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useEnrollBenefit() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Parameters<typeof hrExtras.enrollBenefit>[0]) =>
      unwrap(hrExtras.enrollBenefit(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['benefits'] }),
    onError: () => toast.error('Failed to enroll in benefit'),
  });
}

/* ── 40. useTraining ── */
export function useTraining() {
  return useQuery({
    queryKey: ['training', 'courses'],
    queryFn: async () => {
      const res = await unwrap(hrExtras.listCourses());
      const raw = res ?? [];
      return z.array(trainingCourseSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useTrainingProgress() {
  return useQuery({
    queryKey: ['training', 'progress'],
    queryFn: async () => {
      const res = await unwrap(hrExtras.listTrainingProgress());
      const raw = res ?? [];
      return z.array(trainingProgressSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

/* ── extras ── */
export function useBirthdays() {
  return useQuery({
    queryKey: ['birthdays'],
    queryFn: async () => {
      const res = await unwrap(hrExtras.listBirthdays());
      const raw = res ?? [];
      return z.array(birthdaySchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useAnniversaries() {
  return useQuery({
    queryKey: ['anniversaries'],
    queryFn: async () => {
      const res = await unwrap(hrExtras.listAnniversaries());
      const raw = res ?? [];
      return z.array(anniversarySchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useAvailability() {
  return useQuery({
    queryKey: ['availability'],
    queryFn: async () => {
      const res = await unwrap(hrExtras.listAvailability());
      const raw = res ?? [];
      return z.array(availabilitySchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useSwaps() {
  return useQuery({
    queryKey: ['swaps'],
    queryFn: async () => {
      const res = await unwrap(hrExtras.listSwaps());
      const raw = res ?? [];
      return z.array(swapRequestSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useResolveSwap() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, accepted }: { id: number; accepted: boolean }) =>
      unwrap(hrExtras.resolveSwap(id, accepted)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['swaps'] }),
    onError: () => toast.error('Failed to resolve swap request'),
  });
}

export function useCreateSwap() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Parameters<typeof hrExtras.createSwap>[0]) =>
      unwrap(hrExtras.createSwap(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['swaps'] }),
    onError: () => toast.error('Failed to create swap request'),
  });
}

export function useBreaks() {
  return useQuery({
    queryKey: ['breaks'],
    queryFn: async () => {
      const res = await unwrap(hrExtras.listBreaks());
      const raw = res ?? [];
      return z.array(breakLogSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
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
    }) => unwrap(hrExtras.startBreak(staffId, staffName, type)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['breaks'] }),
    onError: () => toast.error('Failed to start break'),
  });
}

export function useEndBreak() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrExtras.endBreak(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['breaks'] }),
    onError: () => toast.error('Failed to end break'),
  });
}

export function useOvertimeAlerts() {
  return useQuery({
    queryKey: ['overtime-alerts'],
    queryFn: async () => {
      const res = await unwrap(hrExtras.listOvertimeAlerts());
      const raw = res ?? [];
      return z.array(overtimeAlertSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useHrSettings() {
  return useQuery({
    queryKey: ['settings'],
    queryFn: async () => {
      const res = await unwrap(hrExtras.getSettings());
      return hrSettingsSchema.parse(res);
    },
  });
}

export function useUpdateHrSettings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (patch: Parameters<typeof hrExtras.updateSettings>[0]) =>
      unwrap(hrExtras.updateSettings(patch)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['settings'] }),
    onError: () => toast.error('Failed to save settings'),
  });
}

export function useEmployeeProfile(id: number | undefined) {
  return useQuery({
    queryKey: ['profile', id],
    enabled: id !== undefined,
    queryFn: async () => {
      if (id === undefined) return null;
      const res = await unwrap(hrExtras.getProfile(id));
      return employeeProfileSchema.parse(res);
    },
  });
}
