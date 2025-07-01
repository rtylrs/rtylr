import { defineConfig } from "vite";
import path from "path";
import react from "@vitejs/plugin-react";
import svgr from "vite-plugin-svgr";

export default defineConfig({
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
      "@pages": path.resolve(__dirname, "./src/pages"),
      "@components": path.resolve(__dirname, "./src/components"),
      "@hooks": path.resolve(__dirname, "./src/hooks"),
      "@utils": path.resolve(__dirname, "./src/utils"),
      "@styles": path.resolve(__dirname, "./src/styles"),
      "@interfaces": path.resolve(__dirname, "./src/interfaces"),
      "@scss": path.resolve(__dirname, "./src/scss/"),
    },
  },
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: `
          @use "@app/styles/index.scss";
        `,
        api: "modern-compiler",
      },
    },
  },
  build: {
    manifest: true,
    target: "esnext",
    minify: "esbuild",
    cssMinify: true,
    cssCodeSplit: true,
    modulePreload: true,
    sourcemap: false,
    chunkSizeWarningLimit: 1000,
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
            "react-redux",
            "@reduxjs/toolkit",
            "framer-motion",
            "@fortawesome/react-fontawesome",
          ],
          styles: ["@mui/material", "react-toastify", "slick-carousel"],
        },
      },
      input: {
        main: path.resolve(__dirname, "index.html"),
      },
    },
  },
  server: {
    host: true,
    port: 5173,
    strictPort: true,
    hmr: {
      overlay: false,
    },
  },
  preview: {
    port: 5173,
    strictPort: true,
  },
  optimizeDeps: {
    include: [
      "react",
      "react-dom",
      "react-router-dom",
      "react-redux",
      "@reduxjs/toolkit",
      "framer-motion",
      "@fortawesome/react-fontawesome",
    ],
  },
});
