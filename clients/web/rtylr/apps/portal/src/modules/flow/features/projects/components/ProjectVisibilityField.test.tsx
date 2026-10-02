import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

const projectState = vi.hoisted(() => ({
  data: { visibility: 'private', isPersonal: false } as Record<string, unknown> | undefined,
  isLoading: false,
}));
vi.mock('../../projects-v2/useAllProjects', () => ({ useProject: () => projectState }));

const setVisibilityMutate = vi.hoisted(() => vi.fn());
vi.mock('../../projects-v2/useVisibility', () => ({
  useSetProjectVisibility: () => ({ mutate: setVisibilityMutate, isPending: false }),
}));

import ProjectVisibilityField from './ProjectVisibilityField';

beforeEach(() => {
  setVisibilityMutate.mockClear();
  projectState.data = { visibility: 'private', isPersonal: false };
  projectState.isLoading = false;
});

describe('ProjectVisibilityField (#1088)', () => {
  it('is a real radiogroup, not divs wearing a role', () => {
    // The native inputs stay for their semantics and arrow-key behaviour;
    // only the browser's blue fill and OS focus ring go.
    render(<ProjectVisibilityField projectId="p1" canAdminister />);
    const radios = screen.getAllByRole('radio');
    expect(radios).toHaveLength(2);
    expect(radios.every((r) => r.tagName === 'INPUT')).toBe(true);
  });

  it('marks the current visibility as checked', () => {
    render(<ProjectVisibilityField projectId="p1" canAdminister />);
    expect(screen.getByRole('radio', { name: /private/ })).toBeChecked();
    expect(screen.getByRole('radio', { name: /public/ })).not.toBeChecked();
  });

  it('does not carry the selection by colour alone', () => {
    // WCAG 1.4.1: the selected card also gets a tick and bold text.
    const { container } = render(<ProjectVisibilityField projectId="p1" canAdminister />);
    const labels = [...container.querySelectorAll('label')];
    const selected = labels.find((l) => l.querySelector('input')?.checked === true);
    expect(selected?.querySelector('svg.ms-auto')).not.toBeNull();
    expect(selected?.querySelector('.font-semibold')).not.toBeNull();
  });

  it('paints the selected card from the token set', () => {
    const { container } = render(<ProjectVisibilityField projectId="p1" canAdminister />);
    const labels = [...container.querySelectorAll('label')];
    const selected = labels.find((l) => l.querySelector('input')?.checked === true) as HTMLElement;
    // Read off the authored attribute: happy-dom drops the var() from the
    // `border` shorthand when it reflows it into style.border.
    const authored = selected.getAttribute('style') ?? '';
    expect(authored).toContain('--accent-muted');
    expect(authored).toContain('--accent-border');
    // And nothing from outside the token set, which the browser's own blue
    // radio fill was.
    expect(authored).not.toMatch(/#[0-9a-f]{3,6}/i);
  });

  it('sends the change when the other option is chosen', () => {
    render(<ProjectVisibilityField projectId="p1" canAdminister />);
    fireEvent.click(screen.getByRole('radio', { name: /public/ }));
    expect(setVisibilityMutate).toHaveBeenCalledWith({ id: 'p1', visibility: 'public' });
  });

  it('sends nothing when the current option is chosen again', () => {
    render(<ProjectVisibilityField projectId="p1" canAdminister />);
    fireEvent.click(screen.getByRole('radio', { name: /private/ }));
    expect(setVisibilityMutate).not.toHaveBeenCalled();
  });

  it('explains itself and disables when the caller is not a project admin', () => {
    render(<ProjectVisibilityField projectId="p1" canAdminister={false} />);
    expect(screen.getByText('flow.projectSettings.visibilityGateReason')).toBeInTheDocument();
    for (const radio of screen.getAllByRole('radio')) expect(radio).toBeDisabled();
  });

  it('explains itself on the personal project, which the server refuses to flip', () => {
    projectState.data = { visibility: 'private', isPersonal: true };
    render(<ProjectVisibilityField projectId="p1" canAdminister />);
    expect(screen.getByText('flow.projectSettings.visibilityPersonalGate')).toBeInTheDocument();
  });

  it('preselects neither option while the project is still loading', () => {
    projectState.data = undefined;
    projectState.isLoading = true;
    render(<ProjectVisibilityField projectId="p1" canAdminister />);
    for (const radio of screen.getAllByRole('radio')) expect(radio).not.toBeChecked();
  });
});
