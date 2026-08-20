import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const backSpy = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useRouter: () => ({ history: { back: backSpy } }),
}));

import NoMatch from './NoMatch';

describe('NoMatch', () => {
  it('shows the shared EmptyState and goes back on the CTA', () => {
    render(<NoMatch />);

    expect(screen.getByText('title')).toBeInTheDocument();
    expect(screen.getByText('description')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'goBack' }));

    expect(backSpy).toHaveBeenCalledTimes(1);
  });
});
