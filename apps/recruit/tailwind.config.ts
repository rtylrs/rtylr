import type { Config } from 'tailwindcss';

// The recruit app renders multiple visual templates (minimal, editorial,
// bold) - the palette below is only the accent default. Each template
// layers its own tokens on top via CSS custom properties so a business
// can repaint their careers page without touching Tailwind.
const config: Config = {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        accent: {
          DEFAULT: 'var(--recruit-accent, #111)',
          muted: 'color-mix(in srgb, var(--recruit-accent, #111) 12%, transparent)',
          border: 'color-mix(in srgb, var(--recruit-accent, #111) 30%, transparent)',
          focus: 'color-mix(in srgb, var(--recruit-accent, #111) 50%, transparent)',
        },
      },
      fontFamily: {
        // Both resolve to a CSS custom property swapped at runtime by
        // loadFontPairing(). Defaults in globals.css cover the case
        // where the customization row hasn't been read yet.
        sans: ['var(--recruit-font-body)'],
        display: ['var(--recruit-font-display)'],
      },
    },
  },
  plugins: [],
};

export default config;
