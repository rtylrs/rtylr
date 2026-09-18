import { describe, it, expect } from 'vitest';
import { readFileSync, readdirSync } from 'fs';
import { join } from 'path';

/**
 * RTL guard for the Appearance panel.
 *
 * This scans source rather than rendering, which is unusual and deliberate. The
 * property that matters is whether a style is logical or physical, and neither
 * jsdom nor happy-dom resolves `margin-inline-start` to a physical side, so a
 * rendered assertion cannot tell a correct panel from a broken one. Real bidi
 * rendering is covered by the Playwright spec instead.
 *
 * What this catches is the actual regression path: somebody adds `ml-2` or
 * `left: 0` to a component here months from now, the panel silently breaks for
 * Arabic, and no existing test notices because none of them set `dir`.
 */

const PANEL_DIR = join(__dirname, 'components');
const SECTION_FILE = join(__dirname, 'FlowSettings.tsx');

/** Tailwind utilities that pick a physical side instead of a logical one. */
const PHYSICAL_UTILITY =
  /(?:^|[\s'"`{])(?:ml|mr|pl|pr|border-l|border-r|rounded-l|rounded-r|inset-l|inset-r)-[a-z0-9./[\]]+/;

/** Inline style keys and CSS properties that do the same. */
const PHYSICAL_STYLE =
  /\b(?:marginLeft|marginRight|paddingLeft|paddingRight|borderLeft|borderRight|translateX)\b|(?:^|[\s;{])(?:left|right)\s*:/;

/** text-left and text-right, which do not flip. `text-start`/`text-end` do. */
const PHYSICAL_ALIGN = /\btext-(?:left|right)\b/;

/**
 * Comments are stripped before scanning.
 *
 * Not defensive tidying: the first run of this suite failed on the phrase "a
 * literal colour is right:" inside a comment, which the physical-style pattern
 * read as a CSS `right:` declaration. Prose will keep containing the words left
 * and right, so the scan has to look at code only or it cries wolf.
 */
function stripComments(src: string): string {
  return src.replace(/\/\*[\s\S]*?\*\//g, '').replace(/\/\/[^\n]*/g, '');
}

function panelSources(): { name: string; body: string }[] {
  const files = readdirSync(PANEL_DIR)
    .filter((f) => f.endsWith('.tsx') && !f.includes('.test.'))
    .map((f) => ({
      name: `components/${f}`,
      body: stripComments(readFileSync(join(PANEL_DIR, f), 'utf8')),
    }));

  // Only the AppearanceSection part of FlowSettings, since the rest of that file
  // predates this feature and is not what this guard is about.
  const full = readFileSync(SECTION_FILE, 'utf8');
  const start = full.indexOf('const AppearanceSection');
  const end = full.indexOf('\nconst ', start + 10);
  files.push({
    name: 'FlowSettings.tsx (AppearanceSection)',
    body: start === -1 ? '' : stripComments(full.slice(start, end === -1 ? undefined : end)),
  });
  return files;
}

describe('Appearance panel stays direction-agnostic', () => {
  const sources = panelSources();

  it('reads more than one source file, so a passing run means something', () => {
    // Without this, a bad path silently makes every assertion below vacuous.
    expect(sources.length).toBeGreaterThan(3);
    for (const { name, body } of sources) {
      expect(body.length, `${name} is empty, so the scan below proves nothing`).toBeGreaterThan(50);
    }
  });

  it('uses no physical margin, padding, border or inset utilities', () => {
    const offenders = sources
      .filter(({ body }) => PHYSICAL_UTILITY.test(body))
      .map(({ name, body }) => `${name}: ${body.match(PHYSICAL_UTILITY)?.[0]?.trim()}`);
    expect(offenders, 'use the inline variants: ms/me, ps/pe, border-s/border-e').toEqual([]);
  });

  it('uses no physical direction in inline styles', () => {
    const offenders = sources
      .filter(({ body }) => PHYSICAL_STYLE.test(body))
      .map(({ name, body }) => `${name}: ${body.match(PHYSICAL_STYLE)?.[0]?.trim()}`);
    expect(offenders, 'use insetInlineStart/End and marginInlineStart/End').toEqual([]);
  });

  it('uses no text-left or text-right', () => {
    const offenders = sources
      .filter(({ body }) => PHYSICAL_ALIGN.test(body))
      .map(({ name }) => name);
    expect(offenders, 'use text-start and text-end, which flip with direction').toEqual([]);
  });

  it('pins the hex field to ltr, because its value is a literal not prose', () => {
    const section = sources.find((x) => x.name.startsWith('FlowSettings.tsx'));
    expect(section, 'AppearanceSection not found by the scan').toBeTruthy();

    // Observed in a browser before this was added: with dir=rtl on the document
    // the field inherited RTL, and the leading '#' is a neutral character that
    // resolves to the visual right, so "#f56d8f" rendered as "f56d8f#".
    //
    // Only the value is pinned. The labels and descriptions around it are prose
    // and must keep following the reader's direction.
    const hexBlock = section!.body.slice(
      section!.body.indexOf('id="appearance-accent-hex"'),
      section!.body.indexOf('id="appearance-accent-hex"') + 500
    );
    expect(hexBlock).toMatch(/dir="ltr"/);
  });

  it('moves the switch knob with a logical property', () => {
    const sw = sources.find((s) => s.name.endsWith('AppearanceSwitch.tsx'));
    expect(sw, 'AppearanceSwitch.tsx not found by the scan').toBeTruthy();
    // The knob is the one element in the panel whose position is animated, so it
    // is the one most likely to be written with translateX and travel the wrong
    // way in Arabic.
    expect(sw!.body).toMatch(/insetInlineStart/);
    expect(sw!.body).not.toMatch(/translate-x-|translateX/);
  });
});
