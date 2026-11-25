/**
 * Story points scales (T-381). The server stores a number; T-shirt sizes map to
 * the numbers below so cycle totals and velocity can add them up.
 */
export type PointsScale = 'fibonacci' | 'linear' | 'tshirt';

const SCALE_VALUES: Record<PointsScale, number[]> = {
  fibonacci: [1, 2, 3, 5, 8, 13],
  linear: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
  tshirt: [1, 2, 3, 5, 8],
};

const TSHIRT_LABELS: Record<number, string> = { 1: 'XS', 2: 'S', 3: 'M', 5: 'L', 8: 'XL' };

export const POINTS_SCALES: PointsScale[] = ['fibonacci', 'linear', 'tshirt'];

/** The project's scale; anything unknown (or absent) is Fibonacci, the default. */
export function toPointsScale(value: string | undefined): PointsScale {
  return value === 'linear' || value === 'tshirt' ? value : 'fibonacci';
}

export function pointsOptions(scale: PointsScale): number[] {
  return SCALE_VALUES[scale];
}

/** How a stored value reads on a scale: "XL" on T-shirt, the number otherwise. */
export function formatPoints(points: number, scale: PointsScale): string {
  return scale === 'tshirt' ? (TSHIRT_LABELS[points] ?? String(points)) : String(points);
}
