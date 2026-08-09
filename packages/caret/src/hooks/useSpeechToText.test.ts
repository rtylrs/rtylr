import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useSpeechToText } from './useSpeechToText';

interface MockResult {
  isFinal: boolean;
  0: { transcript: string };
}

class MockRecognition {
  static instance: MockRecognition | null = null;
  continuous = false;
  interimResults = false;
  lang = '';
  onresult: ((event: { results: ArrayLike<MockResult> }) => void) | null = null;
  onend: (() => void) | null = null;
  onerror: ((event: { error: string }) => void) | null = null;
  start = vi.fn();
  stop = vi.fn();

  constructor() {
    MockRecognition.instance = this;
  }
}

const speechRecognitionKey = 'SpeechRecognition' as keyof Window;

beforeEach(() => {
  Object.defineProperty(window, speechRecognitionKey, {
    configurable: true,
    value: MockRecognition,
    writable: true,
  });
});

afterEach(() => {
  delete (window as unknown as Record<string, unknown>)[speechRecognitionKey];
  MockRecognition.instance = null;
});

describe('useSpeechToText', () => {
  it('rebuilds final results instead of appending repeated interim results', () => {
    const { result } = renderHook(() => useSpeechToText());
    const recognition = MockRecognition.instance;

    expect(recognition).not.toBeNull();
    expect(result.current.supported).toBe(true);

    act(() => {
      recognition?.onresult?.({
        results: [{ isFinal: false, 0: { transcript: 'sub' } }],
      });
    });
    expect(result.current.transcript).toBe('');
    expect(result.current.interimTranscript).toBe('sub');

    act(() => {
      recognition?.onresult?.({
        results: [
          { isFinal: true, 0: { transcript: 'sub ' } },
          { isFinal: false, 0: { transcript: 'sub' } },
        ],
      });
    });
    expect(result.current.transcript).toBe('sub ');
    expect(result.current.interimTranscript).toBe('sub');

    act(() => {
      recognition?.onresult?.({
        results: [
          { isFinal: true, 0: { transcript: 'sub ' } },
          { isFinal: true, 0: { transcript: 'again' } },
        ],
      });
    });
    expect(result.current.transcript).toBe('sub again');
    expect(result.current.interimTranscript).toBe('');
  });
});
