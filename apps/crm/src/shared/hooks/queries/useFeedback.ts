import { createOneKitClient } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { z } from 'zod';
import { FeedbackServiceClient } from '@voxire/voxie/crm/feedback/v1/client';
import type { Feedback } from '@voxire/voxie/crm/feedback/v1/types';
import { FEEDBACK_TYPES, feedbackItemSchema, type FeedbackType } from '../../schemas/feedback';
import type { FeedbackItem } from '../../schemas/feedback';
import { appendUniqueById } from '../../lib/pagination';

const feedbackClient = createOneKitClient(FeedbackServiceClient, 'crm');

export interface ListFeedbackParams {
  type?: FeedbackType;
  source?: string;
  minScore?: number;
  maxScore?: number;
}

function fromApiFeedback(f: Feedback): FeedbackItem {
  const type = FEEDBACK_TYPES.includes(f.type as FeedbackType)
    ? (f.type as FeedbackType)
    : 'review';
  return feedbackItemSchema.parse({
    id: f.id ?? '',
    customerName: f.customerName ?? 'Unknown customer',
    type,
    score: Number.isFinite(f.score) ? f.score : 0,
    comment: f.comment ?? '',
    createdAt: f.createdAt ?? '',
    source: f.source ?? '',
  });
}

export const useListFeedback = (params?: ListFeedbackParams): UseQueryResult<FeedbackItem[]> =>
  useQuery<FeedbackItem[]>({
    queryKey: ['crm-feedback', params],
    queryFn: async (): Promise<FeedbackItem[]> => {
      const pageSize = 100;
      const feedback: FeedbackItem[] = [];
      let page = 1;
      const maxPages = 100;

      for (let pageIndex = 0; pageIndex < maxPages; pageIndex += 1) {
        const res = await feedbackClient.listFeedback({
          type: params?.type ?? '',
          source: params?.source ?? '',
          minScore: params?.minScore,
          maxScore: params?.maxScore,
          page,
          pageSize,
        });
        const currentPage = z
          .array(feedbackItemSchema)
          .parse((res.feedback ?? []).map(fromApiFeedback));
        const added = appendUniqueById(feedback, currentPage);
        const total = Number(res.total);
        if (
          added === 0 ||
          currentPage.length < pageSize ||
          (Number.isFinite(total) && feedback.length >= total)
        ) {
          break;
        }
        page += 1;
      }

      return feedback;
    },
  });

export type { FeedbackItem } from '../../schemas/feedback';
