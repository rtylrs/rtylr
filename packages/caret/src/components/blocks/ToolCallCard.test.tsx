import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { ToolCallBlock } from '../../types/message';
import ToolCallCard from './ToolCallCard';

describe('ToolCallCard', () => {
  it('keeps technical tool metadata out of the user-facing details', () => {
    const call: ToolCallBlock = {
      id: 'tool-1',
      type: 'tool_call',
      name: 'insights.get_product_kpis',
      input: { business_id: 42 },
      status: 'success',
      output: { product_kpis: { total_skus: 10 } },
      summary: 'Checked product data.',
    };

    render(<ToolCallCard call={call} />);
    fireEvent.click(screen.getByRole('button', { name: /Completed: check business data/i }));

    expect(screen.queryByText('technical name')).toBeNull();
    expect(screen.queryByText('request details')).toBeNull();
    expect(screen.queryByText('insights.get_product_kpis')).toBeNull();
    expect(screen.queryByText(/business_id/)).toBeNull();
    expect(screen.getByText('Checked product data.')).not.toBeNull();
  });
});
