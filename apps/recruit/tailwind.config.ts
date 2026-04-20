import type { Config } from 'tailwindcss';

// The recruit app renders multiple visual templates (minimal, editorial,
// bold) — the palette below is only the accent default. Each template
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
        // Editorial template uses a serif display face, others fall
        // back to Plus Jakarta Sans.
        sans: ['Plus Jakarta Sans', 'Inter', 'system-ui', 'sans-serif'],
        display: ['"Fraunces"', '"Plus Jakarta Sans"', 'serif'],
      },
    },
  },
  plugins: [],
};

export default config;
