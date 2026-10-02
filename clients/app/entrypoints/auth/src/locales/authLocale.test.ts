import { describe, expect, it } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';
import en from './en/auth.json';
import ar from './ar/auth.json';
import fr from './fr/auth.json';
import {
  isVerificationToken,
  normalizeVerificationToken,
  VERIFICATION_TOKEN_MAX_LENGTH,
} from '../features/email-verification/verificationToken';

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
 *
 * A second failure mode is below: the verification copy stayed pinned to a
 * six-digit code after the server moved to an account-bound token, and every
 * check here still passed. Shape claims are therefore checked against the
 * validator the field itself runs.
 */
type Group = Record<string, string>;

const groups = Object.keys(en) as Array<keyof typeof en>;
const authLocales: Record<string, { verify: Group }> = { en, fr, ar };

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
        // A verification-token sample is digits and a separator, which no
        // language translates: the shape is the whole point of the string. The
        // exemption is narrower than the numeric one above, because it only
        // holds for a value the field's own validator accepts — a placeholder
        // that drifts out of shape is reported by the shape test below instead
        // of being quietly excused here.
        if (isVerificationToken(value)) continue;
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

/**
 * How many code digits the field takes, asked of the validator the component
 * runs rather than written down here: a hardcoded ten would sit here passing
 * through the next format change exactly as this one did.
 *
 * Scans for the length that is accepted rather than stopping at the first one
 * that is refused — the field wants a fixed width, so the first failure is at
 * one digit, not at the boundary.
 */
function acceptedCodeDigits(): number {
  for (let n = 1; n <= VERIFICATION_TOKEN_MAX_LENGTH; n += 1) {
    if (isVerificationToken(`1.${'0'.repeat(n)}`)) return n;
  }
  throw new Error(
    `the field accepted no code length up to its maxLength of ${VERIFICATION_TOKEN_MAX_LENGTH}`
  );
}

describe('verification copy against the accepted token shape', () => {
  const codeDigits = acceptedCodeDigits();

  it('shows a sample in every language the field itself would accept', () => {
    for (const [lang, auth] of Object.entries(authLocales)) {
      const sample = auth.verify.codePlaceholder;

      expect(isVerificationToken(sample), `${lang} codePlaceholder is a token`).toBe(true);
      expect(
        sample.length,
        `${lang} codePlaceholder fits the field's maxLength`
      ).toBeLessThanOrEqual(VERIFICATION_TOKEN_MAX_LENGTH);
      // Every keystroke runs the value through the field's normaliser, so a
      // sample it would rewrite is a sample the user cannot type back.
      expect(normalizeVerificationToken(sample), `${lang} codePlaceholder survives typing`).toBe(
        sample
      );
    }
  });

  it('never promises a code shorter than the one the field takes', () => {
    // Arabic-Indic digits are read too: ٦ أرقام is the same claim as 6 digits
    // and must not slip past a check that only sees ASCII.
    const tooShort: string[] = [];

    for (const [lang, auth] of Object.entries(authLocales)) {
      for (const [key, value] of Object.entries(auth.verify) as Array<[string, string]>) {
        // A full sample is exempt: its short half is the user id, not a count.
        if (isVerificationToken(value)) continue;
        for (const run of value.match(/[0-9\u0660-\u0669\u06f0-\u06f9]+/g) ?? []) {
          if (run.length < codeDigits) tooShort.push(`${lang}.verify.${key} = "${value}"`);
        }
      }
    }

    expect(tooShort, `the field takes a user id then ${codeDigits} digits`).toEqual([]);
  });
});
