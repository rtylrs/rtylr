import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const auditData = vi.hoisted(() => ({
  data: { entries: [] as { id: string }[], total: 0 } as
    { entries: { id: string }[]; total: number } | undefined,
  isLoading: false,
  isError: false,
  isFetching: false,
  refetch: vi.fn(),
  params: [] as unknown[],
}));

vi.mock('../../shared/hooks/queries/useActivity', () => ({
  useAuditLog: (params: unknown) => {
    auditData.params.push(params);
    return auditData;
  },
}));

import Audit from './Audit';

beforeEach(() => {
  vi.clearAllMocks();
  auditData.data = { entries: [], total: 0 };
  auditData.isLoading = false;
  auditData.isError = false;
  auditData.isFetching = false;
  auditData.params = [];
});

describe('Audit log load failure', () => {
  it('shows a retryable error instead of an empty log when loading fails', () => {
    auditData.data = undefined;
    auditData.isError = true;
    render(<Audit />);
    expect(screen.getByRole('alert')).toHaveTextContent('activity.audit.loadFailedTitle');
    expect(screen.queryByText('activity.audit.emptyTitle')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'common.refresh' }));
    expect(auditData.refetch).toHaveBeenCalled();
  });
});

describe('Audit log empty state', () => {
  it('does not advertise an IP field that the activity contract does not provide', () => {
    render(<Audit />);

    expect(screen.getByPlaceholderText('Search actor, action, target...')).toBeInTheDocument();
    expect(screen.queryByPlaceholderText(/IP/)).not.toBeInTheDocument();
  });

  it('shows the illustration, title and a refresh call-to-action when there are no entries', () => {
    render(<Audit />);

    expect(screen.getByText('activity.audit.emptyTitle')).toBeInTheDocument();
    expect(screen.getByText('activity.audit.emptyDescription')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'common.refresh' })).toBeInTheDocument();
  });

  it('refetches the audit log when the refresh call-to-action is clicked', () => {
    render(<Audit />);

    fireEvent.click(screen.getByRole('button', { name: 'common.refresh' }));

    expect(auditData.refetch).toHaveBeenCalledTimes(1);
  });

  it('does not show the empty state when there are entries', () => {
    auditData.data = {
      entries: [
        {
          id: '1',
          actor: 'jane@rtylr.com',
          action: 'login',
          target: 'session',
          app: 'dash',
          createdAt: new Date().toISOString(),
        } as unknown as { id: string },
      ],
      total: 1,
    };
    render(<Audit />);

    expect(screen.queryByText('activity.audit.emptyTitle')).not.toBeInTheDocument();
  });

  it('narrows rows to the query and labels the footer as page-filtered', () => {
    auditData.data = {
      entries: [
        {
          id: '1',
          actor: 'jane@rtylr.com',
          action: 'login',
          target: 'session',
          app: 'dash',
          severity: 'warning',
          createdAt: new Date().toISOString(),
        },
        {
          id: '2',
          actor: 'ops@rtylr.com',
          action: 'delete menu',
          target: 'menu-1',
          app: 'dash',
          severity: 'critical',
          createdAt: new Date().toISOString(),
        },
      ] as unknown as { id: string }[],
      total: 2,
    };
    render(<Audit />);

    fireEvent.change(screen.getByPlaceholderText('Search actor, action, target...'), {
      target: { value: 'jane' },
    });

    expect(screen.getByText('jane@rtylr.com')).toBeInTheDocument();
    expect(screen.queryByText('ops@rtylr.com')).not.toBeInTheDocument();
    expect(screen.getByText('activity.audit.filteredShowing')).toBeInTheDocument();
  });

  it('resets to the first page when the query or severity changes', () => {
    auditData.data = { entries: [], total: 120 };
    render(<Audit />);

    fireEvent.click(screen.getByRole('button', { name: 'Next' }));
    expect(auditData.params[auditData.params.length - 1]).toMatchObject({ page: 2 });

    fireEvent.change(screen.getByPlaceholderText('Search actor, action, target...'), {
      target: { value: 'jane' },
    });
    expect(auditData.params[auditData.params.length - 1]).toMatchObject({ page: 1 });

    fireEvent.click(screen.getByRole('button', { name: 'Next' }));
    fireEvent.click(screen.getByRole('button', { name: 'critical' }));
    expect(auditData.params[auditData.params.length - 1]).toMatchObject({ page: 1 });
  });
});
