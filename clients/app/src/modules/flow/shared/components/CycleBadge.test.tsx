// @vitest-environment happy-dom
import React from 'react';
import { beforeAll, describe, expect, it } from 'vitest';
import { render, screen } from '@testing-library/react';

import CycleBadge from './CycleBadge';
import { initTestI18n } from '../testing/i18n';
import type { Cycle } from '../stores/flowStore';

const cycle = (status: Cycle['status']): Cycle =>
  ({
    id: '1',
    name: 'Sprint 12',
    status,
    startDate: '2026-11-01',
    endDate: '2026-11-14',
    scope: 4,
    completed: 1,
  }) as Cycle;

beforeAll(async () => {
  await initTestI18n();
});

// The cycles list prints the name beside the badge, and the badge repeated it.
describe('CycleBadge', () => {
  it.each([
    ['active', 'Active'],
    ['upcoming', 'Upcoming'],
    ['completed', 'Completed'],
  ] as const)('shows the %s status rather than the name', (status, label) => {
    render(<CycleBadge cycle={cycle(status)} compact />);
    expect(screen.getByText(label)).toBeInTheDocument();
    expect(screen.queryByText('Sprint 12')).not.toBeInTheDocument();
  });
});
