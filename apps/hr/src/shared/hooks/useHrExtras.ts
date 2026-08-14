import { createOneKitClient, unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { StaffServiceClient } from '@voxire/voxie/hr/staff/v1/client';
import { DepartmentServiceClient } from '@voxire/voxie/hr/department/v1/client';
import { LeaveServiceClient } from '@voxire/voxie/hr/leave/v1/client';
import type {
  ChecklistTemplatePayload,
  EmployeeChecklistPayload,
  ListChecklistTemplatesResponse,
  ListEmployeeChecklistsResponse,
} from '@voxire/voxie/hr/onboarding/v1/types';
import { TimeOffServiceClient } from '@voxire/voxie/hr/time_off/v1/client';
import { ReviewServiceClient } from '@voxire/voxie/hr/review/v1/client';
import { OneOnOneServiceClient } from '@voxire/voxie/hr/one_on_one/v1/client';
import { TrainingServiceClient } from '@voxire/voxie/hr/training/v1/client';
import { BenefitServiceClient } from '@voxire/voxie/hr/benefit/v1/client';
import { OvertimeServiceClient } from '@voxire/voxie/hr/overtime/v1/client';
import { ShiftServiceClient } from '@voxire/voxie/hr/shift/v1/client';
import type { AvailabilityPayload } from '@voxire/voxie/hr/shift/v1/types';
import { z } from 'zod';
import { toast } from 'sonner';
import { hrExtras } from '../services/hrExtras';
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
import { documentItemSchema, type DocumentItem } from '../schemas/documents';
import { benefitPlanSchema, benefitEnrollmentSchema } from '../schemas/benefits';
import { trainingCourseSchema, trainingProgressSchema } from '../schemas/training';
import { birthdaySchema, anniversarySchema, employeeProfileSchema } from '../schemas/staff';
import { hrSettingsSchema } from '../schemas/settings';

const staffClient = createOneKitClient(StaffServiceClient, 'hr');
const departmentClient = createOneKitClient(DepartmentServiceClient, 'hr');
const leaveClient = createOneKitClient(LeaveServiceClient, 'hr');
const timeOffClient = createOneKitClient(TimeOffServiceClient, 'hr');
const reviewClient = createOneKitClient(ReviewServiceClient, 'hr');
const oneOnOneClient = createOneKitClient(OneOnOneServiceClient, 'hr');
const trainingClient = createOneKitClient(TrainingServiceClient, 'hr');
const benefitClient = createOneKitClient(BenefitServiceClient, 'hr');
const overtimeClient = createOneKitClient(OvertimeServiceClient, 'hr');
const shiftClient = createOneKitClient(ShiftServiceClient, 'hr');

type ChecklistMode = 'onboarding' | 'offboarding';

const checklistTriggerIsOffboarding = (triggerType: string | undefined): boolean => {
  const trigger = (triggerType ?? '').toLowerCase();
  return trigger.includes('offboard') || trigger.includes('terminat');
};

const checklistCategory = (
  title: string,
  description: string,
  mode: ChecklistMode
): string => {
  const text = `${title} ${description}`.toLowerCase();
  if (/access|account|permission|badge|system|it/.test(text)) return 'Access';
  if (/equipment|laptop|device|asset|uniform|return/.test(text)) return 'Equipment';
  if (mode === 'onboarding' && /training|course|safety|orientation/.test(text)) {
    return 'Training';
  }
  if (mode === 'onboarding' && /team|introduc|welcome|buddy/.test(text)) {
    return 'Meet the team';
  }
  return mode === 'onboarding' ? 'Paperwork' : 'Knowledge transfer';
};

const checklistTasks = async (mode: ChecklistMode): Promise<unknown[]> => {
  const [checklistsResponse, templatesResponse, staffResponse] = await Promise.all([
    unwrap(hrExtras.listOnboardingChecklists({ limit: 100, offset: 0 })),
    unwrap(hrExtras.listOnboardingTemplates({ limit: 100, offset: 0 })),
    staffClient.listStaff({ search: '', departmentId: '', status: '' }),
  ]);

  const checklists = (checklistsResponse as ListEmployeeChecklistsResponse).checklists ?? [];
  const templates = (templatesResponse as ListChecklistTemplatesResponse).templates ?? [];
  const templatesById = new Map<number, ChecklistTemplatePayload>(
    templates.map((template) => [Number(template.id), template])
  );
  const itemDetailsById = new Map<
    number,
    { title: string; description: string }
  >();
  for (const template of templates) {
    for (const item of template.items ?? []) {
      itemDetailsById.set(Number(item.id), {
        title: item.title ?? `Checklist item #${item.id}`,
        description: item.description ?? '',
      });
    }
  }
  const namesByStaffId = displayNameByStaffId(staffResponse.staff);

  const mapped = (checklists as EmployeeChecklistPayload[]).flatMap((checklist) => {
    const template = templatesById.get(Number(checklist.templateId));
    if (mode === 'offboarding' && !checklistTriggerIsOffboarding(template?.triggerType)) {
      return [];
    }
    if (mode === 'onboarding' && checklistTriggerIsOffboarding(template?.triggerType)) {
      return [];
    }

    const staffId = Number(checklist.employeeId);
    const staffName = namesByStaffId.get(staffId) ?? `#${staffId}`;
    return (checklist.items ?? []).map((item) => {
      const details = itemDetailsById.get(Number(item.checklistItemId));
      const title = details?.title ?? `Checklist item #${item.checklistItemId}`;
      return {
        id: Number(item.id),
        staffId,
        staffName,
        task: title,
        category: checklistCategory(title, details?.description ?? '', mode),
        completed: item.status === 'complete',
        dueDate: item.dueDate,
      };
    });
  });

  return mapped;
};

const inclusiveDays = (startDate: string, endDate: string): number => {
  const start = Date.parse(`${startDate}T00:00:00Z`);
  const end = Date.parse(`${endDate}T00:00:00Z`);
  if (!Number.isFinite(start) || !Number.isFinite(end) || end < start) return 0;
  return Math.floor((end - start) / 86_400_000) + 1;
};

const legacyTimeOffStatus = (status: string | undefined): string => {
  const normalized = (status ?? '').toLowerCase();
  if (normalized === 'approved') return 'Approved';
  if (normalized === 'rejected') return 'Rejected';
  return 'Pending';
};

const legacyReviewStatus = (status: string | undefined): string => {
  const normalized = (status ?? '').toLowerCase();
  if (normalized === 'in_progress') return 'In Progress';
  if (normalized === 'submitted') return 'Submitted';
  if (normalized === 'completed') return 'Completed';
  return 'Draft';
};

const legacyOneOnOneStatus = (status: string | undefined): string => {
  const normalized = (status ?? '').toLowerCase();
  if (normalized === 'completed') return 'Completed';
  if (normalized === 'cancelled' || normalized === 'canceled') return 'Cancelled';
  return 'Scheduled';
};

const displayNameByStaffId = (
  staff: Awaited<ReturnType<typeof staffClient.listStaff>>['staff']
): Map<number, string> =>
  new Map(
    (staff ?? []).map((member) => [
      Number(member.id),
      `${member.firstName ?? ''} ${member.lastName ?? ''}`.trim() ||
        member.email ||
        `#${member.id}`,
    ])
  );

const normalizeDocumentType = (value: unknown): DocumentItem['type'] => {
  const normalized = String(value ?? '')
    .trim()
    .toLowerCase()
    .replace(/[_-]+/g, ' ');
  if (normalized === 'i 9' || normalized === 'i9') return 'I-9';
  if (normalized === 'w 4' || normalized === 'w4') return 'W4';
  if (normalized === 'contract') return 'Contract';
  if (normalized === 'nda') return 'NDA';
  if (normalized === 'offer letter') return 'Offer Letter';
  return 'Other';
};

const documentTimestamp = (value: unknown): string => {
  if (typeof value === 'number' || (typeof value === 'string' && /^\d+$/.test(value))) {
    const numeric = Number(value);
    const milliseconds = numeric > 10_000_000_000 ? numeric : numeric * 1000;
    return new Date(milliseconds).toISOString();
  }
  return String(value ?? '');
};

const documentRows = (value: unknown): unknown[] => {
  if (Array.isArray(value)) return value;
  if (value && typeof value === 'object') {
    const rows = (value as { documents?: unknown }).documents;
    return Array.isArray(rows) ? rows : [];
  }
  return [];
};

const mapDocument = (value: unknown, namesByStaffId: Map<number, string>): DocumentItem => {
  const row = (value && typeof value === 'object' ? value : {}) as Record<string, unknown>;
  const staffId = Number(row.employee_id ?? row.employeeId ?? row.staff_id ?? row.staffId ?? 0);
  const expiry = row.expiry_date ?? row.expiryDate;
  const url = row.external_url ?? row.externalUrl ?? row.file_url ?? row.fileUrl;
  return documentItemSchema.parse({
    id: Number(row.id),
    staffId,
    staffName: namesByStaffId.get(staffId) ?? `#${staffId}`,
    name: String(row.name ?? row.title ?? ''),
    type: normalizeDocumentType(row.document_type ?? row.documentType ?? row.type),
    uploadedAt: documentTimestamp(
      row.uploaded_at ?? row.uploadedAt ?? row.created_at ?? row.createdAt
    ),
    expiresAt: expiry == null ? undefined : String(expiry),
    sizeKb: Number(row.size_kb ?? row.sizeKb ?? 0),
    url: url == null || url === '' ? undefined : String(url),
  });
};

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
      return staffMemberSchema.parse({
        id: staff.id,
        name: `${staff.firstName ?? ''} ${staff.lastName ?? ''}`.trim(),
        role: staff.role ?? '',
        department: department?.name ?? '',
        email: staff.email ?? '',
        employmentType: staff.employmentType,
        status: staff.status,
      });
    },
  });
}

/* ── 32. useTimeOff (extras) ── */
export function useTimeOff() {
  return useQuery({
    queryKey: ['time-off-extras'],
    queryFn: async () => {
      const [res, staffRes] = await Promise.all([
        timeOffClient.listTimeOff({}),
        staffClient.listStaff({ search: '', departmentId: '', status: '' }),
      ]);
      const namesById = displayNameByStaffId(staffRes.staff);
      return z.array(timeOffEntrySchema).parse(
        (res.requests ?? []).map((request) => ({
          id: Number(request.id),
          staffId: Number(request.staffId),
          staffName: namesById.get(Number(request.staffId)) ?? `#${request.staffId}`,
          type: request.type,
          startDate: request.startDate,
          endDate: request.endDate,
          reason: request.reason,
          status: legacyTimeOffStatus(request.status),
          note: request.notes,
          createdAt: request.createdAt ?? new Date().toISOString(),
        }))
      );
    },
  });
}

export function useSubmitTimeOff() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Parameters<typeof hrExtras.submitTimeOff>[0]) => {
      const daysCount = inclusiveDays(data.startDate, data.endDate);
      if (daysCount <= 0) throw new Error('end date must not be before start date');
      return timeOffClient.submitTimeOff({
        staffId: String(data.staffId),
        type: data.type,
        startDate: data.startDate,
        endDate: data.endDate,
        daysCount,
        reason: data.reason.trim(),
        notes: data.note,
      });
    },
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
      leaveClient.approveLeave({
        requestId: String(id),
        body: { approved, rejectionReason: note },
      }),
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
      const [res, staffRes] = await Promise.all([
        reviewClient.listReviews({}),
        staffClient.listStaff({ search: '', departmentId: '', status: '' }),
      ]);
      const namesById = displayNameByStaffId(staffRes.staff);
      return z.array(reviewSchema).parse(
        (res.reviews ?? []).map((review) => ({
          id: Number(review.id),
          staffId: Number(review.staffId),
          staffName: namesById.get(Number(review.staffId)) ?? `#${review.staffId}`,
          reviewerId: Number(review.reviewerId),
          reviewer: namesById.get(Number(review.reviewerId)) ?? `#${review.reviewerId}`,
          period: review.period ?? '',
          status: legacyReviewStatus(review.status),
          overall: review.rating ?? 0,
          goals: review.goalsRating ?? review.rating ?? 0,
          values: review.valuesRating ?? review.rating ?? 0,
          skills: review.skillsRating ?? review.rating ?? 0,
          notes: review.notes ?? '',
          date: review.createdAt ?? new Date().toISOString(),
        }))
      );
    },
  });
}

export function useCreateReview() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Parameters<typeof hrExtras.createReview>[0]) =>
      reviewClient.createReview({
        staffId: String(data.staffId),
        reviewerId: String(data.reviewerId),
        period: data.period,
        rating: data.overall,
        goalsRating: data.goals,
        valuesRating: data.values,
        skillsRating: data.skills,
        notes: data.notes,
        status: data.status,
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['reviews'] }),
    onError: () => toast.error('Failed to create review'),
  });
}

/* ── 36. useOneOnOnes ── */
export function useOneOnOnes() {
  return useQuery({
    queryKey: ['one-on-ones'],
    queryFn: async () => {
      const [res, staffRes] = await Promise.all([
        oneOnOneClient.listOneOnOnes({}),
        staffClient.listStaff({ search: '', departmentId: '', status: '' }),
      ]);
      const namesById = displayNameByStaffId(staffRes.staff);
      return z.array(oneOnOneSchema).parse(
        (res.oneOnOnes ?? []).map((meeting) => ({
          id: Number(meeting.id),
          managerId: Number(meeting.managerId),
          manager: namesById.get(Number(meeting.managerId)) ?? `#${meeting.managerId}`,
          reportId: Number(meeting.staffId),
          report: namesById.get(Number(meeting.staffId)) ?? `#${meeting.staffId}`,
          scheduledAt: meeting.scheduledAt ?? '',
          duration: 30,
          status: legacyOneOnOneStatus(meeting.status),
          topics: meeting.actionItems
            ? meeting.actionItems.split('\n').map((topic) => topic.trim()).filter(Boolean)
            : [],
          notes: meeting.notes,
        }))
      );
    },
  });
}

export function useCreateOneOnOne() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Parameters<typeof hrExtras.createOneOnOne>[0]) =>
      oneOnOneClient.createOneOnOne({
        managerId: String(data.managerId),
        staffId: String(data.reportId),
        scheduledAt: data.scheduledAt,
        notes: data.notes,
        actionItems: data.topics.join('\n'),
        status: data.status,
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['one-on-ones'] }),
    onError: () => toast.error('Failed to schedule 1:1'),
  });
}

/* ── 37. useOnboardingChecklist ── */
export function useOnboardingChecklist() {
  return useQuery({
    queryKey: ['onboarding'],
    queryFn: async () => z.array(onboardingTaskSchema).parse(await checklistTasks('onboarding')),
  });
}

export function useToggleOnboardingTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrExtras.completeOnboardingItem(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['onboarding'] }),
    onError: () => toast.error('Failed to update onboarding task'),
  });
}

export function useOffboardingChecklist() {
  return useQuery({
    queryKey: ['offboarding'],
    queryFn: async () => z.array(offboardingTaskSchema).parse(await checklistTasks('offboarding')),
  });
}

export function useToggleOffboardingTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrExtras.completeOnboardingItem(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['offboarding'] }),
    onError: () => toast.error('Failed to update offboarding task'),
  });
}

/* ── 38. useDocuments ── */
export function useDocuments() {
  return useQuery({
    queryKey: ['documents'],
    queryFn: async () => {
      const [res, staffRes] = await Promise.all([
        unwrap(hrExtras.listDocuments()),
        staffClient.listStaff({ search: '', departmentId: '', status: '' }),
      ]);
      const namesByStaffId = displayNameByStaffId(staffRes.staff);
      return documentRows(res).map((document) => mapDocument(document, namesByStaffId));
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
      const [res, staffRes] = await Promise.all([
        unwrap(hrExtras.expiringDocuments(daysAhead)),
        staffClient.listStaff({ search: '', departmentId: '', status: '' }),
      ]);
      const namesByStaffId = displayNameByStaffId(staffRes.staff);
      return documentRows(res).map((document) => mapDocument(document, namesByStaffId));
    },
  });
}

/* ── 39. useBenefits ── */
export function useBenefits() {
  return useQuery({
    queryKey: ['benefits', 'plans'],
    queryFn: async () => {
      const res = await benefitClient.listBenefits({});
      return z.array(benefitPlanSchema).parse(
        (res.plans ?? []).map((plan) => ({
          id: Number(plan.id),
          name: plan.name ?? '',
          category: plan.type,
          carrier: plan.provider ?? '',
          monthlyCost: plan.costEmployee ?? plan.costEmployer ?? 0,
          enrolledCount: 0,
        }))
      );
    },
  });
}

export function useBenefitEnrollments() {
  return useQuery({
    queryKey: ['benefits', 'enrollments'],
    queryFn: async () => {
      const [res, plansRes] = await Promise.all([
        benefitClient.listMyBenefitEnrollments({}),
        benefitClient.listBenefits({}),
      ]);
      const namesById = new Map(
        (plansRes.plans ?? []).map((plan) => [Number(plan.id), plan.name ?? ''])
      );
      return z.array(benefitEnrollmentSchema).parse(
        (res.enrollments ?? []).map((enrollment) => ({
          id: Number(enrollment.id),
          staffId: Number(enrollment.employeeId),
          staffName: 'You',
          planId: Number(enrollment.planId),
          planName: namesById.get(Number(enrollment.planId)) ?? `#${enrollment.planId}`,
          startedAt: enrollment.enrolledAt ?? new Date().toISOString(),
          status:
            enrollment.status === 'active'
              ? 'Active'
              : enrollment.status === 'waived'
                ? 'Waived'
                : 'Pending',
        }))
      );
    },
  });
}

export function useEnrollBenefit() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { planId: number }) =>
      benefitClient.enrollInBenefit({ planId: String(data.planId) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['benefits'] }),
    onError: () => toast.error('Failed to enroll in benefit'),
  });
}

/* ── 40. useTraining ── */
export function useTraining() {
  return useQuery({
    queryKey: ['training', 'courses'],
    queryFn: async () => {
      const res = await trainingClient.listTrainingCourses({});
      return z.array(trainingCourseSchema).parse(
        (res.courses ?? []).map((course) => ({
          id: Number(course.id),
          title: course.name ?? '',
          category: course.description,
          durationMin: Math.round((course.durationHours ?? 0) * 60),
          modules: 0,
          required: course.isRequired ?? false,
        }))
      );
    },
  });
}

export function useTrainingProgress() {
  return useQuery({
    queryKey: ['training', 'progress'],
    queryFn: async () => {
      const res = await unwrap(hrExtras.listTrainingProgress());
      const raw = Array.isArray(res) ? res : [];
      return z.array(trainingProgressSchema).parse(
        raw.map((value) => {
          const row = value as Record<string, unknown>;
          return {
            id: Number(row.id),
            staffId: Number(row.staff_id ?? row.staffId),
            staffName: String(row.staff_name ?? row.staffName ?? ''),
            courseId: Number(row.course_id ?? row.courseId),
            courseName: String(row.course_name ?? row.courseName ?? ''),
            progress: Number(row.progress ?? 0),
            completedAt:
              typeof (row.completed_at ?? row.completedAt) === 'string'
                ? (row.completed_at ?? row.completedAt)
                : undefined,
          };
        })
      );
    },
  });
}

/* ── extras ── */
export function useBirthdays() {
  return useQuery({
    queryKey: ['birthdays'],
    queryFn: async () => {
      const res = await unwrap(hrExtras.listBirthdays());
      const payload = (res ?? {}) as { employees?: Array<Record<string, unknown>> };
      return z.array(birthdaySchema).parse(
        (payload.employees ?? [])
          .filter((employee) => /^\d{4}-\d{2}-\d{2}/.test(String(employee.date_of_birth ?? '')))
          .map((employee) => ({
            id: Number(employee.id),
            staffId: Number(employee.id),
            name: `${String(employee.first_name ?? '')} ${String(employee.last_name ?? '')}`.trim() || String(employee.email ?? ''),
            date: String(employee.date_of_birth).slice(0, 10),
            department: String(employee.job_title ?? '') || undefined,
          }))
      );
    },
  });
}

export function useAnniversaries() {
  return useQuery({
    queryKey: ['anniversaries'],
    queryFn: async () => {
      const res = await staffClient.listStaff({ search: '', departmentId: '', status: '' });
      const today = new Date();
      return z.array(anniversarySchema).parse(
        (res.staff ?? [])
          .filter((member) => /^\d{4}-\d{2}-\d{2}/.test(member.startDate ?? ''))
          .map((member) => {
            const date = member.startDate!.slice(0, 10);
            return {
              id: Number(member.id),
              staffId: Number(member.id),
              name:
                `${member.firstName ?? ''} ${member.lastName ?? ''}`.trim() ||
                member.email ||
                `#${member.id}`,
              date,
              years: Math.max(0, today.getFullYear() - Number(date.slice(0, 4))),
            };
          })
      );
    },
  });
}

export function useAvailability() {
  return useQuery({
    queryKey: ['availability'],
    queryFn: async () => {
      const [res, staffRes] = await Promise.all([
        shiftClient.listAvailability({ limit: 100, offset: 0 }),
        staffClient.listStaff({ search: '', departmentId: '', status: '' }),
      ]);
      const namesById = displayNameByStaffId(staffRes.staff);
      const weekdays = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'];
      return z.array(availabilitySchema).parse(
        (res.entries ?? []).map((entry: AvailabilityPayload) => {
          const staffId = Number(entry.staffId);
          const day = entry.specificDate ?? weekdays[Number(entry.dayOfWeek)] ?? 'Unknown';
          return {
            id: Number(entry.id),
            staffId,
            staffName: namesById.get(staffId) ?? `#${staffId}`,
            day,
            canWork: entry.isAvailable ?? false,
            startTime: entry.startTime,
            endTime: entry.endTime,
          };
        })
      );
    },
  });
}

export function useSwaps() {
  return useQuery({
    queryKey: ['swaps'],
    queryFn: async () => {
      const [res, staffRes] = await Promise.all([
        unwrap(hrExtras.listSwaps()),
        staffClient.listStaff({ search: '', departmentId: '', status: '' }),
      ]);
      const namesById = displayNameByStaffId(staffRes.staff);
      return z.array(swapRequestSchema).parse(
        (res.requests ?? []).map((request) => ({
          ...request,
          requester: request.requester || namesById.get(Number(request.requesterId)) || `#${request.requesterId}`,
          target: request.target || namesById.get(Number(request.targetId)) || `#${request.targetId}`,
        }))
      );
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
      const [res, staffRes] = await Promise.all([
        unwrap(hrExtras.listBreaks()),
        staffClient.listStaff({ search: '', departmentId: '', status: '' }),
      ]);
      const namesById = displayNameByStaffId(staffRes.staff);
      return z.array(breakLogSchema).parse(
        (res.breaks ?? []).map((entry) => ({
          ...entry,
          staffName: entry.staffName || namesById.get(Number(entry.staffId)) || `#${entry.staffId}`,
        }))
      );
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
      const [res, staffRes] = await Promise.all([
        overtimeClient.listOvertimeRequests({}),
        staffClient.listStaff({ search: '', departmentId: '', status: '' }),
      ]);
      const namesById = displayNameByStaffId(staffRes.staff);
      return z.array(overtimeAlertSchema).parse(
        (res.requests ?? [])
          .filter((request) => ['pending', 'flagged'].includes((request.status ?? '').toLowerCase()))
          .map((request) => ({
            id: Number(request.id),
            staffId: Number(request.employeeId),
            staffName:
              namesById.get(Number(request.employeeId)) ?? `#${request.employeeId}`,
            date: request.date ?? '',
            plannedExtraMinutes: Number(request.plannedExtraMinutes ?? 0),
            reason: request.reason ?? '',
            status: request.status ?? 'pending',
            triggeredAt: request.createdAt ?? request.date ?? '',
          }))
      );
    },
  });
}

export function useHrSettings() {
  return useQuery({
    queryKey: ['settings'],
    queryFn: async () => {
      const res = await unwrap(hrExtras.getSettings());
      const raw = ((res as unknown as { config?: Record<string, unknown> }).config ?? res) as Record<
        string,
        unknown
      >;
      return hrSettingsSchema.parse({
        workWeekStart: raw.workWeekStart ?? 'Monday',
        overtimeAfterHours: raw.overtimeAfterHours ?? 40,
        ptoPolicyDays: raw.ptoPolicyDays ?? 0,
        currency: raw.currency ?? raw.defaultCurrency ?? 'USD',
        weeklyHours: raw.weeklyHours ?? 40,
      });
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
      const response = await staffClient.getStaff({ id: String(id) });
      const staff = response.staff;
      if (!staff) return null;
      return employeeProfileSchema.parse({
        id: Number(staff.id),
        phone: staff.phone,
        address: staff.address,
        hiredAt: staff.startDate,
        managerId: staff.userId ? Number(staff.userId) : undefined,
        bio: staff.notes,
        emergency: staff.emergencyContactName
          ? {
              name: staff.emergencyContactName,
              relation: '',
              phone: staff.emergencyContactPhone ?? '',
            }
          : undefined,
      });
    },
  });
}
