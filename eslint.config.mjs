// Flat ESLint config for the monorepo. Intentionally conservative: the rules
// that catch real bugs (rules-of-hooks, no-debugger) are errors; stylistic and
// large-surface rules are warnings so they surface in the editor/CI logs
// without turning the whole existing codebase red. Tighten over time.
import tseslint from 'typescript-eslint';
import reactHooks from 'eslint-plugin-react-hooks';
import prettier from 'eslint-config-prettier';

export default tseslint.config(
  {
    ignores: [
      '**/dist/**',
      '**/build/**',
      '**/node_modules/**',
      '**/coverage/**',
      '**/*.config.{js,ts,mjs,cjs}',
      '**/*.d.ts',
      'voxie/**',
      'packages/voxie/**',
      'dist/**',
      // Scratch files written by the Remember plugin, in this repo and in the
      // packages submodule. They are gitignored but still reachable by the
      // linter, and one of them fails `no-unused-expressions`, which failed
      // `pnpm lint` and the pre-commit hook for every unrelated change.
      '**/.remember/**',
    ],
  },
  ...tseslint.configs.recommended,
  {
    files: ['**/*.{ts,tsx,js,jsx,mjs,cjs}'],
    plugins: { 'react-hooks': reactHooks },
    rules: {
      // Real-bug rules — errors.
      'react-hooks/rules-of-hooks': 'error',
      'no-debugger': 'error',

      // High-signal but large-surface — warn for now.
      'react-hooks/exhaustive-deps': 'warn',
      'no-console': ['warn', { allow: ['warn', 'error'] }],
      '@typescript-eslint/no-explicit-any': 'warn',
      '@typescript-eslint/no-unused-vars': [
        'warn',
        { argsIgnorePattern: '^_', varsIgnorePattern: '^_', caughtErrorsIgnorePattern: '^_' },
      ],
      '@typescript-eslint/ban-ts-comment': 'warn',
      '@typescript-eslint/no-empty-object-type': 'warn',
      '@typescript-eslint/no-non-null-assertion': 'off',
    },
  },
  prettier
);
