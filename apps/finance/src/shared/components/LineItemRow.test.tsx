import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeAll } from 'vitest';
import LineItemRow, { LineItem } from './LineItemRow';
import { initTestI18n } from '../testing/i18n';

// i18next is not mocked here: the real locale files mean an assertion
// written in English also proves the key resolves (rtylr#1269).
beforeAll(async () => {
  await initTestI18n('en');
});

describe('LineItemRow', () => {
  const sampleItem: LineItem = {
    id: 'item-1',
    description: 'Web Development',
    quantity: 2,
    unitPrice: 150,
    taxRate: 10,
  };

  it('calculates total with tax correctly (2 * 150 = 300 + 10% = 330)', () => {
    render(
      <table>
        <tbody>
          <LineItemRow item={sampleItem} onChange={vi.fn()} onRemove={vi.fn()} />
        </tbody>
      </table>
    );
    expect(screen.getByDisplayValue('Web Development')).toBeInTheDocument();
    expect(screen.getByText('$330.00')).toBeInTheDocument();
  });

  it('invokes onChange when quantity or unit price changes', () => {
    const handleChange = vi.fn();
    render(
      <table>
        <tbody>
          <LineItemRow item={sampleItem} onChange={handleChange} onRemove={vi.fn()} />
        </tbody>
      </table>
    );

    const qtyInput = screen.getByDisplayValue('2');
    fireEvent.change(qtyInput, { target: { value: '3' } });
    expect(handleChange).toHaveBeenCalledWith({ ...sampleItem, quantity: 3 });
  });

  it('invokes onRemove when trash button is clicked', () => {
    const handleRemove = vi.fn();
    render(
      <table>
        <tbody>
          <LineItemRow item={sampleItem} onChange={vi.fn()} onRemove={handleRemove} />
        </tbody>
      </table>
    );

    const removeBtn = screen.getByTitle('Remove line');
    fireEvent.click(removeBtn);
    expect(handleRemove).toHaveBeenCalled();
  });
});
