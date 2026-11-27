import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

const updateReviewStatusMutate = vi.hoisted(() => vi.fn());
const deleteReviewMutate = vi.hoisted(() => vi.fn());
const mockReviews = vi.hoisted(() => [
  {
    id: '11',
    locationId: '0',
    authorName: 'Nadia',
    rating: 5,
    comment: 'Best latte!',
    status: 'pending',
    createdAt: '2026-11-27T10:00:00Z',
    updatedAt: '2026-11-27T10:00:00Z',
  },
]);

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

vi.mock('@voxire/api', () => ({
  uploadAsset: vi.fn(),
}));

vi.mock('../../shared/hooks/queries/useMenuBuilder', () => ({
  useMenuReviews: () => ({ data: mockReviews, isLoading: false }),
  useUpdateMenuReviewStatus: () => ({ mutate: updateReviewStatusMutate, isPending: false }),
  useDeleteMenuReview: () => ({ mutate: deleteReviewMutate, isPending: false }),
}));

import PagesEditor from './PagesEditor';

describe('PagesEditor', () => {
  beforeEach(() => {
    updateReviewStatusMutate.mockReset();
    deleteReviewMutate.mockReset();
  });

  it('edits About content, adds Offers, and moderates Reviews', () => {
    const onPatchPages = vi.fn();
    const onPatchContent = vi.fn();

    render(
      <PagesEditor
        selectedLocationId={0}
        pages={{ menu: true, about: true, gallery: true, offers: true, reviews: true }}
        content={{ gallery: ['https://cdn.example.com/a.jpg', 'https://cdn.example.com/b.jpg'] }}
        defaultAddress="123 Main St"
        onPatchPages={onPatchPages}
        onPatchContent={onPatchContent}
      />
    );

    // Edit About story
    fireEvent.change(screen.getByPlaceholderText('design.pagesEditor.storyPlaceholder'), {
      target: { value: 'Specialty coffee roasters.' },
    });
    expect(onPatchContent).toHaveBeenCalledWith({
      about: { story: 'Specialty coffee roasters.' },
    });

    // Add Offer
    fireEvent.change(screen.getByPlaceholderText('design.pagesEditor.offerTitlePlaceholder'), {
      target: { value: 'Morning Deal' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'design.pagesEditor.addOffer' }));
    expect(onPatchContent).toHaveBeenCalledWith(
      expect.objectContaining({
        offers: [expect.objectContaining({ title: 'Morning Deal' })],
      })
    );

    // Approve & delete Review
    expect(screen.getByText('Best latte!')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'design.pagesEditor.approve' }));
    expect(updateReviewStatusMutate).toHaveBeenCalledWith({ id: '11', status: 'approved' });

    fireEvent.click(screen.getByRole('button', { name: 'design.pagesEditor.deleteReview' }));
    expect(deleteReviewMutate).toHaveBeenCalledWith('11');
  });
});
