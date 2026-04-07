/** @type {import('next').NextConfig} */
const path = require('path');

const nextConfig = {
  reactStrictMode: true,

  images: {
    remotePatterns: [],
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
      'react-icons',
    ],
  },

  // Transpile packages if needed
  transpilePackages: ['@voxire/assets', '@voxire/components', '@voxire/config', '@voxire/utils'],

  // Environment variables
  env: {
    CUSTOM_KEY: process.env.CUSTOM_KEY,
  },
};

module.exports = nextConfig;