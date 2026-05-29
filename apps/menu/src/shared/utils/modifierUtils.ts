/**
 * Menu item modifier helpers for the Menu app.
 * Modifiers are add-ons or customisations attached to a cart item.
 */

export interface Modifier {
  id: number;
  name: string;
  priceDelta: number;
}

export interface ModifierGroup {
  id: number;
  name: string;
  required: boolean;
  minSelections: number;
  maxSelections: number;
  modifiers: Modifier[];
}

/** Return the total extra cost from a list of selected modifiers. */
export function modifierTotal(selected: Modifier[]): number {
  return parseFloat(selected.reduce((sum, m) => sum + m.priceDelta, 0).toFixed(2));
}

/** Return true if all required groups have at least one selection. */
export function isModifierSelectionComplete(
  groups: ModifierGroup[],
  selectedIds: Set<number>
): boolean {
  return groups.every((group) => {
    if (!group.required) return true;
    const count = group.modifiers.filter((m) => selectedIds.has(m.id)).length;
    return count >= group.minSelections;
  });
}

/** Return the selected modifiers that belong to a given group. */
export function getGroupSelections(group: ModifierGroup, selectedIds: Set<number>): Modifier[] {
  return group.modifiers.filter((m) => selectedIds.has(m.id));
}

/** Return true if an additional modifier can be selected given the group's max. */
export function canSelectMore(group: ModifierGroup, selectedIds: Set<number>): boolean {
  const current = getGroupSelections(group, selectedIds).length;
  return current < group.maxSelections;
}

/** Return a display string listing the selected modifier names. */
export function formatModifierNames(selected: Modifier[]): string {
  if (selected.length === 0) return '';
  return selected.map((m) => m.name).join(', ');
}
