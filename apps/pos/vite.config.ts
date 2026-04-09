/// <reference types='vitest' />
import { defineConfig } from "vite";
import path from "path";
import react from "@vitejs/plugin-react";
import svgr from "vite-plugin-svgr";

export default defineConfig(() => ({
  root: __dirname,
  cacheDir: "./node_modules/.vite",
  base: "/",
  plugins: [
    react(),
    svgr({
      svgrOptions: {
        ref: true,
      },
    }),
  ],
  resolve: {
    alias: {
      "@app": path.resolve(__dirname, "./src"),
      "@screens": path.resolve(__dirname, "./src/screens"),
      "@components": path.resolve(__dirname, "./src/components"),
      "@hooks": path.resolve(__dirname, "./src/hooks"),
      "@stores": path.resolve(__dirname, "./src/stores"),
      "@lib": path.resolve(__dirname, "./src/lib"),
    },
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
    host: "localhost",
    strictPort: true,
  },
  build: {
    outDir: "./dist",
    emptyOutDir: true,
    reportCompressedSize: true,
    manifest: true,
    target: "esnext",
    minify: "esbuild" as const,
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
        entryFileNames: "assets/[name]-[hash].js",
        chunkFileNames: "assets/[name]-[hash].js",
        assetFileNames: "assets/[name]-[hash].[ext]",
        manualChunks: {
          vendor: [
            "react",
            "react-dom",
            "react-router-dom",
            "zustand",
            "framer-motion",
            "lucide-react",
          ],
        },
      },
    },
  },
  optimizeDeps: {
    include: [
      "react",
      "react-dom",
      "react-router-dom",
      "zustand",
      "framer-motion",
      "lucide-react",
    ],
  },
}));
