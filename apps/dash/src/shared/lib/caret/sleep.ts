/** Resolve after `ms` milliseconds. Use sparingly — prefer abort + signals. */
export function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
