import { vi } from 'vitest';
import type { AlarmEnvironment } from '../alarm';

/**
 * Stand-ins for the browser pieces the alarm drives, so a test can decide
 * whether the bell plays, fails or ends, and count what was asked of it.
 */
export function createFakeAlarmEnvironment() {
  const bell = {
    src: '',
    muted: false,
    currentTime: 0,
    preload: '',
    onended: null as (() => void) | null,
    play: vi.fn((): Promise<void> => Promise.resolve()),
    pause: vi.fn(),
  };

  const oscillators: { start: ReturnType<typeof vi.fn>; stop: ReturnType<typeof vi.fn> }[] = [];
  const context = {
    state: 'running',
    currentTime: 0,
    destination: {},
    resume: vi.fn(() => Promise.resolve()),
    createOscillator: vi.fn(() => {
      const oscillator = {
        type: 'sine',
        frequency: { value: 0 },
        connect: vi.fn(),
        start: vi.fn(),
        stop: vi.fn(),
      };
      oscillators.push(oscillator);
      return oscillator;
    }),
    createGain: vi.fn(() => ({
      gain: { value: 1, setValueAtTime: vi.fn(), exponentialRampToValueAtTime: vi.fn() },
      connect: vi.fn(),
    })),
  };

  const wakeLocks: { release: ReturnType<typeof vi.fn>; drop: () => void }[] = [];
  const navigator = {
    audioSession: { type: 'auto' },
    vibrate: vi.fn(() => true),
    wakeLock: {
      request: vi.fn(async () => {
        const target = new EventTarget();
        const sentinel = Object.assign(target, {
          release: vi.fn(async () => {
            target.dispatchEvent(new Event('release'));
          }),
          /** What the browser does when the page is hidden. */
          drop: () => {
            target.dispatchEvent(new Event('release'));
          },
        });
        wakeLocks.push(sentinel);
        return sentinel;
      }),
    },
  };

  let visibility: DocumentVisibilityState = 'visible';
  const document = Object.assign(new EventTarget(), { title: 'Page' });
  Object.defineProperty(document, 'visibilityState', { get: () => visibility });

  const env = {
    createAudio: vi.fn(() => bell),
    createAudioContext: vi.fn(() => context),
    navigator,
    document,
  } as unknown as AlarmEnvironment;

  return {
    env,
    bell,
    context,
    oscillators,
    navigator,
    document,
    wakeLocks,
    /** The bell reached its end, as the audio element would report it. */
    endBell: () => bell.onended?.(),
    setVisibility: (next: DocumentVisibilityState) => {
      visibility = next;
      document.dispatchEvent(new Event('visibilitychange'));
    },
  };
}
