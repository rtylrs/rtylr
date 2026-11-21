/**
 * The colours a match code is shown on, by the `color` index the server sends
 * (0 to 5). The phone and the staff board use the same list, so the customer
 * and the cashier see the same colour. White text passes contrast on each.
 */
export const MATCH_COLORS = ['#1d4ed8', '#b91c1c', '#15803d', '#7e22ce', '#c2410c', '#0f766e'] as const;

export function matchColor(index: number | undefined): string {
  return MATCH_COLORS[index ?? 0] ?? MATCH_COLORS[0];
}
