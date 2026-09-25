import { createOneKitClient, unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { PerformanceServiceClient } from '@voxire/voxie/hr/performance/v1/client';
import type {
  ReviewCyclePayload,
  PerformanceReviewPayload,
  ReviewGoalPayload,
  CreateReviewCycleRequest,
  CreateReviewCycleResponse,
  OpenReviewCycleRequest,
  OpenReviewCycleResponse,
  SubmitReviewRequest,
  SubmitReviewResponse,
  ListReviewCyclesResponse,
  ListGoalsRequest as GeneratedListGoalsRequest,
  ListGoalsResponse,
} from '@voxire/voxie/hr/performance/v1/types';
import { hrTransport } from '../../lib/clients';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

const performanceClient = createOneKitClient(PerformanceServiceClient, 'hr');

export interface ListPerformanceReviewsParams {
  reviewCycleId?: number;
  status?: string;
  limit?: number;
  offset?: number;
}

export interface ListPerformanceReviewsResponse {
  reviews: PerformanceReviewPayload[];
  total: number;
  limit: number;
  offset: number;
}

export interface ListGoalsParams {
  employeeId?: number | string;
  reviewCycleId?: number | string;
}

export interface CreateGoalRequest {
  employeeId: number;
  reviewCycleId: number;
  title: string;
  description: string;
  weight?: number;
  successCriteria?: string;
}

interface CreateGoalResponse {
  goal?: ReviewGoalPayload;
}

interface UpdateGoalRequest {
  goalId: number;
  title?: string;
  description?: string;
  weight?: number;
  status?: string;
}

interface UpdateGoalResponse {
  goal?: ReviewGoalPayload;
}

interface ShareReviewRequest {
  reviewId: number;
}

interface ShareReviewResponse {
  review: PerformanceReviewPayload;
}

function positiveInteger(value: number | string | undefined): number | undefined {
  if (value === undefined || value === null || value === '') return undefined;
  const parsed = Number(value);
  return Number.isSafeInteger(parsed) && parsed > 0 ? parsed : undefined;
}

function requiredPositiveInteger(value: number | string | undefined, label: string): number {
  const parsed = positiveInteger(value);
  if (!parsed) throw new Error(`${label} is required`);
  return parsed;
}

export const useListReviewCycles = () => {
  return useQuery<ListReviewCyclesResponse>({
    queryKey: ['review-cycles'],
    queryFn: () => performanceClient.listReviewCycles({}),
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
    enabled: id !== undefined && id > 0,
  });
};

export const useCreateReviewCycle = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: (data: CreateReviewCycleRequest): Promise<ReviewCyclePayload | undefined> =>
      performanceClient
        .createReviewCycle(data)
        .then((response: CreateReviewCycleResponse) => response.cycle),
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
    mutationFn: (data: OpenReviewCycleRequest): Promise<ReviewCyclePayload | undefined> => {
      const cycleId = requiredPositiveInteger(data.cycleId, 'cycleId');
      return performanceClient
        .openReviewCycle({ cycleId: String(cycleId) })
        .then((response: OpenReviewCycleResponse) => response.cycle);
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

export const useListGoals = (params?: ListGoalsParams) => {
  return useQuery<ListGoalsResponse>({
    queryKey: ['review-goals', 'filter', params],
    queryFn: () => {
      const request: GeneratedListGoalsRequest = {};
      const employeeId = positiveInteger(params?.employeeId);
      const reviewCycleId = positiveInteger(params?.reviewCycleId);
      if (employeeId !== undefined) request.employeeId = String(employeeId);
      if (reviewCycleId !== undefined) request.reviewCycleId = String(reviewCycleId);
      return performanceClient.listGoals(request);
    },
  });
};

export const useCreateGoal = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateGoalRequest): Promise<ReviewGoalPayload> => {
      const employeeId = requiredPositiveInteger(data.employeeId, 'employeeId');
      const reviewCycleId = requiredPositiveInteger(data.reviewCycleId, 'reviewCycleId');
      const res = await unwrap(
        hrTransport.post<CreateGoalResponse>('/hr/performance/goals', {
          employeeId,
          reviewCycleId,
          title: data.title,
          description: data.description,
          weight: data.weight ?? 0,
          successCriteria: data.successCriteria,
        })
      );
      return (res as CreateGoalResponse).goal as ReviewGoalPayload;
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
      const goalId = requiredPositiveInteger(data.goalId, 'goalId');
      const source = data as UpdateGoalRequest & { successCriteria?: string };
      const res = await unwrap(
        hrTransport.put<UpdateGoalResponse>(`/hr/performance/goals/${goalId}`, {
          id: goalId,
          title: data.title,
          description: data.description,
          weight: data.weight,
          successCriteria: source.successCriteria,
          status: data.status,
        })
      );
      return (res as UpdateGoalResponse).goal as ReviewGoalPayload;
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
  const status = params?.status?.trim();
  const serverParams: {
    cycleId?: number;
    employeeId?: number;
    limit?: number;
    offset?: number;
  } = {};
  const cycleId = positiveInteger(params?.reviewCycleId);
  if (cycleId !== undefined) serverParams.cycleId = cycleId;
  if (params?.limit !== undefined) serverParams.limit = params.limit;
  if (params?.offset !== undefined) serverParams.offset = params.offset;

  return useQuery<ListPerformanceReviewsResponse>({
    // Status is deliberately absent from the server query. The live handler
    // accepts cycle_id but has no status filter, so status is applied below to
    // the returned page without changing the request contract.
    queryKey: ['performance-reviews', 'filter', serverParams],
    queryFn: async (): Promise<ListPerformanceReviewsResponse> => {
      const res = await unwrap(
        hrTransport.get<ListPerformanceReviewsResponse>('/hr/performance/reviews', {
          params: serverParams,
        })
      );
      return res as ListPerformanceReviewsResponse;
    },
    select: (response) => {
      if (!status) return response;
      const wanted = status.toLowerCase();
      const reviews = response.reviews.filter(
        (review) => (review.status ?? '').toLowerCase() === wanted
      );
      return { ...response, reviews, total: reviews.length };
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
      const response: SubmitReviewResponse = await performanceClient.submitReview(data);
      return response.review;
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
      const reviewId = requiredPositiveInteger(data.reviewId, 'reviewId');
      const res = await unwrap(
        hrTransport.post<ShareReviewResponse>(`/hr/performance/reviews/${reviewId}/share`, {
          id: reviewId,
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
