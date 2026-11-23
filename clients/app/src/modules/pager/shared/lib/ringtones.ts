import bellSrc from '../../../../shared/alarm/bell.wav';
import beepSrc from '../../../../shared/alarm/tones/beep.wav';
import chimeSrc from '../../../../shared/alarm/tones/chime.wav';
import marimbaSrc from '../../../../shared/alarm/tones/marimba.wav';
import ringSrc from '../../../../shared/alarm/tones/ring.wav';

/** The sounds a ready order can ring with; the server accepts these names. */
export const RINGTONES = [
  { id: 'bell', src: bellSrc },
  { id: 'chime', src: chimeSrc },
  { id: 'marimba', src: marimbaSrc },
  { id: 'ring', src: ringSrc },
  { id: 'beep', src: beepSrc },
] as const;

export type RingtoneId = (typeof RINGTONES)[number]['id'];

export const DEFAULT_RINGTONE: RingtoneId = 'bell';

export function asRingtone(id: string | null | undefined): RingtoneId {
  return RINGTONES.find((tone) => tone.id === id)?.id ?? DEFAULT_RINGTONE;
}

export function ringtoneSrc(id: string | null | undefined): string {
  const name = asRingtone(id);
  return RINGTONES.find((tone) => tone.id === name)?.src ?? bellSrc;
}

const STORAGE_KEY = 'pager:ringtone';

/** The tone this phone's customer picked, kept across venues; null when none. */
export function rememberedRingtone(): RingtoneId | null {
  try {
    const stored = window.localStorage.getItem(STORAGE_KEY);
    return stored ? asRingtone(stored) : null;
  } catch {
    return null;
  }
}

export function rememberRingtone(id: RingtoneId): void {
  try {
    window.localStorage.setItem(STORAGE_KEY, id);
  } catch {
    // Without storage the pick lasts as long as the page.
  }
}

/** Plays a tone once, from the tap that chose it. */
export function previewRingtone(id: RingtoneId): void {
  try {
    void new Audio(ringtoneSrc(id)).play()?.catch(() => undefined);
  } catch {
    // No audio here; choosing still works.
  }
}
