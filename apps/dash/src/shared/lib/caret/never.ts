/** Exhaustiveness assertion. Use in switch's default branch. */
export function assertNever(x: never): never {
  throw new Error(`Unreachable: ${JSON.stringify(x)}`);
}
