import { describe, it, expect } from 'vitest';
import {
  filterByChannel,
  filterByCategory,
  collectCategories,
  sortByName,
  sortByRecency,
  interpolate,
} from './templateUtils';
import type { Template } from '../schemas/templates';

const makeTemplate = (overrides: Partial<Template> = {}): Template => ({
  id: 1,
  name: 'Welcome Email',
  channel: 'email',
  category: 'onboarding',
  body: 'Hello {{name}}!',
  updatedAt: '2024-05-01',
  ...overrides,
});

const templates: Template[] = [
  makeTemplate({
    id: 1,
    name: 'Welcome Email',
    channel: 'email',
    category: 'onboarding',
    updatedAt: '2024-05-10',
  }),
  makeTemplate({
    id: 2,
    name: 'Birthday SMS',
    channel: 'sms',
    category: 'birthday',
    updatedAt: '2024-04-15',
  }),
  makeTemplate({
    id: 3,
    name: 'Promo Email',
    channel: 'email',
    category: 'promotion',
    updatedAt: '2024-05-20',
  }),
];

describe('filterByChannel', () => {
  it('returns only email templates', () => {
    expect(filterByChannel(templates, 'email')).toHaveLength(2);
  });

  it('returns only sms templates', () => {
    expect(filterByChannel(templates, 'sms')).toHaveLength(1);
  });
});

describe('filterByCategory', () => {
  it('returns templates matching the category', () => {
    expect(filterByCategory(templates, 'onboarding')).toHaveLength(1);
  });

  it('is case-insensitive', () => {
    expect(filterByCategory(templates, 'ONBOARDING')).toHaveLength(1);
  });
});

describe('collectCategories', () => {
  it('returns unique categories sorted alphabetically', () => {
    expect(collectCategories(templates)).toEqual(['birthday', 'onboarding', 'promotion']);
  });
});

describe('sortByName', () => {
  it('sorts templates alphabetically by name', () => {
    const sorted = sortByName(templates);
    expect(sorted[0].name).toBe('Birthday SMS');
  });
});

describe('sortByRecency', () => {
  it('sorts templates with newest updatedAt first', () => {
    const sorted = sortByRecency(templates);
    expect(sorted[0].updatedAt).toBe('2024-05-20');
  });
});

describe('interpolate', () => {
  it('replaces known placeholders', () => {
    expect(interpolate('Hello {{name}}!', { name: 'Alice' })).toBe('Hello Alice!');
  });

  it('leaves unknown placeholders unchanged', () => {
    expect(interpolate('Hello {{name}}!', {})).toBe('Hello {{name}}!');
  });
});
