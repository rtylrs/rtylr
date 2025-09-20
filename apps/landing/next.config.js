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

  // Static export configuration for hosting providers like HostGator
  output: 'export',
  trailingSlash: true,
  distDir: '../../dist/landing',

  // Asset and image configuration
  images: {
    domains: [],
    unoptimized: true, // Required for static export
  },

  // Webpack configuration for aliases
  webpack: (config) => {
    config.resolve.alias = {
      ...config.resolve.alias,
      '@app': path.resolve(__dirname, '.'),
      '@pages': path.resolve(__dirname, './pages'),
      '@components': path.resolve(__dirname, './components'),
      '@hooks': path.resolve(__dirname, './hooks'),
      '@utils': path.resolve(__dirname, './utils'),
      '@styles': path.resolve(__dirname, './styles'),
      '@interfaces': path.resolve(__dirname, './interfaces'),
      '@scss': path.resolve(__dirname, './scss/'),
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