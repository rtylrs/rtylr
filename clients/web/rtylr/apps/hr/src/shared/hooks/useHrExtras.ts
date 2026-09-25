import { createOneKitClient, getOneKitErrorMessage, unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { StaffServiceClient } from '@voxire/voxie/hr/staff/v1/client';
import { DepartmentServiceClient } from '@voxire/voxie/hr/department/v1/client';
import { BenefitServiceClient } from '@voxire/voxie/hr/benefit/v1/client';
import { z } from 'zod';
import { toast } from 'sonner';
import { hrExtras } from '../services/hrExtras';
import { staffDetailSchema } from '../schemas/staff';
import { timeOffEntrySchema } from '../schemas/timeOff';
import {
  scheduleShiftSchema,
  swapRequestSchema,
  breakLogSchema,
  overtimeAlertSchema,
} from '../schemas/shifts';
import { timesheetRowSchema } from '../schemas/timesheets';
import { reviewSchema, oneOnOneSchema } from '../schemas/reviews';
import { onboardingTaskSchema, offboardingTaskSchema } from '../schemas/onboarding';
import { documentItemSchema } from '../schemas/documents';
import {
  benefitPlanSchema,
  benefitEnrollmentSchema,
  myBenefitEnrollmentSchema,
} from '../schemas/benefits';
import type { MyBenefitEnrollment } from '../schemas/benefits';
import { toTitleCase } from '../lib/hrStringUtils';
import { trainingCourseSchema, trainingProgressSchema } from '../schemas/training';
import { birthdaySchema, anniversarySchema, myEmployeeSchema } from '../schemas/staff';
import { hrSettingsSchema } from '../schemas/settings';

const staffClient = createOneKitClient(StaffServiceClient, 'hr');
const departmentClient = createOneKitClient(DepartmentServiceClient, 'hr');
const benefitClient = createOneKitClient(BenefitServiceClient, 'hr');

/**
 * The caller's own hr_employees.id - not the same id as the hub auth user id
 * (`useUserStore((s) => s.userId)`). hr.employee.v1 has no self-lookup RPC, so
 * `/hr/employees/me` is the only correct source for this: sending the auth
 * user id where an employee_id is expected fails leave/overtime/equipment
 * requests outright once the receiving endpoint validates ownership.
 */
export function useMyEmployeeId() {
  return useQuery({
    queryKey: ['my-employee-id'],
    queryFn: async () => {
      const res = await unwrap(hrExtras.getMyEmployee());
      return myEmployeeSchema.parse(res.employee).id;
    },
  });
}

/* ── 31. useStaffById ── */
export function useStaffById(id: number | undefined) {
  return useQuery({
    queryKey: ['staff', id],
    enabled: id !== undefined,
    queryFn: async () => {
      if (id === undefined) return null;
      const [staffResponse, departmentsResponse] = await Promise.all([
        staffClient.getStaff({ id: String(id) }),
        departmentClient.listDepartments({}),
      ]);
      const staff = staffResponse.staff;
      if (!staff) return null;
      const department = (departmentsResponse.departments ?? []).find(
        (item) => item.id === staff.departmentId
      );
      // Keep the contact fields getStaff already returned. Dropping them here
      // is what sent the detail screen to /hr/staff/detail for values it was
      // holding all along.
      return staffDetailSchema.parse({
        id: staff.id,
        name: `${staff.firstName ?? ''} ${staff.lastName ?? ''}`.trim(),
        role: staff.role ?? '',
        department: department?.name ?? '',
        email: staff.email ?? '',
        employmentType: staff.employmentType,
        status: staff.status,
        photoUrl: staff.photoUrl || undefined,
        bio: staff.bio || undefined,
        phone: staff.phone || undefined,
        address: staff.address || undefined,
        startDate: staff.startDate || undefined,
        emergencyContactName: staff.emergencyContactName || undefined,
        emergencyContactPhone: staff.emergencyContactPhone || undefined,
      });
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
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Parameters<typeof hrExtras.submitTimeOff>[0]) =>
      unwrap(hrExtras.submitTimeOff(data)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['time-off-extras'] });
      qc.invalidateQueries({ queryKey: ['time-off'] });
    },
    onError: () => toast.error(t('errors.submitTimeOff')),
  });
}

export function useApproveTimeOff() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, approved, note }: { id: number; approved: boolean; note?: string }) =>
      unwrap(hrExtras.approveTimeOff(id, approved, note)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['time-off-extras'] });
      qc.invalidateQueries({ queryKey: ['time-off'] });
    },
    onError: () => toast.error(t('errors.updateTimeOff')),
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
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (shift: Parameters<typeof hrExtras.upsertScheduleShift>[0]) =>
      unwrap(hrExtras.upsertScheduleShift(shift)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['schedule'] }),
    onError: () => toast.error(t('errors.saveShift')),
  });
}

export function useDeleteShift() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrExtras.deleteScheduleShift(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['schedule'] }),
    onError: () => toast.error(t('errors.deleteShift')),
  });
}

export function usePublishWeek() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (weekStart: string) => unwrap(hrExtras.publishWeek(weekStart)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['schedule'] }),
    onError: () => toast.error(t('errors.publishSchedule')),
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
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrExtras.approveTimesheet(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['timesheets'] }),
    onError: () => toast.error(t('errors.approveTimesheet')),
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
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Parameters<typeof hrExtras.createReview>[0]) =>
      unwrap(hrExtras.createReview(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['reviews'] }),
    onError: () => toast.error(t('errors.createReview')),
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
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Parameters<typeof hrExtras.createOneOnOne>[0]) =>
      unwrap(hrExtras.createOneOnOne(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['one-on-ones'] }),
    onError: () => toast.error(t('errors.scheduleOneOnOne')),
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
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrExtras.toggleOnboardingTask(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['onboarding'] }),
    onError: () => toast.error(t('errors.updateOnboardingTask')),
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
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrExtras.toggleOffboardingTask(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['offboarding'] }),
    onError: () => toast.error(t('errors.updateOffboardingTask')),
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
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (doc: Parameters<typeof hrExtras.addDocument>[0]) =>
      unwrap(hrExtras.addDocument(doc)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['documents'] }),
    onError: () => toast.error(t('errors.uploadDocument')),
  });
}

export function useDeleteDocument() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrExtras.deleteDocument(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['documents'] }),
    onError: () => toast.error(t('errors.deleteDocument')),
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
      const res = await benefitClient.listBenefits({});
      const raw = res.plans ?? [];
      return z.array(benefitPlanSchema).parse(
        raw.map((plan) => ({
          id: plan.id ?? '',
          name: plan.name ?? '',
          category: toTitleCase(plan.type ?? ''),
          carrier: plan.provider ?? '',
          monthlyCost: plan.costEmployee ?? 0,
        }))
      );
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

const MY_BENEFIT_ENROLLMENT_PAGE_SIZE = 500;

async function fetchAllMyBenefitEnrollments(): Promise<MyBenefitEnrollment[]> {
  const all: unknown[] = [];
  const seenPages = new Set<string>();
  let page = 1;

  for (;;) {
    const res = await benefitClient.listMyBenefitEnrollments({
      page,
      limit: MY_BENEFIT_ENROLLMENT_PAGE_SIZE,
    });
    const raw = res.enrollments ?? [];
    if (raw.length === 0) {
      if (
        res.pagination?.hasNext === true ||
        (typeof res.pagination?.total === 'number' && all.length < res.pagination.total)
      ) {
        throw new Error('Benefit enrollment pagination ended before the reported total');
      }
      break;
    }

    const pageKey = raw.map((enrollment) => enrollment.id ?? JSON.stringify(enrollment)).join('|');
    if (seenPages.has(pageKey)) {
      throw new Error('Benefit enrollment pagination did not advance');
    }
    seenPages.add(pageKey);
    all.push(...raw);

    const pagination = res.pagination;
    const hasNext =
      pagination?.hasNext ??
      (typeof pagination?.total === 'number'
        ? all.length < pagination.total
        : raw.length === MY_BENEFIT_ENROLLMENT_PAGE_SIZE);
    if (!hasNext) break;

    page += 1;
  }

  return z.array(myBenefitEnrollmentSchema).parse(all);
}

/**
 * The caller's own benefit enrollments, resolved server-side from the auth
 * token via hr.benefit.v1's ListMyBenefitEnrollments - never from a
 * client-supplied staff id. The generated endpoint is paginated, so merge
 * every page instead of presenting the first page as the full history.
 */
export function useMyBenefitEnrollments() {
  return useQuery({
    queryKey: ['benefits', 'my-enrollments'],
    queryFn: fetchAllMyBenefitEnrollments,
  });
}

export function useEnrollInBenefit() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
    mutationFn: (planId: string) => benefitClient.enrollInBenefit({ planId }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['benefits'] });
      toast.success(t('benefits.enrolled'));
    },
    onError: (error) => toast.error(getOneKitErrorMessage(error, t('errors.generic'))),
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
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, accepted }: { id: number; accepted: boolean }) =>
      unwrap(hrExtras.resolveSwap(id, accepted)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['swaps'] }),
    onError: () => toast.error(t('errors.resolveSwap')),
  });
}

export function useCreateSwap() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Parameters<typeof hrExtras.createSwap>[0]) =>
      unwrap(hrExtras.createSwap(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['swaps'] }),
    onError: () => toast.error(t('errors.createSwap')),
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
  const { t } = useTranslation();
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
    onError: () => toast.error(t('errors.startBreak')),
  });
}

export function useEndBreak() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrExtras.endBreak(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['breaks'] }),
    onError: () => toast.error(t('errors.endBreak')),
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
      return hrSettingsSchema.parse(res.config);
    },
  });
}

export function useUpdateHrSettings() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (patch: Parameters<typeof hrExtras.updateSettings>[0]) =>
      unwrap(hrExtras.updateSettings(patch)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['settings'] }),
    onError: () => toast.error(t('errors.saveSettings')),
  });
}
