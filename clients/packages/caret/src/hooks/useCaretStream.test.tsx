import { act, render, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useCaretStream } from './useCaretStream';
import type { UseCaretStreamParams, UseCaretStreamResult } from './useCaretStream';
import { useCaretStore } from '../stores/caretStore';

const params: UseCaretStreamParams = {
  userId: 1,
  apiBaseUrl: 'http://127.0.0.1:9',
};

let api: UseCaretStreamResult | null = null;

function Harness(): null {
  api = useCaretStream(params);
  return null;
}

let streamController: ReadableStreamDefaultController<Uint8Array> | null = null;
let abortSignal: AbortSignal | null = null;
let fetchMock: ReturnType<typeof vi.fn>;

const encoder = new TextEncoder();

function pushSseEvent(line: string): void {
  streamController?.enqueue(encoder.encode(`data: ${line}\n\n`));
}

function resetCaretStore(): void {
  const state = useCaretStore.getState();
  for (const tab of state.tabs.slice(1)) state.closeTab(tab.id);
  const activeTabId = useCaretStore.getState().activeTabId;
  useCaretStore.getState().reset(activeTabId);
  useCaretStore.getState().setActiveTab(activeTabId);
}

beforeEach(() => {
  api = null;
  streamController = null;
  abortSignal = null;
  fetchMock = vi.fn((_input: RequestInfo | URL, init?: RequestInit) => {
    abortSignal = init?.signal ?? null;
    const body = new ReadableStream<Uint8Array>({
      start(controller) {
        streamController = controller;
      },
    });
    abortSignal?.addEventListener('abort', () => {
      streamController?.error(new DOMException('The operation was aborted.', 'AbortError'));
    });
    return Promise.resolve({ ok: true, status: 200, statusText: 'OK', body } as Response);
  });
  globalThis.fetch = fetchMock as unknown as typeof fetch;
  resetCaretStore();
});

describe('useCaretStream', () => {
  it('stops a stream that was started before the hook remounted', async () => {
    const first = render(<Harness />);
    const tabId = useCaretStore.getState().activeTabId;

    await act(async () => {
      void api!.sendMessage('Summarize the report');
    });

    pushSseEvent('{"type":"text_delta","messageId":"m","delta":"Hello"}');
    await waitFor(() => {
      const assistant = useCaretStore
        .getState()
        .tabs.find((tab) => tab.id === tabId)
        ?.messages.find((message) => message.role === 'assistant');
      expect(assistant?.content).toBe('Hello');
    });

    expect(abortSignal?.aborted).toBe(false);
    first.unmount();
    render(<Harness />);

    act(() => {
      api!.stopStreaming(tabId);
    });

    expect(abortSignal?.aborted).toBe(true);
    await waitFor(() => {
      const tab = useCaretStore.getState().tabs.find((candidate) => candidate.id === tabId);
      expect(tab?.isStreaming).toBe(false);
    });

    const tab = useCaretStore.getState().tabs.find((candidate) => candidate.id === tabId);
    const lastAssistant = tab?.messages.find((message) => message.role === 'assistant');
    expect(lastAssistant?.content).toBe('Hello');
  });
});
