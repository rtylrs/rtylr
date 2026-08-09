import { createOneKitClient } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { z } from 'zod';
import { FeedbackServiceClient } from '@voxire/voxie/crm/feedback/v1/client';
import type { Feedback } from '@voxire/voxie/crm/feedback/v1/types';
import { feedbackItemSchema, type FeedbackType } from '../../schemas/feedback';
import type { FeedbackItem } from '../../schemas/feedback';

const feedbackClient = createOneKitClient(FeedbackServiceClient, 'crm');

export interface ListFeedbackParams {
  type?: FeedbackType;
  source?: string;
  minScore?: number;
  maxScore?: number;
}

function fromApiFeedback(f: Feedback): FeedbackItem {
  return feedbackItemSchema.parse({
    id: f.id,
    customerName: f.customerName,
    type: f.type,
    score: f.score,
    comment: f.comment,
    createdAt: f.createdAt,
    source: f.source,
  });
}

export const useListFeedback = (params?: ListFeedbackParams): UseQueryResult<FeedbackItem[]> =>
  useQuery<FeedbackItem[]>({
    queryKey: ['crm-feedback', params],
    queryFn: async (): Promise<FeedbackItem[]> => {
      const res = await feedbackClient.listFeedback({
        type: params?.type ?? '',
        source: params?.source ?? '',
        minScore: params?.minScore,
        maxScore: params?.maxScore,
        page: 1,
        pageSize: 100,
      });
      return z.array(feedbackItemSchema).parse((res.feedback ?? []).map(fromApiFeedback));
    },
  });

export type { FeedbackItem } from '../../schemas/feedback';
