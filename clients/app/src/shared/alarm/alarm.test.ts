import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  BELL_WATCHDOG_MS,
  RING_CAP_MS,
  RING_GAP_MS,
  TITLE_FLASH_MS,
  VIBRATION_PATTERN,
  createAlarm,
} from './alarm';
import { createFakeAlarmEnvironment } from './testing/fakeAlarmEnvironment';

function setup() {
  const fake = createFakeAlarmEnvironment();
  const alarm = createAlarm({ bellSrc: 'bell.wav', env: fake.env });
  return { ...fake, alarm };
}

const flush = (): Promise<unknown> => vi.advanceTimersByTimeAsync(0);
const ring = (key: string | null, title?: string) => ({ keepAwake: true, ringKey: key, title });

beforeEach(() => {
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
});

describe('unlock', () => {
  it('primes the bell muted, rewinds it and starts the audio context', async () => {
    const { alarm, bell, context, navigator, oscillators } = setup();

    alarm.unlock();

    expect(navigator.audioSession.type).toBe('playback');
    expect(bell.play).toHaveBeenCalledTimes(1);
    expect(bell.muted).toBe(true);
    expect(context.createOscillator).toHaveBeenCalledTimes(1);
    expect(oscillators[0].start).toHaveBeenCalled();

    await flush();

    expect(bell.pause).toHaveBeenCalledTimes(1);
    expect(bell.currentTime).toBe(0);
    expect(bell.muted).toBe(false);
    expect(alarm.getSnapshot()).toEqual({ unlocked: true, ringing: false });
  });

  it('resumes a suspended audio context', () => {
    const { alarm, context } = setup();
    context.state = 'suspended';

    alarm.unlock();

    expect(context.resume).toHaveBeenCalled();
  });

  it('still unlocks where there is no audio session, vibration or Web Audio', async () => {
    const fake = createFakeAlarmEnvironment();
    const alarm = createAlarm({
      bellSrc: 'bell.wav',
      env: { ...fake.env, navigator: {} as Navigator, createAudioContext: () => null },
    });

    alarm.unlock();
    alarm.sync(ring('a'));
    await flush();

    expect(alarm.getSnapshot()).toEqual({ unlocked: true, ringing: true });
    expect(fake.bell.play).toHaveBeenCalledTimes(2);
  });

  it('plays out loud straight away when the tap comes after the ring started', async () => {
    const { alarm, bell } = setup();
    bell.play.mockRejectedValueOnce(new DOMException('no gesture', 'NotAllowedError'));
    alarm.sync(ring('a'));
    await flush();

    alarm.unlock();
    await flush();

    expect(bell.play).toHaveBeenCalledTimes(2);
    expect(bell.muted).toBe(false);
    expect(bell.pause).not.toHaveBeenCalled();
  });
});

describe('ringing', () => {
  it('plays the bell, waits for it to end plus a gap, and repeats', async () => {
    const { alarm, bell, endBell, navigator } = setup();
    alarm.unlock();
    await flush();
    bell.play.mockClear();

    alarm.sync(ring('a'));

    expect(alarm.getSnapshot().ringing).toBe(true);
    expect(bell.play).toHaveBeenCalledTimes(1);
    expect(bell.muted).toBe(false);
    expect(navigator.vibrate).toHaveBeenLastCalledWith(VIBRATION_PATTERN);

    endBell();
    await vi.advanceTimersByTimeAsync(RING_GAP_MS - 1);
    expect(bell.play).toHaveBeenCalledTimes(1);
    await vi.advanceTimersByTimeAsync(1);
    expect(bell.play).toHaveBeenCalledTimes(2);
  });

  it('moves on when the bell never reports that it ended', async () => {
    const { alarm, bell } = setup();
    alarm.sync(ring('a'));

    await vi.advanceTimersByTimeAsync(BELL_WATCHDOG_MS);

    expect(bell.play).toHaveBeenCalledTimes(2);
  });

  it('does not ring twice for the same key', async () => {
    const { alarm, bell } = setup();
    alarm.sync(ring('a'));
    alarm.sync(ring('a'));
    alarm.sync(ring('a'));
    expect(bell.play).toHaveBeenCalledTimes(1);

    alarm.silence();
    alarm.sync(ring('a'));
    await vi.advanceTimersByTimeAsync(BELL_WATCHDOG_MS * 2);

    expect(bell.play).toHaveBeenCalledTimes(1);
    expect(alarm.getSnapshot().ringing).toBe(false);
  });

  it('rings again for a new key', () => {
    const { alarm, bell } = setup();
    alarm.sync(ring('a'));
    alarm.silence();

    alarm.sync(ring('b'));

    expect(bell.play).toHaveBeenCalledTimes(2);
    expect(alarm.getSnapshot().ringing).toBe(true);
  });

  it('stops as soon as the key is withdrawn', async () => {
    const { alarm, bell, navigator } = setup();
    alarm.unlock();
    await flush();
    alarm.sync(ring('a'));
    bell.play.mockClear();
    bell.pause.mockClear();

    alarm.sync(ring(null));
    await vi.advanceTimersByTimeAsync(BELL_WATCHDOG_MS * 2);

    expect(alarm.getSnapshot().ringing).toBe(false);
    expect(bell.pause).toHaveBeenCalledTimes(1);
    expect(bell.play).not.toHaveBeenCalled();
    expect(navigator.vibrate).toHaveBeenLastCalledWith(0);
  });

  it('gives up after two minutes', async () => {
    const { alarm, bell } = setup();
    alarm.sync(ring('a'));

    await vi.advanceTimersByTimeAsync(RING_CAP_MS - 1);
    expect(alarm.getSnapshot().ringing).toBe(true);

    await vi.advanceTimersByTimeAsync(1);
    expect(alarm.getSnapshot().ringing).toBe(false);

    bell.play.mockClear();
    await vi.advanceTimersByTimeAsync(BELL_WATCHDOG_MS * 2);
    expect(bell.play).not.toHaveBeenCalled();
  });

  it('falls back to a synthesised chime when the bell will not play', async () => {
    const { alarm, bell, oscillators } = setup();
    alarm.unlock();
    await flush();
    const afterUnlock = oscillators.length;
    bell.play.mockRejectedValue(new DOMException('blocked', 'NotAllowedError'));

    alarm.sync(ring('a'));
    await flush();

    expect(oscillators.length - afterUnlock).toBe(3);

    // The bell is tried again on the next round rather than written off.
    bell.play.mockClear();
    await vi.advanceTimersByTimeAsync(BELL_WATCHDOG_MS);
    expect(bell.play).toHaveBeenCalled();
  });

  it('does not vibrate before a tap has unlocked the page', () => {
    const { alarm, navigator } = setup();

    alarm.sync(ring('a'));

    expect(navigator.vibrate).not.toHaveBeenCalled();
  });

  it('flashes the page title and puts it back', async () => {
    const { alarm, document } = setup();

    alarm.sync(ring('a', 'Order 42 is ready'));
    expect(document.title).toBe('Order 42 is ready');

    await vi.advanceTimersByTimeAsync(TITLE_FLASH_MS);
    expect(document.title).toBe('Page');

    await vi.advanceTimersByTimeAsync(TITLE_FLASH_MS);
    expect(document.title).toBe('Order 42 is ready');

    alarm.silence();
    expect(document.title).toBe('Page');
  });
});

describe('screen wake lock', () => {
  it('holds the lock while asked to and lets go afterwards', async () => {
    const { alarm, navigator, wakeLocks } = setup();

    alarm.sync({ keepAwake: true, ringKey: null });
    await flush();
    alarm.sync({ keepAwake: true, ringKey: null });
    await flush();

    expect(navigator.wakeLock.request).toHaveBeenCalledTimes(1);
    expect(navigator.wakeLock.request).toHaveBeenCalledWith('screen');

    alarm.sync({ keepAwake: false, ringKey: null });

    expect(wakeLocks[0].release).toHaveBeenCalledTimes(1);
  });

  it('asks again when the page comes back to the foreground', async () => {
    const { alarm, navigator, wakeLocks, setVisibility } = setup();
    alarm.sync({ keepAwake: true, ringKey: null });
    await flush();

    setVisibility('hidden');
    wakeLocks[0].drop();
    await flush();
    expect(navigator.wakeLock.request).toHaveBeenCalledTimes(1);

    setVisibility('visible');
    await flush();
    expect(navigator.wakeLock.request).toHaveBeenCalledTimes(2);
  });

  it('releases the lock and stops the ring when the screen leaves', async () => {
    const { alarm, navigator, wakeLocks, setVisibility } = setup();
    alarm.sync(ring('a'));
    await flush();

    alarm.release();

    expect(alarm.getSnapshot().ringing).toBe(false);
    expect(wakeLocks[0].release).toHaveBeenCalledTimes(1);

    setVisibility('visible');
    await flush();
    expect(navigator.wakeLock.request).toHaveBeenCalledTimes(1);
  });

  it('survives a refused lock and asks again only when the page is looked at again', async () => {
    const { alarm, navigator, setVisibility } = setup();
    navigator.wakeLock.request.mockRejectedValueOnce(new DOMException('low battery'));

    alarm.sync({ keepAwake: true, ringKey: null });
    await flush();
    alarm.sync({ keepAwake: true, ringKey: null });
    await flush();
    expect(navigator.wakeLock.request).toHaveBeenCalledTimes(1);

    setVisibility('visible');
    await flush();
    expect(navigator.wakeLock.request).toHaveBeenCalledTimes(2);
  });
});
