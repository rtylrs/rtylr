/** Count how many of the inputs are truthy. */
export function booleanCount(...inputs: unknown[]): number {
  let n = 0;
  for (const i of inputs) if (i) n++;
  return n;
}
