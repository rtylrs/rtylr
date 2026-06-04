import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

const currentDir = dirname(fileURLToPath(import.meta.url));
const source = readFileSync(join(currentDir, 'General.tsx'), 'utf8');

describe('General settings slug action layout', () => {
  it('keeps the save slug button from shrinking or wrapping', () => {
    expect(source).toMatch(/className="[^"]*shrink-0[^"]*whitespace-nowrap[^"]*"/);
  });
});
