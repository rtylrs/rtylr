import { unwrap } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { z } from 'zod';
import { safe } from '../../lib/api';
import { unwrapList } from '../../lib/unwrapList';
import { feedbackItemSchema } from '../../schemas/feedback';
import type { FeedbackItem } from '../../schemas/feedback';

export interface ListFeedbackParams {
  rating?: number;
}

export const useListFeedback = (params?: ListFeedbackParams): UseQueryResult<FeedbackItem[]> =>
  useQuery<FeedbackItem[]>({
    queryKey: ['crm-feedback', params],
    queryFn: async (): Promise<FeedbackItem[]> => {
      const res = await unwrap(safe.get<unknown>('/crm/feedback', params ? { params } : undefined));
      return z.array(feedbackItemSchema).parse(unwrapList(res, 'feedback', 'items'));
    },
  });

export type { FeedbackItem } from '../../schemas/feedback';
