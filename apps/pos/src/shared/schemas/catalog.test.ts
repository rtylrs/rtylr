import { describe, expect, it } from 'vitest';
import { comboSchema } from './combos';
import { menuItemSchema } from './menuItems';
import { modifierGroupSchema } from './modifiers';

describe('catalog schemas', () => {
  it('coerces menu item numeric fields from API strings', () => {
    expect(
      menuItemSchema.parse({
        id: 1,
        name: 'Burger',
        price: '12.50',
        stock: '8',
        taxRate: '0.11',
      }),
    ).toMatchObject({
      id: '1',
      price: 12.5,
      stock: 8,
      taxRate: 0.11,
    });
  });

  it('maps common menu item API aliases', () => {
    expect(
      menuItemSchema.parse({
        id: 2,
        name: 'Fries',
        salePrice: '4.50',
        categoryName: 'Sides',
        imageUrl: 'fries.png',
        isActive: false,
      }),
    ).toMatchObject({
      id: '2',
      price: 4.5,
      category: 'Sides',
      image: 'fries.png',
      available: false,
    });
  });

  it('defaults invalid menu item numeric fields to safe values', () => {
    expect(menuItemSchema.parse({ id: 1, price: 'bad', stock: 'bad', taxRate: 'bad' })).toMatchObject(
      {
        price: 0,
        stock: undefined,
        taxRate: undefined,
      },
    );
  });

  it('coerces modifier and combo numeric fields from API strings', () => {
    expect(
      modifierGroupSchema.parse({
        id: 1,
        min: '1',
        max: '3',
        modifiers: [{ id: 2, priceDelta: '-0.5' }],
      }),
    ).toMatchObject({
      min: 1,
      max: 3,
      modifiers: [{ id: '2', priceDelta: -0.5 }],
    });

    expect(comboSchema.parse({ id: 4, price: '20', items: [{ itemId: 5, qty: '2' }] })).toMatchObject(
      {
        id: '4',
        price: 20,
        items: [{ itemId: '5', qty: 2 }],
      },
    );
  });
});
