import { describe, expect, it } from 'vitest';
import { FEATURE_MATRIX, getRtylrAdvantages, getSupportSummary } from './featureMatrix';

describe('rtylr POS feature matrix', () => {
  it('rtylr supports every feature that Omega supports', () => {
    const omegaFeatures = FEATURE_MATRIX.filter((f) => f.omega === 'yes');
    const rtylrMissing = omegaFeatures.filter((f) => f.rtylr !== 'yes');

    expect(
      rtylrMissing.map((f) => `${f.category}: ${f.feature}`),
    ).toEqual([]);
  });

  it('rtylr has more full-support features than Omega', () => {
    const summary = getSupportSummary();
    expect(summary.rtylr.yes).toBeGreaterThan(summary.omega.yes);
  });

  it('rtylr has fewer gaps (no) than Omega', () => {
    const summary = getSupportSummary();
    expect(summary.rtylr.no).toBeLessThan(summary.omega.no);
  });

  it('rtylr has at least 10 exclusive advantages over Omega', () => {
    const advantages = getRtylrAdvantages();
    expect(advantages.length).toBeGreaterThanOrEqual(10);
  });

  it('all features list a category and feature name', () => {
    for (const entry of FEATURE_MATRIX) {
      expect(entry.category).toBeTruthy();
      expect(entry.feature).toBeTruthy();
    }
  });

  it('hardware detection covers ESC/POS, ZPL, Epson ePOS, Star, KDS, and LPR/IPP', () => {
    const hwFeatures = FEATURE_MATRIX.filter((f) => f.category === 'Hardware' && f.rtylr === 'yes');
    const labels = hwFeatures.map((f) => f.feature);

    expect(labels.some((l) => l.includes('ESC/POS'))).toBe(true);
    expect(labels.some((l) => l.includes('ZPL'))).toBe(true);
    expect(labels.some((l) => l.includes('Epson ePOS'))).toBe(true);
    expect(labels.some((l) => l.includes('Star Micronics'))).toBe(true);
    expect(labels.some((l) => l.includes('KDS'))).toBe(true);
    expect(labels.some((l) => l.includes('LPR'))).toBe(true);
    expect(labels.some((l) => l.includes('IPP'))).toBe(true);
  });

  it('offline queue is an exclusive rtylr advantage not present in Omega', () => {
    const entry = FEATURE_MATRIX.find(
      (f) => f.feature.includes('Offline queue'),
    );
    expect(entry?.rtylr).toBe('yes');
    expect(entry?.omega).toBe('no');
  });
});
