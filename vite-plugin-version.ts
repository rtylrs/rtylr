import { execSync } from 'child_process';
import { mkdirSync, readFileSync, writeFileSync } from 'fs';
import { join, resolve } from 'path';
import type { Plugin } from 'vite';

function getCommit(): string {
  if (process.env.GITHUB_SHA) return process.env.GITHUB_SHA.slice(0, 7);
  try {
    return execSync('git rev-parse --short HEAD', { stdio: ['pipe', 'pipe', 'pipe'] })
      .toString()
      .trim();
  } catch {
    return 'unknown';
  }
}

export function versionPlugin(): Plugin {
  let outDir = 'dist';
  let isBuild = false;

  const buildVersionData = () => {
    const pkg = JSON.parse(readFileSync('./package.json', 'utf8')) as { version: string };
    return {
      version: pkg.version,
      commit: getCommit(),
      buildTime: new Date().toISOString(),
    };
  };

  return {
    name: 'vite-plugin-version',
    configResolved(config) {
      outDir = resolve(config.root, config.build.outDir);
      isBuild = config.command === 'build';
    },
    configureServer(server) {
      server.middlewares.use('/version.json', (_req, res) => {
        res.setHeader('Content-Type', 'application/json');
        res.end(JSON.stringify(buildVersionData(), null, 2) + '\n');
      });
    },
    closeBundle() {
      // Only on a real build. Vitest resolves a config with
      // `outDir: "dummy-non-existing-folder"` as a sentinel and still triggers
      // closeBundle, so without this guard every test run created
      // apps/<app>/dummy-non-existing-folder/version.json inside the source
      // tree. That folder is not gitignored, so `git add apps/<app>` picks it
      // up silently: it has been committed by accident at least once already and
      // had to be removed in a follow-up commit.
      if (!isBuild) return;
      mkdirSync(outDir, { recursive: true });
      writeFileSync(
        join(outDir, 'version.json'),
        JSON.stringify(buildVersionData(), null, 2) + '\n'
      );
    },
  };
}
