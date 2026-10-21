/// <reference types='vitest' />
import { defineConfig } from 'vite';
import path from 'path';
import react from '@vitejs/plugin-react';
import svgr from 'vite-plugin-svgr';
import tsconfigPaths from 'vite-tsconfig-paths';

export default defineConfig(() => ({
  root: __dirname,
  cacheDir: '../../node_modules/.vite/auth',
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
    },
    dedupe: ['react', 'react-dom', 'react/jsx-runtime', 'react-dom/client'],
  },
  server: {
    port: 5176,
    host: true,
    strictPort: true,
    hmr: {
      overlay: false,
    },
    ...(process.env.VITE_API_PROXY_TARGET
      ? {
          proxy: {
            '/api': {
              target: process.env.VITE_API_PROXY_TARGET,
              changeOrigin: true,
              rewrite: (requestPath: string): string => requestPath.replace(/^\/api/, ''),
            },
          },
        }
      : {}),
  },
  preview: {
    port: 5176,
    host: 'localhost',
    strictPort: true,
  },
  build: {
    outDir: '../../dist/auth',
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
        manualChunks(id: string): string | undefined {
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
    include: ['react', 'react-dom', 'motion', '@voxire/config', '@voxire/utils'],
    exclude: ['@voxire/components', '@voxire/ui'],
  },
}));
