/**
 * Per-prompt cooldown ring used so accidental double-Enter doesn't fire
 * the same prompt twice in quick succession.
 */
const cooldown = new Map<string, number>();

const MIN_INTERVAL = 800;

export function canSend(prompt: string): boolean {
  const now = Date.now();
  const last = cooldown.get(prompt);
  if (last && now - last < MIN_INTERVAL) return false;
  cooldown.set(prompt, now);
  return true;
}

export function resetCooldown(): void {
  cooldown.clear();
}
