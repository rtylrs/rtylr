// Skills assessment helpers for the recruit admin panel.

export type AssessmentStatus = 'pending' | 'in_progress' | 'completed' | 'expired';
export type AssessmentType = 'coding' | 'written' | 'personality' | 'aptitude' | 'technical';

export interface Assessment {
  id: string;
  candidateId: string;
  type: AssessmentType;
  status: AssessmentStatus;
  score?: number;
  maxScore: number;
  timeLimitMinutes: number;
  sentAt: string;
  completedAt?: string;
}

/**
 * assessmentTypeLabel returns a display label for an assessment type.
 */
export function assessmentTypeLabel(type: AssessmentType | string): string {
  const labels: Record<string, string> = {
    coding: 'Coding Challenge',
    written: 'Written Assessment',
    personality: 'Personality Assessment',
    aptitude: 'Aptitude Test',
    technical: 'Technical Interview',
  };
  return labels[type] ?? type;
}

/**
 * assessmentScorePct returns the score as a percentage of the max score.
 */
export function assessmentScorePct(assessment: Assessment): number {
  if (assessment.maxScore === 0 || assessment.score === undefined) return 0;
  return Math.round((assessment.score / assessment.maxScore) * 1000) / 10;
}

/**
 * isPassing returns true when the score percentage meets the passing threshold.
 */
export function isPassing(assessment: Assessment, threshold = 70): boolean {
  return assessmentScorePct(assessment) >= threshold;
}

/**
 * completedAssessments returns assessments with completed status.
 */
export function completedAssessments(assessments: Assessment[]): Assessment[] {
  return assessments.filter((a) => a.status === 'completed');
}

/**
 * avgAssessmentScore returns the average score percentage across completed assessments.
 */
export function avgAssessmentScore(assessments: Assessment[]): number {
  const completed = completedAssessments(assessments).filter((a) => a.score !== undefined);
  if (completed.length === 0) return 0;
  const sum = completed.reduce((acc, a) => acc + assessmentScorePct(a), 0);
  return Math.round((sum / completed.length) * 10) / 10;
}

/**
 * assessmentPassRate returns the percentage of completed assessments that passed.
 */
export function assessmentPassRate(assessments: Assessment[], threshold = 70): number {
  const completed = completedAssessments(assessments);
  if (completed.length === 0) return 0;
  const passed = completed.filter((a) => isPassing(a, threshold)).length;
  return Math.round((passed / completed.length) * 1000) / 10;
}
