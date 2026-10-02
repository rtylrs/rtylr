import { z } from 'zod';

// Backed by hr.benefit.v1's BenefitPlan (listBenefits). id stays a wire string
// for the same reason as myBenefitEnrollmentSchema below - protobuf-JSON
// int64 encoding. category is the backend's free-form `type` field
// title-cased for display, not a fixed client enum: the RPC has no enrolled
// employee count either, so that stat is not rendered rather than faked.
export const benefitPlanSchema = z.object({
  id: z.string(),
  name: z.string(),
  category: z.string(),
  carrier: z.string(),
  monthlyCost: z.number(),
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
