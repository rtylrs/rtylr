/**
 * Tiny audio-feedback helpers for chat events. Uses the Web Audio API
 * directly so we don't ship an audio file. Sounds are intentionally short
 * and quiet; they noop if the user has not interacted with the page yet.
 */

let ctx: AudioContext | null = null;

function getContext(): AudioContext | null {
  if (typeof window === 'undefined') return null;
  const w = window as unknown as {
    AudioContext?: typeof AudioContext;
    webkitAudioContext?: typeof AudioContext;
  };
  const Ctor = w.AudioContext ?? w.webkitAudioContext;
  if (!Ctor) return null;
  try {
    if (!ctx) ctx = new Ctor();
    return ctx;
  } catch {
    return null;
  }
}

function tone(frequency: number, durationMs: number, gain = 0.04): void {
  const c = getContext();
  if (!c) return;
  try {
    const osc = c.createOscillator();
    const g = c.createGain();
    osc.type = 'sine';
    osc.frequency.value = frequency;
    g.gain.value = gain;
    osc.connect(g);
    g.connect(c.destination);
    osc.start();
    g.gain.exponentialRampToValueAtTime(0.0001, c.currentTime + durationMs / 1000);
    osc.stop(c.currentTime + durationMs / 1000);
  } catch {
    // ignore — user hasn't interacted yet, or audio is disabled
  }
}

export function playSendSound(): void {
  tone(660, 60);
}

export function playReceiveSound(): void {
  tone(440, 60);
}

export function playErrorSound(): void {
  tone(180, 140, 0.05);
}
