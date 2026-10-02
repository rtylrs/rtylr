import React from 'react';
import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  listReviewCycles: vi.fn(),
  createReviewCycle: vi.fn(),
  openReviewCycle: vi.fn(),
  listGoals: vi.fn(),
  submitReview: vi.fn(),
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => ({
    listReviewCycles: mocks.listReviewCycles,
    createReviewCycle: mocks.createReviewCycle,
    openReviewCycle: mocks.openReviewCycle,
    listGoals: mocks.listGoals,
    submitReview: mocks.submitReview,
  }),
  unwrap: (value: unknown) => value,
}));
vi.mock('../../lib/clients', () => ({
  hrTransport: { get: mocks.get, post: mocks.post, put: mocks.put },
}));
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));
vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

import {
  useCreateGoal,
  useCreateReviewCycle,
  useListReviewCycles,
  useListReviews,
  useSubmitReview,
} from './usePerformance';

const wrapper = ({ children }: { children: React.ReactNode }): React.JSX.Element => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return React.createElement(QueryClientProvider, { client }, children);
};

describe('performance contract hooks', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.listReviewCycles.mockResolvedValue({ cycles: [] });
    mocks.createReviewCycle.mockResolvedValue({ cycle: { id: '7' } });
    mocks.listGoals.mockResolvedValue({ goals: [] });
    mocks.submitReview.mockResolvedValue({ review: { id: '7' } });
    mocks.post.mockResolvedValue({ goal: { id: '8' } });
    mocks.get.mockResolvedValue({ reviews: [], total: 0, limit: 20, offset: 0 });
  });

  it('uses the versioned service for cycle list and creation', async () => {
    const list = renderHook(() => useListReviewCycles(), { wrapper });
    await waitFor(() => expect(list.result.current.isSuccess).toBe(true));
    expect(mocks.listReviewCycles).toHaveBeenCalledWith({});

    const create = renderHook(() => useCreateReviewCycle(), { wrapper });
    await act(async () => {
      await create.result.current.mutateAsync({
        name: 'Q4',
        type: 'quarterly',
        periodStart: '2026-10-01',
        periodEnd: '2026-12-31',
      });
    });
    expect(mocks.createReviewCycle).toHaveBeenCalledWith({
      name: 'Q4',
      type: 'quarterly',
      periodStart: '2026-10-01',
      periodEnd: '2026-12-31',
    });
    expect(mocks.get).not.toHaveBeenCalledWith('/hr/performance/cycles', expect.anything());
  });

  it('includes the employee identity when creating a goal', async () => {
    const { result } = renderHook(() => useCreateGoal(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({
        employeeId: 42,
        reviewCycleId: 7,
        title: 'Ship the release',
        description: 'Finish the release',
      });
    });

    expect(mocks.post).toHaveBeenCalledWith('/hr/performance/goals', {
      employeeId: 42,
      reviewCycleId: 7,
      title: 'Ship the release',
      description: 'Finish the release',
      weight: 0,
      successCriteria: undefined,
    });
  });

  it('maps cycle_id locally and filters status without sending an unsupported status', async () => {
    mocks.get.mockResolvedValue({
      reviews: [
        { id: '1', status: 'draft' },
        { id: '2', status: 'submitted' },
      ],
      total: 2,
      limit: 20,
      offset: 0,
    });

    const { result } = renderHook(() => useListReviews({ reviewCycleId: 7, status: 'submitted' }), {
      wrapper,
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mocks.get).toHaveBeenCalledWith('/hr/performance/reviews', {
      params: { cycleId: 7 },
    });
    expect(result.current.data?.reviews.map((review) => review.id)).toEqual(['2']);
  });

  it('submits through the generated service so a string id is not put in an int64 body', async () => {
    const { result } = renderHook(() => useSubmitReview(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({
        reviewId: '7',
        overallRating: '4',
        narrative: 'A detailed review narrative that is long enough.',
      });
    });

    expect(mocks.submitReview).toHaveBeenCalledWith({
      reviewId: '7',
      overallRating: '4',
      narrative: 'A detailed review narrative that is long enough.',
    });
    expect(mocks.post).not.toHaveBeenCalledWith(
      '/hr/performance/reviews/7/submit',
      expect.anything()
    );
  });
});
