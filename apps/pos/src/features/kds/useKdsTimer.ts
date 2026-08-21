import { useEffect, useState, useSyncExternalStore } from 'react';

// A single shared 1s interval drives every subscriber, so N ticket cards
// cost one timer - not N timers - and each card can re-render independently.
type TickListener = () => void;
const listeners = new Set<TickListener>();
let intervalId: ReturnType<typeof setInterval> | null = null;
let currentTick = 0;

function subscribeTick(listener: TickListener): () => void {
  listeners.add(listener);
  if (intervalId === null) {
    intervalId = setInterval(() => {
      currentTick += 1;
      for (const listener of listeners) listener();
    }, 1000);
  }
  return () => {
    listeners.delete(listener);
    if (listeners.size === 0 && intervalId !== null) {
      clearInterval(intervalId);
      intervalId = null;
    }
  };
}

function getTickSnapshot(): number {
  return currentTick;
}

/**
 * Returns a tick counter that increments every second.
 * Components use this to trigger elapsed-time recalculations without
 * storing per-order timers - one shared counter re-renders subscribers.
 */
export function useKdsTick(): number {
  return useSyncExternalStore(subscribeTick, getTickSnapshot);
}

/** Format elapsed seconds as MM:SS, turning amber at 5 min and red at 10 min. */
export function formatElapsed(seconds: number): { label: string; urgency: 'ok' | 'warn' | 'crit' } {
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  const label = `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
  const urgency = seconds >= 600 ? 'crit' : seconds >= 300 ? 'warn' : 'ok';
  return { label, urgency };
}

export function elapsedSeconds(createdAt: string): number {
  return Math.max(0, Math.floor((Date.now() - new Date(createdAt).getTime()) / 1000));
}
