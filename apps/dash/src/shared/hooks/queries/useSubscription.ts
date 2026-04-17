import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { safe } from '../../lib/api';

export interface SubscriptionPayload {
  id: number;
  planName: string;
  planSlug: string;
  apps: string[];
  status: 'active' | 'trial' | 'cancelled' | 'expired';
}

export interface PlanPayload {
  id: number;
  name: string;
  slug: string;
  apps: string[];
  isActive: boolean;
}

interface MySubscriptionResponse {
  subscription: SubscriptionPayload | null;
}

interface ListPlansResponse {
  plans: PlanPayload[];
}

interface SetSubscriptionRequest {
  planSlug: string;
}

export const useMySubscription = () =>
  useQuery({
    queryKey: ['subscription'],
    queryFn: () => unwrap(safe.get<MySubscriptionResponse>('/subscriptions/my')),
    select: (data) => data.subscription,
  });

export const useListPlans = () =>
  useQuery({
    queryKey: ['subscription', 'plans'],
    queryFn: () => unwrap(safe.get<ListPlansResponse>('/subscriptions/plans')),
    select: (data) => data.plans,
  });

export const useSetSubscription = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: SetSubscriptionRequest) =>
      unwrap(safe.post('/subscriptions/my', { plan_slug: data.planSlug })),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['subscription'] });
    },
  });
};
