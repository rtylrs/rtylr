// Candidate scoring helpers for the recruit admin panel.

import type { Submission } from './types';

export interface ScoredSubmission {
  submission: Submission;
  score: number;
  reasons: string[];
}

/**
 * completenessScore returns a 0-100 score based on filled optional fields.
 */
export function completenessScore(submission: Submission): number {
  let filled = 0;
  const hasResume = submission.hasResume ?? Boolean(submission.resumeUrl);
  const checks = [
    hasResume,
    Boolean(submission.coverLetter?.trim()),
    Boolean(submission.phone?.trim()),
  ];
  for (const check of checks) {
    if (check) filled++;
  }
  return Math.round((filled / checks.length) * 100);
}

/**
 * completenessReasons returns human-readable reasons for the completeness score.
 */
export function completenessReasons(submission: Submission): string[] {
  const reasons: string[] = [];
  if (submission.hasResume ?? Boolean(submission.resumeUrl)) reasons.push('Resume attached');
  if (submission.coverLetter?.trim()) reasons.push('Cover letter provided');
  if (submission.phone?.trim()) reasons.push('Phone number provided');
  return reasons;
}

/**
 * scoreSubmission returns a ScoredSubmission with completeness details.
 */
export function scoreSubmission(submission: Submission): ScoredSubmission {
  return {
    submission,
    score: completenessScore(submission),
    reasons: completenessReasons(submission),
  };
}

/**
 * rankSubmissions returns submissions sorted by completeness score descending.
 */
export function rankSubmissions(submissions: Submission[]): ScoredSubmission[] {
  return submissions.map(scoreSubmission).sort((a, b) => b.score - a.score);
}
