import { useQuery } from '@tanstack/react-query';
import { z } from 'zod';
import { configClient } from '../../lib/clients';

const industryTypeSchema = z.object({
  slug: z.string(),
  label: z.string(),
  icon: z.string(),
});

const industryGroupSchema = z.object({
  category: z.string(),
  types: z.array(industryTypeSchema),
});

const industriesResponseSchema = z.object({
  groups: z.array(industryGroupSchema),
});

export type IndustryType = z.infer<typeof industryTypeSchema>;
export type IndustryGroup = z.infer<typeof industryGroupSchema>;

export const useIndustries = () => {
  return useQuery({
    queryKey: ['industries'],
    queryFn: async () => {
      return industriesResponseSchema.parse(await configClient.getIndustries({}));
    },
    staleTime: 24 * 60 * 60 * 1000,
  });
};
