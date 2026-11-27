import React from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import type { MenuItem } from '@voxire/voxie/pos/menu_item/v1/types';
import ItemTile from './ItemTile';

function makeItem(overrides: Partial<MenuItem> = {}): MenuItem {
  return {
    id: '1',
    ownerId: '1',
    categoryId: '1',
    name: 'Butter Croissant',
    sku: 'CRO-1',
    salePrice: 3.5,
    isActive: true,
    createdAt: '',
    updatedAt: '',
    imageUrl: '',
    category: 'Bakery',
    barcode: '',
    description: '',
    taxRate: 0,
    stock: 0,
    ...overrides,
  };
}

describe('ItemTile stock enforcement', () => {
  afterEach(() => {
    cleanup();
  });

  it('disables and marks tracked items with zero stock as Out of stock', () => {
    const onClick = vi.fn();
    render(
      <ItemTile
        item={makeItem({ trackInventory: true, allowBackorder: false, stock: 0 })}
        onClick={onClick}
      />
    );

    const button = screen.getByRole('button');
    expect(button.hasAttribute('disabled')).toBe(true);
    expect(screen.getByText('Out of stock')).toBeTruthy();

    fireEvent.click(button);
    expect(onClick).not.toHaveBeenCalled();
  });

  it('keeps untracked zero-stock items clickable without showing Out of stock', () => {
    const onClick = vi.fn();
    render(
      <ItemTile
        item={makeItem({ trackInventory: false, allowBackorder: false, stock: 0 })}
        onClick={onClick}
      />
    );

    const button = screen.getByRole('button');
    expect(button.hasAttribute('disabled')).toBe(false);
    expect(screen.queryByText('Out of stock')).toBeNull();

    fireEvent.click(button);
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it('keeps backorder-enabled tracked items clickable even at zero stock', () => {
    const onClick = vi.fn();
    render(
      <ItemTile
        item={makeItem({ trackInventory: true, allowBackorder: true, stock: 0 })}
        onClick={onClick}
      />
    );

    const button = screen.getByRole('button');
    expect(button.hasAttribute('disabled')).toBe(false);

    fireEvent.click(button);
    expect(onClick).toHaveBeenCalledTimes(1);
  });
});
