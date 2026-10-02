import React from 'react';
import {
  Briefcase,
  Building2,
  CheckCircle2,
  ClipboardList,
  Folder,
  Globe,
  Inbox,
  Rocket,
  Star,
  Target,
  Wrench,
  Zap,
} from 'lucide-react';

/**
 * A project's icon, drawn as a glyph.
 *
 * The stored value was painted straight into the DOM, so a project whose icon
 * is the string "building 2" rendered the words "building 2" in coloured
 * 14px text, wrapping onto two lines and colliding with the title.
 *
 * The column holds three kinds of value. Measured against the live database:
 * 2,322 projects store an emoji, 18 store a tick, and 10 store a lucide name.
 * All three map here, so the common cases get a glyph that means what the
 * emoji meant rather than a generic fallback, and only a genuinely unknown
 * value lands on Folder.
 *
 * Rendering only. Nothing here writes, so the stored values are untouched and
 * the sidebar's `project.icon === '✅'` checks for the My Tasks and Inbox
 * pseudo-projects still work. Migrating the column to lucide names is the
 * separate decision recorded on rtylr#1085.
 */
type Glyph = React.ComponentType<{
  size?: number;
  className?: string;
  style?: React.CSSProperties;
}>;

const BY_NAME: Record<string, Glyph> = {
  briefcase: Briefcase,
  'building-2': Building2,
  'building 2': Building2,
  building2: Building2,
  building: Building2,
  'check-circle': CheckCircle2,
  clipboard: ClipboardList,
  'clipboard-list': ClipboardList,
  folder: Folder,
  globe: Globe,
  inbox: Inbox,
  rocket: Rocket,
  star: Star,
  target: Target,
  wrench: Wrench,
  zap: Zap,
};

// The emoji actually in the column, each mapped to the glyph that carries the
// same meaning. Without these, 2,340 projects would all fall back to Folder.
const BY_EMOJI: Record<string, Glyph> = {
  '\u{1F4CB}': ClipboardList,
  '✅': CheckCircle2,
  '\u{1F4E5}': Inbox,
  '\u{1F680}': Rocket,
  '\u{1F310}': Globe,
  '\u{1F3E2}': Building2,
  '⭐': Star,
  '\u{1F4BC}': Briefcase,
  '\u{1F527}': Wrench,
  '⚡': Zap,
  '\u{1F3AF}': Target,
  '\u{1F4C1}': Folder,
};

/** The glyph for a stored icon value, or Folder when nothing matches. */
export function glyphForProjectIcon(icon: string | undefined): Glyph {
  if (!icon) return Folder;
  const trimmed = icon.trim();
  return BY_EMOJI[trimmed] ?? BY_NAME[trimmed.toLowerCase()] ?? Folder;
}

interface ProjectIconProps {
  icon: string | undefined;
  size?: number;
  className?: string;
  /** For tinting the glyph with the project's own colour. */
  style?: React.CSSProperties;
}

export const ProjectIcon: React.FC<ProjectIconProps> = ({ icon, size = 16, className, style }) => {
  const Glyph = glyphForProjectIcon(icon);
  return <Glyph size={size} className={className} style={style} />;
};

export default ProjectIcon;
