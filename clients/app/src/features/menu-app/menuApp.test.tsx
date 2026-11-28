import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import { DEFAULT_LAYOUT_CONFIG } from '@voxire/menu';
import { createEffectiveAccess } from '../../../test/effectiveAccess';
import type { StoredUser } from '../../shared/stores/userStore';
import { canAccessPortalPath } from '../../shared/utils/canAccessPortalPath';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

const navigateMock = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigateMock,
}));

const publishLiveMock = vi.hoisted(() => vi.fn().mockResolvedValue(undefined));
const setSelectedLocationIdMock = vi.hoisted(() => vi.fn());

const branchProfileState = vi.hoisted(() => ({
  user: { id: 1, businessId: 42, businessSlug: 'voxire' },
  businessSlug: 'voxire',
  branches: [{ id: 10, name: 'Voxire Cafe', slug: 'voxire-cafe', address: 'Main St' }],
  selectedLocationId: 0,
  setSelectedLocationId: setSelectedLocationIdMock,
  selectedBranch: undefined as { id: number; name: string; slug: string } | undefined,
  remoteProfile: { publishedAt: '2026-11-27T10:00:00Z' },
  draftConfig: {
    pages: { menu: true, about: true, gallery: true, offers: true, reviews: true },
  } as typeof DEFAULT_LAYOUT_CONFIG,
  isSaving: false,
  isPublishing: false,
  publishLive: publishLiveMock,
}));

vi.mock('./useMenuBranchProfile', () => ({
  useMenuBranchProfile: () => branchProfileState,
}));

const updateStatusMutate = vi.hoisted(() => vi.fn());
const deleteReviewMutate = vi.hoisted(() => vi.fn());

const builderMocks = vi.hoisted(() => ({
  catalogItems: [
    { id: 1, name: 'Espresso', isActive: true },
    { id: 2, name: 'Croissant', isActive: true },
    { id: 3, name: 'Seasonal', isActive: false },
  ],
  reviews: [
    {
      id: 101,
      authorName: 'Layla',
      rating: 5,
      comment: 'Best matcha in town!',
      status: 'approved',
      createdAt: '2026-11-27T10:00:00Z',
    },
    {
      id: 102,
      authorName: 'Omar',
      rating: 4,
      comment: 'Great espresso, friendly staff.',
      status: 'pending',
      createdAt: '2026-11-27T11:00:00Z',
    },
  ],
  analytics: {
    qrScans: 42,
    menuViews: 120,
    itemViews: 85,
    addToCarts: 30,
    checkoutStarts: 18,
    ordersPlaced: 12,
  },
}));

vi.mock('../../shared/hooks/queries/useMenuBuilder', () => ({
  useMenuCatalogItems: () => ({ data: builderMocks.catalogItems }),
  useMenuReviews: () => ({ data: builderMocks.reviews, isLoading: false }),
  useMenuAnalyticsSummary: () => ({ data: builderMocks.analytics }),
  useUpdateMenuReviewStatus: () => ({ mutate: updateStatusMutate, isPending: false }),
  useDeleteMenuReview: () => ({ mutate: deleteReviewMutate, isPending: false }),
}));

import MenuOverviewPage from './MenuOverviewPage';
import MenuReviewsPage from './MenuReviewsPage';

beforeEach(() => {
  vi.clearAllMocks();
  branchProfileState.draftConfig = {
    ...DEFAULT_LAYOUT_CONFIG,
    pages: { menu: true, about: true, gallery: true, offers: true, reviews: true },
  };
});

describe('MenuOverviewPage', () => {
  it('shows live storefront URL, active item and review KPIs, and navigates to menu sections', () => {
    render(<MenuOverviewPage />);

    expect(screen.getByText(/voxire-cafe/i)).toBeInTheDocument();
    expect(screen.getByText('42')).toBeInTheDocument();
    expect(screen.getByText('120')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: /common\.nav\.menuPages/i }));
    expect(navigateMock).toHaveBeenCalledWith({ to: '/menus/pages' });

    fireEvent.click(screen.getByRole('button', { name: /common\.nav\.menuReviews/i }));
    expect(navigateMock).toHaveBeenCalledWith({ to: '/menus/reviews' });

    fireEvent.click(screen.getByRole('button', { name: /common\.nav\.menuSettings/i }));
    expect(navigateMock).toHaveBeenCalledWith({ to: '/menus/settings' });
  });
});

describe('MenuReviewsPage', () => {
  it('filters reviews by status, moderates pending reviews, and toggles the public reviews tab', () => {
    render(<MenuReviewsPage />);

    expect(screen.getByText('Layla')).toBeInTheDocument();
    expect(screen.getByText('Omar')).toBeInTheDocument();

    // Filter to pending only
    fireEvent.click(screen.getByRole('button', { name: 'design.pagesEditor.filterPending' }));
    expect(screen.queryByText('Layla')).not.toBeInTheDocument();
    expect(screen.getByText('Omar')).toBeInTheDocument();

    // Approve pending review
    fireEvent.click(screen.getByRole('button', { name: 'design.pagesEditor.approve' }));
    expect(updateStatusMutate).toHaveBeenCalledWith({ id: '102', status: 'approved' });

    // Delete review
    fireEvent.click(screen.getByRole('button', { name: 'design.pagesEditor.deleteReview' }));
    expect(deleteReviewMutate).toHaveBeenCalledWith('102');

    // Toggle public reviews tab off -> publishes live
    fireEvent.click(screen.getByRole('switch'));
    expect(publishLiveMock).toHaveBeenCalledWith(
      expect.objectContaining({
        pages: expect.objectContaining({ reviews: false }),
      })
    );
  });
});

describe('Menu app route permissions', () => {
  it('gates branding-only menu subpages separately from read/catalog menu subpages', () => {
    const viewer = {
      id: 2,
      effectiveAccess: createEffectiveAccess('menu', {
        permissions: ['menu:menus:view'],
      }),
    } as unknown as StoredUser;

    expect(canAccessPortalPath(viewer, '/menus')).toBe(true);
    expect(canAccessPortalPath(viewer, '/menus/items')).toBe(true);
    expect(canAccessPortalPath(viewer, '/menus/promos')).toBe(true);
    expect(canAccessPortalPath(viewer, '/menus/qr-codes')).toBe(true);
    expect(canAccessPortalPath(viewer, '/menus/analytics')).toBe(true);
    expect(canAccessPortalPath(viewer, '/menus/studio')).toBe(false);
    expect(canAccessPortalPath(viewer, '/menus/pages')).toBe(false);
    expect(canAccessPortalPath(viewer, '/menus/reviews')).toBe(false);
    expect(canAccessPortalPath(viewer, '/menus/settings')).toBe(false);

    const noMenuAccess = {
      id: 3,
      effectiveAccess: createEffectiveAccess(undefined, { dashPermissions: [] }),
    } as unknown as StoredUser;
    expect(canAccessPortalPath(noMenuAccess, '/menus')).toBe(false);
    expect(canAccessPortalPath(noMenuAccess, '/menus/studio')).toBe(false);
  });
});
