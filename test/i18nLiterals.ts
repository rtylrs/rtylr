import fs from 'node:fs';
import path from 'node:path';

/**
 * Shared machinery for the "no hardcoded English in the source" guard.
 *
 * Every rtylr app ships en, ar and fr, so a literal in JSX is not latent: an
 * Arabic or French reader sees English, and nothing fails, because English is
 * what the key would have returned anyway (rtylr#1266).
 *
 * Counting `>Text<` on one line finds a fraction of them, and the gap is not
 * small. On three recruit files where both numbers are known, that pattern
 * read 7, 33 and 56 where the real counts were 11, 47 and 74. Copy hides in
 * four places a markup sweep cannot see:
 *
 *   - text on its own line, which is where prettier puts anything too long to
 *     fit, and which is most of it
 *   - `toast.*` arguments, including the label on a confirmation action
 *   - `title=`, `alt=`, `aria-label=`, `placeholder=`, `label=`, `hint=`, the
 *     only text an icon-only button has
 *   - strings inside data objects: nav item labels, per-kind metadata, option
 *     lists. Four separate ones in recruit alone
 */

/**
 * Text between tags on one line.
 *
 * Only applied to `.tsx`, because in a `.ts` file this shape is a type
 * annotation: `): Promise<Result>` reads as a `>` followed by `Promise<`, and
 * matching it reports `Promise` as user-facing copy.
 */
const INLINE = />\s*([A-Z][A-Za-z][^<>{}]*?)\s*</;

/** A user-facing prop whose value is copy rather than a token. */
const PROP = /(?:placeholder|title|label|aria-label|alt|hint)="([A-Z][^"]{2,})"/;

/** A string handed to a toast or an error, which reaches the user directly. */
const CALL =
  /(?:toast\.(?:error|success|info|warning|message)|new Error)\(\s*['"]([A-Z][^'"]{3,})['"]/;

/** A line holding nothing but prose, between two tags. */
const TEXT_LINE = /^[A-Z][A-Za-z0-9 ,.'’!?&:/-]*$/;

const COMMENT_LINE = /^\s*(\/\/|\*|\/\*|\{\/\*)/;

export interface Literal {
  file: string;
  line: number;
  text: string;
}

function sourceFiles(dir: string, into: string[] = []): string[] {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      if (entry.name === 'node_modules' || entry.name === 'locales') continue;
      sourceFiles(full, into);
      continue;
    }
    if (/\.tsx?$/.test(entry.name) && !/\.test\.tsx?$/.test(entry.name)) into.push(full);
  }
  return into;
}

/**
 * Every hardcoded user-facing string under `srcDir`.
 *
 * `allowed` holds text that is deliberately not translated. Three kinds earn
 * a place there and nothing else should: a format example (`https://…`,
 * `MM-DD`), a key cap or other physical label (`ESC`), and a specimen that
 * must stay in the script it demonstrates, such as a pangram under a Latin
 * typeface picker, where translated text would render through a fallback and
 * misrepresent the very thing being previewed.
 */
export function collectLiterals(srcDir: string, allowed: readonly string[] = []): Literal[] {
  const exempt = new Set(allowed);
  const found: Literal[] = [];

  for (const file of sourceFiles(srcDir)) {
    const isJsx = file.endsWith('.tsx');
    const lines = fs.readFileSync(file, 'utf8').split('\n');
    lines.forEach((line, index) => {
      if (COMMENT_LINE.test(line.trim())) return;
      const hasCall = line.includes('t(');
      const inline = hasCall || !isJsx ? null : INLINE.exec(line);
      const prop = PROP.exec(line);
      const call = CALL.exec(line);

      let standalone: string | null = null;
      const trimmed = line.trim();
      if (isJsx && !hasCall && TEXT_LINE.test(trimmed) && trimmed.length > 2) {
        const prev =
          lines
            .slice(0, index)
            .reverse()
            .find((l) => l.trim() !== '')
            ?.trim() ?? '';
        const next =
          lines
            .slice(index + 1)
            .find((l) => l.trim() !== '')
            ?.trim() ?? '';
        if (/[>{}]$/.test(prev) && next.startsWith('<')) standalone = trimmed;
      }

      const text = call?.[1] ?? prop?.[1] ?? inline?.[1] ?? standalone;
      if (text === null || text === undefined || exempt.has(text)) return;
      found.push({ file: path.relative(srcDir, file), line: index + 1, text });
    });
  }
  return found;
}

/** `file:line  text` lines, for an assertion message that says where to look. */
export function formatLiterals(literals: Literal[]): string[] {
  return literals.map(({ file, line, text }) => `${file}:${line}  ${text}`);
}
