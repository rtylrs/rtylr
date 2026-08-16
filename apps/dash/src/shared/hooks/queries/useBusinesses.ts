import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { safe } from '../../lib/api';
import { useConfigStore } from '../../stores/configStore';
import { useUserStore } from '../../stores/userStore';

export interface OwnedBusiness {
  id: number;
  name: string;
  logoUrl: string;
  parentId: number;
  isActive: boolean;
  /** False when the current user only has team access to the business. */
  isOwned?: boolean;
}

interface ListMyBusinessesResponse {
  businesses?: Array<OwnedBusiness & { id?: number | string; parentId?: number | string }>;
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
    queryFn: async () => {
      const response = await unwrap(
        safe.get<ListMyBusinessesResponse>('/hub/business/v1/businesses/mine')
      );

      return {
        businesses: (response.businesses ?? []).map((business) => ({
          ...business,
          id: Number(business.id),
          parentId: Number(business.parentId ?? 0),
          isOwned: business.isOwned ?? true,
        })),
      } satisfies ListMyBusinessesResponse;
    },
    staleTime: 30_000,
  });

export const useCreateSubBusiness = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateSubBusinessRequest) =>
      unwrap(safe.post<CreateSubBusinessResponse>('/hub/business/v1/businesses', data)),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['businesses', 'mine'] });
      void queryClient.invalidateQueries({ queryKey: ['business'] });
    },
  });
};

interface RenameBusinessRequest {
  businessId: number;
  name: string;
}

export const useRenameBusiness = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ businessId, name }: RenameBusinessRequest) =>
      unwrap(
        safe.patch<{ business?: unknown }>(
          `/hub/business/v1/businesses/${encodeURIComponent(String(businessId))}/name`,
          { name }
        )
      ),
    onSuccess: (_response, variables) => {
      void queryClient.invalidateQueries({ queryKey: ['businesses', 'mine'] });
      void queryClient.invalidateQueries({ queryKey: ['business'] });
      void queryClient.invalidateQueries({ queryKey: ['user'] });
      const currentUser = useUserStore.getState().user;
      if (currentUser && Number(currentUser.businessId) === variables.businessId) {
        useUserStore.getState().setUser({ ...currentUser, businessName: variables.name });
      }
      const currentWebsiteData = useConfigStore.getState().websiteData ?? {};
      useConfigStore.getState().setWebsiteData({
        ...currentWebsiteData,
        siteName: variables.name,
      });
    },
  });
};

interface SetBusinessLogoRequest {
  businessId: number;
  logoUrl: string;
}

export const useSetBusinessLogo = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ businessId, logoUrl }: SetBusinessLogoRequest) =>
      unwrap(
        safe.patch<{ business?: unknown }>(
          `/hub/business/v1/businesses/${encodeURIComponent(String(businessId))}/logo`,
          { logoUrl }
        )
      ),
    onSuccess: (_response, variables) => {
      void queryClient.invalidateQueries({ queryKey: ['businesses', 'mine'] });
      void queryClient.invalidateQueries({ queryKey: ['business'] });
      void queryClient.invalidateQueries({ queryKey: ['user'] });
      const currentUser = useUserStore.getState().user;
      if (currentUser && Number(currentUser.businessId) === variables.businessId) {
        useUserStore.getState().setUser({ ...currentUser, logo: variables.logoUrl });
      }
      const currentWebsiteData = useConfigStore.getState().websiteData ?? {};
      useConfigStore.getState().setWebsiteData({
        ...currentWebsiteData,
        logoUrl: variables.logoUrl,
      });
    },
  });
};

export const useDeleteBusiness = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (businessId: number) =>
      unwrap(safe.delete(`/hub/business/v1/businesses/${encodeURIComponent(String(businessId))}`)),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['businesses', 'mine'] });
      void queryClient.invalidateQueries({ queryKey: ['business'] });
      void queryClient.invalidateQueries({ queryKey: ['user'] });
    },
  });
};

export const useSwitchBusiness = () =>
  useMutation({
    mutationFn: (businessId: number) =>
      unwrap(
        safe.post<SwitchBusinessResponse>('/hub/user/v1/users/switch-business', {
          // OneKit serializes int64 request fields as JSON strings so values
          // remain lossless across JavaScript clients and the Go server.
          business_id: String(businessId),
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
        safe.get<{ exists: boolean }>('/hub/business/v1/businesses/name-exists', {
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
      unwrap(safe.post<OnboardBusinessResponse>('/hub/business/v1/businesses/onboard', { name })),
  });
