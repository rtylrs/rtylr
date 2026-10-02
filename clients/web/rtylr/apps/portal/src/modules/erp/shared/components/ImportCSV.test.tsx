import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import ImportCSV from './ImportCSV';

const toast = vi.hoisted(() => ({
  error: vi.fn(),
  success: vi.fn(),
}));

vi.mock('sonner', () => ({ toast }));

describe('ImportCSV', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('waits for the import mutation before showing success', async () => {
    const user = userEvent.setup();
    let resolveImport: (() => void) | undefined;
    const onImport = vi.fn(
      () =>
        new Promise<number>((resolve) => {
          resolveImport = () => resolve(2);
        })
    );

    render(<ImportCSV onImport={onImport} />);
    const input = screen.getByTestId('csv-file-input');
    await user.upload(
      input,
      new File(['name,sku\nCoffee,COF-1'], 'products.csv', { type: 'text/csv' })
    );
    await screen.findByText('Preview (1 rows)');

    fireEvent.click(screen.getByRole('button', { name: 'Import 1 rows' }));
    expect(onImport).toHaveBeenCalledWith([{ name: 'Coffee', sku: 'COF-1' }]);
    expect(toast.success).not.toHaveBeenCalled();

    resolveImport?.();
    await waitFor(() => expect(toast.success).toHaveBeenCalledWith('Imported 2 rows'));
  });

  it('keeps quoted commas, escaped quotes and embedded newlines in one field', async () => {
    const user = userEvent.setup();
    const onImport = vi.fn().mockResolvedValue(undefined);

    render(<ImportCSV onImport={onImport} />);
    const input = screen.getByTestId('csv-file-input');
    await user.upload(
      input,
      new File(
        ['name,description,sku\n"Widget, Large","Say ""hi""\nsecond line",COF-1'],
        'products.csv',
        { type: 'text/csv' }
      )
    );
    await screen.findByText('Preview (1 rows)');

    await user.click(screen.getByRole('button', { name: 'Import 1 rows' }));
    await waitFor(() =>
      expect(onImport).toHaveBeenCalledWith([
        { name: 'Widget, Large', description: 'Say "hi"\nsecond line', sku: 'COF-1' },
      ])
    );
  });

  it('keeps the preview and reports a failed mutation', async () => {
    const user = userEvent.setup();
    const onImport = vi.fn().mockRejectedValue(new Error('Row 2: supplier_id is required'));

    render(<ImportCSV onImport={onImport} />);
    const input = screen.getByTestId('csv-file-input');
    await user.upload(
      input,
      new File(['name,sku\nCoffee,COF-1'], 'products.csv', { type: 'text/csv' })
    );
    await screen.findByText('Preview (1 rows)');

    await user.click(screen.getByRole('button', { name: 'Import 1 rows' }));
    await waitFor(() => expect(toast.error).toHaveBeenCalledWith('Row 2: supplier_id is required'));
    expect(screen.getByText('Preview (1 rows)')).toBeInTheDocument();
    expect(toast.success).not.toHaveBeenCalled();
  });
});
