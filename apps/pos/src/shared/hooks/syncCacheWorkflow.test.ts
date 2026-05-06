import { describe, expect, it } from 'vitest';
import {
  normalizeCachedCombos,
  normalizeCachedModifiers,
  normalizeCachedProducts,
} from './syncCacheWorkflow';

describe('syncCacheWorkflow', () => {
  it('normalizes product cache rows and drops records without valid IDs', () => {
    expect(
      normalizeCachedProducts(
        [
          {
            id: '10',
            name: ' Burger ',
            sku: 'BRG',
            category_id: '3',
            category: 'Mains',
            salePrice: '12.50',
            cost_price: '5',
            taxClass: '',
            imageUrl: 'burger.png',
          },
          { id: 'draft', name: 'Invalid' },
        ],
        123,
      ),
    ).toEqual([
      {
        id: 10,
        name: 'Burger',
        sku: 'BRG',
        barcode: undefined,
        categoryId: 3,
        categoryName: 'Mains',
        price: 12.5,
        costPrice: 5,
        taxClass: 'standard',
        imageUrl: 'burger.png',
        available: true,
        syncedAt: 123,
      },
    ]);
  });

  it('normalizes modifier cache rows from flat and grouped API shapes', () => {
    expect(
      normalizeCachedModifiers(
        [
          { id: '1', name: 'Cheese', group_name: 'Toppings', price_delta: '1.5' },
          {
            id: 'group-1',
            name: 'Sauces',
            modifiers: [{ id: 2, name: 'Aioli', priceDelta: 0.75, required: true }],
          },
        ],
        456,
      ),
    ).toEqual([
      {
        id: 1,
        name: 'Cheese',
        groupName: 'Toppings',
        priceDelta: 1.5,
        isRequired: false,
        syncedAt: 456,
      },
      {
        id: 2,
        name: 'Aioli',
        groupName: 'Sauces',
        priceDelta: 0.75,
        isRequired: true,
        syncedAt: 456,
      },
    ]);
  });

  it('normalizes combo cache rows and filters invalid combo item IDs', () => {
    expect(
      normalizeCachedCombos(
        [
          {
            id: '8',
            name: 'Lunch',
            description: null,
            price: '20',
            items: [
              { productId: '10', name: 'Burger', quantity: '2' },
              { itemId: 'bad', name: 'Draft', qty: 1 },
            ],
          },
        ],
        789,
      ),
    ).toEqual([
      {
        id: 8,
        name: 'Lunch',
        description: '',
        price: 20,
        items: [{ productId: 10, name: 'Burger', quantity: 2 }],
        isActive: true,
        syncedAt: 789,
      },
    ]);
  });
});
