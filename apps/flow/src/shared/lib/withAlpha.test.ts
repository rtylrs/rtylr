import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';
import { withAlpha } from './colorUtils';

/**
 * A colour at an opacity has to survive the colour being a token.
 *
 * The codebase built these by sticking a hex alpha suffix on the end,
 * `${health.color}1a`. That works only while the colour is a literal hex. Five
 * sites had already crossed that line and were writing
 * `var(--accent)66`, which is not a colour, so the browser dropped the whole
 * declaration and those fills and borders **did not render at all**.
 */
const SRC = path.resolve(__dirname, '../..');

function sourceFiles(dir: string, into: string[] = []): string[] {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      if (entry.name === 'node_modules' || entry.name === 'locales') continue;
      sourceFiles(full, into);
      continue;
    }
    if (!/\.tsx?$/.test(entry.name)) continue;
    if (/\.test\.tsx?$/.test(entry.name)) continue;
    into.push(full);
  }
  return into;
}

describe('withAlpha', () => {
  it('works on a hex', () => {
    expect(withAlpha('#f56d8f', 13)).toBe('color-mix(in srgb, #f56d8f 13%, transparent)');
  });

  it('works on a token, which a hex suffix cannot', () => {
    expect(withAlpha('var(--flow-status-approved)', 20)).toBe(
      'color-mix(in srgb, var(--flow-status-approved) 20%, transparent)'
    );
  });

  it('produces a value the browser will parse, not a concatenation', () => {
    const value = withAlpha('var(--accent)', 40);
    expect(value).toMatch(/^color-mix\(in srgb, .+ \d+%, transparent\)$/);
    expect(value).not.toMatch(/\)[0-9a-f]{2}$/i);
  });
});

describe('no colour gets an alpha by string concatenation', () => {
  it('nowhere in the app', () => {
    // `${color}1a` is the shape. It is silently correct for a hex and
    // silently broken for a token, which is the worst combination: it works
    // until someone tokenises the colour, then stops rendering with no error.
    const offenders: string[] = [];
    for (const file of sourceFiles(SRC)) {
      // Comments stripped: withAlpha's own doc comment quotes the shape it
      // replaces, and a naive scan matches that rather than any code.
      const lines = fs
        .readFileSync(file, 'utf8')
        .replace(/\/\*[\s\S]*?\*\//g, (block) => block.replace(/[^\n]/g, ' '))
        .split('\n');
      lines.forEach((line, index) => {
        if (line.trim().startsWith('//')) return;
        for (const match of line.matchAll(/\$\{[^}]+\}[0-9a-fA-F]{2}\b/g)) {
          offenders.push(`${path.relative(SRC, file)}:${index + 1}  ${match[0]}`);
        }
      });
    }
    expect(offenders).toEqual([]);
  });
});

describe('Flow assigns tones from tokens, not raw hex', () => {
  it.each([
    ['shared/stores/flowStore.ts', 'done'],
    ['shared/lib/progressUtils.ts', 'on_track'],
    ['shared/lib/workloadUtils.ts', 'ok'],
    ['shared/components/CycleBadge.tsx', 'active'],
  ])('%s maps %s to the approved token', (file, key) => {
    // The exact pair, not the first line starting with the key:
    // progressUtils has an on_track in its label map as well as its colour
    // map, and the label one comes first.
    const source = fs.readFileSync(path.join(SRC, file), 'utf8');
    expect(source, `${key} in ${file}`).toContain(`${key}: 'var(--flow-status-approved)'`);
  });

  // The palette arrays keep their hex: those are the colours a user picks from,
  // not a tone the app assigns. Everywhere else a tone carries meaning, so it
  // belongs to a token that can hold a per-mode value.
  const PALETTES = [
    'features/settings/FlowSettings.tsx',
    'features/projects/components/StatusSettingsPanel.tsx',
    'shared/lib/colorUtils.ts',
  ];

  it.each([
    ['green', '#22c55e', '--flow-status-approved'],
    ['blue', '#3b82f6', '--flow-info-fg'],
  ])('leaves no %s as a semantic value', (_name, hex, token) => {
    const offenders: string[] = [];
    for (const file of sourceFiles(SRC)) {
      const relative = path.relative(SRC, file);
      if (PALETTES.includes(relative)) continue;
      if (fs.readFileSync(file, 'utf8').includes(hex)) offenders.push(relative);
    }
    expect(offenders, `use var(${token}) instead of ${hex}`).toEqual([]);
  });

  // Each of these carries a literal hex chosen per mode, so a tone added to one
  // block only resolves to nothing in the other. An unresolved var() paints the
  // browser default rather than failing, so nothing catches it but a reader.
  it.each(['--flow-info-fg', '--flow-warn-fg', '--flow-danger-fg', '--flow-locked-fg'])(
    'declares %s in both modes',
    (token) => {
      const css = fs.readFileSync(path.join(SRC, 'globals.css'), 'utf8');
      const declaredUnder = (mode: string) =>
        css
          .split(/:root\[data-theme='/)
          .filter((block) => block.startsWith(`${mode}'`))
          .some((block) => block.includes(`${token}:`));

      expect(declaredUnder('dark'), `${token} is missing from dark`).toBe(true);
      expect(declaredUnder('light'), `${token} is missing from light`).toBe(true);
    }
  );
});
