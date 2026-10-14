import { describe, it, expect } from 'vitest';
import {
  menuHomePath,
  menuLocationPath,
  menuCheckoutPath,
  menuOrderStatusPath,
  menuByBusinessIdPath,
} from './menuPaths';

describe('menuPaths', () => {
  it('builds the slug location picker', () => {
    expect(menuHomePath('acme')).toBe('/acme/menu');
  });

  it('builds a location menu path', () => {
    expect(menuLocationPath('acme', 'downtown')).toBe('/acme/menu/downtown');
  });

  it('builds the slug-free checkout path', () => {
    expect(menuCheckoutPath()).toBe('/menu/checkout');
  });

  it('builds the slug-free order status path', () => {
    expect(menuOrderStatusPath('999')).toBe('/menu/order-status/999');
  });

  it('builds the business-id menu path', () => {
    expect(menuByBusinessIdPath(69)).toBe('/b/69/menu');
  });
});
