import React from 'react';
import { Compass, Wand2, BarChart3, LifeBuoy } from 'lucide-react';
import type { CaretMode } from '../types/mode';

/** Shared per-mode icon set - keep in sync across ModePicker and suggestion cards. */
export const CARET_MODE_ICONS: Record<CaretMode, React.ReactNode> = {
  ask: <Compass size={11} />,
  agent: <Wand2 size={11} />,
  analyst: <BarChart3 size={11} />,
  support: <LifeBuoy size={11} />,
};
