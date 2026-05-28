import { describe, it, expect } from 'vitest';

import {
  findWorkspaceById,
  findWorkspaceByName,
  sortWorkspacesAlpha,
  workspaceNameExists,
  workspaceInitials,
  searchWorkspaces,
} from './workspaceUtils';
import type { Workspace } from '../stores/flowStore';

function makeWorkspace(overrides: Partial<Workspace>): Workspace {
  return {
    id: 'w1',
    name: 'Acme Corp',
    color: '#3b82f6',
    ...overrides,
  } as Workspace;
}

describe('findWorkspaceById', () => {
  it('finds workspace by id', () => {
    const workspaces = [makeWorkspace({ id: 'a' }), makeWorkspace({ id: 'b', name: 'Beta' })];
    expect(findWorkspaceById(workspaces, 'b')?.name).toBe('Beta');
  });

  it('returns undefined when not found', () => {
    expect(findWorkspaceById([], 'x')).toBeUndefined();
  });
});

describe('findWorkspaceByName', () => {
  it('finds by name case-insensitively', () => {
    const workspaces = [makeWorkspace({ name: 'Acme Corp' })];
    expect(findWorkspaceByName(workspaces, 'acme corp')?.id).toBe('w1');
  });

  it('returns undefined when not found', () => {
    expect(findWorkspaceByName([], 'x')).toBeUndefined();
  });
});

describe('sortWorkspacesAlpha', () => {
  it('sorts alphabetically', () => {
    const workspaces = [
      makeWorkspace({ name: 'Zebra' }),
      makeWorkspace({ name: 'Alpha' }),
      makeWorkspace({ name: 'Mango' }),
    ];
    expect(sortWorkspacesAlpha(workspaces).map((w) => w.name)).toEqual(['Alpha', 'Mango', 'Zebra']);
  });

  it('does not mutate original', () => {
    const workspaces = [makeWorkspace({ name: 'Z' }), makeWorkspace({ name: 'A' })];
    const original = [...workspaces];
    sortWorkspacesAlpha(workspaces);
    expect(workspaces).toEqual(original);
  });
});

describe('workspaceNameExists', () => {
  it('returns true for existing name (case-insensitive)', () => {
    const workspaces = [makeWorkspace({ name: 'Acme Corp' })];
    expect(workspaceNameExists(workspaces, 'acme corp')).toBe(true);
  });

  it('returns false for new name', () => {
    const workspaces = [makeWorkspace({ name: 'Acme Corp' })];
    expect(workspaceNameExists(workspaces, 'Beta Inc')).toBe(false);
  });
});

describe('workspaceInitials', () => {
  it('returns first two letters for single word', () => {
    expect(workspaceInitials('Acme')).toBe('AC');
  });

  it('returns initials of first two words', () => {
    expect(workspaceInitials('Acme Corp')).toBe('AC');
  });

  it('handles extra spaces', () => {
    expect(workspaceInitials('  Beta   Inc  ')).toBe('BI');
  });
});

describe('searchWorkspaces', () => {
  it('returns matching workspaces', () => {
    const workspaces = [makeWorkspace({ name: 'Acme Corp' }), makeWorkspace({ name: 'Beta Inc' })];
    expect(searchWorkspaces(workspaces, 'acme').map((w) => w.name)).toEqual(['Acme Corp']);
  });

  it('returns all workspaces for empty query', () => {
    const workspaces = [makeWorkspace(), makeWorkspace({ id: 'w2' })];
    expect(searchWorkspaces(workspaces, '')).toHaveLength(2);
  });
});
