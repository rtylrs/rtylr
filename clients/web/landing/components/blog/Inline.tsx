import Link from 'next/link';
import React from 'react';

/**
 * Renders the two inline forms post text may carry: [label](href) and
 * **strong**. Anything else is plain text, so a stray bracket in prose
 * renders as itself rather than as broken markup.
 */
const TOKEN = /\[([^\]]+)\]\(([^)\s]+)\)|\*\*([^*]+)\*\*/g;

export function Inline({ text }: { text: string }): React.JSX.Element {
  const out: React.ReactNode[] = [];
  let last = 0;
  let key = 0;
  for (const m of text.matchAll(TOKEN)) {
    const at = m.index ?? 0;
    if (at > last) out.push(text.slice(last, at));
    if (m[1] !== undefined) {
      const href = m[2];
      out.push(
        href.startsWith('/') ? (
          <Link key={key++} href={href}>
            {m[1]}
          </Link>
        ) : (
          <a key={key++} href={href} rel="noopener" target="_blank">
            {m[1]}
          </a>
        )
      );
    } else {
      out.push(<strong key={key++}>{m[3]}</strong>);
    }
    last = at + m[0].length;
  }
  if (last < text.length) out.push(text.slice(last));
  return <>{out}</>;
}

/** The same text with the markup removed, for meta tags and JSON-LD. */
export const plainText = (text: string): string =>
  text.replace(TOKEN, (_, label, _href, strong) => label ?? strong);
