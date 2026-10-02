import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const refetchFeedback = vi.fn();
const feedbackList: unknown[] = [];

vi.mock('../../shared/hooks/queries/useFeedback', () => ({
  useListFeedback: () => ({
    data: feedbackList,
    isLoading: false,
    refetch: refetchFeedback,
  }),
}));

import Feedback from './Feedback';

function renderFeedback(): void {
  render(<Feedback />);
}

beforeEach(() => {
  refetchFeedback.mockClear();
});

describe('Feedback empty state', () => {
  it('shows a "refresh" call-to-action when there is no feedback yet', () => {
    renderFeedback();

    expect(screen.getByText('crm.feedback.empty')).toBeInTheDocument();
    expect(screen.getByText('crm.feedback.emptyDescription')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'crm.actions.refresh' }));

    expect(refetchFeedback).toHaveBeenCalledTimes(1);
  });
});
