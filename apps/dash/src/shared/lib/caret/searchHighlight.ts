import { escapeHtml } from './escapeHtml';
import { highlight } from './highlight';

/** Escape user content + wrap matches - safe to dangerouslySetInnerHTML. */
export function safeHighlight(text: string, query: string): string {
  return highlight(escapeHtml(text), query);
}
