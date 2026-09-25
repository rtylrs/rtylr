import React from 'react';
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Input } from '../index';

describe('Input', () => {
  it('associates the label with the input even when no id is passed', () => {
    render(<Input label="Email" />);
    // getByLabelText only succeeds when htmlFor/id (or aria-labelledby)
    // actually resolve to the same element - the association this test exists
    // to pin, not just that both a label and an input happen to be present.
    expect(screen.getByLabelText('Email')).toBeInTheDocument();
  });

  it('respects an explicitly passed id instead of generating one', () => {
    render(<Input label="Email" id="explicit-email-id" />);
    const input = screen.getByLabelText('Email');
    expect(input).toHaveAttribute('id', 'explicit-email-id');
  });
});
