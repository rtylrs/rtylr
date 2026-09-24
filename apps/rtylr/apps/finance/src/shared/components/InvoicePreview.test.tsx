// @vitest-environment happy-dom
import React from 'react';
import { render, screen } from '@testing-library/react';
import { beforeAll, describe, expect, it } from 'vitest';
import InvoicePreview from './InvoicePreview';
import { initTestI18n } from '../testing/i18n';

beforeAll(async () => {
  await initTestI18n('en');
});

describe('InvoicePreview persisted totals', () => {
  it('uses the server total for a percentage discount', () => {
    render(
      <InvoicePreview
        invoiceNumber="INV-1"
        client={{ id: 'client-1', name: 'Acme' }}
        issuedDate="2026-08-01"
        dueDate="2026-08-31"
        items={[
          {
            id: 'line-1',
            description: 'Consulting',
            quantity: 1,
            unitPrice: 200,
            taxRate: 0,
          },
        ]}
        discount={10}
        discountType="percentage"
        discountAmount={20}
        subtotal={200}
        tax={0}
        total={180}
      />
    );

    expect(screen.getByText('$180.00')).toBeInTheDocument();
    expect(screen.getByText('Discount').parentElement).toHaveTextContent('-$20.00');
  });
});
