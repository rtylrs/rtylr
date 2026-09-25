/** @type {import('next').NextConfig} */
const { PHASE_DEVELOPMENT_SERVER } = require('next/constants');

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
});

module.exports = createNextConfig;
