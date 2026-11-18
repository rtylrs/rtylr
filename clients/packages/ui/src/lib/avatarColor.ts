/**
 * A person's avatar colour, chosen from their id so it is the same on every
 * screen and every device, and different from most of their colleagues'.
 *
 * Twelve solid colours, each dark enough for white initials to pass WCAG AA
 * (4.5:1) at the small sizes avatars use. Solid rather than a tint, so the
 * circle reads the same on light and dark surfaces.
 */
export interface AvatarColor {
  background: string;
  foreground: string;
}

export const AVATAR_PALETTE: readonly string[] = [
  "#B42318", // red
  "#C4320A", // orange
  "#A15C07", // amber
  "#4D7C0F", // lime
  "#027A48", // green
  "#0E7090", // teal
  "#1D4ED8", // blue
  "#4338CA", // indigo
  "#6D28D9", // violet
  "#A21CAF", // fuchsia
  "#BE185D", // pink
  "#475467", // slate
];

/** FNV-1a: tiny, stable across runs and engines, and well spread for short ids. */
function hash(value: string): number {
  let h = 0x811c9dc5;
  for (let i = 0; i < value.length; i += 1) {
    h ^= value.charCodeAt(i);
    h = Math.imul(h, 0x01000193);
  }
  return h >>> 0;
}

/** The palette index for a seed (a user id, ideally). */
export function avatarColorIndex(seed: string): number {
  return hash(seed.trim().toLowerCase()) % AVATAR_PALETTE.length;
}

export function avatarColor(seed: string): AvatarColor {
  return {
    background: AVATAR_PALETTE[avatarColorIndex(seed)],
    foreground: "#FFFFFF",
  };
}
