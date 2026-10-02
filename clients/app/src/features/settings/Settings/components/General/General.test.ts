import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

const currentDir = dirname(fileURLToPath(import.meta.url));
const source = readFileSync(join(currentDir, 'General.tsx'), 'utf8');

describe('General settings save actions', () => {
  it('uses a centered save bar for pending changes', () => {
    expect(source).toMatch(/fixed bottom-4 left-1\/2/);
    expect(source).toMatch(/common\.unsavedChanges/);
    expect(source).toMatch(/common\.cancel/);
    expect(source).toMatch(/const hasChanges = isFormDirty \|\| isSlugDirty/);
  });

  it('does not keep a separate inline slug save button', () => {
    expect(source).not.toMatch(/Save slug/);
  });
});
