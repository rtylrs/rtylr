/**
 * Dashboard widget layout and visibility helpers.
 * All functions are pure with no side-effects.
 */

export type WidgetSize = 'sm' | 'md' | 'lg' | 'xl';
export type WidgetVariant = 'stat' | 'chart' | 'table' | 'list' | 'map';

export interface WidgetConfig {
  id: string;
  title: string;
  size: WidgetSize;
  variant: WidgetVariant;
  visible: boolean;
  order: number;
}

export function colSpanForSize(size: WidgetSize): number {
  switch (size) {
    case 'sm': return 3;
    case 'md': return 6;
    case 'lg': return 9;
    case 'xl': return 12;
  }
}

export function rowSpanForSize(size: WidgetSize): number {
  switch (size) {
    case 'sm': return 1;
    case 'md': return 2;
    case 'lg': return 3;
    case 'xl': return 4;
  }
}

export function visibleWidgets(widgets: WidgetConfig[]): WidgetConfig[] {
  return widgets.filter(w => w.visible).sort((a, b) => a.order - b.order);
}

export function toggleWidgetVisibility(widgets: WidgetConfig[], id: string): WidgetConfig[] {
  return widgets.map(w => w.id === id ? { ...w, visible: !w.visible } : w);
}

export function setAllVisible(widgets: WidgetConfig[], visible: boolean): WidgetConfig[] {
  return widgets.map(w => ({ ...w, visible }));
}

export function reorderWidget(widgets: WidgetConfig[], id: string, newOrder: number): WidgetConfig[] {
  const target = widgets.find(w => w.id === id);
  if (!target) return widgets;
  const others = widgets.filter(w => w.id !== id).sort((a, b) => a.order - b.order);
  const clamped = Math.max(1, Math.min(newOrder, widgets.length));
  others.splice(clamped - 1, 0, { ...target, order: clamped });
  return others.map((w, i) => ({ ...w, order: i + 1 }));
}

export function countByVariant(widgets: WidgetConfig[]): Record<WidgetVariant, number> {
  const counts: Record<WidgetVariant, number> = { stat: 0, chart: 0, table: 0, list: 0, map: 0 };
  for (const w of widgets) { if (w.visible) counts[w.variant]++; }
  return counts;
}

export function sizeLabel(size: WidgetSize): string {
  switch (size) {
    case 'sm': return 'Small';
    case 'md': return 'Medium';
    case 'lg': return 'Large';
    case 'xl': return 'Extra Large';
  }
}
