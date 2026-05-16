/**
 * Extract `@app` mentions from a composer string. Used to scope a question
 * to a specific app or feature (e.g. "@crm find john").
 */
export interface Mention {
  app: string;
  start: number;
  end: number;
}

const MENTION_RE = /(^|\s)@([a-z][a-z0-9_-]{1,20})\b/gi;

export function parseMentions(text: string): Mention[] {
  const out: Mention[] = [];
  let m: RegExpExecArray | null;
  while ((m = MENTION_RE.exec(text)) !== null) {
    out.push({
      app: m[2].toLowerCase(),
      start: m.index + (m[1] ? m[1].length : 0),
      end: m.index + m[0].length,
    });
  }
  return out;
}
