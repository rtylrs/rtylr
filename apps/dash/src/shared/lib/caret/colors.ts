/** Deterministic hash → hue for chat-avatar coloring. */
function hashString(input: string): number {
  let h = 5381;
  for (let i = 0; i < input.length; i++) {
    h = (h * 33) ^ input.charCodeAt(i);
  }
  return Math.abs(h);
}

/** Pick an HSL color in a fixed palette based on a stable hash of `seed`. */
export function paletteColor(seed: string): { bg: string; fg: string } {
  const palette: Array<{ bg: string; fg: string }> = [
    { bg: 'rgba(100,197,185,0.16)', fg: '#64c5b9' },
    { bg: 'rgba(14,165,233,0.16)', fg: '#38bdf8' },
    { bg: 'rgba(16,185,129,0.16)', fg: '#34d399' },
    { bg: 'rgba(245,158,11,0.16)', fg: '#fbbf24' },
    { bg: 'rgba(239,68,68,0.16)', fg: '#f87171' },
    { bg: 'rgba(168,85,247,0.16)', fg: '#c084fc' },
    { bg: 'rgba(99,102,241,0.16)', fg: '#818cf8' },
    { bg: 'rgba(34,211,238,0.16)', fg: '#67e8f9' },
  ];
  return palette[hashString(seed) % palette.length];
}
