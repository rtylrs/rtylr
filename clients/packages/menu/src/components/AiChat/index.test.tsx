import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import AiChat from './index';

describe('AiChat', () => {
  it('sends the latest user message in the history passed to onSendMessage', async () => {
    const onSendMessage = vi.fn(async () => 'All set');
    const user = userEvent.setup();

    render(<AiChat onSendMessage={onSendMessage} />);

    await user.click(screen.getByRole('button', { name: /open ai chat/i }));
    await user.type(screen.getByRole('textbox'), 'Is the burger spicy?');
    await user.click(screen.getByRole('button', { name: /send message/i }));

    await waitFor(() => expect(onSendMessage).toHaveBeenCalledTimes(1));

    const history = (
      onSendMessage.mock.calls[0] as unknown as [string, Array<{ role: string; content: string }>]
    )[1];
    expect(history).toHaveLength(2);
    expect(history[0]?.role).toBe('assistant');
    expect(history[1]).toMatchObject({
      role: 'user',
      content: 'Is the burger spicy?',
    });
  });
});
