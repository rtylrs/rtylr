import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { useCaretUIStore } from '../../stores/uiStore';
import HistoryHover from './HistoryHover';

describe('HistoryHover', () => {
  beforeEach(() => {
    useCaretUIStore.setState({ view: 'chat' });
  });

  it('opens the full history view without rendering a recent-chat overlay', () => {
    render(<HistoryHover />);

    const historyButton = screen.getByRole('button', { name: 'History' });
    fireEvent.mouseEnter(historyButton);
    expect(screen.queryByRole('menu')).toBeNull();

    fireEvent.click(historyButton);
    expect(useCaretUIStore.getState().view).toBe('history');
  });
});
