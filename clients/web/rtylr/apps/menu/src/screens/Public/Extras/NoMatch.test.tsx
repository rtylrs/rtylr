import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeAll } from 'vitest';

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves,
// which is how this suite passed while every string on screen was a dotted
// path (rtylr#1269). The real locale files mean the copy below also proves the
// lookup.

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const backSpy = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useRouter: () => ({ history: { back: backSpy } }),
}));

import NoMatch from './NoMatch';
import { initTestI18n } from '../../../shared/testing/i18n';

beforeAll(async () => {
  await initTestI18n('en');
});

describe('NoMatch', () => {
  it('shows the shared EmptyState and goes back on the CTA', () => {
    render(<NoMatch />);

    expect(screen.getByText('Page not found')).toBeInTheDocument();
    expect(
      screen.getByText("The page you're looking for doesn't exist or may have moved.")
    ).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Go back' }));

    expect(backSpy).toHaveBeenCalledTimes(1);
  });
});
