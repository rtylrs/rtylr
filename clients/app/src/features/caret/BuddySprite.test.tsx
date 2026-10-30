import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import BuddySprite from './BuddySprite';

afterEach(cleanup);

describe('BuddySprite', () => {
  it('shows its mood and squishes when poked', () => {
    render(<BuddySprite mood="talking" pokeable title="Caret Buddy" />);
    const sprite = screen.getByRole('img', { name: 'Caret Buddy' });
    expect(sprite).toHaveAttribute('data-mood', 'talking');
    fireEvent.pointerDown(sprite);
    expect(sprite).toHaveAttribute('data-mood', 'poked');
  });

  it('is decorative without a title', () => {
    const { container } = render(<BuddySprite />);
    expect(container.querySelector('svg')).toHaveAttribute('aria-hidden', 'true');
  });
});
