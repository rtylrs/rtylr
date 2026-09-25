import { describe, expect, it } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';
import en from './en/auth.json';
import ar from './ar/auth.json';
import fr from './fr/auth.json';

/**
 * Auth had no locale test, which is how a whole screen stayed in English.
 *
 * `/verify/$token`, the page the verification email links to, is the most
 * travelled path in this app and `EmailVerification.tsx` used **no `t()` at
 * all**: "Email verified", "Verification failed", "Continue to sign in",
 * "Back to sign in" and "Verifying your email..." were literals, while
 * `/verify` beside it was fully translated. Four more literals sat in the
 * OAuth consent screen, the reset-password confirm screen and the register
 * form's duplicate-email error.
 *
 * An English default hides this: only Arabic and French readers ever see it,
 * and nothing failed. So both halves are checked here, the keys and the source.
 */
type Group = Record<string, string>;

const groups = Object.keys(en) as Array<keyof typeof en>;

describe('auth translations', () => {
  it('has the same keys in every language', () => {
    const missing: string[] = [];

    for (const group of groups) {
      for (const key of Object.keys(en[group] as Group)) {
        if (!(key in ((ar[group] ?? {}) as Group))) missing.push(`ar: ${group}.${key}`);
        if (!(key in ((fr[group] ?? {}) as Group))) missing.push(`fr: ${group}.${key}`);
      }
    }

    expect(missing, 'add the key to every locale, not just en').toEqual([]);
  });

  it('does not leave a string in English in Arabic or French', () => {
    // A copied English value is the same defect as a missing key: the reader
    // sees English either way, and a key-presence check alone passes.
    const untranslated: string[] = [];

    for (const group of groups) {
      for (const [key, value] of Object.entries(en[group] as Group)) {
        // Values that are legitimately identical across languages.
        //
        // A numeric placeholder ("123456"), a brand name, and a **format
        // example**: an email address is Latin script whatever the reader's
        // language, so `you@company.com` staying put in Arabic is a defensible
        // choice rather than an untranslated string.
        //
        // Worth flagging separately, though: French did translate it, to
        // `vous@entreprise.com`. One of the two is wrong and it is a copy
        // decision, not a code one, so this skips the shape and leaves the
        // question rather than silently picking a side.
        if (/^[0-9]+$/.test(value)) continue;
        if (['google', 'github', 'microsoft', 'apple'].includes(key)) continue;
        if (/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) continue;
        if (/^https?:\/\//.test(value)) continue;

        if (((ar[group] ?? {}) as Group)[key] === value) untranslated.push(`ar: ${group}.${key}`);
        if (((fr[group] ?? {}) as Group)[key] === value) untranslated.push(`fr: ${group}.${key}`);
      }
    }

    expect(untranslated, 'translate the value, do not copy the English').toEqual([]);
  });

  it('leaves no user-facing English literal in the source', () => {
    const SRC = path.resolve(__dirname, '..');

    const walk = (dir: string): string[] =>
      fs.readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
        const full = path.join(dir, entry.name);
        if (entry.isDirectory()) {
          return entry.name === 'locales' || entry.name === 'node_modules' ? [] : walk(full);
        }
        return /\.tsx$/.test(entry.name) && !/\.test\.tsx$/.test(entry.name) ? [full] : [];
      });

    const offenders: string[] = [];

    for (const file of walk(SRC)) {
      const lines = fs.readFileSync(file, 'utf8').split('\n');
      lines.forEach((line, i) => {
        const trimmed = line.trim();
        if (trimmed.startsWith('*') || trimmed.startsWith('//') || trimmed.startsWith('{/*')) {
          return;
        }
        // Text between JSX tags: a capitalised word followed by at least one
        // more word. One word on its own is too often a component name or an
        // enum value to flag usefully.
        if (!/>\s*[A-Z][a-z]+(\s+[a-z']+){1,8}[.?!]?\s*</.test(line)) return;
        if (line.includes('{t(')) return;
        offenders.push(`${path.relative(SRC, file)}:${i + 1}`);
      });
    }

    expect(offenders, 'put the string through t()').toEqual([]);
  });
});
