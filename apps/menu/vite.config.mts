/// <reference types='vitest' />
import { defineConfig } from 'vite';
import path from 'path';
import react from '@vitejs/plugin-react';
import svgr from 'vite-plugin-svgr';
import tsconfigPaths from 'vite-tsconfig-paths';

export default defineConfig(() => ({
  root: __dirname,
  cacheDir: '../../node_modules/.vite/menu',
  base: '/',
  plugins: [
    react(),
    svgr({
      svgrOptions: {
        ref: true,
      },
    }),
    tsconfigPaths(),
  ],
  resolve: {
    alias: {
      '@app': path.resolve(__dirname, './src'),
      '@pages': path.resolve(__dirname, './src/pages'),
      '@components': path.resolve(__dirname, './src/components'),
      '@hooks': path.resolve(__dirname, './src/hooks'),
      '@utils': path.resolve(__dirname, './src/utils'),
      '@styles': path.resolve(__dirname, './src/styles'),
      '@interfaces': path.resolve(__dirname, './src/interfaces'),
      '@scss': path.resolve(__dirname, './src/scss/'),
      '@root': path.resolve(__dirname, '../..'),
    },
  },
  server: {
    port: 5174,
    host: true,
    strictPort: true,
    hmr: {
      overlay: false,
    },
    proxy: {
      '/api': {
        target: 'https://onion.voxire.dev',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''),
        secure: true,
      },
    },
  },
  preview: {
    port: 5174,
    host: 'localhost',
    strictPort: true,
  },
  build: {
    outDir: './dist',
    emptyOutDir: true,
    reportCompressedSize: true,
    manifest: true,
    target: 'esnext',
    minify: 'esbuild' as const,
    cssMinify: true,
    cssCodeSplit: true,
    modulePreload: true,
    sourcemap: false,
    chunkSizeWarningLimit: 1000,
    commonjsOptions: {
      transformMixedEsModules: true,
    },
    rollupOptions: {
      output: {
        entryFileNames: 'assets/[name]-[hash].js',
        chunkFileNames: 'assets/[name]-[hash].js',
        assetFileNames: 'assets/[name]-[hash].[ext]',
        manualChunks: {
          vendor: [
            'react',
            'react-dom',
            'react-router-dom',
            'zustand',
            'motion',
          ],
        },
      },
    },
  },
  optimizeDeps: {
    include: [
      'react',
      'react-dom',
      'react-router-dom',
      'zustand',
      'motion',
    ],
    exclude: ['@voxire/components', '@voxire/config', '@voxire/menu', '@voxire/ui', '@voxire/tailwind-config'],
  },
}));
