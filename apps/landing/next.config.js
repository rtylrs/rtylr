/** @type {import('next').NextConfig} */
const { PHASE_DEVELOPMENT_SERVER } = require('next/constants');

/**
 * Landing-site Next config. Turbopack is the default bundler in
 * Next 16 - we rely on it.
 *
 * The previous `webpack: (config) => ...` block defined path
 * aliases (`@app`, `@pages`, `@components`, …) that were declared
 * defensively but never actually imported by any file in this app
 * (grep across .ts/.tsx/.scss/.css came back empty). Dropping the
 * block is cleaner than migrating dead code to `turbopack.rules` /
 * `turbopack.resolveAlias`. If a future screen needs a short alias,
 * add the single entry to `turbopack.resolveAlias` then - don't
 * re-introduce the kitchen-sink list.
 */
const createNextConfig = (phase) => ({
  reactStrictMode: true,
  output: 'export',
  // Keep dev output app-local so server-side module resolution can find
  // app dependencies. The production export stays inside this project too, so
  // the deploy pipeline can read it from the workspace's working directory.
  distDir: phase === PHASE_DEVELOPMENT_SERVER ? '.next' : 'dist/landing',
  trailingSlash: true,

  images: {
    unoptimized: true,
  },

  experimental: {
    optimizePackageImports: ['react-icons'],
  },

  transpilePackages: [
    '@voxire/components',
    '@voxire/observability',
    '@voxire/config',
    '@voxire/consent',
    '@voxire/utils',
  ],

  env: {
    CUSTOM_KEY: process.env.CUSTOM_KEY,
  },
});

module.exports = createNextConfig;
