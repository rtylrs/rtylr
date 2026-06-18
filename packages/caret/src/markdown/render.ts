import { parseBlocks, renderListItems, renderCodeBlock, renderTable } from './block';
import { renderInlineMarkdown } from './inline';

/** Render a complete markdown document to safe HTML. */
export function renderMarkdown(src: string): string {
  const blocks = parseBlocks(src);
  const out: string[] = [];
  for (const b of blocks) {
    switch (b.type) {
      case 'heading': {
        const lvl = Math.min(Math.max(b.level ?? 2, 1), 6);
        out.push(`<h${lvl} class="caret-h">${renderInlineMarkdown(b.raw)}</h${lvl}>`);
        break;
      }
      case 'paragraph':
        out.push(`<p class="caret-p">${renderInlineMarkdown(b.raw)}</p>`);
        break;
      case 'list': {
        const tag = b.listType === 'ordered' ? 'ol' : 'ul';
        out.push(`<${tag} class="caret-list">${renderListItems(b.raw)}</${tag}>`);
        break;
      }
      case 'quote':
        out.push(`<blockquote class="caret-quote">${renderInlineMarkdown(b.raw)}</blockquote>`);
        break;
      case 'code':
        out.push(renderCodeBlock(b.raw, b.language));
        break;
      case 'table':
        out.push(renderTable(b.raw));
        break;
      case 'hr':
        out.push('<hr class="caret-hr" />');
        break;
    }
  }
  return out.join('');
}
