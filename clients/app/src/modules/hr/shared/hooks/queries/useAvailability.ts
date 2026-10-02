import { createOneKitClient, unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type {
  ListAvailabilityResponse,
  SetAvailabilityRequest,
  SetAvailabilityResponse,
} from '@voxire/voxie/hr/shift/v1/types';
import { StaffServiceClient } from '@voxire/voxie/hr/staff/v1/client';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { hrTransport } from '../../lib/clients';
import { availabilitySchema, type Availability } from '../../schemas/shifts';

// The staff module mounts its generated server, so the generated client's
// versioned path is the one onion actually serves. That is not true of
// hr_shifts - see AVAILABILITY_PATH below.
const staffClient = createOneKitClient(StaffServiceClient, 'hr');

/**
 * The availability endpoints as onion mounts them.
 *
 * `hr.shift.v1` declares these as `@get("/availability")` and
 * `@post("/availability")`, which onekit generates as `/hr/shift/v1/availability`.
 * onion does not serve that path for this service: `hr_shifts` hand-registers its
 * routes (`/hr/shifts`, `/hr/shifts/roster/publish`), unlike `staff`, which
 * mounts its generated server and therefore does serve versioned paths. So the
 * generated `ShiftServiceClient` cannot be used here the way `useStaff` uses
 * `StaffServiceClient`.
 *
 * Which of the two onion picks for availability is not yet decided, so the path
 * lives here alone and is the single line to change once the backend lands.
 * Tracked on tryrtylr/onion#69.
 */
const AVAILABILITY_PATH = '/hr/shifts/availability';

/**
 * The availability table loads every staff member's windows in one call, so the
 * page size has to cover a whole team rather than a screenful. Kept explicit
 * because the endpoint is paginated and omitting it would silently accept
 * whatever default the server applies.
 */
const AVAILABILITY_PAGE_SIZE = 200;

/**
 * A window is either a recurring weekly pattern or a single-date override, never
 * both and never neither. `.onk` has no cross-field decorator to say so, so the
 * contract leans on the onion handler to reject a bad request. This union moves
 * that check to compile time for our own call sites: supplying both keys, or
 * neither, does not typecheck.
 */
export type SetAvailabilityInput = {
  /** Absent creates the window, or upserts it on its natural key. */
  id?: number;
  staffId: number;
  startTime: string;
  endTime: string;
  /** Absent means available. False marks a deliberate block. */
  isAvailable?: boolean;
} & ({ dayOfWeek: number; specificDate?: never } | { specificDate: string; dayOfWeek?: never });

export const useAvailability = () =>
  useQuery({
    queryKey: ['hr', 'availability'],
    queryFn: async (): Promise<Availability[]> => {
      // AvailabilityPayload carries only staff_id, so the display name has to be
      // joined in from the staff service, as useTimeOff does for its requests.
      const [res, staffRes] = await Promise.all([
        unwrap(
          hrTransport.get<ListAvailabilityResponse>(AVAILABILITY_PATH, {
            params: { limit: AVAILABILITY_PAGE_SIZE, offset: 0 },
          })
        ),
        staffClient.listStaff({ search: '', departmentId: '', status: '' }),
      ]);
      const staffById = new Map((staffRes.staff ?? []).map((s) => [s.id, s]));

      return (res.entries ?? []).map((entry) => {
        const member = entry.staffId === undefined ? undefined : staffById.get(entry.staffId);
        return availabilitySchema.parse({
          id: entry.id,
          staffId: Number(entry.staffId),
          staffName: member
            ? `${member.firstName} ${member.lastName}`.trim()
            : `Staff #${entry.staffId}`,
          dayOfWeek: entry.dayOfWeek,
          specificDate: entry.specificDate,
          // Passed through on purpose: the schema defaults an absent key to
          // false, which is what the server dropping a false value means.
          isAvailable: entry.isAvailable,
          startTime: entry.startTime,
          endTime: entry.endTime,
        });
      });
    },
  });

export const useSetAvailability = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation('hrCommon');

  return useMutation({
    mutationFn: (input: SetAvailabilityInput): Promise<SetAvailabilityResponse> => {
      // Typed as the generated request so the field set cannot drift from the
      // contract. Sent camelCase: the @voxire/api request interceptor converts
      // body keys to snake_case, which is what the handler binds.
      const body: SetAvailabilityRequest = {
        id: input.id === undefined ? undefined : String(input.id),
        staffId: String(input.staffId),
        dayOfWeek: input.dayOfWeek,
        specificDate: input.specificDate,
        startTime: input.startTime,
        endTime: input.endTime,
        isAvailable: input.isAvailable,
      };
      return unwrap(hrTransport.post<SetAvailabilityResponse>(AVAILABILITY_PATH, body));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hr', 'availability'] });
      toast.success(t('hrOperations.availability.saved'));
    },
    onError: () => toast.error(t('hrOperations.availability.saveFailed')),
  });
};
