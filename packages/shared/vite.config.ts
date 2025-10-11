import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import svgr from "vite-plugin-svgr";
import dts from "vite-plugin-dts";
import { resolve, dirname } from "path";
import { fileURLToPath } from "url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

export default defineConfig(({ command }) => {
  const isDev = command === "serve" || process.env.NODE_ENV === "development";

  return {
    plugins: [
      react(),
      svgr({
        svgrOptions: { icon: true },
        include: ["**/*.svg?react"],
      }),
      !isDev &&
        dts({
          insertTypesEntry: true,
          exclude: ["src/dev.tsx"],
        }),
    ].filter(Boolean),
    // Development server config
    root: isDev ? "." : undefined,
    build: isDev
      ? undefined
      : {
          outDir: resolve(__dirname, "./dist"),
          lib: {
            entry: {
              "shared/index": resolve(__dirname, "src/shared/index.ts"),
              "utils/index": resolve(__dirname, "src/utils/index.ts"),
            },
            name: "VoxireMonorepo",
            formats: ["es", "cjs"],
            fileName: (format, entryName) =>
              `${entryName}.${format === "es" ? "esm" : format}.js`,
          },
          rollupOptions: {
            external: [
              "react",
              "react-dom",
              "react/jsx-runtime",
              "@emotion/react",
              "@emotion/styled",
              "@mui/material",
              "@mui/icons-material",
              "classnames",
              "framer-motion",
              "react-helmet",
              "date-fns",
              "lodash",
            ],
            output: {
              globals: {
                react: "React",
                "react-dom": "ReactDOM",
                "@emotion/react": "emotionReact",
                "@emotion/styled": "emotionStyled",
                "@mui/material": "muiMaterial",
                "@mui/icons-material": "muiIcons",
                classnames: "classNames",
                "framer-motion": "framerMotion",
                "react-helmet": "ReactHelmet",
                "date-fns": "dateFns",
                lodash: "_",
              },
            },
          },
          cssCodeSplit: false,
        },
    css: {
      preprocessorOptions: {
        scss: {
          api: "modern",
          quietDeps: true,
        },
      },
      postcss: {
        plugins: [],
      },
    },
  };
});
