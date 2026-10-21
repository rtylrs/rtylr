import { common, createLowlight } from "lowlight";

/** The parts of lowlight's syntax tree that rendering needs. */
export type HighlightNode =
  | { type: "text"; value: string }
  | {
      type: "element";
      properties?: { className?: string[] };
      children: HighlightNode[];
    };

/** Blocks past this size stay plain; highlighting reruns as a reply streams. */
const MAX_HIGHLIGHT_CHARS = 20_000;

const lowlight = createLowlight(common);

/**
 * Syntax-highlights `code` in `language` (a name or alias such as "py" or
 * "ts"). Returns null for an unknown language or an oversized block, so the
 * caller shows plain text.
 */
export function highlightCode(
  code: string,
  language: string,
): HighlightNode[] | null {
  const name = language.toLowerCase();
  if (code.length > MAX_HIGHLIGHT_CHARS || !lowlight.registered(name))
    return null;
  try {
    return lowlight.highlight(name, code).children as HighlightNode[];
  } catch {
    return null;
  }
}
