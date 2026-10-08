import { createOneKitClient } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { z } from 'zod';
import { FeedbackServiceClient } from '@voxire/voxie/crm/feedback/v1/client';
import type { Feedback } from '@voxire/voxie/crm/feedback/v1/types';
import { feedbackItemSchema, type FeedbackType } from '../../schemas/feedback';
import type { FeedbackItem } from '../../schemas/feedback';
import { unwrapList } from '../../lib/unwrapList';

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
      const feedback: FeedbackItem[] = [];
      let page = 1;
      let total = 0;

      do {
        const res = await feedbackClient.listFeedback({
          type: params?.type ?? '',
          source: params?.source ?? '',
          minScore: params?.minScore,
          maxScore: params?.maxScore,
          page,
          pageSize: 100,
        });
        const pageItems = unwrapList<Feedback>(res, 'feedback').map(fromApiFeedback);
        feedback.push(...pageItems);
        total = res.total ?? feedback.length;
        page += 1;
        if (pageItems.length === 0) {
          if (feedback.length < total) throw new Error('Feedback pagination ended early');
          break;
        }
      } while (feedback.length < total);

      return z.array(feedbackItemSchema).parse(feedback);
    },
  });

export type { FeedbackItem } from '../../schemas/feedback';
