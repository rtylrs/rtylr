import { describe, it, expect, beforeEach } from 'vitest';
import { useBusinessStore, type BusinessType } from './businessStore';

const { getState } = useBusinessStore;

beforeEach(() => {
  useBusinessStore.setState({
    businessType: null,
    businessPreferences: {},
  });
});

describe('businessStore', () => {
  it('has correct initial state', () => {
    expect(getState().businessType).toBeNull();
    expect(getState().businessPreferences).toEqual({});
  });

  it('setBusinessType sets restaurant', () => {
    getState().setBusinessType('restaurant');
    expect(getState().businessType).toBe('restaurant');
  });

  it('setBusinessType sets gym', () => {
    getState().setBusinessType('gym');
    expect(getState().businessType).toBe('gym');
  });

  it('setBusinessType changes type', () => {
    getState().setBusinessType('restaurant');
    getState().setBusinessType('cafe');
    expect(getState().businessType).toBe('cafe');
  });

  it('setBusinessPreferences sets preferences', () => {
    const prefs = { currency: 'USD', timezone: 'UTC' };
    getState().setBusinessPreferences(prefs);
    expect(getState().businessPreferences).toEqual(prefs);
  });

  it('setBusinessPreferences replaces existing', () => {
    getState().setBusinessPreferences({ a: 1 });
    getState().setBusinessPreferences({ b: 2 });
    expect(getState().businessPreferences).toEqual({ b: 2 });
  });

  it('updateBusinessPreference adds single key', () => {
    getState().updateBusinessPreference('theme', 'dark');
    expect(getState().businessPreferences).toEqual({ theme: 'dark' });
  });

  it('updateBusinessPreference preserves other keys', () => {
    getState().setBusinessPreferences({ a: 1, b: 2 });
    getState().updateBusinessPreference('c', 3);
    expect(getState().businessPreferences).toEqual({ a: 1, b: 2, c: 3 });
  });

  it('updateBusinessPreference overwrites existing key', () => {
    getState().setBusinessPreferences({ theme: 'light' });
    getState().updateBusinessPreference('theme', 'dark');
    expect(getState().businessPreferences.theme).toBe('dark');
  });

  it('setBusinessType does not affect preferences', () => {
    getState().setBusinessPreferences({ key: 'value' });
    getState().setBusinessType('hotel');
    expect(getState().businessPreferences).toEqual({ key: 'value' });
  });

  it('handles all business types', () => {
    const types: BusinessType[] = [
      'restaurant',
      'gym',
      'retail',
      'salon',
      'cafe',
      'bar',
      'hotel',
      'other',
    ];
    for (const type of types) {
      getState().setBusinessType(type);
      expect(getState().businessType).toBe(type);
    }
  });
});
