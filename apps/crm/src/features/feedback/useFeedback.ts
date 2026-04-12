import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { z } from 'zod';
import { crmApi } from '../../shared/services/crmApi';
import { feedbackItemSchema } from '../../shared/schemas/feedback';

export function useFeedback() {
  return useQuery({
    queryKey: ['feedback'],
    queryFn: async () => {
      const res = await unwrap(crmApi.getFeedback());
      const raw = res ?? [];
      return z.array(feedbackItemSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}
