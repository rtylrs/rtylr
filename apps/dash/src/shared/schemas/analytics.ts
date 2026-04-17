import { z } from 'zod';

export const analyticsResponseSchema = z.object({}).passthrough();

export type AnalyticsResponse = z.infer<typeof analyticsResponseSchema>;
