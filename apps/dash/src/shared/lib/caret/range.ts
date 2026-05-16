/** Generate an inclusive integer range — handy for tests + virtual lists. */
export function range(start: number, end: number, step = 1): number[] {
  if (step === 0) return [];
  const out: number[] = [];
  if (step > 0) for (let i = start; i <= end; i += step) out.push(i);
  else for (let i = start; i >= end; i += step) out.push(i);
  return out;
}
