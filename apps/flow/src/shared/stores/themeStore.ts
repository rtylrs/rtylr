import type { ThemeMode } from '@voxire/utils';
import { useAppearanceStore } from './appearanceStore';

/**
 * A forwarding shim, kept only until its last consumer moves.
 *
 * This file used to be a store in its own right: it held the mode and it wrote
 * `data-theme` to documentElement. Both jobs moved. The mode lives in
 * `useAppearanceStore`, and the only writer of the appearance tokens and
 * attributes is `AppearanceProvider` from `@voxire/ui`. Its shipped behaviour is
 * preserved there rather than here: dark is still the default, because that is
 * the mode Flow is designed in, and a stored preference still wins, including
 * the `flow-theme` entry this file used to own, which the appearance store reads
 * once as `legacyModeKey`.
 *
 * What is deliberately gone is the DOM write. Two stores each stamping
 * `data-theme` from their own copy of the mode is how the palette and the tokens
 * end up disagreeing, so this forwards to the one store instead of shadowing it.
 *
 * @deprecated Read `mode` and `setMode` from `useAppearanceStore`, or render
 * `shared/components/ModeCards`. This file exists so
 * `features/settings/FlowSettings.tsx` keeps compiling while its Appearance
 * section is rewritten, and should be deleted with that rewrite.
 */
export function useThemeStore(): { theme: ThemeMode; setTheme: (theme: ThemeMode) => void } {
  const theme = useAppearanceStore((s) => s.mode);
  const setTheme = useAppearanceStore((s) => s.setMode);
  return { theme, setTheme };
}
