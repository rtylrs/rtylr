// @vitest-environment happy-dom
import React from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';

import FlowDatePicker from './FlowDatePicker';
import { useTimeSettingsStore } from '../stores/timeSettingsStore';

const firstWeekday = (): string | null =>
  screen.getByRole('dialog').querySelector('.grid > span')?.textContent ?? null;

describe('FlowDatePicker', () => {
  afterEach(() => useTimeSettingsStore.getState().setWeekStart('monday'));

  it.each([
    ['monday', 'Mon'],
    ['sunday', 'Sun'],
  ] as const)('starts the week on %s when that is the time setting', (weekStart, label) => {
    useTimeSettingsStore.getState().setWeekStart(weekStart);
    render(<FlowDatePicker value="2026-08-16" onChange={vi.fn()} locale="en-US" />);
    fireEvent.click(screen.getByRole('button', { name: /aug 16, 2026/i }));
    expect(firstWeekday()).toBe(label);
  });
});
