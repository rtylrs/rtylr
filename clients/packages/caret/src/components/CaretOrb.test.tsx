import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import CaretOrb, { orbStateForTool, StreamingOrb } from './CaretOrb';

describe('orbStateForTool', () => {
  it('maps read, analysis, write and memory tools to distinct states', () => {
    expect(orbStateForTool('erp.list_products')).toBe('searching');
    expect(orbStateForTool('business.snapshot')).toBe('searching');
    expect(orbStateForTool('insights.get_dashboard')).toBe('solving');
    expect(orbStateForTool('analytics.query')).toBe('solving');
    expect(orbStateForTool('documents.map_form_fields')).toBe('searching');
    expect(orbStateForTool('erp.create_product')).toBe('working');
    expect(orbStateForTool('erp.set_stock')).toBe('working');
    expect(orbStateForTool('finance.create_invoice')).toBe('working');
    expect(orbStateForTool('insights.schedule_report')).toBe('connecting');
    expect(orbStateForTool('crm.create_campaign')).toBe('connecting');
    expect(orbStateForTool('memory.remember')).toBe('weaving');
  });

  it('falls back to working for unknown tools', () => {
    expect(orbStateForTool('something.unknown')).toBe('working');
  });
});

describe('CaretOrb', () => {
  it('renders the labelled animation for the requested state', () => {
    render(<CaretOrb state="listening" />);

    expect(screen.getByRole('img', { name: 'Listening…' })).not.toBeNull();
  });
});

describe('StreamingOrb', () => {
  it('thinks before the first token and composes once text arrives', () => {
    const { unmount } = render(<StreamingOrb hasText={false} />);
    expect(screen.getByRole('img', { name: 'Thinking…' })).not.toBeNull();
    unmount();

    render(<StreamingOrb hasText />);
    expect(screen.getByRole('img', { name: 'Composing…' })).not.toBeNull();
  });
});
