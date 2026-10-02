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
}));

const customersInSegment: unknown[] = [];

vi.mock('../../shared/hooks/queries/useSegments', () => ({
  useGetSegment: () => ({
    data: segmentMocks.segment,
    isLoading: false,
    error: segmentMocks.segmentError,
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
  segmentMocks.segment = undefined;
  segmentMocks.segmentError = undefined;
  customersInSegment.length = 0;
});

describe('SegmentDetail not-found state', () => {
  it('shows a "back to segments" call-to-action when the segment failed to load', () => {
    segmentMocks.segmentError = new Error('not found');
    renderSegmentDetail();

    expect(screen.getByText('crm.segmentDetail.notFoundTitle')).toBeInTheDocument();
    expect(screen.getByText('crm.segmentDetail.notFoundDescription')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'crm.segmentDetail.backToSegments' }));

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/crm/segments' });
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
