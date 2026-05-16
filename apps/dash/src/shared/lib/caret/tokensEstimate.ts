/** Estimate the total token cost of an array of strings. Useful for the
 *  composer "remaining context" gauge. */
import { approxTokens } from './contextWindow';

export function totalTokens(parts: string[]): number {
  let sum = 0;
  for (const p of parts) sum += approxTokens(p);
  return sum;
}
