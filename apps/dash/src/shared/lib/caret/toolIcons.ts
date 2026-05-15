/**
 * Map MCP tool names to lucide icons + brand colors so the ToolCallCard
 * can render a quick visual hint about which module the tool belongs to.
 */
import {
  Users,
  Package,
  CreditCard,
  MonitorSpeaker,
  Layers,
  BarChart3,
  UserCheck,
  Building2,
  Mail,
  Wrench,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';

export interface ToolVisual {
  icon: LucideIcon;
  color: string;
  bg: string;
}

const MODULE_VISUALS: Record<string, ToolVisual> = {
  crm: { icon: Users, color: '#a855f7', bg: 'rgba(168,85,247,0.14)' },
  pos: { icon: MonitorSpeaker, color: '#0ea5e9', bg: 'rgba(14,165,233,0.14)' },
  erp: { icon: Package, color: '#10b981', bg: 'rgba(16,185,129,0.14)' },
  finance: { icon: CreditCard, color: '#f59e0b', bg: 'rgba(245,158,11,0.14)' },
  flow: { icon: Layers, color: '#22d3ee', bg: 'rgba(34,211,238,0.14)' },
  insights: { icon: BarChart3, color: '#6366f1', bg: 'rgba(99,102,241,0.14)' },
  hr: { icon: UserCheck, color: '#ef4444', bg: 'rgba(239,68,68,0.14)' },
  business: { icon: Building2, color: '#94a3b8', bg: 'rgba(148,163,184,0.14)' },
  email: { icon: Mail, color: '#f43f5e', bg: 'rgba(244,63,94,0.14)' },
};

const FALLBACK: ToolVisual = { icon: Wrench, color: '#a1a1aa', bg: 'rgba(161,161,170,0.14)' };

export function visualForTool(toolName: string): ToolVisual {
  const module = toolName.split('.')[0] || '';
  return MODULE_VISUALS[module] ?? FALLBACK;
}
