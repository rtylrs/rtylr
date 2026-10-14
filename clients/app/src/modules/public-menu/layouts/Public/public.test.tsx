import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';

vi.mock('./PublicHeader', () => ({
  default: () => <header data-testid="mock-header">Header</header>,
}));

vi.mock('./PublicFooter/PublicFooter', () => ({
  default: () => <footer data-testid="mock-footer">Footer</footer>,
}));

import Public from './public';

describe('Public layout wrapper', () => {
  it('renders header, children and footer by default', () => {
    render(
      <Public>
        <main data-testid="main-content">Content</main>
      </Public>
    );

    expect(screen.getByTestId('mock-header')).toBeInTheDocument();
    expect(screen.getByTestId('main-content')).toBeInTheDocument();
    expect(screen.getByTestId('mock-footer')).toBeInTheDocument();
  });

  it('respects hasHeader=false and hasFooter=false', () => {
    render(
      <Public hasHeader={false} hasFooter={false}>
        <main data-testid="main-content">Content</main>
      </Public>
    );

    expect(screen.queryByTestId('mock-header')).not.toBeInTheDocument();
    expect(screen.getByTestId('main-content')).toBeInTheDocument();
    expect(screen.queryByTestId('mock-footer')).not.toBeInTheDocument();
  });
});
