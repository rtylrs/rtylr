import type { Template } from '../schemas/templates';

/** Returns templates filtered by channel. */
export function filterByChannel(templates: Template[], channel: 'email' | 'sms'): Template[] {
  return templates.filter((t) => t.channel === channel);
}

/** Returns templates filtered by category (case-insensitive). */
export function filterByCategory(templates: Template[], category: string): Template[] {
  return templates.filter((t) => t.category.toLowerCase() === category.toLowerCase());
}

/** Returns the unique categories from a list of templates. */
export function collectCategories(templates: Template[]): string[] {
  const seen = new Set<string>();
  templates.forEach((t) => seen.add(t.category));
  return [...seen].sort();
}

/** Returns templates sorted by name alphabetically. */
export function sortByName(templates: Template[]): Template[] {
  return [...templates].sort((a, b) => a.name.localeCompare(b.name));
}

/** Returns templates sorted by updatedAt, newest first. */
export function sortByRecency(templates: Template[]): Template[] {
  return [...templates].sort((a, b) => b.updatedAt.localeCompare(a.updatedAt));
}

/** Interpolates placeholders in a template body. */
export function interpolate(body: string, vars: Record<string, string>): string {
  return body.replace(/\{\{(\w+)\}\}/g, (_match, key: string) => vars[key] ?? `{{${key}}}`);
}
