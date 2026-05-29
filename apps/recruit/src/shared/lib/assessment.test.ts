import { describe, it, expect } from 'vitest';
import {
  assessmentTypeLabel,
  assessmentScorePct,
  isPassing,
  completedAssessments,
  avgAssessmentScore,
  assessmentPassRate,
} from './assessment';
import type { Assessment } from './assessment';

function makeAssessment(
  status: 'pending' | 'in_progress' | 'completed' | 'expired',
  score?: number,
  maxScore = 100
): Assessment {
  return {
    id: 'a1',
    candidateId: 'c1',
    type: 'coding',
    status,
    score,
    maxScore,
    timeLimitMinutes: 60,
    sentAt: '2026-01-01',
  };
}

const assessments = [
  makeAssessment('completed', 85),
  makeAssessment('completed', 60),
  makeAssessment('completed', 90),
  makeAssessment('pending'),
];

describe('assessmentTypeLabel', () => {
  it('returns display labels for known types', () => {
    expect(assessmentTypeLabel('coding')).toBe('Coding Challenge');
    expect(assessmentTypeLabel('aptitude')).toBe('Aptitude Test');
  });

  it('returns raw value for unknown type', () => {
    expect(assessmentTypeLabel('custom')).toBe('custom');
  });
});

describe('assessmentScorePct', () => {
  it('returns score as percentage', () => {
    expect(assessmentScorePct(makeAssessment('completed', 85))).toBe(85);
  });

  it('returns 0 when score is undefined', () => {
    expect(assessmentScorePct(makeAssessment('pending'))).toBe(0);
  });

  it('returns 0 when maxScore is 0', () => {
    expect(assessmentScorePct(makeAssessment('completed', 50, 0))).toBe(0);
  });
});

describe('isPassing', () => {
  it('returns true when score meets threshold', () => {
    expect(isPassing(makeAssessment('completed', 75), 70)).toBe(true);
  });

  it('returns false when score is below threshold', () => {
    expect(isPassing(makeAssessment('completed', 60), 70)).toBe(false);
  });
});

describe('completedAssessments', () => {
  it('returns only completed assessments', () => {
    expect(completedAssessments(assessments)).toHaveLength(3);
  });
});

describe('avgAssessmentScore', () => {
  it('returns average score percentage', () => {
    expect(avgAssessmentScore(assessments)).toBeCloseTo(78.3, 0);
  });

  it('returns 0 for empty array', () => {
    expect(avgAssessmentScore([])).toBe(0);
  });
});

describe('assessmentPassRate', () => {
  it('returns percentage of assessments passing threshold', () => {
    expect(assessmentPassRate(assessments, 70)).toBeCloseTo(66.7, 0);
  });

  it('returns 0 when no completed assessments', () => {
    expect(assessmentPassRate([makeAssessment('pending')])).toBe(0);
  });
});
