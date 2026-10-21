import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const refetchFeedback = vi.fn();
const feedbackList: unknown[] = [];
const feedbackState = { isError: false };

vi.mock('../../shared/hooks/queries/useFeedback', () => ({
  useListFeedback: () => ({
    data: feedbackList,
    isLoading: false,
    isError: feedbackState.isError,
    refetch: refetchFeedback,
  }),
}));

import Feedback from './Feedback';

function renderFeedback(): void {
  render(<Feedback />);
}

beforeEach(() => {
  refetchFeedback.mockClear();
  feedbackState.isError = false;
});

describe('Feedback empty state', () => {
  it('shows a "refresh" call-to-action when there is no feedback yet', () => {
    renderFeedback();

    expect(screen.getByText('crm.feedback.empty')).toBeInTheDocument();
    expect(screen.getByText('crm.feedback.emptyDescription')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'crm.actions.refresh' }));

    expect(refetchFeedback).toHaveBeenCalledTimes(1);
  });

  it('shows a load error instead of treating a failed request as an empty list', () => {
    feedbackState.isError = true;

    renderFeedback();

    expect(screen.getByRole('alert')).toHaveTextContent('crmErrors.generic');
    expect(screen.queryByText('crm.feedback.empty')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'crm.actions.refresh' }));
    expect(refetchFeedback).toHaveBeenCalledTimes(1);
  });
});
