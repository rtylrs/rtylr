import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';

const source = readFileSync(join(import.meta.dirname, 'ui.tsx'), 'utf8');

describe('crm field spacing', () => {
  it('uses a slightly roomier gap between labels and controls', () => {
    expect(source).toContain("gap: '6px'");
  });
});
