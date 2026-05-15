import { CARET_MODELS, type CaretModel } from '@voxire/caret';

/** Approximate token count using a 4-chars-per-token rule of thumb. */
export function approxTokens(text: string): number {
  if (!text) return 0;
  return Math.ceil(text.length / 4);
}

/** Percentage of the model's context window we'd consume if we sent this text. */
export function contextWindowPercent(model: CaretModel, text: string): number {
  const def = CARET_MODELS[model];
  if (!def || def.contextWindow === 0) return 0;
  return Math.min(100, (approxTokens(text) / def.contextWindow) * 100);
}
