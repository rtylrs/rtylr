import { z } from 'zod';

// crm.feedback.v1.Feedback.type is either "nps" or "review" - a single row is
// one or the other, never both. score's scale/meaning depends on this: a 0-10
// NPS response, or a 1-5 star rating.
export const FEEDBACK_TYPES = ['nps', 'review'] as const;
export type FeedbackType = (typeof FEEDBACK_TYPES)[number];

// Every field mirrors crm.feedback.v1.Feedback, where all fields are
// optional on the wire (omitempty). Keeping them required here would make a
// single row without e.g. `comment` throw and kill the whole list parse.
export const feedbackItemSchema = z.object({
  id: z.string(),
  customerName: z.string().default(''),
  type: z.enum(FEEDBACK_TYPES),
  score: z.number().default(0),
  comment: z.string().default(''),
  createdAt: z.string(),
  source: z.string().default(''),
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
  return item.type === 'review' && item.score >= 4;
}

export function isNegativeFeedback(item: FeedbackItem): boolean {
  return item.type === 'review' && item.score <= 2;
}

export function isPromoter(item: FeedbackItem): boolean {
  return item.type === 'nps' && item.score >= 9;
}

export function isDetractor(item: FeedbackItem): boolean {
  return item.type === 'nps' && item.score <= 6;
}
