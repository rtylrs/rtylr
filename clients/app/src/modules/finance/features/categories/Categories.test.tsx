import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import Categories from './Categories';
import { initTestI18n } from '../../shared/testing/i18n';

/**
 * finance.report.v1.CreateExpenseCategoryDetailRequest has no budget column and
 * the category list returns no budget/spent, so the UI must not collect or show
 * a budget it cannot persist or read back.
 */
const { expenseCategoriesMock, createCategoryMock, updateCategoryMock } = vi.hoisted(() => ({
  expenseCategoriesMock: vi.fn(),
  createCategoryMock: vi.fn(),
  updateCategoryMock: vi.fn(),
}));

vi.mock('../../shared/hooks/useExpenseCategories', () => ({
  useExpenseCategories: () => expenseCategoriesMock(),
  useCreateExpenseCategory: () => ({ mutate: createCategoryMock, isPending: false }),
  useUpdateExpenseCategory: () => ({ mutate: updateCategoryMock, isPending: false }),
  useDeleteExpenseCategory: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

vi.mock('@voxire/ui', () => ({ EmptyState: () => null }));

vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div> },
}));

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

vi.mock('../../shared/components/PageHeader', () => ({
  default: ({ actions }: { actions?: React.ReactNode }) => <div>{actions}</div>,
}));

beforeAll(async () => {
  await initTestI18n('en');
});

beforeEach(() => {
  vi.clearAllMocks();
  expenseCategoriesMock.mockReturnValue({
    data: [{ id: 1, name: 'Travel', color: '#14b8a6' }],
    isLoading: false,
  });
});

describe('Expense categories without budget support', () => {
  it('collects and sends no budget, which the category contract cannot carry', () => {
    render(<Categories />);

    fireEvent.click(screen.getByRole('button', { name: /New Category/ }));
    expect(screen.queryByLabelText('Monthly Budget (Optional)')).not.toBeInTheDocument();

    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Travel' } });
    fireEvent.click(screen.getByRole('button', { name: /Create category/ }));

    expect(createCategoryMock).toHaveBeenCalledTimes(1);
    expect(createCategoryMock.mock.calls[0][0]).toStrictEqual({
      name: 'Travel',
      color: '#14b8a6',
    });

    fireEvent.click(screen.getByRole('button', { name: 'Edit Travel' }));
    expect(screen.queryByLabelText('Monthly Budget')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: /Save Changes/ }));
    expect(updateCategoryMock.mock.calls[0][0]).toStrictEqual({
      id: 1,
      data: { name: 'Travel', color: '#14b8a6' },
    });
  });

  it('renders no budget progress from data the category list never returns', () => {
    render(<Categories />);

    expect(screen.queryByText('0%')).not.toBeInTheDocument();
  });
});
