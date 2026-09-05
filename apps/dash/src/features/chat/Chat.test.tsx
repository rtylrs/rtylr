import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: (props: React.ComponentProps<'div'>) => <div {...props} />,
    button: (props: React.ComponentProps<'button'>) => <button type="button" {...props} />,
  },
  AnimatePresence: ({ children }: { children?: React.ReactNode }) => <>{children}</>,
}));

vi.mock('../../shared/stores/userStore', () => ({
  useUserStore: () => ({ user: { id: '1', businessName: 'Acme', role: 'owner', apps: [] } }),
}));

vi.mock('../../shared/lib/dashCaretContext', () => ({
  buildDashSystemContext: () => 'system-context',
  buildDashPageContext: () => 'page-context',
}));

type CaretMessage = { id: string; role: 'user' | 'assistant'; content: string; blocks: unknown[] };

const caretState = vi.hoisted(() => ({
  messages: [] as CaretMessage[],
  isStreaming: false,
  sendMessage: vi.fn(),
}));

vi.mock('@voxire/caret', () => ({
  useCaretStore: () => ({ messages: caretState.messages, isStreaming: caretState.isStreaming }),
  useCaretStream: () => ({ sendMessage: caretState.sendMessage }),
  MessageActions: () => null,
  MessageBlocks: () => null,
  StreamingDot: () => null,
}));

import Chat from './Chat';

beforeEach(() => {
  vi.clearAllMocks();
  caretState.messages = [];
  caretState.isStreaming = false;
});

describe('Chat: no-messages welcome empty state', () => {
  it('shows the illustration, title and description when there are no messages yet', () => {
    render(<Chat />);

    expect(screen.getByText('chat.welcomeTitle')).toBeInTheDocument();
    expect(screen.getByText('chat.welcomeDescription')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'chat.startChatting' })).toBeInTheDocument();
  });

  it('focuses the message input when the "start chatting" call-to-action is clicked', () => {
    render(<Chat />);

    const messageInput = screen.getByPlaceholderText('Message Caret...');
    expect(messageInput).not.toHaveFocus();

    fireEvent.click(screen.getByRole('button', { name: 'chat.startChatting' }));

    expect(messageInput).toHaveFocus();
  });

  it('does not show the welcome empty state when there are messages', () => {
    caretState.messages = [{ id: '1', role: 'user', content: 'Hello', blocks: [] }];
    render(<Chat />);

    expect(screen.queryByText('chat.welcomeTitle')).not.toBeInTheDocument();
    expect(screen.getByText('Hello')).toBeInTheDocument();
  });
});

describe('messageCopyText', () => {
  const base = {
    id: 'm1',
    role: 'assistant',
    createdAt: new Date().toISOString(),
  } as const;

  it('prefers message content when present', async () => {
    const { messageCopyText } = await import('./Chat');

    expect(
      messageCopyText({
        ...base,
        content: 'Hello there',
        blocks: [{ id: 'b1', type: 'text', content: 'Stale' }],
      })
    ).toBe('Hello there');
  });

  it('falls back to block text when content is an empty string', async () => {
    const { messageCopyText } = await import('./Chat');

    expect(
      messageCopyText({
        ...base,
        content: '',
        blocks: [{ id: 'b1', type: 'text', content: 'Streamed text' }],
      })
    ).toBe('Streamed text');
  });

  it('returns empty string when neither content nor blocks carry text', async () => {
    const { messageCopyText } = await import('./Chat');

    expect(messageCopyText({ ...base, content: '', blocks: [] })).toBe('');
  });
});
