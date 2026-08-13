import { renderInlineMarkdown } from './inline';
import { escapeHtml } from './escape';

export interface MarkdownBlock {
  type: 'paragraph' | 'heading' | 'list' | 'quote' | 'code' | 'table' | 'hr';
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

    if (isTableStart(lines, i)) {
      const buf: string[] = [line, lines[i + 1]];
      i += 2;
      while (
        i < lines.length &&
        !/^\s*$/.test(lines[i]) &&
        hasTablePipe(lines[i]) &&
        !lines[i].startsWith('```')
      ) {
        buf.push(lines[i]);
        i++;
      }
      blocks.push({ type: 'table', raw: buf.join('\n') });
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
      !isTableStart(lines, i) &&
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

function hasTablePipe(line: string): boolean {
  return line.includes('|');
}

function isTableStart(lines: string[], index: number): boolean {
  if (index + 1 >= lines.length) return false;
  const header = splitTableRow(lines[index]);
  const separator = splitTableRow(lines[index + 1]);
  return (
    header.some((cell) => cell.length > 0) &&
    separator.length > 0 &&
    separator.every(isTableSeparatorCell)
  );
}

function splitTableRow(line: string): string[] {
  const trimmed = line.trim();
  if (!trimmed.includes('|')) return [];
  const withoutOuterPipes = trimmed.replace(/^\|/, '').replace(/\|$/, '');
  return withoutOuterPipes.split('|').map((cell) => cell.trim());
}

function isTableSeparatorCell(cell: string): boolean {
  const normalized = cell.trim().replace(/\s+/g, '').replace(/[–—]/g, '-');
  return /^:?-{2,}:?$/.test(normalized);
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

export function renderTable(raw: string): string {
  const rows = raw
    .split('\n')
    .map(splitTableRow)
    .filter((row) => row.length > 0);
  if (rows.length < 2 || !rows[1].every(isTableSeparatorCell)) {
    return `<p class="caret-p">${renderInlineMarkdown(raw)}</p>`;
  }

  const headers = rows[0];
  const bodyRows = rows.slice(2);
  if (bodyRows.length === 0) {
    return '<p class="caret-table-empty">No matching records.</p>';
  }
  const columnCount = Math.max(headers.length, ...bodyRows.map((row) => row.length), 0);
  const normalizedHeaders = normalizeRow(headers, columnCount);
  const thead = normalizedHeaders.map((cell) => `<th>${renderInlineMarkdown(cell)}</th>`).join('');
  const tbody = bodyRows
    .map((row) => {
      const cells = normalizeRow(row, columnCount)
        .map((cell) => `<td>${renderInlineMarkdown(cell)}</td>`)
        .join('');
      return `<tr>${cells}</tr>`;
    })
    .join('');

  return `<div class="caret-table-wrap"><table class="caret-table"><thead><tr>${thead}</tr></thead><tbody>${tbody}</tbody></table></div>`;
}

function normalizeRow(row: string[], columnCount: number): string[] {
  const normalized = row.slice(0, columnCount);
  while (normalized.length < columnCount) normalized.push('');
  return normalized;
}
