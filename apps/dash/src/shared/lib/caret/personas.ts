/**
 * Display metadata for Caret modes/personas. The runtime mode catalog
 * lives in the @voxire/caret package; this file only adds dash-specific
 * descriptors (illustrative tasks, recommended apps) for the help UI.
 */

export interface PersonaInfo {
  id: 'ask' | 'agent' | 'analyst' | 'support';
  emoji: string;
  tagline: string;
  goodFor: string[];
  recommendedApps: string[];
}

export const PERSONA_INFO: PersonaInfo[] = [
  {
    id: 'ask',
    emoji: '💬',
    tagline: 'Conversational, read-only.',
    goodFor: ['Pulling KPIs', 'Looking up customers', 'Quick how-tos'],
    recommendedApps: ['dash', 'crm', 'erp'],
  },
  {
    id: 'agent',
    emoji: '🪄',
    tagline: 'Plans and acts across your apps.',
    goodFor: ['Creating tasks', 'Refunding orders', 'Updating inventory'],
    recommendedApps: ['pos', 'erp', 'flow'],
  },
  {
    id: 'analyst',
    emoji: '📊',
    tagline: 'Builds reports and charts.',
    goodFor: ['Sales breakdowns', 'Cohort exports', 'Trend analysis'],
    recommendedApps: ['insights', 'finance'],
  },
  {
    id: 'support',
    emoji: '🧭',
    tagline: 'Helps you navigate the product.',
    goodFor: ['Finding settings', 'Tutorial answers'],
    recommendedApps: ['dash'],
  },
];
