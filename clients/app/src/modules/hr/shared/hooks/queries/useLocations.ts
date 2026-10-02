import { unwrap } from '@voxire/api';
import { useMutation } from '@tanstack/react-query';
import { hrTransport } from '../../lib/clients';

export interface CreateLocationInput {
  name: string;
  address?: string;
}

interface LocationPayload {
  id: number;
  name: string;
  address?: string;
}

interface CreateLocationResponse {
  location: LocationPayload;
}

// hr.location.v1.LocationService is defined in voxie but was never wired into
// any onion module's RegisterAPI, so it 404s in production. This calls the
// legacy chi-routed /hr/locations path instead, which is the one actually
// live. Its handler binds JSON by exact snake_case tag with no camelCase
// fallback (see the same note that used to live on the department legacy
// path), so keys are converted by hand rather than left to the usual
// camelCase-to-snake_case interceptor.
export function useCreateLocation() {
  return useMutation({
    mutationFn: (data: CreateLocationInput) =>
      unwrap(
        hrTransport.post<CreateLocationResponse>('/hr/locations', {
          name: data.name,
          address: data.address,
        })
      ),
  });
}
