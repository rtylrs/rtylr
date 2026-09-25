/// <reference types='vitest' />
import { defineConfig } from 'vite';
import path from 'path';
import react from '@vitejs/plugin-react';

export default defineConfig(() => ({
  root: __dirname,
  cacheDir: './node_modules/.vite',
  base: '/',
  plugins: [react()],
  resolve: {
    dedupe: ['react', 'react-dom', 'react/jsx-runtime', 'react-dom/client'],
    alias: {
      '@app': path.resolve(__dirname, './src'),
      '@features': path.resolve(__dirname, './src/features'),
      '@layouts': path.resolve(__dirname, './src/layouts'),
      '@shared': path.resolve(__dirname, './src/shared'),
    },
  },
  server: {
    port: 5180,
    host: true,
    strictPort: true,
    hmr: {
      overlay: false,
    },
  },
  preview: {
    port: 5180,
    host: 'localhost',
    strictPort: true,
  },
  build: {
    outDir: '../../dist/hr',
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
        // Vite 8 (Rolldown) only accepts function form of manualChunks.
        manualChunks(id: string): string | undefined {
          if (/@react-pdf\/|pdfjs-dist/.test(id)) {
            return 'pdf-vendor';
          }
          if (/node_modules\/(react|react-dom|@tanstack\/react-router|zustand)\//.test(id)) {
            return 'react-vendor';
          }
          if (/node_modules\/(motion|lucide-react)\//.test(id)) {
            return 'ui-vendor';
          }
          return undefined;
        },
      },
    },
  },
  optimizeDeps: {
    include: ['react', 'react-dom', 'zustand', 'motion', 'lucide-react'],
  },
}));
