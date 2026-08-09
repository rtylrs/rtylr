import type { Task, Project, Member } from '../stores/flowStore';

/** Normalizes a string for comparison (lowercase, trimmed). */
function normalize(s: string): string {
  return s.toLowerCase().trim();
}

/** Returns true if the haystack contains the needle (case-insensitive). */
function contains(haystack: string, needle: string): boolean {
  return normalize(haystack).includes(normalize(needle));
}

/** Search tasks by title, identifier, or project name. */
export function searchTasks(tasks: Task[], query: string): Task[] {
  if (!query.trim()) return tasks;
  return tasks.filter(
    (t) =>
      contains(t.title, query) || contains(t.identifier, query) || contains(t.projectName, query)
  );
}

/** Search projects by name. */
export function searchProjects(projects: Project[], query: string): Project[] {
  if (!query.trim()) return projects;
  return projects.filter((p) => contains(p.name, query));
}

/** Search members by name or email. */
export function searchMembers(members: Member[], query: string): Member[] {
  if (!query.trim()) return members;
  return members.filter((m) => contains(m.name, query) || contains(m.email, query));
}

/** Returns true if a task matches all provided token substrings (multi-word search). */
export function taskMatchesTokens(task: Task, tokens: string[]): boolean {
  return tokens.every(
    (token) =>
      contains(task.title, token) ||
      contains(task.identifier, token) ||
      contains(task.projectName, token)
  );
}

/** Splits a query into lowercase tokens for multi-token search. */
export function tokenizeQuery(query: string): string[] {
  return query
    .trim()
    .toLowerCase()
    .split(/\s+/)
    .filter((t) => t.length > 0);
}

/** Multi-token task search (all words must appear somewhere in the task). */
export function searchTasksMultiToken(tasks: Task[], query: string): Task[] {
  const tokens = tokenizeQuery(query);
  if (tokens.length === 0) return tasks;
  return tasks.filter((t) => taskMatchesTokens(t, tokens));
}

/** Returns tasks whose identifier starts with the given prefix (e.g. "FLOW-"). */
export function tasksByIdentifierPrefix(tasks: Task[], prefix: string): Task[] {
  const p = prefix.toUpperCase();
  return tasks.filter((t) => t.identifier.toUpperCase().startsWith(p));
}
