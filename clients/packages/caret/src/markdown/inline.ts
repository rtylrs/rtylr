import { escapeHtml } from './escape';

/**
 * Render minimal inline markdown to safe HTML:
 *   **bold**  __bold__
 *   *italic*  _italic_
 *   `code`
 *   [link](http://...)
 *
 * Input must already be HTML-escaped via {@link escapeHtml} when invoked
 * standalone; callers in this module handle escaping themselves.
 */
export function renderInlineMarkdown(input: string): string {
  let s = escapeHtml(input);

  // inline code first so its content isn't formatted further
  s = s.replace(/`([^`]+)`/g, (_, code) => `<code class="caret-inline-code">${code}</code>`);

  // bold (** or __)
  s = s.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
  s = s.replace(/__([^_]+)__/g, '<strong>$1</strong>');

  // italics (* or _) - avoid eating bold markers (already replaced)
  s = s.replace(/(^|[^*])\*([^*\n]+)\*([^*]|$)/g, '$1<em>$2</em>$3');
  s = s.replace(/(^|[^_])_([^_\n]+)_([^_]|$)/g, '$1<em>$2</em>$3');

  // links [text](url) - only http(s) and relative paths allowed
  s = s.replace(/\[([^\]]+)\]\(((?:https?:\/\/|\/)[^\s)]+)\)/g, (_, text: string, url: string) => {
    return `<a href="${url}" target="_blank" rel="noopener noreferrer" class="caret-link">${text}</a>`;
  });

  return s;
}
