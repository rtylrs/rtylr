import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import dts from 'vite-plugin-dts';
import { resolve } from 'path';

export default defineConfig({
  plugins: [
    react(),
    dts({
      insertTypesEntry: true,
    }),
  ],
  build: {
    outDir: resolve(__dirname, '../../../dist/@voxire/shared'),
    lib: {
      entry: resolve(__dirname, 'src/index.ts'),
      name: 'VoxireShared',
      formats: ['es', 'cjs'],
      fileName: (format) => `index.${format === 'es' ? 'esm' : format}.js`,
    },
    rollupOptions: {
      external: [
        'react',
        'react-dom',
        'react/jsx-runtime',
        '@emotion/react',
        '@emotion/styled',
        '@mui/material',
        '@mui/icons-material',
        'classnames',
        'framer-motion',
      ],
      output: {
        globals: {
          react: 'React',
          'react-dom': 'ReactDOM',
          '@emotion/react': 'emotionReact',
          '@emotion/styled': 'emotionStyled',
          '@mui/material': 'muiMaterial',
          '@mui/icons-material': 'muiIcons',
          classnames: 'classNames',
          'framer-motion': 'framerMotion',
        },
      },
    },
    cssCodeSplit: false,
  },
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: `@import "${resolve(
          __dirname,
          'src/styles/_variables.scss'
        )}";`,
      },
    },
  },
});
