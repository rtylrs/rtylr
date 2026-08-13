import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import Markdown from './Markdown';

describe('Markdown tables', () => {
  it('replaces a header-only table with an empty result message', () => {
    render(
      <Markdown
        source={`| Customer | Total Spend |
| --- | --- |`}
      />
    );

    expect(screen.getByText('No matching records.')).toBeTruthy();
    expect(document.querySelector('table')).toBeNull();
  });

  it('keeps tables that contain data rows', () => {
    render(
      <Markdown
        source={`| Customer | Total Spend |
| --- | --- |
| Acme | $100 |`}
      />
    );

    expect(screen.getByText('Acme')).toBeTruthy();
    expect(document.querySelector('table.caret-table')).toBeTruthy();
  });
});
