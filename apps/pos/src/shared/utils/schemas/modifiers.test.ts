import { describe, it, expect } from 'vitest';
import { modifierSchema, modifierGroupSchema } from './modifiers';

describe('modifierSchema', () => {
  const base = { id: '1', name: 'Extra Cheese', priceDelta: 1.5 };

  it('parses a valid modifier', () => {
    const result = modifierSchema.parse(base);
    expect(result.id).toBe('1');
    expect(result.name).toBe('Extra Cheese');
    expect(result.priceDelta).toBe(1.5);
  });

  it('coerces a numeric id to string', () => {
    const result = modifierSchema.parse({ ...base, id: 7 });
    expect(result.id).toBe('7');
  });

  it('defaults name to "Option" when missing', () => {
    const result = modifierSchema.parse({ ...base, name: undefined });
    expect(result.name).toBe('Option');
  });

  it('defaults priceDelta to 0 when missing', () => {
    const result = modifierSchema.parse({ ...base, priceDelta: undefined });
    expect(result.priceDelta).toBe(0);
  });

  it('defaults required to false when missing', () => {
    const result = modifierSchema.parse(base);
    expect(result.required).toBe(false);
  });

  it('defaults defaultSelected to false when missing', () => {
    const result = modifierSchema.parse(base);
    expect(result.defaultSelected).toBe(false);
  });

  it('accepts a group field', () => {
    const result = modifierSchema.parse({ ...base, group: 'Cheese Options' });
    expect(result.group).toBe('Cheese Options');
  });
});

describe('modifierGroupSchema', () => {
  const base = { id: '10', name: 'Toppings' };

  it('parses a minimal valid modifier group', () => {
    const result = modifierGroupSchema.parse(base);
    expect(result.id).toBe('10');
    expect(result.name).toBe('Toppings');
  });

  it('defaults name to "Modifiers" when missing', () => {
    const result = modifierGroupSchema.parse({ ...base, name: undefined });
    expect(result.name).toBe('Modifiers');
  });

  it('defaults min to 0', () => {
    const result = modifierGroupSchema.parse(base);
    expect(result.min).toBe(0);
  });

  it('defaults max to 1', () => {
    const result = modifierGroupSchema.parse(base);
    expect(result.max).toBe(1);
  });

  it('defaults modifiers to an empty array', () => {
    const result = modifierGroupSchema.parse(base);
    expect(result.modifiers).toEqual([]);
  });

  it('parses nested modifiers', () => {
    const result = modifierGroupSchema.parse({
      ...base,
      modifiers: [{ id: '1', name: 'Extra Cheese', priceDelta: 1.5 }],
    });
    expect(result.modifiers).toHaveLength(1);
    expect(result.modifiers[0].name).toBe('Extra Cheese');
  });
});
