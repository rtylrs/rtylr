import { describe, it, expect } from 'vitest';
import { colSpanForSize, rowSpanForSize, visibleWidgets, toggleWidgetVisibility, setAllVisible, reorderWidget, countByVariant, sizeLabel, type WidgetConfig } from './widgetUtils';

const w = (overrides: Partial<WidgetConfig> = {}): WidgetConfig => ({ id: 'w1', title: 'W', size: 'md', variant: 'stat', visible: true, order: 1, ...overrides });

describe('colSpanForSize', () => {
  it('sm=3', () => expect(colSpanForSize('sm')).toBe(3));
  it('xl=12', () => expect(colSpanForSize('xl')).toBe(12));
});

describe('rowSpanForSize', () => {
  it('sm=1', () => expect(rowSpanForSize('sm')).toBe(1));
  it('xl=4', () => expect(rowSpanForSize('xl')).toBe(4));
});

describe('visibleWidgets', () => {
  it('returns visible sorted', () => {
    const widgets = [w({ id: 'a', visible: true, order: 2 }), w({ id: 'b', visible: false, order: 1 }), w({ id: 'c', visible: true, order: 1 })];
    const r = visibleWidgets(widgets);
    expect(r).toHaveLength(2);
    expect(r[0].id).toBe('c');
  });
});

describe('toggleWidgetVisibility', () => {
  it('toggles visible to false', () => expect(toggleWidgetVisibility([w({ id: 'w1', visible: true })], 'w1')[0].visible).toBe(false));
  it('toggles false to visible', () => expect(toggleWidgetVisibility([w({ id: 'w1', visible: false })], 'w1')[0].visible).toBe(true));
});

describe('setAllVisible', () => {
  it('hides all', () => {
    const result = setAllVisible([w({ visible: true }), w({ id: 'w2', visible: true })], false);
    expect(result.every(x => !x.visible)).toBe(true);
  });
});

describe('reorderWidget', () => {
  it('moves widget to position', () => {
    const widgets = [w({ id: 'a', order: 1 }), w({ id: 'b', order: 2 }), w({ id: 'c', order: 3 })];
    expect(reorderWidget(widgets, 'c', 1)[0].id).toBe('c');
  });
});

describe('countByVariant', () => {
  it('counts by variant', () => {
    const widgets = [w({ variant: 'stat', visible: true }), w({ id: 'w2', variant: 'stat', visible: true }), w({ id: 'w3', variant: 'chart', visible: false })];
    expect(countByVariant(widgets).stat).toBe(2);
  });
});

describe('sizeLabel', () => {
  it('Small for sm', () => expect(sizeLabel('sm')).toBe('Small'));
  it('Extra Large for xl', () => expect(sizeLabel('xl')).toBe('Extra Large'));
});
