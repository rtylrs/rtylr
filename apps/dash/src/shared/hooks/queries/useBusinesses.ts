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
    queryFn: () => unwrap(safe.get<ListMyBusinessesResponse>('/businesses/mine')),
    staleTime: 30_000,
  });

export const useCreateSubBusiness = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateSubBusinessRequest) =>
      unwrap(safe.post<CreateSubBusinessResponse>('/businesses', data)),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['businesses', 'mine'] });
    },
  });
};

export const useSwitchBusiness = () =>
  useMutation({
    mutationFn: (businessId: number) =>
      unwrap(safe.post<SwitchBusinessResponse>('/auth/switch-business', { businessId })),
  });
