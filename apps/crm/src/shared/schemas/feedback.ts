import { z } from 'zod';

export const feedbackItemSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  customer: z.string(),
  rating: z.number(),
  nps: z.number(),
  comment: z.string(),
  createdAt: z.string(),
  source: z.string(),
});

export type FeedbackItem = z.infer<typeof feedbackItemSchema>;

// -- Domain constants ---------------------------------------------------------

export const FEEDBACK_SOURCES = ['survey', 'review', 'pos', 'email', 'app'] as const;
export type FeedbackSource = (typeof FEEDBACK_SOURCES)[number];

// -- Helpers ------------------------------------------------------------------

export function isFeedbackSource(value: string): value is FeedbackSource {
  return (FEEDBACK_SOURCES as readonly string[]).includes(value);
}

export function isPositiveFeedback(item: FeedbackItem): boolean {
  return item.rating >= 4;
}

export function isNegativeFeedback(item: FeedbackItem): boolean {
  return item.rating <= 2;
}

export function isPromoter(item: FeedbackItem): boolean {
  return item.nps >= 9;
}

export function isDetractor(item: FeedbackItem): boolean {
  return item.nps <= 6;
}
