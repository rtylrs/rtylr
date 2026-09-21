import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import CaretPanel from './CaretPanel';
import { useCaretStore } from '../stores/caretStore';
import { useCaretSettingsStore } from '../stores/settingsStore';

const memoryClient = vi.hoisted(() => ({
  getChatMessages: vi.fn(),
  listChats: vi.fn(),
  deleteChat: vi.fn(),
}));

vi.mock('../lib/memoryClient', () => ({
  createMemoryClient: () => memoryClient,
}));

type HistoryResolver = (value: {
  messages: Array<{ role: string; content: string; createdAt: string }>;
}) => void;

let resolveHistory: HistoryResolver | null = null;

function renderPanel(): void {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <QueryClientProvider client={queryClient}>
      <CaretPanel appId="dash" userId={1} apiBaseUrl="http://127.0.0.1:9" />
    </QueryClientProvider>
  );
}

beforeEach(() => {
  resolveHistory = null;
  useCaretSettingsStore.setState({ voiceEnabled: false, mode: 'ask' });
  memoryClient.listChats.mockResolvedValue({ chats: [] });
  memoryClient.getChatMessages.mockImplementation(
    () =>
      new Promise((resolve) => {
        resolveHistory = resolve;
      })
  );

  const state = useCaretStore.getState();
  for (const tab of state.tabs.slice(1)) state.closeTab(tab.id);
  const emptyTabId = useCaretStore.getState().activeTabId;
  useCaretStore.getState().reset(emptyTabId);
  // Open a saved chat: its transcript is fetched when the panel mounts.
  useCaretStore.getState().newTab({ sessionId: 'chat-1', title: 'Weekly report' });
});

describe('CaretPanel history loading', () => {
  it('keeps a prompt sent while the history request is in flight', async () => {
    renderPanel();

    const tabId = useCaretStore.getState().activeTabId;
    await waitFor(() => expect(memoryClient.getChatMessages).toHaveBeenCalledTimes(1));

    fireEvent.change(screen.getByLabelText('Message Caret'), {
      target: { value: 'What changed this week?' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Send message' }));
    await waitFor(() => {
      const tab = useCaretStore.getState().tabs.find((candidate) => candidate.id === tabId);
      expect(tab?.messages.some((message) => message.content === 'What changed this week?')).toBe(
        true
      );
    });

    await act(async () => {
      resolveHistory?.({
        messages: [
          {
            role: 'user',
            content: 'Previous question about revenue',
            createdAt: '2026-09-18T10:00:00.000Z',
          },
          {
            role: 'assistant',
            content: 'Previous answer about revenue',
            createdAt: '2026-09-18T10:00:05.000Z',
          },
        ],
      });
    });

    const tab = useCaretStore.getState().tabs.find((candidate) => candidate.id === tabId);
    expect(tab?.messages.some((message) => message.content === 'What changed this week?')).toBe(
      true
    );
    expect(
      tab?.messages.some((message) => message.content === 'Previous answer about revenue')
    ).toBe(false);
    expect(tab?.isLoaded).toBe(true);
    // Marking the tab loaded must stop the effect from refetching the history.
    expect(memoryClient.getChatMessages).toHaveBeenCalledTimes(1);
  });
});
