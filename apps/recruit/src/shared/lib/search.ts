// Search helpers for the recruit careers and admin panel.

import type { RecruitCareer, Submission } from './types';

/**
 * searchCareers returns careers where the title, department, or location
 * contains the query (case-insensitive).
 */
export function searchCareers(careers: RecruitCareer[], query: string): RecruitCareer[] {
  if (!query.trim()) return careers;
  const q = query.toLowerCase();
  return careers.filter(
    (c) =>
      c.title.toLowerCase().includes(q) ||
      c.department.toLowerCase().includes(q) ||
      c.location.toLowerCase().includes(q)
  );
}

/**
 * searchSubmissions returns submissions where the candidate name or email
 * contains the query (case-insensitive).
 */
export function searchSubmissions(submissions: Submission[], query: string): Submission[] {
  if (!query.trim()) return submissions;
  const q = query.toLowerCase();
  return submissions.filter(
    (s) => s.candidateName.toLowerCase().includes(q) || s.email.toLowerCase().includes(q)
  );
}

/**
 * highlightQuery wraps the first occurrence of query in the text with a
 * simple marker string for rendering. Returns original text when no match.
 */
export function highlightQuery(text: string, query: string): string {
  if (!query.trim()) return text;
  const idx = text.toLowerCase().indexOf(query.toLowerCase());
  if (idx === -1) return text;
  return (
    text.slice(0, idx) +
    `[HIGHLIGHT]${text.slice(idx, idx + query.length)}[/HIGHLIGHT]` +
    text.slice(idx + query.length)
  );
}

/**
 * tokenizeQuery splits a search query into individual lowercase tokens.
 */
export function tokenizeQuery(query: string): string[] {
  return query
    .toLowerCase()
    .trim()
    .split(/\s+/)
    .filter((t) => t.length > 0);
}
