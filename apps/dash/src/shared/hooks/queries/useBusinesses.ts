import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { safe } from '../../lib/api';

export interface OwnedBusiness {
  id: number;
  name: string;
  logoUrl: string;
  parentId: number;
  isActive: boolean;
}

interface ListMyBusinessesResponse {
  businesses: OwnedBusiness[];
}

interface CreateSubBusinessRequest {
  name: string;
  parentBusinessId: number;
}

interface CreateSubBusinessResponse {
  id: number;
}

interface SwitchBusinessResponse {
  token: string;
}

export const useListMyBusinesses = () =>
  useQuery({
    queryKey: ['businesses', 'mine'],
    queryFn: () =>
      unwrap(safe.get<ListMyBusinessesResponse>('/api/hub/business/v1/businesses/mine')),
    staleTime: 30_000,
  });

export const useCreateSubBusiness = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateSubBusinessRequest) =>
      unwrap(safe.post<CreateSubBusinessResponse>('/api/hub/business/v1/businesses', data)),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['businesses', 'mine'] });
    },
  });
};

export const useSwitchBusiness = () =>
  useMutation({
    mutationFn: (businessId: number) =>
      unwrap(
        safe.post<SwitchBusinessResponse>('/api/hub/user/v1/users/switch-business', {
          business_id: businessId,
        })
      ),
  });

interface OnboardBusinessResponse {
  token: string;
}

// Live availability check for the onboarding screen (business names are globally unique).
export const useBusinessExists = (name: string) =>
  useQuery({
    queryKey: ['business-exists', name.trim()],
    queryFn: () =>
      unwrap(
        safe.get<{ exists: boolean }>('/api/hub/business/v1/businesses/name-exists', {
          params: { name: name.trim() },
        })
      ),
    enabled: name.trim().length > 1,
    staleTime: 10_000,
  });

// Creates the owner's first business during onboarding; returns a fresh token
// carrying the new business_id.
export const useOnboardBusiness = () =>
  useMutation({
    mutationFn: (name: string) =>
      unwrap(
        safe.post<OnboardBusinessResponse>('/api/hub/business/v1/businesses/onboard', { name })
      ),
  });
