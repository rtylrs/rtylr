/** Generate a numeric sequence starting at 0. */
export function sequence(length: number): number[] {
  if (length <= 0) return [];
  const out = new Array<number>(length);
  for (let i = 0; i < length; i++) out[i] = i;
  return out;
}
