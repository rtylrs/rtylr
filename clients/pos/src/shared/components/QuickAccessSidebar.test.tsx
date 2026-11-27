import React from 'react';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import QuickAccessSidebar from './QuickAccessSidebar';
import { useSettingsStore } from '../stores/settingsStore';

describe('QuickAccessSidebar', () => {
  beforeEach(() => {
    useSettingsStore.setState({
      quickAccessSidebar: true,
      businessProfile: 'restaurant',
      features: {
        ...useSettingsStore.getState().features,
        tables: true,
        kitchenDisplay: true,
        inventory: true,
        customers: true,
        giftCards: true,
      },
    });
  });

  afterEach(() => {
    cleanup();
  });

  it('orders navigation into Sales, Catalog, Guests, and Shift groups', () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <QuickAccessSidebar />
      </MemoryRouter>
    );

    const groups = screen.getAllByRole('group');
    expect(groups.map((g) => g.getAttribute('aria-label'))).toEqual([
      'Sales',
      'Catalog',
      'Guests',
      'Shift',
    ]);

    const links = screen.getAllByRole('link').map((a) => a.textContent?.trim());
    expect(links).toEqual([
      'Terminal',
      'Orders',
      'Held',
      'Tables',
      'KDS',
      'Menu',
      'Inventory',
      'Customers',
      'Gift Cards',
      'Receipts',
      'Day',
      'Offline',
      'Settings',
    ]);
  });

  it('omits empty groups when their features are disabled and uses Products label in retail profile', () => {
    useSettingsStore.setState({
      businessProfile: 'retail',
      features: {
        ...useSettingsStore.getState().features,
        tables: false,
        kitchenDisplay: false,
        customers: false,
        giftCards: false,
      },
    });

    render(
      <MemoryRouter initialEntries={['/']}>
        <QuickAccessSidebar />
      </MemoryRouter>
    );

    const groups = screen.getAllByRole('group');
    expect(groups.map((g) => g.getAttribute('aria-label'))).toEqual(['Sales', 'Catalog', 'Shift']);
    expect(screen.getByRole('link', { name: 'Products' })).toBeTruthy();
  });
});
