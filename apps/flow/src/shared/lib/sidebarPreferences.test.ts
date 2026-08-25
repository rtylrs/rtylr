import { describe, expect, it } from 'vitest';
import {
  createDefaultSidebarPreferences,
  parseSidebarPreferences,
  recordRecentSidebarPage,
  serializeSidebarPreferences,
  updatePreferencesCopy,
} from './sidebarPreferences';

describe('sidebar preferences', () => {
  it('starts with the system groups and top navigation', () => {
    const preferences = createDefaultSidebarPreferences();

    expect(preferences.groupOrder).toEqual(['plan', 'team', 'time', 'projects']);
    expect(preferences.topLevelItemIds).toEqual([
      'builtin:inbox',
      'builtin:my-issues',
      'builtin:today',
      'builtin:projects-shortcut',
      'builtin:triage',
    ]);
  });

  it('keeps personal changes while adding new system items', () => {
    const current = createDefaultSidebarPreferences();
    current.items.find((item) => item.id === 'builtin:inbox')!.visible = false;
    const parsed = parseSidebarPreferences(serializeSidebarPreferences(current));

    expect(parsed.items.find((item) => item.id === 'builtin:inbox')?.visible).toBe(false);
    expect(parsed.items.some((item) => item.id === 'builtin:archive')).toBe(true);
  });

  it('limits recent pages to five and keeps the newest first', () => {
    let preferences = createDefaultSidebarPreferences();
    for (let index = 0; index < 6; index += 1) {
      preferences = recordRecentSidebarPage(preferences, `/page-${index}`, `Page ${index}`);
    }

    const recent = preferences.groups.find((group) => group.id === 'recent');
    expect(recent?.itemIds).toHaveLength(5);
    expect(recent?.itemIds[0]).toBe('recent:/page-5');
    expect(preferences.items.some((item) => item.id === 'recent:/page-0')).toBe(false);
  });

  it('does not mutate the current preference document while copying it', () => {
    const current = createDefaultSidebarPreferences();
    const copy = updatePreferencesCopy(current);
    copy.groups[0].itemIds.reverse();
    copy.items[0].visible = false;

    expect(current.groups[0].itemIds).not.toEqual(copy.groups[0].itemIds);
    expect(current.items[0].visible).toBe(true);
  });
});
