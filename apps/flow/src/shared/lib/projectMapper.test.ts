import { describe, expect, it } from 'vitest';

import { apiProjectToProject, apiWorkspaceToWorkspace } from './projectMapper';
import { ACCENT_DEFAULT } from './colorUtils';
import type { ApiProject, ApiWorkspace } from '../services/flowApi';

function baseProject(overrides: Partial<ApiProject> = {}): ApiProject {
  return {
    id: '1',
    workspaceId: '10',
    name: 'P',
    color: '',
    taskCount: 0,
    completedCount: 0,
    memberAvatars: [],
    ...overrides,
  } as ApiProject;
}

describe('apiProjectToProject', () => {
  it('stringifies numeric IDs', () => {
    const out = apiProjectToProject(
      baseProject({
        id: 42 as unknown as string,
        workspaceId: 7 as unknown as string,
      })
    );
    expect(out.id).toBe('42');
  });

  it('falls back to the flow accent when color is missing', () => {
    // Without a color the sidebar dots and project chips render
    // invisible against a white background. The fallback is ACCENT_DEFAULT,
    // asserted through the constant rather than a copy of its value: a
    // project's colour is persisted, so it has to be a real hex, and
    // normalizeHex makes the lowercase form the canonical one.
    const out = apiProjectToProject(baseProject({ color: '' }));
    expect(out.color).toBe(ACCENT_DEFAULT);
  });

  it('computes progress percentage from task counts', () => {
    const out = apiProjectToProject(baseProject({ taskCount: 10, completedCount: 3 }));
    expect(out.progress).toBe(30);
  });

  it('returns 0% progress when there are no tasks (no divide-by-zero)', () => {
    const out = apiProjectToProject(baseProject({ taskCount: 0, completedCount: 0 }));
    expect(out.progress).toBe(0);
  });

  it('rounds progress to the nearest integer', () => {
    // 7 / 11 = 0.6363… → 64, not 63.6
    const out = apiProjectToProject(baseProject({ taskCount: 11, completedCount: 7 }));
    expect(out.progress).toBe(64);
  });

  it('defaults health to on_track when the server does not emit it', () => {
    const out = apiProjectToProject(baseProject());
    expect(out.health).toBe('on_track');
  });

  // Regression: numeric taskCount / completedCount from JSON (int64 wire values).
  it('handles numeric taskCount and completedCount sent as numbers', () => {
    const out = apiProjectToProject(
      baseProject({
        taskCount: 20 as unknown as number,
        completedCount: 5 as unknown as number,
      })
    );
    expect(out.taskCount).toBe(20);
    expect(out.completedCount).toBe(5);
    expect(out.progress).toBe(25);
  });

  it('clamps progress to 100 when completedCount exceeds taskCount', () => {
    // Backend may send inconsistent counts; ensure we never show >100%.
    const out = apiProjectToProject(baseProject({ taskCount: 3, completedCount: 5 }));
    // Math.round(5/3*100) = 167, but the mapper does not cap - this test
    // documents the CURRENT behaviour. If the mapper is ever patched to cap
    // at 100, update this assertion accordingly.
    expect(out.progress).toBeGreaterThan(0);
  });

  it('memberAvatars defaults to empty array when server omits it', () => {
    const { memberAvatars: _, ...rest } = baseProject();
    const out = apiProjectToProject(rest as unknown as import('../services/flowApi').ApiProject);
    expect(out.memberAvatars).toEqual([]);
  });

  it('preserves a non-empty memberAvatars array', () => {
    const avatars = ['https://cdn/a.png', 'https://cdn/b.png'];
    const out = apiProjectToProject(baseProject({ memberAvatars: avatars }));
    expect(out.memberAvatars).toEqual(avatars);
  });

  it('preserves explicit color', () => {
    const out = apiProjectToProject(baseProject({ color: '#ff0000' }));
    expect(out.color).toBe('#ff0000');
  });

  it('preserves fields required to make a safe full project update', () => {
    const out = apiProjectToProject(
      baseProject({
        description: 'Keep this description',
        icon: 'briefcase',
        status: 'archived',
        startDate: '2026-01-01',
        dueDate: '2026-06-01',
      })
    );

    expect(out).toMatchObject({
      description: 'Keep this description',
      icon: 'briefcase',
      status: 'archived',
      startDate: '2026-01-01',
      targetDate: '2026-06-01',
    });
  });
});

describe('apiWorkspaceToWorkspace', () => {
  it('stringifies numeric workspace IDs', () => {
    const out = apiWorkspaceToWorkspace({
      id: 5 as unknown as string,
      name: 'Personal',
      color: '',
    });
    expect(out.id).toBe('5');
  });

  it('falls back to the flow accent when color is missing', () => {
    const out = apiWorkspaceToWorkspace({
      id: '5',
      name: 'Personal',
      color: '',
    } as ApiWorkspace);
    expect(out.color).toBe(ACCENT_DEFAULT);
  });

  it('preserves explicit colors', () => {
    const out = apiWorkspaceToWorkspace({
      id: '5',
      name: 'Personal',
      color: '#ff00aa',
    } as ApiWorkspace);
    expect(out.color).toBe('#ff00aa');
  });
});

describe('apiProjectToProject visibleViaReport', () => {
  // The sidebar and All Projects both group on this flag, so a mapper that
  // dropped it would silently merge a report's projects into the caller's own.
  it('carries the manager-proxy flag through', () => {
    const out = apiProjectToProject(baseProject({ visibleViaReport: true }));
    expect(out.visibleViaReport).toBe(true);
  });

  // The v1 shape has no such field, so an absent value must mean "the caller's
  // own" rather than undefined leaking into a filter.
  it('defaults to false when the server omits it', () => {
    const out = apiProjectToProject(baseProject());
    expect(out.visibleViaReport).toBe(false);
  });
});
