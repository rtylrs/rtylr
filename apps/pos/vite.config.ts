/// <reference types='vitest' />
import { defineConfig } from 'vite';
import path from 'path';
import react from '@vitejs/plugin-react';
import { versionPlugin } from './vite-plugin-version.ts';

export default defineConfig(() => ({
  root: __dirname,
  cacheDir: './node_modules/.vite',
  base: '/',
  plugins: [react(), versionPlugin()],
  resolve: {
    alias: {
      '@app': path.resolve(__dirname, './src'),
      '@features': path.resolve(__dirname, './src/features'),
      '@shared': path.resolve(__dirname, './src/shared'),
      '@stores': path.resolve(__dirname, './src/shared/stores'),
      '@hooks': path.resolve(__dirname, './src/shared/hooks'),
      '@layouts': path.resolve(__dirname, './src/layouts'),
      '@styles': path.resolve(__dirname, './src/styles'),
      '@services': path.resolve(__dirname, './src/shared/services'),
      '@utils': path.resolve(__dirname, './src/shared/utils'),
      '@tanstack/react-query': path.resolve(__dirname, './node_modules/@tanstack/react-query'),
    },
    dedupe: [
      'react',
      'react-dom',
      'react/jsx-runtime',
      'react-dom/client',
      '@tanstack/react-query',
    ],
  },
  server: {
    port: 5178,
    host: true,
    strictPort: true,
    hmr: {
      overlay: false,
    },
  },
  preview: {
    port: 5178,
    host: 'localhost',
    strictPort: true,
  },
  build: {
    outDir: 'dist',
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
        manualChunks(id) {
          if (/node_modules\/(react|react-dom|react-router-dom|zustand|motion)\//.test(id)) {
            return 'vendor';
          }
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
      '@tanstack/react-query',
      '@voxire/config',
    ],
  },
  test: {
    environment: 'happy-dom',
    globals: false,
    setupFiles: ['./src/test/setup.ts'],
    include: ['src/**/*.{test,spec}.{ts,tsx}'],
  },
}));
