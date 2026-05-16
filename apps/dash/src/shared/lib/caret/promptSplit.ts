/**
 * Tear a long prompt into ~N-character segments at sensible boundaries
 * (paragraph → sentence → space). Used for follow-up streaming so we can
 * show approximate progress before the LLM responds.
 */
export function splitPrompt(prompt: string, target = 400): string[] {
  if (prompt.length <= target) return [prompt];
  const out: string[] = [];
  let remaining = prompt;
  while (remaining.length > target) {
    const slice = remaining.slice(0, target);
    let cut = slice.lastIndexOf('\n\n');
    if (cut === -1) cut = slice.lastIndexOf('. ');
    if (cut === -1) cut = slice.lastIndexOf(' ');
    if (cut === -1) cut = target;
    out.push(remaining.slice(0, cut).trim());
    remaining = remaining.slice(cut).trim();
  }
  if (remaining) out.push(remaining);
  return out.filter(Boolean);
}
