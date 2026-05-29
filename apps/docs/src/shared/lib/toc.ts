/** Table-of-contents parsing helpers for the docs app. */

export interface TocEntry {
  level: number;
  title: string;
  slug: string;
}

/**
 * Parses markdown headings into a flat TOC entry list.
 * Only h2 (##) and h3 (###) headings are extracted.
 */
export function parseToc(markdown: string): TocEntry[] {
  const lines = markdown.split('\n');
  const entries: TocEntry[] = [];
  for (const line of lines) {
    const match = line.match(/^(#{2,3})\s+(.+)$/);
    if (!match) continue;
    const level = match[1].length;
    const title = match[2].trim();
    const slug = title
      .toLowerCase()
      .replace(/[^\w\s-]/g, '')
      .replace(/\s+/g, '-');
    entries.push({ level, title, slug });
  }
  return entries;
}

/**
 * Filters a TOC to only include entries at or shallower than maxDepth.
 * Level 2 = h2, level 3 = h3.
 */
export function filterTocDepth(toc: TocEntry[], maxDepth: number): TocEntry[] {
  return toc.filter((e) => e.level <= maxDepth);
}

/**
 * Returns true if the TOC has at least one entry.
 */
export function hasToc(toc: TocEntry[]): boolean {
  return toc.length > 0;
}

/**
 * Returns the title of the first TOC entry, or undefined if empty.
 */
export function tocTitle(toc: TocEntry[]): string | undefined {
  return toc[0]?.title;
}
