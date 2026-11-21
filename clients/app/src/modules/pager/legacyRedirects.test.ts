import { readFileSync } from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';

/**
 * Printed counter posters point at portal.rtylr.com/<slug>/pager for as long
 * as they hang in a venue, so the Portal's Cloudflare Pages `_redirects` must
 * keep sending them to the venue on Rennili. This applies the rules the way
 * Pages does: first match wins, `:name` takes one segment, `*` the rest.
 */

interface Rule {
  from: string;
  to: string;
  status: number;
}

function readRules(): Rule[] {
  const file = path.resolve(__dirname, '../../../public/_redirects');
  return readFileSync(file, 'utf8')
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => line !== '' && !line.startsWith('#'))
    .map((line) => {
      const [from, to, status] = line.split(/\s+/);
      return { from, to, status: Number(status) };
    });
}

function resolve(rules: Rule[], pathname: string): { to: string; status: number } | null {
  for (const rule of rules) {
    const names: string[] = [];
    const pattern = rule.from
      .split('/')
      .map((part) => {
        if (part === '*') return '(.*)';
        if (part.startsWith(':')) {
          names.push(part.slice(1));
          return '([^/]+)';
        }
        return part.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
      })
      .join('/');
    const match = new RegExp(`^${pattern}$`).exec(pathname);
    if (!match) continue;
    let to = rule.to;
    names.forEach((name, index) => {
      to = to.replace(`:${name}`, match[index + 1]);
    });
    return { to, status: rule.status };
  }
  return null;
}

describe('Portal redirects for the old pager URLs', () => {
  const rules = readRules();

  it.each([
    ['/acme/pager', 'https://rennili.rtylr.com/acme'],
    ['/cafe-du-port/pager', 'https://rennili.rtylr.com/cafe-du-port'],
    ['/pager/t/tok_123', 'https://rennili.rtylr.com/t/tok_123'],
    ['/pager', 'https://rennili.rtylr.com/board'],
    ['/pager/display', 'https://rennili.rtylr.com/display'],
    ['/pager/poster', 'https://rennili.rtylr.com/poster'],
    ['/pager/setup', 'https://rennili.rtylr.com/setup'],
    ['/pager/try', 'https://rennili.rtylr.com/try'],
  ])('sends %s to %s', (from, to) => {
    expect(resolve(rules, from)).toEqual({ to, status: 302 });
  });

  it.each(['/', '/flow', '/acme/menu', '/settings/pos', '/pagers'])(
    'leaves the Portal route %s to the app',
    (from) => {
      expect(resolve(rules, from)).toEqual({ to: '/index.html', status: 200 });
    }
  );
});
