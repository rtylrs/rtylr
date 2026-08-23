import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';

// t returns the key, so a label that goes through i18n is distinguishable from
// one that renders a raw contract value: 'flow.projectStatus.active' vs 'active'.
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
}));

vi.mock('./useAllProjects', () => ({
  useAllProjects: () => ({ data: { projects: [], total: 0 }, isLoading: false }),
}));

import AllProjects from './AllProjects';

describe('AllProjects filters', () => {
  // The status filter rendered the bare contract value while the visibility
  // filter beside it went through t(), so the two dropdowns disagreed about
  // whether they were translated.
  it('translates every status option instead of showing the contract value', () => {
    render(<AllProjects />);

    expect(screen.getByRole('option', { name: 'flow.projectStatus.active' })).toBeInTheDocument();
    expect(
      screen.getByRole('option', { name: 'flow.projectStatus.completed' })
    ).toBeInTheDocument();
    expect(screen.queryByRole('option', { name: 'active' })).not.toBeInTheDocument();
    expect(screen.queryByRole('option', { name: 'completed' })).not.toBeInTheDocument();
  });

  it('translates the visibility options too', () => {
    render(<AllProjects />);

    expect(screen.getByRole('option', { name: 'flow.visibility.private' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'flow.visibility.public' })).toBeInTheDocument();
  });
});
