import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { businessClient, userClient } from '../../lib/clients';
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

interface CreateSubBusinessRequest {
  name: string;
  parentBusinessId: number;
}

export const useListMyBusinesses = () =>
  useQuery({
    queryKey: ['businesses', 'mine'],
    queryFn: async () => {
      const response = await businessClient.listMyBusinesses({});

      return {
        businesses: (response.businesses ?? []).map((business) => ({
          ...business,
          id: Number(business.id),
          name: business.name ?? '',
          logoUrl: business.logoUrl ?? '',
          parentId: Number(business.parentId ?? 0),
          isActive: business.isActive ?? true,
          isOwned: business.isOwned ?? true,
        })),
      };
    },
    staleTime: 30_000,
  });

export const useCreateSubBusiness = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateSubBusinessRequest) =>
      businessClient.createSubBusiness({
        name: data.name,
        // OneKit wire-encodes int64 request fields as JSON strings. A bare
        // number 400s with "cannot unmarshal number into Go struct field
        // .parent_business_id of type string", same as switch-business.
        parentBusinessId: String(data.parentBusinessId),
      }),
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
      businessClient.renameBusiness({ businessId: String(businessId), name }),
    onSuccess: (_response, variables) => {
      void queryClient.invalidateQueries({ queryKey: ['businesses', 'mine'] });
      void queryClient.invalidateQueries({ queryKey: ['business'] });
      void queryClient.invalidateQueries({ queryKey: ['user'] });
      const currentUser = useUserStore.getState().user;
      if (currentUser && Number(currentUser.businessId) === variables.businessId) {
        useUserStore.getState().setUser({ ...currentUser, businessName: variables.name });
        // websiteData holds the current business's config; writing another
        // business's name into it would relabel the current one.
        const currentWebsiteData = useConfigStore.getState().websiteData ?? {};
        useConfigStore.getState().setWebsiteData({
          ...currentWebsiteData,
          siteName: variables.name,
        });
      }
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
      businessClient.setLogo({ businessId: String(businessId), logoUrl }),
    onSuccess: (_response, variables) => {
      void queryClient.invalidateQueries({ queryKey: ['businesses', 'mine'] });
      void queryClient.invalidateQueries({ queryKey: ['business'] });
      void queryClient.invalidateQueries({ queryKey: ['user'] });
      const currentUser = useUserStore.getState().user;
      if (currentUser && Number(currentUser.businessId) === variables.businessId) {
        useUserStore.getState().setUser({ ...currentUser, logo: variables.logoUrl });
        // Same as rename: only the current business's cached config may change.
        const currentWebsiteData = useConfigStore.getState().websiteData ?? {};
        useConfigStore.getState().setWebsiteData({
          ...currentWebsiteData,
          logoUrl: variables.logoUrl,
        });
      }
    },
  });
};

export const useDeleteBusiness = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (businessId: number) =>
      businessClient.deleteBusiness({ businessId: String(businessId) }),
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
      userClient.switchBusiness({
        // OneKit serializes int64 request fields as JSON strings so values
        // remain lossless across JavaScript clients and the Go server.
        businessId: String(businessId),
      }),
  });

// Live availability check for the onboarding screen (business names are globally unique).
export const useBusinessExists = (name: string) =>
  useQuery({
    queryKey: ['business-exists', name.trim()],
    queryFn: () => businessClient.checkBusinessNameExists({ name: name.trim() }),
    enabled: name.trim().length > 1,
    staleTime: 10_000,
  });

// Creates the owner's first business during onboarding; returns a fresh token
// carrying the new business_id.
export const useOnboardBusiness = () =>
  useMutation({
    mutationFn: (name: string) => businessClient.onboardBusiness({ name }),
  });
