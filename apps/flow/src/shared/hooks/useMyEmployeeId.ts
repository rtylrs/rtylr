import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { safe } from '../lib/api';
import { myEmployeeSchema } from '../schemas/employee';

/**
 * The caller's own hr_employees.id, resolved via the same self-lookup
 * endpoint HR's own frontend uses (`/hr/employees/me`) - never the hub auth
 * user id, which is a different id-space and gets rejected outright by
 * hr.leave.v1's write endpoints.
 */
export function useMyEmployeeId() {
  return useQuery({
    queryKey: ['my-employee-id'],
    queryFn: async () => {
      const res = await unwrap(safe.get<unknown>('/hr/employees/me'));
      const parsed = myEmployeeSchema.parse((res as { employee: unknown }).employee);
      return parsed.id;
    },
  });
}
