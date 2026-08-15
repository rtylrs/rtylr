import { z } from 'zod';

/**
 * `/hr/staff/detail` is an ad hoc REST endpoint outside the `hr.staff.v1`
 * contract - photo and bio have no `.onk` field anywhere, so this is the only
 * source for them. Mirrors `apps/hr/src/shared/schemas/staff.ts`'s
 * `employeeProfileSchema`, which reads the same endpoint.
 */
export const profileDetailSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  photoUrl: z.string().optional(),
  bio: z.string().optional(),
});

export type ProfileDetail = z.infer<typeof profileDetailSchema>;
