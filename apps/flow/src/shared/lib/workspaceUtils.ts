import type { Workspace } from '../stores/flowStore';

/** Find a workspace by its ID. */
export function findWorkspaceById(workspaces: Workspace[], id: string): Workspace | undefined {
  return workspaces.find((w) => w.id === id);
}

/** Find a workspace by its name (case-insensitive). */
export function findWorkspaceByName(workspaces: Workspace[], name: string): Workspace | undefined {
  const lower = name.trim().toLowerCase();
  return workspaces.find((w) => w.name.toLowerCase() === lower);
}

/** Sort workspaces alphabetically. */
export function sortWorkspacesAlpha(workspaces: Workspace[]): Workspace[] {
  return [...workspaces].sort((a, b) => a.name.localeCompare(b.name));
}

/** Returns true when a workspace name already exists (case-insensitive). */
export function workspaceNameExists(workspaces: Workspace[], name: string): boolean {
  const lower = name.trim().toLowerCase();
  return workspaces.some((w) => w.name.toLowerCase() === lower);
}

/** Returns initials for a workspace name (first letter of each word, up to 2). */
export function workspaceInitials(name: string): string {
  const words = name.trim().split(/\s+/);
  if (words.length === 1) return words[0].slice(0, 2).toUpperCase();
  return (words[0][0] + words[1][0]).toUpperCase();
}

/** Returns workspaces that match a search query. */
export function searchWorkspaces(workspaces: Workspace[], query: string): Workspace[] {
  const lower = query.trim().toLowerCase();
  if (!lower) return workspaces;
  return workspaces.filter((w) => w.name.toLowerCase().includes(lower));
}
