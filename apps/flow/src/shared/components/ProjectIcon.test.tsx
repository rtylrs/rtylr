import fs from 'node:fs';
import path from 'node:path';
import React from 'react';
import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { Building2, CheckCircle2, ClipboardList, Folder, Globe, Inbox, Rocket } from 'lucide-react';
import ProjectIcon, { glyphForProjectIcon } from './ProjectIcon';

/**
 * The stored value was painted straight into the DOM, so a project whose icon
 * is "building 2" rendered those words as coloured 14px text.
 *
 * The column holds three kinds of value. Measured against the live database:
 * 2,322 rows hold an emoji, 18 hold a tick, 10 hold a lucide name.
 */
describe('glyphForProjectIcon', () => {
  it('maps a lucide name to its glyph', () => {
    expect(glyphForProjectIcon('rocket')).toBe(Rocket);
    expect(glyphForProjectIcon('globe')).toBe(Globe);
  });

  it('maps the spaced and hyphenated spellings of the same name', () => {
    // "building 2" is what the column actually holds, and it is the value the
    // issue shows wrapping onto two lines.
    expect(glyphForProjectIcon('building-2')).toBe(Building2);
    expect(glyphForProjectIcon('building 2')).toBe(Building2);
    expect(glyphForProjectIcon('building2')).toBe(Building2);
  });

  it('is not case or whitespace sensitive', () => {
    expect(glyphForProjectIcon('  Rocket  ')).toBe(Rocket);
  });

  it('maps the common emoji to a glyph that means the same thing', () => {
    // Without this the 2,340 emoji rows would all fall back to Folder, which
    // is the cost the decision on #1085 was weighed against.
    // Compared by identity: lucide exports aliases, so CheckCircle2 reports
    // a displayName of 'CircleCheck' and a name assertion would be testing
    // lucide's naming rather than this map.
    expect(glyphForProjectIcon('\u{1F4CB}')).toBe(ClipboardList);
    expect(glyphForProjectIcon('✅')).toBe(CheckCircle2);
    expect(glyphForProjectIcon('\u{1F4E5}')).toBe(Inbox);
  });

  it('falls back to Folder for an unknown value', () => {
    expect(glyphForProjectIcon('not-a-real-icon')).toBe(Folder);
    expect(glyphForProjectIcon('\u{1F984}')).toBe(Folder);
  });

  it('falls back for a missing or empty value', () => {
    expect(glyphForProjectIcon(undefined)).toBe(Folder);
    expect(glyphForProjectIcon('')).toBe(Folder);
    expect(glyphForProjectIcon('   ')).toBe(Folder);
  });

  it('never returns undefined, whatever it is handed', () => {
    for (const value of ['', ' ', 'ROCKET', '📋', 'building 2', 'zzz', undefined]) {
      expect(glyphForProjectIcon(value as string | undefined)).toBeTruthy();
    }
  });
});

describe('ProjectIcon', () => {
  it('renders an svg, not the stored text', () => {
    const { container } = render(<ProjectIcon icon="building 2" />);
    expect(container.querySelector('svg')).not.toBeNull();
    expect(container.textContent).toBe('');
  });

  it('renders no emoji for an emoji value', () => {
    const { container } = render(<ProjectIcon icon={'\u{1F4CB}'} />);
    expect(container.textContent).toBe('');
    expect(container.querySelector('svg')).not.toBeNull();
  });
});

describe('the project card uses it', () => {
  const card = fs.readFileSync(
    path.resolve(__dirname, '../../features/projects-v2/AllProjects.tsx'),
    'utf8'
  );

  it('renders the component rather than the raw value', () => {
    expect(card).toContain('<ProjectIcon icon={project.icon}');
  });

  it('keeps no emoji literal as a fallback', () => {
    const code = card
      .replace(/\/\*[\s\S]*?\*\//g, '')
      .split('\n')
      .filter((line) => !line.trim().startsWith('//'))
      .join('\n');
    expect(code).not.toContain("project.icon || '\u{1F4CB}'");
  });
});
