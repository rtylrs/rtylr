import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { useCaretStore } from '../stores/caretStore';
import { caretTokens } from '../ui/tokens';
import CaretButton, { caretButtonStyle } from './CaretButton';

describe('caretButtonStyle', () => {
  it('matches the header button size and stays neutral when closed', () => {
    const style = caretButtonStyle(false, false);

    expect(style.height).toBe('28px');
    expect(style.backgroundColor).toBe('transparent');
    expect(style.border).toBe(`1px solid ${caretTokens.border}`);
    expect(style.color).toBe(caretTokens.textMuted);
  });

  it('shows the accent treatment while the panel is open', () => {
    const style = caretButtonStyle(true, false);

    expect(style.backgroundColor).toBe(caretTokens.accentSoft);
    expect(style.border).toBe(`1px solid ${caretTokens.accentBorder}`);
    expect(style.color).toBe(caretTokens.accent);
  });

  it('tints on hover without the open accent', () => {
    const style = caretButtonStyle(false, true);

    expect(style.backgroundColor).toBe(caretTokens.hover);
    expect(style.color).toBe(caretTokens.textMuted);
  });
});

describe('CaretButton', () => {
  beforeEach(() => {
    useCaretStore.setState({ isOpen: false });
  });

  it('reflects the panel state for assistive tech and toggles on click', () => {
    render(<CaretButton />);

    const button = screen.getByRole('button', { name: 'Open Caret AI' });
    expect(button.getAttribute('aria-expanded')).toBe('false');

    fireEvent.click(button);
    expect(useCaretStore.getState().isOpen).toBe(true);
    expect(button.getAttribute('aria-expanded')).toBe('true');
  });
});
