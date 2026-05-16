/**
 * Parse a leading slash-command from the composer (e.g. "/report sales").
 * Returns the command name and the rest of the prompt.
 */
export interface SlashCommand {
  command: string;
  args: string;
}

export function parseSlash(input: string): SlashCommand | null {
  const m = /^\/([a-z][a-z0-9_-]{0,20})(?:\s+(.*))?$/i.exec(input.trimStart());
  if (!m) return null;
  return { command: m[1].toLowerCase(), args: m[2] ?? '' };
}
