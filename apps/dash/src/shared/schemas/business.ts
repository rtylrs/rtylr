import { z } from 'zod';

export const businessSchema = z.object({}).passthrough();

export type Business = z.infer<typeof businessSchema>;
