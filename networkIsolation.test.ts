import { describe, it, expect } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';

/**
 * The suite used to reach the live internet on every run.
 *
 * `@voxire/config`'s `loadRuntimeConfig` fetched `${VITE_API_URL}/runtime-config`
 * and `VITE_API_URL` defaults to `https://api.voxire.com`, so `pnpm test` sent
 * requests to production. Separately, recruit's font pairing injected a
 * `<link>` pointing at fonts.googleapis.com, which happy-dom then fetched.
 *
 * Both left a request in flight when the test file's window was torn down.
 * happy-dom's `abortAll` throws a `DOMException [AbortError]` out of teardown,
 * which no application `try`/`catch` can see, because it is not the rejection
 * of the awaited promise. An unhandled throw during teardown is a plausible
 * route to a worker dying and taking its queued files with it, which is the
 * signature #988 reported: one failure and 59 tests that never ran.
 */
describe('a test run cannot reach the network', () => {
  const config = fs.readFileSync(path.resolve(__dirname, 'vitest.config.ts'), 'utf8');

  it('points the API base at loopback, not at a real host', () => {
    // Read from the running environment, not from the file, so this fails if
    // the substitution stops reaching the code that reads it.
    const base = import.meta.env.VITE_API_URL;
    expect(base, 'VITE_API_URL must be set for tests').toBeTruthy();
    expect(base).toMatch(/^https?:\/\/(127\.0\.0\.1|localhost)(:\d+)?$/);
  });

  it('substitutes the base at transform time, since @voxire/config is external', () => {
    // `test.env` alone reaches process.env but not the `import.meta.env` read
    // inside an externalised dependency, which is why the first attempt at
    // this fix looked applied and changed nothing.
    expect(config).toMatch(/define:\s*\{/);
    expect(config).toContain("'import.meta.env.VITE_API_URL'");
  });

  it('stops happy-dom loading external stylesheets and scripts', () => {
    expect(config).toContain('disableCSSFileLoading: true');
    expect(config).toContain('disableJavaScriptFileLoading: true');
  });
});
