/**
 * Rings a phone from a web page, later, after one earlier tap.
 *
 * Browsers only let a page make sound once the user has touched it, so
 * `unlock()` has to run synchronously inside a click handler, before any
 * await. It survives router navigation but not a document reload.
 *
 * No React in here: a screen feeds `sync()` from whatever it polls and reads
 * `getSnapshot()` through `useAlarm`.
 */

export const RING_CAP_MS = 120_000;
export const RING_GAP_MS = 700;
/** How long to wait for the bell's `ended` event before moving on anyway. */
export const BELL_WATCHDOG_MS = 4_000;
export const TITLE_FLASH_MS = 1_000;
export const VIBRATION_PATTERN = [280, 120, 280, 120, 520];

const CHIME_NOTES_HZ = [1046.5, 1318.5, 1568];
const CHIME_NOTE_SPACING_S = 0.18;
const CHIME_NOTE_LENGTH_S = 0.5;
const CHIME_MS = 900;

export interface AlarmState {
  /** Hold a screen wake lock: the page is waiting for, or showing, an alert. */
  keepAwake: boolean;
  /**
   * Identifies one reason to ring, or null for silence. A key rings once: the
   * same key again is ignored even after the ring was silenced or timed out.
   */
  ringKey: string | null;
  /** Alternates with the page title while ringing. */
  title?: string;
}

export interface AlarmSnapshot {
  unlocked: boolean;
  ringing: boolean;
}

export interface AlarmEnvironment {
  createAudio: (src: string) => HTMLAudioElement;
  createAudioContext: () => AudioContext | null;
  navigator: Navigator;
  document: Document;
}

export interface AlarmOptions {
  bellSrc: string;
  env?: Partial<AlarmEnvironment>;
}

export interface Alarm {
  unlock: () => void;
  /** Short friendly confirmation tone right after the customer unlocks sound. */
  blip: () => void;
  sync: (state: AlarmState) => void;
  /** Stops the current ring on this device only. */
  silence: () => void;
  /** Stops everything and forgets what was rung; call when the screen leaves. */
  release: () => void;
  subscribe: (listener: () => void) => () => void;
  getSnapshot: () => AlarmSnapshot;
}

type AudioSessionNavigator = Navigator & { audioSession?: { type: string } };
type WebkitWindow = Window & { webkitAudioContext?: typeof AudioContext };

const noop = (): void => {};

function defaultAudioContext(): AudioContext | null {
  const Ctor = window.AudioContext ?? (window as WebkitWindow).webkitAudioContext;
  return Ctor ? new Ctor() : null;
}

export function createAlarm(options: AlarmOptions): Alarm {
  const env = options.env ?? {};
  const nav = (): Navigator => env.navigator ?? navigator;
  const doc = (): Document => env.document ?? document;
  const createAudio = env.createAudio ?? ((src: string) => new Audio(src));
  const createAudioContext = env.createAudioContext ?? defaultAudioContext;

  let snapshot: AlarmSnapshot = { unlocked: false, ringing: false };
  const listeners = new Set<() => void>();
  const update = (patch: Partial<AlarmSnapshot>): void => {
    snapshot = { ...snapshot, ...patch };
    listeners.forEach((listener) => listener());
  };

  let bell: HTMLAudioElement | null = null;
  let context: AudioContext | null = null;
  let chimeVoices: OscillatorNode[] = [];
  let lastRingKey: string | null = null;
  // Bumped whenever a bell cycle is superseded, so its late callbacks no-op.
  let cycle = 0;
  let cycleTimer: ReturnType<typeof setTimeout> | undefined;
  let capTimer: ReturnType<typeof setTimeout> | undefined;
  let titleTimer: ReturnType<typeof setInterval> | undefined;
  let flashTitle = '';
  let restoreTitle: string | null = null;

  let keepAwake = false;
  let wakeLock: WakeLockSentinel | null = null;
  let acquiringWakeLock = false;

  function resumeContext(): void {
    if (context && context.state !== 'running') void context.resume().catch(noop);
  }

  function unlock(): void {
    try {
      // Lets an iPhone play through the speaker with the silent switch on.
      const session = (nav() as AudioSessionNavigator).audioSession;
      if (session) session.type = 'playback';
    } catch {
      // Not every browser lets the type be set.
    }

    try {
      bell ??= createAudio(options.bellSrc);
      bell.preload = 'auto';
      // While ringing, the restart below plays the bell out loud instead.
      if (!snapshot.ringing) {
        const primed = bell;
        const unmute = (): void => {
          primed.muted = false;
        };
        primed.muted = true;
        void primed.play()?.then(() => {
          if (!snapshot.ringing) {
            primed.pause();
            primed.currentTime = 0;
          }
          unmute();
        }, unmute);
      }
    } catch {
      // The chime below is the fallback.
    }

    try {
      context ??= createAudioContext();
      if (context) {
        resumeContext();
        const oscillator = context.createOscillator();
        const gain = context.createGain();
        gain.gain.value = 0.0001;
        oscillator.connect(gain);
        gain.connect(context.destination);
        oscillator.start();
        oscillator.stop(context.currentTime + 0.02);
      }
    } catch {
      // No Web Audio: the bell element is all there is.
    }

    update({ unlocked: true });
    if (snapshot.ringing) startCycle();
  }

  function blip(): void {
    try {
      context ??= createAudioContext();
      if (!context) return;
      resumeContext();
      const ctx = context;
      [880, 1318.5].forEach((hz, index) => {
        const at = ctx.currentTime + index * 0.09;
        const oscillator = ctx.createOscillator();
        const gain = ctx.createGain();
        oscillator.type = 'sine';
        oscillator.frequency.value = hz;
        gain.gain.setValueAtTime(0.0001, at);
        gain.gain.exponentialRampToValueAtTime(0.2, at + 0.015);
        gain.gain.exponentialRampToValueAtTime(0.0001, at + 0.16);
        oscillator.connect(gain);
        gain.connect(ctx.destination);
        oscillator.start(at);
        oscillator.stop(at + 0.17);
      });
    } catch {
      // Audible confirmation is a bonus.
    }
  }

  function playChime(): void {
    if (!context) return;
    try {
      resumeContext();
      const ctx = context;
      CHIME_NOTES_HZ.forEach((hz, index) => {
        const at = ctx.currentTime + index * CHIME_NOTE_SPACING_S;
        const oscillator = ctx.createOscillator();
        const gain = ctx.createGain();
        oscillator.type = 'triangle';
        oscillator.frequency.value = hz;
        gain.gain.setValueAtTime(0.0001, at);
        gain.gain.exponentialRampToValueAtTime(0.9, at + 0.01);
        gain.gain.exponentialRampToValueAtTime(0.0001, at + CHIME_NOTE_LENGTH_S);
        oscillator.connect(gain);
        gain.connect(ctx.destination);
        oscillator.start(at);
        oscillator.stop(at + CHIME_NOTE_LENGTH_S);
        chimeVoices.push(oscillator);
      });
    } catch {
      // Nothing left to fall back to; the screen and the title still show it.
    }
  }

  function stopChime(): void {
    for (const oscillator of chimeVoices) {
      try {
        oscillator.stop();
      } catch {
        // Already finished.
      }
    }
    chimeVoices = [];
  }

  function startCycle(): void {
    const token = ++cycle;
    const next = (delayMs: number): void => {
      if (token !== cycle) return;
      clearTimeout(cycleTimer);
      cycleTimer = setTimeout(() => {
        if (token === cycle && snapshot.ringing) startCycle();
      }, delayMs);
    };
    const chime = (): void => {
      if (token !== cycle) return;
      stopChime();
      playChime();
      next(CHIME_MS + RING_GAP_MS);
    };

    // Vibration needs a prior tap too; without one Chrome only logs a warning.
    if (snapshot.unlocked) {
      try {
        const vibrator = nav();
        if (typeof vibrator.vibrate === 'function') vibrator.vibrate(VIBRATION_PATTERN);
      } catch {
        // Vibration is a bonus.
      }
    }

    try {
      bell ??= createAudio(options.bellSrc);
      const audio = bell;
      audio.onended = () => next(RING_GAP_MS);
      audio.muted = false;
      audio.currentTime = 0;
      next(BELL_WATCHDOG_MS);
      void audio.play()?.catch(() => {
        if (token !== cycle) return;
        audio.onended = null;
        chime();
      });
    } catch {
      if (bell) bell.onended = null;
      chime();
    }
  }

  function startTitleFlash(): void {
    if (!flashTitle) return;
    const page = doc();
    const original = page.title;
    restoreTitle = original;
    let flashed = false;
    const tick = (): void => {
      flashed = !flashed;
      page.title = flashed ? flashTitle : original;
    };
    tick();
    titleTimer = setInterval(tick, TITLE_FLASH_MS);
  }

  function stopRing(): void {
    if (!snapshot.ringing) return;
    cycle += 1;
    clearTimeout(cycleTimer);
    clearTimeout(capTimer);
    clearInterval(titleTimer);
    if (restoreTitle !== null) {
      doc().title = restoreTitle;
      restoreTitle = null;
    }
    if (bell) {
      bell.onended = null;
      try {
        bell.pause();
        bell.currentTime = 0;
      } catch {
        // Never started.
      }
    }
    stopChime();
    if (snapshot.unlocked) {
      try {
        const vibrator = nav();
        if (typeof vibrator.vibrate === 'function') vibrator.vibrate(0);
      } catch {
        // Vibration is a bonus.
      }
    }
    update({ ringing: false });
  }

  function startRing(): void {
    update({ ringing: true });
    capTimer = setTimeout(stopRing, RING_CAP_MS);
    startTitleFlash();
    startCycle();
  }

  async function acquireWakeLock(): Promise<void> {
    const wakeLocks = nav().wakeLock;
    if (!wakeLocks || wakeLock || acquiringWakeLock || doc().visibilityState !== 'visible') return;
    acquiringWakeLock = true;
    try {
      const sentinel = await wakeLocks.request('screen');
      if (!keepAwake) {
        void sentinel.release().catch(noop);
        return;
      }
      wakeLock = sentinel;
      // The browser drops the lock whenever the page is hidden.
      sentinel.addEventListener('release', () => {
        if (wakeLock === sentinel) wakeLock = null;
      });
    } catch {
      // Refused: low battery, or the page went to the background meanwhile.
    } finally {
      acquiringWakeLock = false;
    }
  }

  const onVisibilityChange = (): void => {
    if (doc().visibilityState === 'visible') void acquireWakeLock();
  };

  // A refused lock is asked for again only when the page returns to the
  // foreground, not on every sync.
  function setKeepAwake(wanted: boolean): void {
    if (wanted === keepAwake) return;
    keepAwake = wanted;
    if (wanted) {
      doc().addEventListener('visibilitychange', onVisibilityChange);
      void acquireWakeLock();
      return;
    }
    doc().removeEventListener('visibilitychange', onVisibilityChange);
    if (wakeLock) {
      const held = wakeLock;
      wakeLock = null;
      void held.release().catch(noop);
    }
  }

  function sync(state: AlarmState): void {
    flashTitle = state.title ?? '';
    setKeepAwake(state.keepAwake);
    if (state.ringKey === null) {
      stopRing();
      return;
    }
    if (state.ringKey === lastRingKey) return;
    lastRingKey = state.ringKey;
    stopRing();
    startRing();
  }

  function release(): void {
    stopRing();
    setKeepAwake(false);
    lastRingKey = null;
  }

  return {
    unlock,
    blip,
    sync,
    silence: stopRing,
    release,
    subscribe: (listener) => {
      listeners.add(listener);
      return () => {
        listeners.delete(listener);
      };
    },
    getSnapshot: () => snapshot,
  };
}
