import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import PrintPOButton from './PrintPOButton';
import { initTestI18n } from '../testing/i18n';

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

beforeAll(async () => {
  await initTestI18n('en');
});

afterEach(() => {
  vi.restoreAllMocks();
});

function printedHtml(): string {
  const write = vi.fn();
  vi.spyOn(window, 'open').mockReturnValue({
    document: { write, close: vi.fn() },
  } as unknown as Window);
  fireEvent.click(screen.getByRole('button', { name: /Print/ }));
  return write.mock.calls[0][0] as string;
}

describe('PrintPOButton', () => {
  it('escapes user-entered names so they cannot run in the print window', () => {
    render(
      <PrintPOButton
        order={{
          id: '7',
          poNumber: 'PO"><script>alert(1)</script>',
          notes: '<img src=x onerror=alert(2)>',
          items: [{ id: 'i1', productId: 'p1', quantityOrdered: 2, unitCost: 3 }],
        }}
        supplierName={'<img src=x onerror=alert(3)>'}
        productNameById={new Map([['p1', '<b onmouseover=alert(4)>Beans</b>']])}
        currency="USD"
      />
    );

    const html = printedHtml();
    expect(html).not.toContain('<script>alert(1)</script>');
    expect(html).not.toContain('<img src=x');
    expect(html).not.toContain('<b onmouseover');
    expect(html).toContain('&lt;img src=x onerror=alert(3)&gt;');
    expect(html).toContain('&lt;b onmouseover=alert(4)&gt;Beans&lt;/b&gt;');
  });

  it('prints every line item with its cost and the order totals in the business currency', () => {
    render(
      <PrintPOButton
        order={{
          id: '7',
          poNumber: 'ACME-1',
          subtotal: 26,
          taxAmount: 2.6,
          total: 28.6,
          items: [
            { id: 'i1', productId: 'p1', quantityOrdered: 2, unitCost: 4, totalCost: 8 },
            { id: 'i2', productId: 'p2', quantityOrdered: 3, unitCost: 6 },
          ],
        }}
        supplierName="Acme"
        productNameById={new Map([['p1', 'Beans']])}
        currency="EUR"
      />
    );

    const html = printedHtml();
    expect(html).toContain('Purchase Order ACME-1');
    expect(html).toContain('<td>Beans</td>');
    expect(html).toContain('<td>Product #p2</td>');
    expect(html).toContain('€18.00');
    expect(html).toContain('€28.60');
    expect(html).not.toContain('see order details');
  });
});
