/** Map a string seed to a tiny 4×4 b/w identicon as inline SVG. */
function hash(s: string): number {
  let h = 5381;
  for (let i = 0; i < s.length; i++) h = (h * 33) ^ s.charCodeAt(i);
  return Math.abs(h);
}

export function identicon(seed: string, size = 24, color = '#64c5b9'): string {
  const bits = hash(seed).toString(2).padStart(16, '0').slice(-16);
  const rects: string[] = [];
  for (let i = 0; i < 16; i++) {
    if (bits[i] === '0') continue;
    const col = i % 4;
    const row = Math.floor(i / 4);
    const cellSize = size / 4;
    rects.push(
      `<rect x="${col * cellSize}" y="${row * cellSize}" width="${cellSize}" height="${cellSize}" fill="${color}"/>`
    );
  }
  return `<svg width="${size}" height="${size}" viewBox="0 0 ${size} ${size}" xmlns="http://www.w3.org/2000/svg">${rects.join('')}</svg>`;
}
