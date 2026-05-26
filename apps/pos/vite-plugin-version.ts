import { execSync } from 'child_process';
import { readFileSync, writeFileSync } from 'fs';
import { join } from 'path';
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
      outDir = config.build.outDir;
    },
    configureServer(server) {
      server.middlewares.use('/version.json', (_req, res) => {
        res.setHeader('Content-Type', 'application/json');
        res.end(JSON.stringify(buildVersionData(), null, 2) + '\n');
      });
    },
    closeBundle() {
      writeFileSync(
        join(outDir, 'version.json'),
        JSON.stringify(buildVersionData(), null, 2) + '\n',
      );
    },
  };
}
