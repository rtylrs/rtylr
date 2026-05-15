import { renderInlineMarkdown } from './inline';
import { escapeHtml } from './escape';

export interface MarkdownBlock {
  type: 'paragraph' | 'heading' | 'list' | 'quote' | 'code' | 'hr';
  /** Heading level 1..6 */
  level?: number;
  /** For lists: 'ordered' or 'unordered'. */
  listType?: 'ordered' | 'unordered';
  /** For code: the fenced language label, lowercased. */
  language?: string;
  /** Raw text content of the block (un-rendered). */
  raw: string;
}

/**
 * Tokenize streaming markdown into structural blocks. Designed to be fast,
 * incremental-friendly (idempotent on the same input) and not depend on any
 * remote package. Tolerant of partial input from an SSE stream.
 */
export function parseBlocks(src: string): MarkdownBlock[] {
  const lines = src.split('\n');
  const blocks: MarkdownBlock[] = [];
  let i = 0;
  while (i < lines.length) {
    const line = lines[i];

    if (/^\s*$/.test(line)) {
      i++;
      continue;
    }

    if (/^---+\s*$/.test(line)) {
      blocks.push({ type: 'hr', raw: '' });
      i++;
      continue;
    }

    const heading = /^(#{1,6})\s+(.*)$/.exec(line);
    if (heading) {
      blocks.push({ type: 'heading', level: heading[1].length, raw: heading[2].trim() });
      i++;
      continue;
    }

    if (line.startsWith('```')) {
      const lang = line.slice(3).trim().toLowerCase();
      const buf: string[] = [];
      i++;
      while (i < lines.length && !lines[i].startsWith('```')) {
        buf.push(lines[i]);
        i++;
      }
      if (i < lines.length) i++; // consume closing fence
      blocks.push({ type: 'code', language: lang || undefined, raw: buf.join('\n') });
      continue;
    }

    if (/^\s*>/.test(line)) {
      const buf: string[] = [];
      while (i < lines.length && /^\s*>/.test(lines[i])) {
        buf.push(lines[i].replace(/^\s*>\s?/, ''));
        i++;
      }
      blocks.push({ type: 'quote', raw: buf.join('\n') });
      continue;
    }

    if (/^\s*([-*+]|\d+\.)\s+/.test(line)) {
      const ordered = /^\s*\d+\.\s+/.test(line);
      const buf: string[] = [];
      while (i < lines.length && /^\s*([-*+]|\d+\.)\s+/.test(lines[i])) {
        buf.push(lines[i].replace(/^\s*([-*+]|\d+\.)\s+/, ''));
        i++;
      }
      blocks.push({
        type: 'list',
        listType: ordered ? 'ordered' : 'unordered',
        raw: buf.join('\n'),
      });
      continue;
    }

    // paragraph: collect contiguous non-blank lines that aren't another block start
    const buf: string[] = [line];
    i++;
    while (
      i < lines.length &&
      !/^\s*$/.test(lines[i]) &&
      !lines[i].startsWith('```') &&
      !/^#{1,6}\s+/.test(lines[i]) &&
      !/^\s*>/.test(lines[i]) &&
      !/^\s*([-*+]|\d+\.)\s+/.test(lines[i])
    ) {
      buf.push(lines[i]);
      i++;
    }
    blocks.push({ type: 'paragraph', raw: buf.join('\n') });
  }
  return blocks;
}

/** Render a list block's raw lines into <ul>/<ol> HTML. */
export function renderListItems(raw: string): string {
  return raw
    .split('\n')
    .filter((l) => l.trim().length > 0)
    .map((item) => `<li>${renderInlineMarkdown(item)}</li>`)
    .join('');
}

/** Render a code block to HTML. Syntax highlighting is intentionally
 *  delegated to a CSS layer so the markdown module stays dependency-free. */
export function renderCodeBlock(raw: string, language?: string): string {
  const escaped = escapeHtml(raw);
  const langAttr = language ? ` data-lang="${escapeHtml(language)}"` : '';
  return `<pre class="caret-codeblock"${langAttr}><code>${escaped}</code></pre>`;
}
