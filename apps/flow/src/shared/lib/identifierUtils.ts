/** Generate a task identifier string like "PROJ-42". */
export function makeTaskIdentifier(projectName: string, count: number): string {
  const prefix =
    projectName
      .slice(0, 4)
      .toUpperCase()
      .replace(/[^A-Z]/g, 'X') || 'TASK';
  return `${prefix}-${count}`;
}

/** Extract the numeric part from an identifier like "FLOW-42" -> 42. */
export function identifierToNumber(identifier: string): number {
  const parts = identifier.split('-');
  const last = parts[parts.length - 1];
  const n = parseInt(last, 10);
  return Number.isNaN(n) ? 0 : n;
}

/** Sort identifiers numerically (FLOW-9 before FLOW-10). */
export function sortIdentifiers(ids: string[]): string[] {
  return [...ids].sort((a, b) => identifierToNumber(a) - identifierToNumber(b));
}

/** Returns true when a string looks like a valid task identifier (e.g. "FLOW-123"). */
export function isValidIdentifier(s: string): boolean {
  return /^[A-Z]{1,8}-\d+$/.test(s);
}

/** Normalize an identifier to uppercase and trim whitespace. */
export function normalizeIdentifier(id: string): string {
  return id.trim().toUpperCase();
}
