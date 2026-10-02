/// <reference types='vitest' />
import { defineConfig } from 'vite';
import path from 'path';
import react from '@vitejs/plugin-react';
import svgr from 'vite-plugin-svgr';

export default defineConfig(() => ({
  root: __dirname,
  cacheDir: './node_modules/.vite',
  base: '/',
  plugins: [
    react(),
    svgr({
      svgrOptions: {
        ref: true,
      },
    }),
  ],
  resolve: {
    dedupe: ['react', 'react-dom', 'react/jsx-runtime', 'react-dom/client'],
    alias: {
      '@app': path.resolve(__dirname, './src'),
      '@pages': path.resolve(__dirname, './src/pages'),
      '@components': path.resolve(__dirname, './src/components'),
      '@hooks': path.resolve(__dirname, './src/hooks'),
      '@utils': path.resolve(__dirname, './src/utils'),
      '@styles': path.resolve(__dirname, './src/styles'),
      '@interfaces': path.resolve(__dirname, './src/interfaces'),
      '@scss': path.resolve(__dirname, './src/scss'),
      '@root': path.resolve(__dirname, './'),
    },
  },
  server: {
    port: 5175,
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
              ws: true,
              rewriteWsOrigin: true,
              rewrite: (requestPath: string): string => requestPath.replace(/^\/api/, ''),
            },
          },
        }
      : {}),
  },
  preview: {
    port: 5175,
    host: 'localhost',
    strictPort: true,
  },
  build: {
    outDir: './dist/portal',
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
    include: ['react', 'react-dom', 'zustand', 'motion', 'lucide-react'],
    exclude: ['@voxire/menu', '@voxire/ui'],
  },
}));
