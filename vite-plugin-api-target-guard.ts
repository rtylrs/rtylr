import type { Plugin } from 'vite';

/**
 * Dev-only guard against a local dev server silently talking to production.
 *
 * Every app resolves its API base from `VITE_API_URL`, falling back to the
 * hosted default (see `getBootstrapApiBase` in `@voxire/config`'s
 * `runtimeConfig.ts`). No app ships a `.env.local`, so the default state of a
 * fresh checkout is that `pnpm --filter <app> dev` points at production - and
 * nothing on screen says so. Writes issued from that dev server hit real tenant
 * data.
 *
 * This plugin prints a loud banner at dev-server startup, and injects a browser
 * console warning, whenever the resolved API target is not loopback.
 *
 * It is `apply: 'serve'`, so it cannot run during `vite build`: production and
 * self-hosted deployments are unaffected by construction, not by convention.
 */

/**
 * Hosted API default. Mirrors `getBootstrapApiBase()` in
 * `packages/config/src/runtimeConfig.ts`. Duplicated as a literal because this
 * file runs in Vite's Node context, before any workspace package is resolved.
 */
const HOSTED_API_DEFAULT = 'https://api.voxire.com/';

/** Hostnames that mean "a backend on this machine". */
const LOOPBACK_HOSTS = new Set([
  'localhost',
  '127.0.0.1',
  '0.0.0.0',
  '::1',
  'host.docker.internal',
]);

/** How the app arrived at its API target. */
export type ApiTargetSource = 'env' | 'hosted-default';

export interface ApiTarget {
  /** The absolute base URL the app will call. */
  url: string;
  source: ApiTargetSource;
  /** True when `url` points at a backend on this machine. */
  local: boolean;
}

/**
 * Whether an API base points at this machine.
 *
 * Deliberately stricter than `isLocalDevHost` in `@voxire/api`, which treats
 * every IPv4 literal as local: a private-LAN address here is more likely a
 * shared staging box than someone's laptop, and a false "you're safe" is the
 * failure this plugin exists to prevent. An unparseable value is treated as
 * non-local for the same reason.
 */
export function isLocalApiTarget(target: string): boolean {
  let hostname: string;
  try {
    hostname = new URL(target).hostname.toLowerCase();
  } catch {
    return false;
  }
  // `new URL('http://[::1]/')` yields the bracketed form.
  const host = hostname.replace(/^\[/, '').replace(/\]$/, '');
  return LOOPBACK_HOSTS.has(host) || host.endsWith('.localhost');
}

/** Resolve what the app will actually call, and how it got there. */
export function resolveApiTarget(env: Record<string, string | undefined>): ApiTarget {
  const configured = env.VITE_API_URL;
  const url = configured && configured.trim().length > 0 ? configured.trim() : HOSTED_API_DEFAULT;
  return {
    url,
    source: configured && configured.trim().length > 0 ? 'env' : 'hosted-default',
    local: isLocalApiTarget(url),
  };
}

/**
 * The warning body, as plain lines. Kept separate from rendering so the wording
 * and the branch that produces it are testable without ANSI or box drawing.
 */
export function buildWarningLines(appName: string, target: ApiTarget): string[] {
  const origin =
    target.source === 'hosted-default'
      ? 'hosted default (VITE_API_URL is not set)'
      : 'VITE_API_URL';

  return [
    'DEV SERVER IS POINTED AT A NON-LOCAL API',
    '',
    `app     ${appName}`,
    `API     ${target.url}`,
    `source  ${origin}`,
    '',
    'Every request from this dev server - including writes - goes to that',
    'backend. If it is production, you are creating and mutating real',
    'tenant data.',
    '',
    'To point at a local onion instance instead:',
    '',
    `    cp .env.local.example apps/${appName}/.env.local`,
    '',
    'then restart the dev server. See .env.local.example for details.',
  ];
}

/** Draw an ASCII box sized to its widest line. */
function renderBox(lines: string[]): string {
  const pad = 2;
  const inner = Math.max(...lines.map((l) => l.length)) + pad * 2;
  const top = `+${'-'.repeat(inner)}+`;
  const body = lines.map(
    (l) => `|${' '.repeat(pad)}${l.padEnd(inner - pad * 2)}${' '.repeat(pad)}|`
  );
  return [top, ...body, top].join('\n');
}

/** Bold red, unless the stream is not a TTY or NO_COLOR is set. */
function red(text: string): string {
  const plain = process.env.NO_COLOR !== undefined || !process.stdout.isTTY;
  return plain ? text : `\u001b[1;31m${text}\u001b[0m`;
}

/** Dim, for the everything-is-fine confirmation line. */
function dim(text: string): string {
  const plain = process.env.NO_COLOR !== undefined || !process.stdout.isTTY;
  return plain ? text : `\u001b[2m${text}\u001b[0m`;
}

/**
 * Warn at dev-server startup when the resolved API base is not loopback.
 *
 * The app name is derived from the Vite `root` directory, so every app
 * registers it identically with no argument.
 */
export function apiTargetGuardPlugin(): Plugin {
  let appName = 'app';
  let target: ApiTarget = { url: HOSTED_API_DEFAULT, source: 'hosted-default', local: false };

  return {
    name: 'vite-plugin-api-target-guard',
    // Dev only. This plugin must never influence a production build.
    apply: 'serve',
    configResolved(config) {
      appName = config.root.split('/').filter(Boolean).pop() ?? 'app';
      target = resolveApiTarget(config.env as Record<string, string | undefined>);
    },
    configureServer(server) {
      // Print after Vite's own URL block so the banner is the last thing on
      // screen rather than being scrolled away by startup output.
      const printUrls = server.printUrls.bind(server);
      server.printUrls = (): void => {
        printUrls();
        if (target.local) {
          // eslint-disable-next-line no-console
          console.log(dim(`  API target  ${target.url}  (local)`) + '\n');
          return;
        }
        // eslint-disable-next-line no-console
        console.warn('\n' + red(renderBox(buildWarningLines(appName, target))) + '\n');
      };
    },
    transformIndexHtml() {
      if (target.local) return;
      const message = [
        'RTYLR dev server is talking to a NON-LOCAL API.',
        '',
        ...buildWarningLines(appName, target).slice(2),
      ].join('\n');
      return [
        {
          tag: 'script',
          injectTo: 'head-prepend' as const,
          children: `console.warn(${JSON.stringify(message)});`,
        },
      ];
    },
  };
}
