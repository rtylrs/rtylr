import { z } from 'zod';

/**
 * `/hr/employees/me` is an ad hoc REST endpoint outside the `hr.employee.v1`
 * contract (that contract has no self-lookup RPC, and no live server behind
 * it at all). This is the only way to resolve the caller's own
 * `hr_employees.id` - a different id-space from the hub auth user id, and
 * the id `hr.leave.v1`'s `employee_id` actually expects.
 */
export const myEmployeeSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
});
