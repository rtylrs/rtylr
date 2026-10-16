import type { Review, OneOnOne } from '../schemas/reviews';
import type { TrainingCourse } from '../schemas/training';
import type { Anniversary } from '../schemas/staff';

export interface ApiReview {
  id?: string;
  staffId?: string;
  reviewerId?: string;
  period?: string;
  rating?: number;
  goalsRating?: number;
  valuesRating?: number;
  skillsRating?: number;
  notes?: string;
  status?: string;
  createdAt?: string;
}

export interface ApiOneOnOne {
  id?: string;
  managerId?: string;
  staffId?: string;
  scheduledAt?: string;
  notes?: string;
  actionItems?: string;
  status?: string;
}

export interface ApiTrainingCourse {
  id?: string;
  name?: string;
  description?: string;
  durationHours?: number;
  isRequired?: boolean;
}

export type StaffNames = ReadonlyMap<number, string>;

const DEFAULT_ONE_ON_ONE_MINUTES = 30;
const COURSE_CATEGORIES: TrainingCourse['category'][] = [
  'Compliance',
  'Safety',
  'Skills',
  'Leadership',
];

const toNumber = (value: string | number | undefined): number => {
  const parsed = Number(value ?? 0);
  return Number.isFinite(parsed) ? parsed : 0;
};

export function reviewStatusFromApi(status: string | undefined): Review['status'] {
  switch ((status ?? '').toLowerCase()) {
    case 'submitted':
      return 'Submitted';
    case 'acknowledged':
      return 'Completed';
    default:
      return 'Draft';
  }
}

export function reviewStatusToApi(status: Review['status']): string {
  switch (status) {
    case 'Submitted':
      return 'submitted';
    case 'Completed':
      return 'acknowledged';
    default:
      return 'draft';
  }
}

export function mapReview(review: ApiReview, names: StaffNames): Review {
  const staffId = toNumber(review.staffId);
  const reviewerId = toNumber(review.reviewerId);
  return {
    id: toNumber(review.id),
    staffId,
    staffName: names.get(staffId) ?? '',
    reviewerId,
    reviewer: names.get(reviewerId) ?? '',
    period: review.period ?? '',
    status: reviewStatusFromApi(review.status),
    overall: review.rating ?? 0,
    goals: review.goalsRating ?? 0,
    values: review.valuesRating ?? 0,
    skills: review.skillsRating ?? 0,
    notes: review.notes ?? '',
    date: (review.createdAt ?? '').slice(0, 10),
  };
}

export function oneOnOneStatusFromApi(status: string | undefined): OneOnOne['status'] {
  switch ((status ?? '').toLowerCase()) {
    case 'completed':
      return 'Completed';
    case 'cancelled':
      return 'Cancelled';
    default:
      return 'Scheduled';
  }
}

export function topicsToActionItems(topics: string[]): string {
  return JSON.stringify(topics);
}

export function actionItemsToTopics(actionItems: string | undefined): string[] {
  if (!actionItems) return [];
  try {
    const parsed: unknown = JSON.parse(actionItems);
    if (Array.isArray(parsed)) {
      return parsed.filter((item): item is string => typeof item === 'string' && item !== '');
    }
  } catch {
    return actionItems
      .split('\n')
      .map((line) => line.trim())
      .filter(Boolean);
  }
  return [];
}

export function mapOneOnOne(item: ApiOneOnOne, names: StaffNames): OneOnOne {
  const managerId = toNumber(item.managerId);
  const reportId = toNumber(item.staffId);
  return {
    id: toNumber(item.id),
    managerId,
    manager: names.get(managerId) ?? '',
    reportId,
    report: names.get(reportId) ?? '',
    scheduledAt: item.scheduledAt ?? '',
    duration: DEFAULT_ONE_ON_ONE_MINUTES,
    status: oneOnOneStatusFromApi(item.status),
    topics: actionItemsToTopics(item.actionItems),
    notes: item.notes,
  };
}

export function mapTrainingCourse(course: ApiTrainingCourse): TrainingCourse {
  const description = course.description ?? '';
  const category = COURSE_CATEGORIES.find((c) => c.toLowerCase() === description.toLowerCase());
  return {
    id: toNumber(course.id),
    title: course.name ?? '',
    category: category ?? 'Skills',
    durationMin: Math.round((course.durationHours ?? 0) * 60),
    modules: 1,
    required: course.isRequired ?? false,
  };
}

export interface AnniversaryStaffInput {
  id?: string;
  firstName?: string;
  lastName?: string;
  email?: string;
  status?: string;
  startDate?: string;
}

const INACTIVE_STAFF_STATUSES = new Set(['terminated', 'inactive', 'resigned', 'offboarded']);

/**
 * Work anniversaries for the staff who have one in the next `windowDays` days.
 * `date` is the MM-DD the widget counts down to; `years` is the milestone the
 * upcoming anniversary completes, so a hire made this year never appears.
 */
export function buildAnniversaries(
  staff: AnniversaryStaffInput[],
  today: Date,
  windowDays = 60
): Anniversary[] {
  const todayMidnight = new Date(today.getFullYear(), today.getMonth(), today.getDate());
  const result: Anniversary[] = [];
  for (const member of staff) {
    if (INACTIVE_STAFF_STATUSES.has((member.status ?? '').toLowerCase())) continue;
    const match = /^(\d{4})-(\d{2})-(\d{2})/.exec(member.startDate ?? '');
    if (!match) continue;
    const startYear = Number(match[1]);
    const month = Number(match[2]);
    const day = Number(match[3]);
    let next = new Date(todayMidnight.getFullYear(), month - 1, day);
    if (next < todayMidnight) next = new Date(todayMidnight.getFullYear() + 1, month - 1, day);
    const years = next.getFullYear() - startYear;
    if (years < 1) continue;
    const daysAway = Math.round((next.getTime() - todayMidnight.getTime()) / 86_400_000);
    if (daysAway > windowDays) continue;
    const name = `${member.firstName ?? ''} ${member.lastName ?? ''}`.trim() || member.email || '';
    result.push({
      id: toNumber(member.id),
      staffId: toNumber(member.id),
      name,
      date: `${match[2]}-${match[3]}`,
      years,
    });
  }
  return result;
}
