import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type {
  ReviewCyclePayload,
  PerformanceReviewPayload,
  ReviewGoalPayload,
} from '@voxire/voxie/hr/performance/v1/types';
import type {
  CreateReviewCycleRequest,
  CreateReviewCycleResponse,
} from '@voxire/voxie/hr/performance/v1/types';
import type {
  OpenReviewCycleRequest,
  OpenReviewCycleResponse,
} from '@voxire/voxie/hr/performance/v1/types';
import type {
  SubmitReviewRequest,
  SubmitReviewResponse,
} from '@voxire/voxie/hr/performance/v1/types';
import type { ListReviewCyclesResponse } from '@voxire/voxie/hr/performance/v1/types';
import type { ListGoalsRequest, ListGoalsResponse } from '@voxire/voxie/hr/performance/v1/types';
import { hrTransport } from '../../lib/clients';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

/* ── Local param types for endpoints missing from contracts ── */

interface ListReviewCyclesParams {
  limit?: number;
  offset?: number;
}

interface ListPerformanceReviewsParams {
  reviewCycleId?: number;
  status?: string;
  limit?: number;
  offset?: number;
}

interface ListPerformanceReviewsResponse {
  reviews: PerformanceReviewPayload[];
  total: number;
  limit: number;
  offset: number;
}

interface CreateGoalRequest {
  reviewCycleId: number;
  title: string;
  description: string;
  weight?: number;
}

interface CreateGoalResponse {
  goal: ReviewGoalPayload;
}

interface UpdateGoalRequest {
  goalId: number;
  title?: string;
  description?: string;
  weight?: number;
  status?: string;
}

interface UpdateGoalResponse {
  goal: ReviewGoalPayload;
}

interface ShareReviewRequest {
  reviewId: number;
}

interface ShareReviewResponse {
  review: PerformanceReviewPayload;
}

type CreateReviewCycleMutationInput = CreateReviewCycleRequest & {
  type?: string;
  cycleType?: string;
  periodStart?: string;
  periodEnd?: string;
  reviewTypes?: string[];
};

/* ── Queries ── */

export const useListReviewCycles = (params?: ListReviewCyclesParams) => {
  return useQuery<ListReviewCyclesResponse>({
    queryKey: ['review-cycles', 'filter', params],
    queryFn: async (): Promise<ListReviewCyclesResponse> => {
      const res = await unwrap(
        hrTransport.get<ListReviewCyclesResponse>('/hr/performance/cycles', { params })
      );
      return res as ListReviewCyclesResponse;
    },
  });
};

export const useGetReviewCycle = (id: number | undefined) => {
  return useQuery<ReviewCyclePayload>({
    queryKey: ['review-cycle', id],
    queryFn: async (): Promise<ReviewCyclePayload> => {
      const res = await unwrap(
        hrTransport.get<{ cycle: ReviewCyclePayload }>(`/hr/performance/cycles/${id}`)
      );
      return (res as { cycle: ReviewCyclePayload }).cycle;
    },
    enabled: !!id,
  });
};

export const useCreateReviewCycle = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (
      data: CreateReviewCycleMutationInput
    ): Promise<ReviewCyclePayload | undefined> => {
      const res = await unwrap(
        hrTransport.post<CreateReviewCycleResponse>('/hr/performance/cycles', {
          name: data.name,
          cycleType: data.cycleType ?? data.type,
          periodStart: data.periodStart,
          periodEnd: data.periodEnd,
          reviewTypes: data.reviewTypes,
        })
      );
      return (res as CreateReviewCycleResponse).cycle;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['review-cycles'] });
      toast.success(t('performance.cycleCreated'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useOpenReviewCycle = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: OpenReviewCycleRequest): Promise<ReviewCyclePayload | undefined> => {
      const res = await unwrap(
        hrTransport.post<OpenReviewCycleResponse>(`/hr/performance/cycles/${data.cycleId}/open`, {
          id: data.cycleId,
        })
      );
      return (res as OpenReviewCycleResponse).cycle;
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['review-cycles'] });
      queryClient.invalidateQueries({ queryKey: ['review-cycle', variables.cycleId] });
      toast.success(t('performance.cycleOpened'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useListGoals = (params?: ListGoalsRequest) => {
  return useQuery<ListGoalsResponse>({
    queryKey: ['review-goals', 'filter', params],
    queryFn: async (): Promise<ListGoalsResponse> => {
      const res = await unwrap(
        hrTransport.get<ListGoalsResponse>('/hr/performance/goals', { params })
      );
      return res as ListGoalsResponse;
    },
  });
};

export const useCreateGoal = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateGoalRequest): Promise<ReviewGoalPayload> => {
      const res = await unwrap(hrTransport.post<CreateGoalResponse>('/hr/performance/goals', data));
      return (res as CreateGoalResponse).goal;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['review-goals'] });
      toast.success(t('performance.goalCreated'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useUpdateGoal = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: UpdateGoalRequest): Promise<ReviewGoalPayload> => {
      const source = data as UpdateGoalRequest & { successCriteria?: string };
      const res = await unwrap(
        hrTransport.put<UpdateGoalResponse>(`/hr/performance/goals/${data.goalId}`, {
          id: data.goalId,
          title: data.title,
          description: data.description,
          weight: data.weight,
          successCriteria: source.successCriteria,
          status: data.status,
        })
      );
      return (res as UpdateGoalResponse).goal;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['review-goals'] });
      toast.success(t('performance.goalUpdated'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useListReviews = (params?: ListPerformanceReviewsParams) => {
  return useQuery<ListPerformanceReviewsResponse>({
    queryKey: ['performance-reviews', 'filter', params],
    queryFn: async (): Promise<ListPerformanceReviewsResponse> => {
      const res = await unwrap(
        hrTransport.get<ListPerformanceReviewsResponse>('/hr/performance/reviews', { params })
      );
      return res as ListPerformanceReviewsResponse;
    },
  });
};

export const useSubmitReview = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (
      data: SubmitReviewRequest
    ): Promise<PerformanceReviewPayload | undefined> => {
      const source = data as SubmitReviewRequest & {
        rating?: string;
        narrative?: string;
        overallRating?: number;
      };
      const res = await unwrap(
        hrTransport.post<SubmitReviewResponse>(`/hr/performance/reviews/${data.reviewId}/submit`, {
          id: data.reviewId,
          rating: source.rating ?? String(source.overallRating ?? ''),
          narrative: source.narrative,
        })
      );
      return (res as SubmitReviewResponse).review;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['performance-reviews'] });
      toast.success(t('performance.reviewSubmitted'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useShareReview = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: ShareReviewRequest): Promise<PerformanceReviewPayload> => {
      const res = await unwrap(
        hrTransport.post<ShareReviewResponse>(`/hr/performance/reviews/${data.reviewId}/share`, {
          id: data.reviewId,
        })
      );
      return (res as ShareReviewResponse).review;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['performance-reviews'] });
      toast.success(t('performance.reviewShared'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};
