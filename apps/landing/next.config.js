/** @type {import('next').NextConfig} */
const path = require('path');
const { composePlugins, withNx } = require('@nx/next');

const nextConfig = {
  // Nx configuration
  nx: {
    svgr: false,
  },

  // Basic Next.js configuration
  reactStrictMode: true,
  swcMinify: true,

  // Asset and image configuration
  images: {
    domains: [],
    unoptimized: false,
  },

  // Webpack configuration for aliases
  webpack: (config) => {
    config.resolve.alias = {
      ...config.resolve.alias,
      '@app': path.resolve(__dirname, './src'),
      '@pages': path.resolve(__dirname, './src/pages'),
      '@components': path.resolve(__dirname, './src/components'),
      '@hooks': path.resolve(__dirname, './src/hooks'),
      '@utils': path.resolve(__dirname, './src/utils'),
      '@styles': path.resolve(__dirname, './src/styles'),
      '@interfaces': path.resolve(__dirname, './src/interfaces'),
      '@scss': path.resolve(__dirname, './src/scss/'),
      '@root': path.resolve(__dirname, '../..'),
    };

    return config;
  },

  // Experimental features
  experimental: {
    optimizePackageImports: [
      '@mui/material',
      '@mui/icons-material',
      '@fortawesome/react-fontawesome',
    ],
  },

  // Transpile packages if needed
  transpilePackages: ['@voxire/shared'],

  // Environment variables
  env: {
    CUSTOM_KEY: process.env.CUSTOM_KEY,
  },
};

// Export configuration with Nx wrapper
module.exports = composePlugins(withNx)(nextConfig);