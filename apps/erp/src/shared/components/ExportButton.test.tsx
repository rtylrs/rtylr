import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import ExportButton from './ExportButton';

const toast = vi.hoisted(() => ({
  error: vi.fn(),
  success: vi.fn(),
}));

vi.mock('sonner', () => ({ toast }));

describe('ExportButton', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:export');
    vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => undefined);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('exports rows loaded asynchronously instead of only the visible page', async () => {
    const user = userEvent.setup();
    const getData = vi.fn().mockResolvedValue([{ id: '1' }, { id: '2' }]);

    render(<ExportButton data={[{ id: '1' }]} getData={getData} />);
    await user.click(screen.getByRole('button', { name: 'Export CSV' }));

    await waitFor(() => expect(toast.success).toHaveBeenCalledWith('Exported 2 rows'));
    expect(getData).toHaveBeenCalledOnce();
    expect(toast.error).not.toHaveBeenCalled();
  });
});
