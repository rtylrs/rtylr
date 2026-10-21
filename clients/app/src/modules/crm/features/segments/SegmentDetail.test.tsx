import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
}));

vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ id: 'seg-1' }),
  useNavigate: () => mocks.navigate,
  Link: ({ children, to }: React.PropsWithChildren<{ to: string }>) => <a href={to}>{children}</a>,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const refetchSegmentCustomers = vi.fn();

const segmentMocks = vi.hoisted(() => ({
  segment: undefined as
    | undefined
    | {
        id: string;
        name: string;
        description: string;
        color: string;
        rules: string[];
        customerCount: number;
        trend: string;
      },
  segmentError: undefined as Error | undefined,
  segmentIsError: false,
  segmentIsPending: false,
  segmentIsSuccess: false,
  segmentIsPaused: false,
  segmentFailureCount: 0,
  refetchSegment: vi.fn(),
}));

const customersInSegment: unknown[] = [];

vi.mock('../../shared/hooks/queries/useSegments', () => ({
  useGetSegment: () => ({
    data: segmentMocks.segment,
    isLoading: false,
    error: segmentMocks.segmentError,
    isError: segmentMocks.segmentIsError,
    isPending: segmentMocks.segmentIsPending,
    isSuccess: segmentMocks.segmentIsSuccess,
    isPaused: segmentMocks.segmentIsPaused,
    failureCount: segmentMocks.segmentFailureCount,
    refetch: segmentMocks.refetchSegment,
  }),
  useSegmentCustomers: () => ({
    data: { pages: [{ customers: customersInSegment }] },
    isLoading: false,
    isFetchingNextPage: false,
    hasNextPage: false,
    fetchNextPage: vi.fn(),
    refetch: refetchSegmentCustomers,
  }),
}));

import SegmentDetail from './SegmentDetail';

function renderSegmentDetail(): void {
  render(<SegmentDetail />);
}

beforeEach(() => {
  mocks.navigate.mockClear();
  refetchSegmentCustomers.mockClear();
  segmentMocks.refetchSegment.mockClear();
  segmentMocks.segment = undefined;
  segmentMocks.segmentError = undefined;
  segmentMocks.segmentIsError = false;
  segmentMocks.segmentIsPending = false;
  segmentMocks.segmentIsSuccess = false;
  segmentMocks.segmentIsPaused = false;
  segmentMocks.segmentFailureCount = 0;
  customersInSegment.length = 0;
});

describe('SegmentDetail not-found state', () => {
  it('shows a "back to segments" call-to-action when the segment failed to load', () => {
    segmentMocks.segmentError = new Error('not found');
    segmentMocks.segmentIsError = true;
    renderSegmentDetail();

    expect(screen.getByText('crm.segmentDetail.notFoundTitle')).toBeInTheDocument();
    expect(screen.getByText('crm.segmentDetail.notFoundDescription')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'crm.segmentDetail.backToSegments' }));

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/crm/segments' });
  });

  it('shows retry when the segment request failed and paused offline', () => {
    segmentMocks.segmentError = new Error('network unavailable');
    segmentMocks.segmentIsPending = true;
    segmentMocks.segmentIsPaused = true;
    segmentMocks.segmentFailureCount = 1;
    renderSegmentDetail();

    expect(screen.getByText('crmErrors.generic')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'crm.actions.retry' }));
    expect(segmentMocks.refetchSegment).toHaveBeenCalledTimes(1);
  });
});

describe('SegmentDetail empty-segment state', () => {
  it('shows a "refresh" call-to-action when the segment has no matching customers', () => {
    segmentMocks.segment = {
      id: 'seg-1',
      name: 'VIP Customers',
      description: 'High spenders',
      color: '#3b82f6',
      rules: ['spend > 1000'],
      customerCount: 0,
      trend: '+0%',
    };
    renderSegmentDetail();

    expect(screen.getByText('crm.segmentDetail.noCustomersTitle')).toBeInTheDocument();
    expect(screen.getByText('crm.segmentDetail.noCustomersDescription')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'crm.actions.refresh' }));

    expect(refetchSegmentCustomers).toHaveBeenCalledTimes(1);
  });
});
