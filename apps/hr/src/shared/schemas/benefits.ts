import { z } from 'zod';

export const benefitPlanSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  category: z.enum(['Health', 'Dental', 'Vision', '401k', 'Life']).catch('Health'),
  carrier: z.string(),
  monthlyCost: z.number(),
  enrolledCount: z.number(),
});

export const benefitEnrollmentSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  staffId: z.number(),
  staffName: z.string(),
  planId: z.number(),
  planName: z.string(),
  startedAt: z.string(),
  status: z.enum(['Active', 'Pending', 'Waived']).catch('Pending'),
});

export type BenefitPlan = z.infer<typeof benefitPlanSchema>;
export type BenefitEnrollment = z.infer<typeof benefitEnrollmentSchema>;

// The caller's own enrollment, as returned by hr.benefit.v1's
// ListMyBenefitEnrollments / EnrollInBenefit. planId stays a string: the wire
// encodes it using the protobuf-JSON int64 convention, and status is a free
// backend string (e.g. "active", "cancelled") rather than a fixed UI enum.
export const myBenefitEnrollmentSchema = z.object({
  id: z.string(),
  planId: z.string(),
  status: z.string(),
  enrolledAt: z.string(),
});

export type MyBenefitEnrollment = z.infer<typeof myBenefitEnrollmentSchema>;
