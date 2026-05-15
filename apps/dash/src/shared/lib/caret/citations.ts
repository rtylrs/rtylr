/**
 * Citation extraction for Caret responses. Caret formats internal links as
 *   [source: orders/123]
 * inline; this helper pulls them out so we can render dedicated chips
 * below the message and let users navigate.
 */

export interface Citation {
  id: string;
  label: string;
  href: string;
}

const CITATION_RE = /\[source:\s*([a-z0-9/_-]+(?:\s*[0-9]+)?)\]/gi;

const ROUTE_PREFIX: Record<string, string> = {
  orders: '/orders/',
  customers: '/customers/',
  products: '/products/',
  invoices: '/finance/invoices/',
  tasks: '/flow/tasks/',
};

export function extractCitations(text: string): Citation[] {
  const out: Citation[] = [];
  let m: RegExpExecArray | null;
  let i = 0;
  while ((m = CITATION_RE.exec(text)) !== null) {
    const raw = m[1].trim();
    const [kind, id] = raw.split(/[\s/]/);
    const prefix = ROUTE_PREFIX[kind];
    if (!prefix || !id) {
      i++;
      continue;
    }
    out.push({
      id: `${kind}-${id}-${i}`,
      label: `${kind} ${id}`,
      href: prefix + id,
    });
    i++;
  }
  return out;
}

/** Returns the message text with `[source: …]` chunks removed. */
export function stripCitations(text: string): string {
  return text
    .replace(CITATION_RE, '')
    .replace(/\s{2,}/g, ' ')
    .trim();
}
