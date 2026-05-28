// Pure helpers for modifier group validation and selection logic.

import type { Modifier, ModifierGroup } from '../types/pos';

/** Return true if the selection count satisfies the group's min and max constraints. */
export function isValidSelection(group: ModifierGroup, selectedCount: number): boolean {
  return selectedCount >= group.min && selectedCount <= group.max;
}

/** Return true if a group requires at least one selection. */
export function isRequired(group: ModifierGroup): boolean {
  return group.min > 0;
}

/** Return true if a group allows multiple selections. */
export function isMultiSelect(group: ModifierGroup): boolean {
  return group.max > 1;
}

/** Return the default-selected modifiers for a group. */
export function defaultSelections(group: ModifierGroup): Modifier[] {
  return group.modifiers.filter((m) => m.defaultSelected);
}

/** Return the total price delta for a set of selected modifiers. */
export function selectedPriceDelta(selected: Modifier[]): number {
  return selected.reduce((sum, m) => sum + (Number.isFinite(m.priceDelta) ? m.priceDelta : 0), 0);
}

/** Return the modifier with the highest price delta in a group. */
export function mostExpensiveModifier(group: ModifierGroup): Modifier | undefined {
  if (!group.modifiers.length) return undefined;
  return group.modifiers.reduce((best, m) => (m.priceDelta > best.priceDelta ? m : best));
}

/** Return modifiers from a group sorted by name alphabetically. */
export function sortModifiersByName(group: ModifierGroup): Modifier[] {
  return [...group.modifiers].sort((a, b) => a.name.localeCompare(b.name));
}

/** Return a summary string for the selected modifiers (comma-separated names). */
export function modifierSummary(selected: Modifier[]): string {
  return selected.map((m) => m.name).join(', ');
}

/** Check whether a specific modifier id is in the selected list. */
export function isModifierSelected(selected: Modifier[], modifierId: string): boolean {
  return selected.some((m) => m.id === modifierId);
}
