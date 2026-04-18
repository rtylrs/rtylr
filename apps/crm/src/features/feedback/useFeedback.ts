import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { z } from 'zod';
import { crmApi } from '../../shared/services/crmApi';
import { feedbackItemSchema } from '../../shared/schemas/feedback';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useFeedback() {
  return useQuery({
    queryKey: ['feedback'],
    queryFn: async () => {
      const res = await unwrap(crmApi.getFeedback());
      return z.array(feedbackItemSchema).parse(unwrapList(res, 'feedback', 'items'));
    },
  });
}
