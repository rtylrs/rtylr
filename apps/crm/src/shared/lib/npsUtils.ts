/**
 * NPS (Net Promoter Score) calculation utilities.
 */

export interface NpsBreakdown {
  promoters: number;
  passives: number;
  detractors: number;
  total: number;
  score: number;
  promoterPct: number;
  passivePct: number;
  detractorPct: number;
}

/** Classifies a single NPS response. */
export function classifyNps(nps: number): 'promoter' | 'passive' | 'detractor' {
  if (nps >= 9) return 'promoter';
  if (nps >= 7) return 'passive';
  return 'detractor';
}

/** Calculates a full NPS breakdown from an array of NPS scores. */
export function calculateNps(scores: number[]): NpsBreakdown {
  const total = scores.length;
  if (total === 0) {
    return {
      promoters: 0,
      passives: 0,
      detractors: 0,
      total: 0,
      score: 0,
      promoterPct: 0,
      passivePct: 0,
      detractorPct: 0,
    };
  }
  const promoters = scores.filter((s) => s >= 9).length;
  const passives = scores.filter((s) => s >= 7 && s < 9).length;
  const detractors = scores.filter((s) => s < 7).length;
  const score = Math.round(((promoters - detractors) / total) * 100);
  return {
    promoters,
    passives,
    detractors,
    total,
    score,
    promoterPct: (promoters / total) * 100,
    passivePct: (passives / total) * 100,
    detractorPct: (detractors / total) * 100,
  };
}

/** Returns a label for the NPS score range. */
export function npsLabel(score: number): 'Excellent' | 'Good' | 'Fair' | 'Poor' {
  if (score >= 70) return 'Excellent';
  if (score >= 50) return 'Good';
  if (score >= 0) return 'Fair';
  return 'Poor';
}
